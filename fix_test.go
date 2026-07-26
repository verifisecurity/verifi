package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/verifisecurity/verifi/internal/gate"
	"github.com/verifisecurity/verifi/internal/plan"
)

// markPlan flips apply on one package and saves the plan, standing in for the
// human step: opening the file and editing it.
func markPlan(t *testing.T, path, pkg string) plan.Plan {
	t.Helper()
	p, err := plan.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for i := range p.Candidates {
		if p.Candidates[i].Package == pkg {
			p.Candidates[i].Apply = true
			found = true
		}
	}
	if !found {
		t.Fatalf("no candidate for %q to mark", pkg)
	}
	if err := plan.Write(path, p); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestFix_NoPlan points the user at the command that produces one, rather than
// failing with a bare file-not-found.
func TestFix_NoPlan(t *testing.T) {
	isolate(t)
	err := runFix([]string{fixture()})
	if err == nil {
		t.Fatal("expected an error when no plan exists")
	}
	if !strings.Contains(err.Error(), "verifi scan") {
		t.Errorf("error should tell the user to scan first, got: %v", err)
	}
}

// TestFix_NothingMarked is the default outcome of the whole design: scan marks
// nothing, so a fix straight afterwards writes nothing and says why.
func TestFix_NothingMarked(t *testing.T) {
	isolate(t)
	scanFixture(t)

	out := capture(t, func() error { return runFix([]string{fixture()}) })
	if !strings.Contains(out, "Nothing marked") {
		t.Errorf("output should say nothing was marked:\n%s", out)
	}
	if !strings.Contains(out, `"apply": true`) {
		t.Errorf("output should say how to mark a fix:\n%s", out)
	}
}

// TestFix_RefusesWhatTheGateOnlyProposes is the rule that keeps the gate from
// being decorative. Marking a fix says you want it; the gate still has to agree
// that verifi can deliver it.
func TestFix_RefusesWhatTheGateOnlyProposes(t *testing.T) {
	home := isolate(t)
	ws := filepath.Join("testdata", "npm", "transitive")
	capture(t, func() error { return runScan([]string{ws, "--db", fixtureDB(), "--offline"}) })

	planPath := plan.Path(home, ws)
	markPlan(t, planPath, "lodash") // transitive, so the gate only proposes it

	out := capture(t, func() error { return runFix([]string{ws}) })
	if !strings.Contains(out, "Skipping lodash") {
		t.Errorf("a marked propose-only fix should be skipped:\n%s", out)
	}
	if !strings.Contains(out, "overrides") {
		t.Errorf("the skip should carry the gate's reason:\n%s", out)
	}
}

// TestFix_SelectionNeedsBothMarkAndGate states the contract in one place: a fix
// runs only when the human marked it and the gate authorised it.
func TestFix_SelectionNeedsBothMarkAndGate(t *testing.T) {
	confirmed := plan.Candidate{
		Package: "a", Action: "upgrade", Command: []string{"npm", "install", "a@2"},
		Gate: gate.Decision{Authorization: gate.Confirm},
	}
	proposed := plan.Candidate{
		Package: "b", Action: "upgrade", Command: []string{"npm", "install", "b@2"},
		Gate: gate.Decision{Authorization: gate.Propose},
	}

	tests := []struct {
		name    string
		mark    bool
		c       plan.Candidate
		wantRun bool
	}{
		{"marked and confirmed runs", true, confirmed, true},
		{"confirmed but unmarked does not", false, confirmed, false},
		{"marked but only proposed does not", true, proposed, false},
		{"unmarked and only proposed does not", false, proposed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.c
			c.Apply = tt.mark
			p := plan.Plan{Candidates: []plan.Candidate{c}}

			runnable := false
			for _, s := range p.Selected() {
				if s.Gate.Authorization == gate.Confirm {
					runnable = true
				}
			}
			if runnable != tt.wantRun {
				t.Errorf("runnable = %v, want %v", runnable, tt.wantRun)
			}
		})
	}
}

func TestFix_PlanFlag(t *testing.T) {
	isolate(t)
	out := filepath.Join(t.TempDir(), "plan.json")
	scanFixture(t, "--out", out)

	// Reading from an explicit path finds the same plan.
	got := capture(t, func() error { return runFix([]string{fixture(), "--plan", out}) })
	if !strings.Contains(got, "Nothing marked") {
		t.Errorf("--plan did not read the plan written by --out:\n%s", got)
	}

	if err := runFix([]string{fixture(), "--plan"}); err == nil ||
		!strings.Contains(err.Error(), "--plan needs a file path") {
		t.Errorf("error = %v, want a complaint about --plan", err)
	}
}

func TestFix_RejectsUnknownFlag(t *testing.T) {
	isolate(t)
	// --apply is gone: the plan is the preview, so there is nothing to opt into.
	err := runFix([]string{fixture(), "--apply"})
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("error = %v, want an unknown-flag complaint", err)
	}
}

// TestFix_SchemaMismatchIsRefused keeps fix from acting on a plan whose fields
// may have moved, since the file decides what gets written to a real project.
func TestFix_SchemaMismatchIsRefused(t *testing.T) {
	home := isolate(t)
	scanFixture(t)

	planPath := plan.Path(home, fixture())
	p, err := plan.Read(planPath)
	if err != nil {
		t.Fatal(err)
	}
	p.Schema = plan.Schema + 1
	if err := plan.Write(planPath, p); err != nil {
		t.Fatal(err)
	}

	if err := runFix([]string{fixture()}); err == nil {
		t.Fatal("expected fix to refuse a plan from an unknown schema")
	}
}
