package osv

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/verifisecurity/verifi/internal/inventory"
)

var update = flag.Bool("update", false, "regenerate golden files")

func TestInRange(t *testing.T) {
	events := []Event{{Introduced: "0"}, {Fixed: "4.17.21"}}
	cases := map[string]bool{
		"4.17.11": true,
		"4.17.20": true,
		"4.17.21": false,
		"5.0.0":   false,
		"0.1.0":   true,
	}
	for v, want := range cases {
		if got := inRange(v, events); got != want {
			t.Errorf("inRange(%s) = %v, want %v", v, got, want)
		}
	}
}

// TestMatch_Golden is the fixture end-to-end check: parse a vulnerable project,
// match it against the local OSV fixtures, assert the findings against golden.
// The decoy advisory (introduced after the installed version) must not appear.
// Regenerate with: go test ./... -update
func TestMatch_Golden(t *testing.T) {
	lock, err := os.ReadFile(filepath.Join("..", "..", "testdata", "npm", "vuln", "package-lock.json"))
	if err != nil {
		t.Fatalf("read fixture lock: %v", err)
	}
	inv, err := inventory.ParseNpmLock(lock)
	if err != nil {
		t.Fatalf("ParseNpmLock: %v", err)
	}
	db, err := Load(filepath.Join("..", "..", "testdata", "osv"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	findings := db.Match(inv)
	if len(findings) != 2 {
		t.Fatalf("findings = %d, want 2 (lodash, minimist; left-pad and decoy excluded)", len(findings))
	}

	got, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	checkGolden(t, filepath.Join("..", "..", "testdata", "npm", "vuln", "expected.findings.json"), got)
}

func checkGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run: go test ./... -update): %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Errorf("%s mismatch\n--- got ---\n%s\n--- want ---\n%s", filepath.Base(path), got, want)
	}
}
