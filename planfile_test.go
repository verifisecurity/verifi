package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/verifisecurity/verifi/internal/gate"
	"github.com/verifisecurity/verifi/internal/plan"
)

// isolate points the verifi home at a temp directory and freezes the clock, so a
// scan writes a plan that is comparable to a golden file and never touches the
// real home directory.
func isolate(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("VERIFI_HOME", home)

	orig := now
	now = func() time.Time { return time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { now = orig })

	return home
}

func scanFixture(t *testing.T, extra ...string) {
	t.Helper()
	args := append([]string{fixture(), "--db", fixtureDB(), "--offline"}, extra...)
	capture(t, func() error { return runScan(args) })
}

func TestScan_WritesPlan_Golden(t *testing.T) {
	home := isolate(t)
	scanFixture(t)

	got, err := os.ReadFile(plan.Path(home, fixture()))
	if err != nil {
		t.Fatalf("scan wrote no plan: %v", err)
	}
	// The workspace is an absolute path, so it differs per machine and cannot be
	// pinned in a golden file. Everything else is deterministic.
	var p plan.Plan
	if err := json.Unmarshal(got, &p); err != nil {
		t.Fatal(err)
	}
	p.Workspace = "<workspace>"
	normalised, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	checkGolden(t, "expected.plan.json", string(normalised))
}

// TestScan_PlanStartsUnmarked is the safety property of the whole design: a scan
// never marks anything, so a fix straight after a scan can only be a no-op. Every
// change to a project starts with a human editing this file.
func TestScan_PlanStartsUnmarked(t *testing.T) {
	home := isolate(t)
	scanFixture(t)

	p, err := plan.Read(plan.Path(home, fixture()))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Candidates) == 0 {
		t.Fatal("plan has no candidates, the fixture should be vulnerable")
	}
	for _, c := range p.Candidates {
		if c.Apply {
			t.Errorf("scan marked %q for apply; only a human may do that", c.Package)
		}
	}
	if got := p.Selected(); len(got) != 0 {
		t.Errorf("a fresh plan selected %d candidates, want 0", len(got))
	}
}

// TestScan_PlanCarriesGateAndEvidence covers what the MCP removal took out of
// reach: the gate's verdict and the evidence behind it had no CLI surface, and
// the plan is now the only place they appear.
func TestScan_PlanCarriesGateAndEvidence(t *testing.T) {
	home := isolate(t)
	scanFixture(t)

	p, err := plan.Read(plan.Path(home, fixture()))
	if err != nil {
		t.Fatal(err)
	}
	byPkg := map[string]plan.Candidate{}
	for _, c := range p.Candidates {
		byPkg[c.Package] = c
	}

	lodash, ok := byPkg["lodash"]
	if !ok {
		t.Fatal("no candidate for lodash")
	}
	if lodash.Gate.Authorization != gate.Confirm {
		t.Errorf("lodash authorization = %q, want %q", lodash.Gate.Authorization, gate.Confirm)
	}
	if lodash.Gate.Rung != gate.RungAdvisory {
		t.Errorf("lodash rung = %q, want %q", lodash.Gate.Rung, gate.RungAdvisory)
	}
	if len(lodash.Gate.Reasons) == 0 {
		t.Error("gate decision carries no reasons")
	}
	// Evidence and the detail the terminal view has no room for.
	if !lodash.Evidence.Direct || !lodash.Evidence.Imported || !lodash.Evidence.TargetExists {
		t.Errorf("lodash evidence looks wrong: %+v", lodash.Evidence)
	}
	if lodash.Distance != "patch" {
		t.Errorf("lodash distance = %q, want patch", lodash.Distance)
	}
	if len(lodash.Clears) == 0 {
		t.Error("lodash clears nothing, expected an advisory id")
	}
	if len(lodash.Command) == 0 {
		t.Error("lodash carries no command, so fix would have nothing to run")
	}
}

func TestScan_PlanRecordsItsAdvisorySource(t *testing.T) {
	home := isolate(t)
	scanFixture(t)

	p, err := plan.Read(plan.Path(home, fixture()))
	if err != nil {
		t.Fatal(err)
	}
	if p.Advisories.Source != "db" {
		t.Errorf("source = %q, want db (the scan passed --db)", p.Advisories.Source)
	}
	if p.Advisories.Path == "" {
		t.Error("plan does not record which database it matched against")
	}
	if p.Schema != plan.Schema || p.Ecosystem != "npm" {
		t.Errorf("plan header wrong: schema=%d ecosystem=%q", p.Schema, p.Ecosystem)
	}
}

func TestScan_OutOverridesPlanPath(t *testing.T) {
	isolate(t)
	out := filepath.Join(t.TempDir(), "my-plan.json")
	scanFixture(t, "--out", out)

	if _, err := plan.Read(out); err != nil {
		t.Fatalf("--out did not write the plan there: %v", err)
	}
}

// TestScan_PlanStaysOutOfTheWorkspace pins the storage rule that keeps scanning
// someone's repo from dirtying it.
func TestScan_PlanStaysOutOfTheWorkspace(t *testing.T) {
	isolate(t)
	scanFixture(t)

	if _, err := os.Stat(filepath.Join(fixture(), plan.Name)); !os.IsNotExist(err) {
		t.Errorf("scan wrote a plan into the scanned project")
	}
}

func TestScan_OutNeedsAValue(t *testing.T) {
	isolate(t)
	err := runScan([]string{fixture(), "--out"})
	if err == nil || !strings.Contains(err.Error(), "--out needs a file path") {
		t.Errorf("error = %v, want a complaint about --out", err)
	}
}
