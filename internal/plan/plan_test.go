package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/verifisecurity/verifi/internal/gate"
)

func samplePlan() Plan {
	return Plan{
		Schema:    Schema,
		Generated: time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		Verifi:    "test",
		Workspace: "/projects/app",
		Ecosystem: "npm",
		Candidates: []Candidate{
			{
				Package: "lodash", Current: "4.17.11", Action: "upgrade", Target: "4.17.21",
				Command: []string{"npm", "install", "lodash@4.17.21"},
				Gate:    gate.Decision{Rung: gate.RungAdvisory, Authorization: gate.Confirm},
			},
			{
				Package: "minimist", Current: "1.2.0", Action: "remove",
				Command: []string{"npm", "uninstall", "minimist"},
				Gate:    gate.Decision{Rung: gate.RungAdvisory, Authorization: gate.Propose},
			},
			{
				Package: "left-pad", Current: "1.3.0", Action: "none",
				Gate: gate.Decision{Rung: gate.RungAdvisory, Authorization: gate.Propose},
			},
		},
	}
}

func TestWriteRead_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", Name)
	want := samplePlan()
	if err := Write(path, want); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Workspace != want.Workspace || len(got.Candidates) != len(want.Candidates) {
		t.Errorf("round trip lost data: %+v", got)
	}
	if !got.Generated.Equal(want.Generated) {
		t.Errorf("generated = %v, want %v", got.Generated, want.Generated)
	}
}

// TestRead_RejectsUnknownSchema keeps fix from acting on a plan whose fields may
// have moved. Guessing is the one thing it must not do, since the file decides
// what gets written to someone's project.
func TestRead_RejectsUnknownSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), Name)
	p := samplePlan()
	p.Schema = Schema + 99
	if err := Write(path, p); err != nil {
		t.Fatal(err)
	}
	_, err := Read(path)
	if err == nil {
		t.Fatal("expected an error for an unknown schema")
	}
	if !strings.Contains(err.Error(), "verifi scan") {
		t.Errorf("error should tell the user how to recover, got: %v", err)
	}
}

func TestRead_RejectsGarbage(t *testing.T) {
	path := filepath.Join(t.TempDir(), Name)
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); err == nil {
		t.Fatal("expected an error for a malformed plan")
	}
}

// TestSelected is the human loop in one assertion: nothing is selected until a
// human marks it, and marking something with no runnable command selects nothing.
func TestSelected(t *testing.T) {
	p := samplePlan()
	if got := p.Selected(); len(got) != 0 {
		t.Fatalf("a fresh plan selected %d candidates, want 0", len(got))
	}

	p.Candidates[0].Apply = true // lodash, runnable
	p.Candidates[2].Apply = true // left-pad, action none, no command
	got := p.Selected()
	if len(got) != 1 || got[0].Package != "lodash" {
		t.Fatalf("selected = %+v, want just lodash", got)
	}
}

func TestActionable(t *testing.T) {
	tests := []struct {
		name string
		c    Candidate
		want bool
	}{
		{"upgrade with command", Candidate{Action: "upgrade", Command: []string{"npm", "install", "x"}}, true},
		{"remove with command", Candidate{Action: "remove", Command: []string{"npm", "uninstall", "x"}}, true},
		{"no published fix", Candidate{Action: "none"}, false},
		{"action but no command", Candidate{Action: "upgrade"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Actionable(); got != tt.want {
				t.Errorf("Actionable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPathIsOutsideTheWorkspace pins the storage rule: scanning a project must
// not put files in it, so a plan path never lands under the workspace.
func TestPathIsOutsideTheWorkspace(t *testing.T) {
	home, workspace := "/home/u/.verifi", "/projects/app"
	got := Path(home, workspace)
	if !strings.HasPrefix(got, home) {
		t.Errorf("plan path %q is not under the verifi home %q", got, home)
	}
	if strings.HasPrefix(got, workspace) {
		t.Errorf("plan path %q is inside the scanned workspace", got)
	}
}

func TestDirIsStablePerWorkspace(t *testing.T) {
	home := "/home/u/.verifi"
	a1 := Dir(home, "/projects/app")
	a2 := Dir(home, "/projects/app")
	b := Dir(home, "/projects/other")
	if a1 != a2 {
		t.Errorf("same workspace gave different directories: %q and %q", a1, a2)
	}
	if a1 == b {
		t.Errorf("different workspaces collided on %q", a1)
	}
}
