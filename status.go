package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/verifisecurity/verifi/internal/candidate"
	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/osv"
	"github.com/verifisecurity/verifi/internal/reason"
	"github.com/verifisecurity/verifi/internal/registry"
)

// runStatus implements `verifi status <path> [--json] [--db <dir>] [--offline]`:
// resolve the project, match it against a local OSV database, and print what
// needs fixing with a reasoned fix per package. Read-only. npm for now. By
// default it checks the registry so it only recommends versions that exist;
// --offline skips that and trusts OSV's fixed versions.
func runStatus(args []string) error {
	path, dbDir := "", ""
	asJSON, offline := false, false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--json":
			asJSON = true
		case "--offline":
			offline = true
		case "--db":
			if i+1 >= len(args) {
				return fmt.Errorf("--db needs a directory")
			}
			i++
			dbDir = args[i]
		default:
			if len(a) > 0 && a[0] == '-' {
				return fmt.Errorf("unknown flag %q", a)
			}
			path = a
		}
	}
	if path == "" {
		path = "."
	}

	lockPath := filepath.Join(path, "package-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", lockPath, err)
	}
	inv, err := inventory.ParseNpmLock(data)
	if err != nil {
		return err
	}

	if dbDir == "" {
		dbDir = filepath.Join(cacheRoot(), inv.Ecosystem)
	}
	db, err := osv.Load(dbDir)
	if err != nil {
		return fmt.Errorf("no OSV database at %s\nrun `verifi update` to download it, or pass --db <dir>", dbDir)
	}
	findings := db.Match(inv)

	if asJSON {
		out, err := json.MarshalIndent(findings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}
	var exists candidate.Exists
	if !offline {
		exists = registryExists()
	}
	printStatus(inv, findings, exists)
	return nil
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

func printStatus(inv *inventory.Inventory, findings []finding.Finding, exists candidate.Exists) {
	fmt.Printf("%s@%s (%s)\n", inv.Root.Name, inv.Root.Version, inv.Ecosystem)
	if len(findings) == 0 {
		fmt.Printf("%d packages scanned, none vulnerable.\n", len(inv.Packages))
		return
	}

	byPkg := map[string][]finding.Finding{}
	var order []string
	for _, f := range findings {
		if _, ok := byPkg[f.Name]; !ok {
			order = append(order, f.Name)
		}
		byPkg[f.Name] = append(byPkg[f.Name], f)
	}
	sort.Slice(order, func(i, j int) bool {
		ri, rj := worstRank(byPkg[order[i]]), worstRank(byPkg[order[j]])
		if ri != rj {
			return ri < rj
		}
		return order[i] < order[j]
	})

	recs := map[string]reason.Recommendation{}
	for _, r := range reason.Explain(candidate.Compute(findings, exists)) {
		recs[r.Name] = r
	}

	fmt.Printf("%d packages scanned, %d vulnerable.\n\n", len(inv.Packages), len(order))
	for _, name := range order {
		fs := byPkg[name]
		fmt.Printf("%-8s %s %s   %s\n", sevLabel(worstRank(fs)), name, fs[0].Version, directTag(inv, name))
		for _, f := range fs {
			id := f.Advisory
			if len(f.Aliases) > 0 {
				id = f.Aliases[0] + " (" + f.Advisory + ")"
			}
			fmt.Printf("   %s   %s\n", id, f.Summary)
		}
		if r, ok := recs[name]; ok {
			if r.Action == "upgrade" {
				fmt.Printf("   Fix: upgrade to %s   confidence: %s\n", r.Target, r.Confidence)
			} else {
				fmt.Printf("   No published fix   confidence: %s\n", r.Confidence)
			}
			fmt.Printf("        %s\n", r.Reason)
			for _, l := range r.Limitations {
				fmt.Printf("        - %s\n", l)
			}
		}
		fmt.Println()
	}
	fmt.Println("Next: impact (does the fix touch code you use) and behavioural checks.")
}

func sevRank(s string) int {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return 0
	case "HIGH":
		return 1
	case "MEDIUM", "MODERATE":
		return 2
	case "LOW":
		return 3
	default:
		return 4
	}
}

func worstRank(fs []finding.Finding) int {
	r := 4
	for _, f := range fs {
		if x := sevRank(f.Severity); x < r {
			r = x
		}
	}
	return r
}

func sevLabel(rank int) string {
	switch rank {
	case 0:
		return "CRITICAL"
	case 1:
		return "HIGH"
	case 2:
		return "MEDIUM"
	case 3:
		return "LOW"
	default:
		return "UNKNOWN"
	}
}

func directTag(inv *inventory.Inventory, name string) string {
	for _, p := range inv.Packages {
		if p.Name == name {
			if p.Direct {
				return "direct"
			}
			return "transitive"
		}
	}
	return ""
}
