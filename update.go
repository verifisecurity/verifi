package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/verifisecurity/verifi/internal/osv"
)

// downloadDB fetches the OSV advisory database for an ecosystem into the local
// cache, so matching runs offline with no --db flag. It backs `verifi scan
// --download`; the ecosystem comes from the project's lockfile, not a flag.
func downloadDB(ecosystem string) error {
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

// verifiHome is the root of everything verifi stores on the machine: the
// advisory cache and the per-project plans. VERIFI_HOME relocates all of it,
// which is how the tests keep off a real home directory.
func verifiHome() string {
	if h := os.Getenv("VERIFI_HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".verifi")
	}
	return ".verifi"
}

// cacheRoot is where the downloaded OSV database lives, one directory per
// ecosystem underneath. It is a cache: disposable and re-fetchable, unlike the
// plans alongside it.
func cacheRoot() string { return filepath.Join(verifiHome(), "osv") }
