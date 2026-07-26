package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/verifisecurity/verifi/internal/osv"
)

// runUpdate implements `verifi update [--ecosystem npm]`: download the OSV
// advisory database into the local cache so `verifi status` works offline with
// no --db flag.
func runUpdate(args []string) error {
	ecosystem := "npm"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--ecosystem":
			if i+1 >= len(args) {
				return fmt.Errorf("--ecosystem needs a value")
			}
			i++
			ecosystem = args[i]
		default:
			return fmt.Errorf("unknown argument %q", args[i])
		}
	}

	root := cacheRoot()
	fmt.Printf("Downloading the OSV database for %s ...\n", ecosystem)
	n, err := osv.Fetch(ecosystem, root)
	if err != nil {
		return err
	}
	if err := osv.WriteMeta(root, ecosystem, n); err != nil {
		return fmt.Errorf("stamp cache: %w", err)
	}
	fmt.Printf("Stored %d advisories in %s\n", n, filepath.Join(root, ecosystem))
	return nil
}

// cacheRoot is where the downloaded OSV database lives, one directory per
// ecosystem underneath.
func cacheRoot() string {
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".verifi", "osv")
	}
	return filepath.Join(".verifi", "osv")
}
