package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/verifisecurity/verifi/internal/inventory"
)

// runInspect implements `verifi inspect <path> [--json] [--sbom]`: resolve the
// project's dependency tree and print it. Read-only. npm for now.
func runInspect(args []string) error {
	path := ""
	asJSON, asSBOM := false, false
	for _, a := range args {
		switch a {
		case "--json":
			asJSON = true
		case "--sbom":
			asSBOM = true
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

	switch {
	case asSBOM:
		out, err := inv.ToCycloneDX()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	case asJSON:
		out, err := json.MarshalIndent(inv, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	default:
		printInventory(inv)
	}
	return nil
}

func printInventory(inv *inventory.Inventory) {
	direct, dev := 0, 0
	for _, p := range inv.Packages {
		if p.Direct {
			direct++
		}
		if p.Scope == "dev" {
			dev++
		}
	}
	fmt.Printf("%s@%s (%s)\n", inv.Root.Name, inv.Root.Version, inv.Ecosystem)
	fmt.Printf("%d packages: %d direct, %d transitive, %d dev\n",
		len(inv.Packages), direct, len(inv.Packages)-direct, dev)
	for _, p := range inv.Packages {
		tag := "transitive"
		if p.Direct {
			tag = "direct"
		}
		fmt.Printf("  %s@%s  [%s, %s]\n", p.Name, p.Version, tag, p.Scope)
	}
}
