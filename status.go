package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/osv"
)

// runStatus implements `verifi status <path> [--json] [--db <dir>]`: resolve the
// project, match it against a local OSV database, and print what needs fixing.
// Read-only. npm for now. Fix candidates and reasoning are later slices; this
// reports what is vulnerable and the fixed versions the advisory records.
func runStatus(args []string) error {
	path, dbDir := "", ""
	asJSON := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--json":
			asJSON = true
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
	if dbDir == "" {
		dbDir = defaultDBDir()
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
	db, err := osv.Load(dbDir)
	if err != nil {
		return fmt.Errorf("%w\npoint at a local OSV database with --db <dir>", err)
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
	printStatus(inv, findings)
	return nil
}

func defaultDBDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".verifi", "osv")
	}
	return filepath.Join(".verifi", "osv")
}

func printStatus(inv *inventory.Inventory, findings []finding.Finding) {
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
			if len(f.FixedVersions) > 0 {
				fmt.Printf("      fixed in %s\n", strings.Join(f.FixedVersions, ", "))
			} else {
				fmt.Printf("      no fixed version published\n")
			}
		}
		fmt.Println()
	}
	fmt.Println("Fix candidates and reasoning come next. This build reports what is vulnerable.")
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
