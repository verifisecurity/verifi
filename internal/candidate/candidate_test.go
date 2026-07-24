package candidate

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/osv"
)

var update = flag.Bool("update", false, "regenerate golden files")

func TestNearestFix(t *testing.T) {
	cases := []struct {
		current string
		fixes   []string
		want    string
	}{
		{"4.17.11", []string{"4.17.21"}, "4.17.21"},
		{"1.2.0", []string{"1.2.6"}, "1.2.6"},
		{"1.0.0", []string{"1.2.6", "2.0.1"}, "1.2.6"}, // nearest, not latest
		{"1.3.0", []string{"1.2.6"}, ""},               // no fix moves forward
		{"1.0.0", nil, ""},                             // no published fix
	}
	for _, c := range cases {
		if got := nearestFix("pkg", c.current, c.fixes, nil); got != c.want {
			t.Errorf("nearestFix(%s, %v) = %q, want %q", c.current, c.fixes, got, c.want)
		}
	}
}

func TestCompute_MultiAdvisory(t *testing.T) {
	// One package, two advisories with different fixes: the target must clear
	// both, so it is the furthest of the nearest fixes.
	fs := []finding.Finding{
		{Name: "pkg", Version: "1.0.0", Purl: "pkg:npm/pkg@1.0.0", Advisory: "A", FixedVersions: []string{"1.2.0"}},
		{Name: "pkg", Version: "1.0.0", Purl: "pkg:npm/pkg@1.0.0", Advisory: "B", FixedVersions: []string{"1.5.0"}},
	}
	got := Compute(fs, nil)
	if len(got) != 1 {
		t.Fatalf("candidates = %d, want 1", len(got))
	}
	c := got[0]
	if c.Target != "1.5.0" || c.Distance != "minor" || len(c.Clears) != 2 {
		t.Errorf("got %+v, want target 1.5.0, minor, clears both", c)
	}
}

func TestCompute_SkipsUnpublishedFix(t *testing.T) {
	// One advisory's only fix is an unpublished version; it must fall to
	// residual, not be recommended. The other clears to a real version.
	fs := []finding.Finding{
		{Name: "lodash", Version: "4.17.11", Purl: "pkg:npm/lodash@4.17.11", Advisory: "REAL", FixedVersions: []string{"4.17.21"}},
		{Name: "lodash", Version: "4.17.11", Purl: "pkg:npm/lodash@4.17.11", Advisory: "PHANTOM", FixedVersions: []string{"4.18.0"}},
	}
	exists := func(name, version string) bool { return version == "4.17.21" } // 4.18.0 not published
	c := Compute(fs, exists)[0]
	if c.Target != "4.17.21" {
		t.Errorf("target = %q, want 4.17.21 (not the phantom 4.18.0)", c.Target)
	}
	if len(c.Clears) != 1 || c.Clears[0] != "REAL" {
		t.Errorf("clears = %v, want [REAL]", c.Clears)
	}
	if len(c.Residual) != 1 || c.Residual[0] != "PHANTOM" {
		t.Errorf("residual = %v, want [PHANTOM]", c.Residual)
	}
}

// TestCompute_Golden is the fixture end-to-end: match the vulnerable project,
// compute candidates, assert against golden. Regenerate with -update.
func TestCompute_Golden(t *testing.T) {
	base := filepath.Join("..", "..", "testdata")
	lock, err := os.ReadFile(filepath.Join(base, "npm", "vuln", "package-lock.json"))
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	inv, err := inventory.ParseNpmLock(lock)
	if err != nil {
		t.Fatalf("ParseNpmLock: %v", err)
	}
	db, err := osv.Load(filepath.Join(base, "osv"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, err := json.MarshalIndent(Compute(db.Match(inv), nil), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	golden := filepath.Join(base, "npm", "vuln", "expected.candidates.json")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run: go test ./... -update): %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Errorf("candidates mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
