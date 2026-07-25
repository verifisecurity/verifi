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
)

// analyze runs the read-only pipeline for a project: resolve the dependencies,
// match them against the local OSV database, and reason about each fix. Shared
// by `status` and `fix`. npm for now.
func analyze(path, dbDir string, offline bool) (*inventory.Inventory, []finding.Finding, []reason.Recommendation, error) {
	lockPath := filepath.Join(path, "package-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read %s: %w", lockPath, err)
	}
	inv, err := inventory.ParseNpmLock(data)
	if err != nil {
		return nil, nil, nil, err
	}
	if dbDir == "" {
		dbDir = filepath.Join(cacheRoot(), inv.Ecosystem)
	}
	db, err := osv.Load(dbDir)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("no OSV database at %s\nrun `verifi update` to download it, or pass --db <dir>", dbDir)
	}
	findings := db.Match(inv)

	var exists candidate.Exists
	if !offline {
		exists = registryExists()
	}
	recs := reason.Explain(candidate.Compute(findings, exists))
	return inv, findings, recs, nil
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
