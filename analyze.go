package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/verifisecurity/verifi/internal/candidate"
	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/osv"
	"github.com/verifisecurity/verifi/internal/reason"
	"github.com/verifisecurity/verifi/internal/registry"
	usagescan "github.com/verifisecurity/verifi/internal/usage"
)

// analysis is the read-only result for a project.
type analysis struct {
	inv      *inventory.Inventory
	findings []finding.Finding
	cands    []candidate.Candidate // aligned 1:1 with recs; carries distance and residual for the gate
	recs     []reason.Recommendation
	imported map[string]bool // packages the project's own code imports
	dbMeta   *osv.Meta       // cache freshness; nil when --db is used or unstamped
	dbPath   string          // the advisory directory actually matched against
	dbCache  bool            // true when that directory was the local cache, not a --db
}

// analyze runs the read-only pipeline for a project: resolve the dependencies,
// match them against the local OSV database, scan the source for imports, and
// reason about each fix. Shared by `status` and `fix`. npm for now.
func analyze(path, dbDir string, offline bool) (*analysis, error) {
	lockPath := filepath.Join(path, "package-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", lockPath, err)
	}
	inv, err := inventory.ParseNpmLock(data)
	if err != nil {
		return nil, err
	}
	// Only the default cache carries a freshness stamp; a user-supplied --db is
	// their own directory, so we do not second-guess its age.
	usedCache := dbDir == ""
	if usedCache {
		dbDir = filepath.Join(cacheRoot(), inv.Ecosystem)
	}
	db, err := osv.Load(dbDir)
	if err != nil {
		return nil, fmt.Errorf("no OSV database at %s\nrun `verifi scan %s --download` to fetch it, or pass --db <dir>", dbDir, path)
	}
	findings := db.Match(inv)

	var dbMeta *osv.Meta
	if usedCache {
		if m, err := osv.ReadMeta(cacheRoot(), inv.Ecosystem); err == nil {
			dbMeta = &m
		}
	}

	// Best-effort usage scan: it drives the remove recommendation and the view.
	imported, _ := usagescan.Scan(path)
	direct := map[string]bool{}
	for _, p := range inv.Packages {
		if p.Direct {
			direct[p.Name] = true
		}
	}
	removable := func(name string) bool { return direct[name] && !imported[name] }

	var exists candidate.Exists
	if !offline {
		exists = registryExists()
	}
	cands := candidate.Compute(findings, exists, removable)
	recs := reason.Explain(cands)
	return &analysis{
		inv: inv, findings: findings, cands: cands, recs: recs,
		imported: imported, dbMeta: dbMeta, dbPath: dbDir, dbCache: usedCache,
	}, nil
}

// registryExists returns a predicate that reports whether a package version is
// published, memoised per package. If the registry cannot be reached it returns
// true rather than over-filtering, so a lookup failure degrades to OSV-only.
func registryExists() candidate.Exists {
	cache := map[string]map[string]bool{}
	return func(name, version string) bool {
		vs, ok := cache[name]
		if !ok {
			v, err := registry.NpmVersions(name)
			if err != nil {
				v = nil
			}
			cache[name] = v
			vs = v
		}
		if vs == nil {
			return true
		}
		return vs[version]
	}
}
