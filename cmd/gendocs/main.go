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
	"regexp"
	"strings"

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

	// Examples run against a throwaway verifi home so capturing them cannot
	// write into the real one, and so no absolute path from this machine can
	// reach the committed docs.
	home := filepath.Join(dir, "home")
	repo, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	for _, c := range command.All {
		if len(c.Example) == 0 {
			continue
		}
		var buf bytes.Buffer
		run := exec.Command(bin, c.Example...)
		run.Env = append(os.Environ(), "VERIFI_HOME="+home)
		run.Stdout = &buf
		run.Stderr = &buf
		if err := run.Run(); err != nil {
			return nil, fmt.Errorf("example for %q (%v): %w\n%s", c.Name, c.Example, err, buf.String())
		}
		out[c.Name] = scrubPaths(buf.String(), home, repo)
	}
	return out, nil
}

// projectDir matches the per-project directory in a plan path, which is a digest
// of the workspace's absolute path.
var projectDir = regexp.MustCompile(`projects/[0-9a-f]+`)

// scrubPaths removes machine-specific absolute paths from captured output. The
// docs are public and committed, so a username or a checkout location must never
// reach them. The project directory is scrubbed too: it is a digest of where the
// checkout happens to live, so leaving it in would make the committed docs
// disagree with the same command run anywhere else, including in CI.
func scrubPaths(s, home, repo string) string {
	s = strings.ReplaceAll(s, home, "~/.verifi")
	s = strings.ReplaceAll(s, repo+string(os.PathSeparator), "")
	return projectDir.ReplaceAllString(s, "projects/<project>")
}
