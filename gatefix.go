package main

import (
	"github.com/verifisecurity/verifi/internal/candidate"
	"github.com/verifisecurity/verifi/internal/gate"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/reason"
)

// gatedFix pairs one package's recommendation with the evidence the gate saw and
// the authorization it returned. It is the single unit behind both the
// propose_fix and apply_fix MCP tools, so the two cannot disagree about what may
// be applied: propose reports the decision, apply obeys it.
type gatedFix struct {
	Rec      reason.Recommendation
	Evidence gate.Evidence
	Decision gate.Decision
}

// evaluateFixes runs the gate over every recommendation in an analysis, in the
// analysis's order. All the evidence comes from facts already computed upstream
// (candidate, inventory, usage); the gate itself adds no lookups.
func evaluateFixes(res *analysis) []gatedFix {
	byPurl := make(map[string]inventory.Package, len(res.inv.Packages))
	for _, p := range res.inv.Packages {
		byPurl[p.Purl] = p
	}

	out := make([]gatedFix, 0, len(res.recs))
	for i, r := range res.recs {
		// recs and cands are aligned 1:1 (reason.Explain preserves order), so
		// the candidate at the same index carries this package's distance and
		// residual.
		var c candidate.Candidate
		if i < len(res.cands) {
			c = res.cands[i]
		}
		p := byPurl[r.Purl]
		ev := gate.Evidence{
			Action:       r.Action,
			Direct:       p.Direct,
			Imported:     res.imported[r.Name],
			Scope:        p.Scope,
			TargetExists: r.Action == "upgrade" && r.Target != "",
			Distance:     c.Distance,
			Residual:     c.Residual,
		}
		out = append(out, gatedFix{Rec: r, Evidence: ev, Decision: gate.Evaluate(ev)})
	}
	return out
}
