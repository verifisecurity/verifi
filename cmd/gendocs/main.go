// Command gendocs regenerates the generated blocks in the docs from the code:
// the README and the per-command pages. Example output in the docs is captured
// by actually running the CLI against the fixtures, so the docs show what the
// tool really produces. Run it with `make docs`; CI fails if the committed docs
// are stale.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/verifisecurity/verifi/internal/command"
	"github.com/verifisecurity/verifi/internal/docgen"
)

func main() {
	examples, err := captureExamples()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
	if err := docgen.Generate("README.md", "docs", examples); err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
}

// captureExamples builds the CLI once and runs each command's example to capture
// its real output. Examples use offline flags and fixtures so they are
// deterministic and need no network.
func captureExamples() (map[string]string, error) {
	dir, err := os.MkdirTemp("", "verifi-gendocs")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "verifi")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}

	out := map[string]string{}
	for _, c := range command.All {
		if len(c.Example) == 0 {
			continue
		}
		var buf bytes.Buffer
		run := exec.Command(bin, c.Example...)
		run.Stdout = &buf
		run.Stderr = &buf
		if err := run.Run(); err != nil {
			return nil, fmt.Errorf("example for %q (%v): %w\n%s", c.Name, c.Example, err, buf.String())
		}
		out[c.Name] = buf.String()
	}
	return out, nil
}
