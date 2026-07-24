package reason

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/verifisecurity/verifi/internal/candidate"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/osv"
)

var update = flag.Bool("update", false, "regenerate golden files")

func TestExplain_HonestLimits(t *testing.T) {
	// A major bump must always carry the honest "not checked" limitations plus
	// the major-bump caution, and never claim more than advisory confidence.
	cands := []candidate.Candidate{
		{Name: "pkg", Current: "1.0.0", Action: "upgrade", Target: "3.0.0", Distance: "major", Clears: []string{"CVE-X"}},
	}
	r := Explain(cands)[0]
	if r.Confidence != "advisory" {
		t.Errorf("confidence = %q, want advisory (never overclaim at raw-OSV depth)", r.Confidence)
	}
	joined := strings.Join(r.Limitations, " | ")
	for _, want := range []string{"Major version bump", "Behaviour not verified", "Code impact not checked"} {
		if !strings.Contains(joined, want) {
			t.Errorf("limitations missing %q, got: %s", want, joined)
		}
	}
}

func TestExplain_NoFix(t *testing.T) {
	cands := []candidate.Candidate{
		{Name: "pkg", Current: "1.0.0", Action: "none", Residual: []string{"CVE-Y"}},
	}
	r := Explain(cands)[0]
	if r.Action != "none" || !strings.Contains(r.Reason, "CVE-Y") {
		t.Errorf("got %+v, want a no-fix recommendation citing CVE-Y", r)
	}
}

// TestExplain_Golden runs the full chain on the vulnerable fixture.
func TestExplain_Golden(t *testing.T) {
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
	got, err := json.MarshalIndent(Explain(candidate.Compute(db.Match(inv))), "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	golden := filepath.Join(base, "npm", "vuln", "expected.recommendations.json")
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
		t.Errorf("recommendations mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
