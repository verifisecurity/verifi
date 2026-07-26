package main

import (
	"path/filepath"
	"time"

	"github.com/verifisecurity/verifi/internal/candidate"
	"github.com/verifisecurity/verifi/internal/fix"
	"github.com/verifisecurity/verifi/internal/plan"
)

// now is the clock the plan is stamped with, a variable so tests can freeze it
// and compare a generated plan against a golden file.
var now = func() time.Time { return time.Now().UTC() }

// buildPlan assembles the file `verifi scan` writes and `verifi fix` reads: one
// entry per vulnerable package, carrying the concrete command, the evidence
// behind it, and the gate's verdict, every entry marked apply=false.
//
// Nothing here decides anything. It gathers what the pipeline already computed
// and records it, including the parts the terminal view has room to summarise
// but not to show: the version distance, the advisories a fix leaves behind, and
// the gate's reasoning.
func buildPlan(res *analysis, workspace, verifiVersion string) plan.Plan {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}

	// The command per package comes from the same builder fix runs, so the plan
	// records exactly what would be executed rather than a description of it.
	actionFor := map[string]fix.Action{}
	for _, a := range fix.Build(res.recs).Actions {
		actionFor[a.Name] = a
	}
	candFor := map[string]candidate.Candidate{}
	for _, c := range res.cands {
		candFor[c.Name] = c
	}

	cands := make([]plan.Candidate, 0, len(res.recs))
	for _, g := range evaluateFixes(res) {
		c := candFor[g.Rec.Name]
		pc := plan.Candidate{
			Apply:       false,
			Package:     g.Rec.Name,
			Current:     g.Rec.Current,
			Action:      g.Rec.Action,
			Target:      g.Rec.Target,
			Distance:    c.Distance,
			Clears:      c.Clears,
			Residual:    c.Residual,
			Used:        usageLine(res.inv, g.Rec.Name, res.imported),
			Reason:      g.Rec.Reason,
			Limitations: g.Rec.Limitations,
			Gate:        g.Decision,
			Evidence:    g.Evidence,
		}
		if a, ok := actionFor[g.Rec.Name]; ok {
			pc.Command = a.Command
		}
		cands = append(cands, pc)
	}

	adv := plan.Advisories{Source: "db", Path: res.dbPath}
	if res.dbCache {
		adv.Source = "cache"
		if res.dbMeta != nil {
			at := res.dbMeta.FetchedAt
			adv.FetchedAt = &at
			adv.Count = res.dbMeta.Count
		}
	}

	return plan.Plan{
		Schema:     plan.Schema,
		Verifi:     verifiVersion,
		Workspace:  abs,
		Ecosystem:  res.inv.Ecosystem,
		Advisories: adv,
		Candidates: cands,
	}
}
