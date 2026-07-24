// Command gendocs regenerates the generated blocks in the docs from the code.
// Run it with `make docs`. CI runs it and fails if the committed docs are stale.
package main

import (
	"fmt"
	"os"

	"github.com/verifisecurity/verifi/internal/docgen"
)

func main() {
	if err := docgen.UpdateReadme("README.md"); err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
}
