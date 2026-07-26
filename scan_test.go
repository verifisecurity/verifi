package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/verifisecurity/verifi/internal/command"
)

var update = flag.Bool("update", false, "regenerate golden files")

func fixture() string   { return filepath.Join("testdata", "npm", "vuln") }
func fixtureDB() string { return filepath.Join("testdata", "osv") }

// capture runs f with os.Stdout redirected and returns what it printed. The
// reader runs concurrently so output larger than the pipe buffer cannot deadlock.
func capture(t *testing.T, f func() error) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	runErr := f()
	w.Close()
	os.Stdout = orig
	out := <-done
	r.Close()

	if runErr != nil {
		t.Fatalf("runScan: %v", runErr)
	}
	return out
}

func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join(fixture(), name)
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run: go test ./... -update): %v", err)
	}
	if strings.TrimSpace(got) != strings.TrimSpace(string(want)) {
		t.Errorf("%s mismatch\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

// TestScan_Golden is the end-to-end read path over the npm fixture: resolve the
// lockfile, match the fixture OSV database, reason about each fix, render it.
// Regenerate with -update.
func TestScan_Golden(t *testing.T) {
	got := capture(t, func() error {
		return runScan([]string{fixture(), "--db", fixtureDB(), "--offline"})
	})
	checkGolden(t, "expected.scan.txt", got)
}

func TestScan_JSON_Golden(t *testing.T) {
	got := capture(t, func() error {
		return runScan([]string{fixture(), "--db", fixtureDB(), "--offline", "--json"})
	})
	checkGolden(t, "expected.scan.json", got)
}

func TestScan_SBOM_Golden(t *testing.T) {
	got := capture(t, func() error {
		return runScan([]string{fixture(), "--sbom"})
	})
	checkGolden(t, "expected.scan.sbom.json", got)
}

func TestScan_Inventory(t *testing.T) {
	got := capture(t, func() error {
		return runScan([]string{fixture(), "--inventory"})
	})
	for _, want := range []string{"vuln-app@1.0.0", "lodash@4.17.11", "minimist@1.2.0", "left-pad@1.3.0"} {
		if !strings.Contains(got, want) {
			t.Errorf("inventory view missing %q:\n%s", want, got)
		}
	}
}

// TestScan_TreeViewsNeedNoDatabase pins the reason --inventory and --sbom are
// folded into scan rather than gated behind it: describing the tree does not
// need advisories, so a missing database must not break them.
func TestScan_TreeViewsNeedNoDatabase(t *testing.T) {
	for _, view := range []string{"--inventory", "--sbom"} {
		if err := runScan([]string{fixture(), view, "--db", "/nonexistent"}); err != nil {
			t.Errorf("scan %s with no database: %v", view, err)
		}
	}
}

// TestScan_MissingDatabase_PointsAtRealCommand is a regression guard. A missing
// database used to send the user to `verifi update`, which no longer exists,
// leaving a fresh install with no way to obtain one. The message must name a
// command that is actually in the registry.
func TestScan_MissingDatabase_PointsAtRealCommand(t *testing.T) {
	err := runScan([]string{fixture(), "--db", "/nonexistent", "--offline"})
	if err == nil {
		t.Fatal("expected an error for a missing database")
	}
	if !strings.Contains(err.Error(), "--download") {
		t.Errorf("message does not offer --download:\n%s", err)
	}
	if strings.Contains(err.Error(), "verifi update") {
		t.Errorf("message points at the removed update command:\n%s", err)
	}
}

func TestScan_RejectsBadFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"unknown flag", []string{fixture(), "--nope"}, "unknown flag"},
		{"db without value", []string{fixture(), "--db"}, "--db needs a directory"},
		{"download with db", []string{fixture(), "--download", "--db", fixtureDB()}, "cannot be combined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runScan(tt.args)
			if err == nil {
				t.Fatalf("expected an error, got none")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

// TestSurfaceIsScanAndFix pins the command surface. inspect, update, and status
// folded into scan; anything reintroducing them should fail here first.
func TestSurfaceIsScanAndFix(t *testing.T) {
	got := map[string]bool{}
	for _, c := range command.All {
		got[c.Name] = true
	}
	for _, want := range []string{"scan", "fix", "version", "help"} {
		if !got[want] {
			t.Errorf("command %q missing from the registry", want)
		}
	}
	for _, gone := range []string{"inspect", "update", "status", "mcp"} {
		if got[gone] {
			t.Errorf("command %q is retired but still in the registry", gone)
		}
	}
}
