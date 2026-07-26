package main

import (
	"fmt"
	"time"

	"github.com/verifisecurity/verifi/internal/osv"
	"github.com/verifisecurity/verifi/internal/plan"
)

// osvStaleAfter is when we start nudging the user to refresh the advisory
// cache. A week balances signal against noise: OSV changes daily, but a scan a
// few days old is not yet a real risk.
const osvStaleAfter = 7 * 24 * time.Hour

// ageDesc renders a cache age for humans: "updated today", "1 day old", "9 days old".
func ageDesc(d time.Duration) string {
	days := int(d.Hours() / 24)
	switch {
	case days <= 0:
		return "updated today"
	case days == 1:
		return "1 day old"
	default:
		return fmt.Sprintf("%d days old", days)
	}
}

// planStaleAfter is when a plan is old enough to mention before acting on it.
// The project can move under a plan (a dependency added, a lockfile regenerated)
// and fix would still run what the scan decided, so a day is enough to say so.
const planStaleAfter = 24 * time.Hour

// freshnessLine is the informational cache-age line for the scan view, or ""
// when there is no stamp (a user-supplied --db, or a cache from before stamping).
func freshnessLine(m *osv.Meta) string {
	if m == nil {
		return ""
	}
	line := "OSV data " + ageDesc(m.Age())
	if m.Stale(osvStaleAfter) {
		line += ", stale: re-run with --download"
	}
	return line
}

// planWarnings reports what is old about a plan, before fix acts on it. fix
// analyses nothing of its own, so both answers come from what the scan recorded
// in the plan rather than from a fresh look at the data.
//
// Silent staleness is the real risk for a security tool: a plan computed from
// week-old advisories, or against a lockfile that has since changed, still
// applies cleanly and tells you nothing.
func planWarnings(p plan.Plan, at time.Time) []string {
	var out []string
	if f := p.Advisories.FetchedAt; f != nil {
		if age := at.Sub(*f); age > osvStaleAfter {
			out = append(out, fmt.Sprintf(
				"the advisory data behind this plan is %s; re-run `verifi scan --download` for current advisories",
				ageDesc(age)))
		}
	}
	if age := at.Sub(p.Generated); age > planStaleAfter {
		out = append(out, fmt.Sprintf(
			"this plan was written %s; re-run `verifi scan` if the project has changed since",
			ageDesc(age)))
	}
	return out
}
