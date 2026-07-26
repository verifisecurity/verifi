package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/verifisecurity/verifi/internal/inventory"
)

// loadInventory reads a project's lockfile and resolves it to an inventory.
// npm for now.
func loadInventory(path string) (*inventory.Inventory, error) {
	lockPath := filepath.Join(path, "package-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", lockPath, err)
	}
	return inventory.ParseNpmLock(data)
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
