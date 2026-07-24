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
		if got := nearestFix(c.current, c.fixes); got != c.want {
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
	got := Compute(fs)
	if len(got) != 1 {
		t.Fatalf("candidates = %d, want 1", len(got))
	}
	c := got[0]
	if c.Target != "1.5.0" || c.Distance != "minor" || len(c.Clears) != 2 {
		t.Errorf("got %+v, want target 1.5.0, minor, clears both", c)
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
	got, err := json.MarshalIndent(Compute(db.Match(inv)), "", "  ")
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
