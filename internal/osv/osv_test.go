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

func TestSeverity_Precedence(t *testing.T) {
	top := Advisory{DatabaseSpecific: map[string]any{"severity": "high"}}
	if got := severity(top); got != "HIGH" {
		t.Errorf("top-level severity = %q, want HIGH", got)
	}
	perAff := Advisory{Affected: []Affected{{DatabaseSpecific: map[string]any{"severity": "moderate"}}}}
	if got := severity(perAff); got != "MODERATE" {
		t.Errorf("per-affected severity = %q, want MODERATE", got)
	}
	if got := severity(Advisory{}); got != "UNKNOWN" {
		t.Errorf("missing severity = %q, want UNKNOWN", got)
	}
}

func TestMatch_SkipsWithdrawn(t *testing.T) {
	db := &DB{byKey: map[string][]Advisory{
		key("npm", "left-pad"): {{
			ID:        "GHSA-withdrawn",
			Withdrawn: "2021-01-01T00:00:00Z",
			Affected: []Affected{{
				Package: Package{Ecosystem: "npm", Name: "left-pad"},
				Ranges:  []Range{{Type: "SEMVER", Events: []Event{{Introduced: "0"}}}},
			}},
		}},
	}}
	inv := &inventory.Inventory{
		Ecosystem: "npm",
		Packages:  []inventory.Package{{Name: "left-pad", Version: "1.3.0", Purl: "pkg:npm/left-pad@1.3.0"}},
	}
	if got := db.Match(inv); len(got) != 0 {
		t.Errorf("withdrawn advisory should be skipped, got %d findings", len(got))
	}
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
