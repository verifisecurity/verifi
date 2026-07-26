package main

import (
	"strings"
	"testing"
	"time"

	"github.com/verifisecurity/verifi/internal/osv"
	"github.com/verifisecurity/verifi/internal/plan"
)

// TestFreshnessLine_PointsAtRealCommand is the same regression guard the missing
// database has: a nudge is useless if it names a command this build removed.
func TestFreshnessLine_PointsAtRealCommand(t *testing.T) {
	stale := osv.Meta{Ecosystem: "npm", FetchedAt: time.Now().Add(-30 * 24 * time.Hour)}
	line := freshnessLine(&stale)
	if !strings.Contains(line, "--download") {
		t.Errorf("stale line does not offer --download: %q", line)
	}
	if strings.Contains(line, "verifi update") {
		t.Errorf("stale line names the removed update command: %q", line)
	}
}

// TestPlanWarnings covers the staleness fix cannot see for itself. It analyses
// nothing, so the only thing that can tell it the data is old is what the scan
// recorded in the plan.
func TestPlanWarnings(t *testing.T) {
	at := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	fresh := at.Add(-time.Hour)
	ancient := at.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name      string
		p         plan.Plan
		wantCount int
		wantAny   string
	}{
		{
			name:      "fresh plan and fresh advisories say nothing",
			p:         plan.Plan{Generated: fresh, Advisories: plan.Advisories{FetchedAt: &fresh}},
			wantCount: 0,
		},
		{
			name:      "stale advisories are called out",
			p:         plan.Plan{Generated: fresh, Advisories: plan.Advisories{FetchedAt: &ancient}},
			wantCount: 1,
			wantAny:   "--download",
		},
		{
			name:      "an old plan is called out",
			p:         plan.Plan{Generated: ancient, Advisories: plan.Advisories{FetchedAt: &fresh}},
			wantCount: 1,
			wantAny:   "if the project has changed",
		},
		{
			name:      "both can fire at once",
			p:         plan.Plan{Generated: ancient, Advisories: plan.Advisories{FetchedAt: &ancient}},
			wantCount: 2,
		},
		{
			// A --db directory carries no stamp, so there is nothing to judge.
			name:      "no stamp means no advisory warning",
			p:         plan.Plan{Generated: fresh},
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := planWarnings(tt.p, at)
			if len(got) != tt.wantCount {
				t.Fatalf("got %d warnings %v, want %d", len(got), got, tt.wantCount)
			}
			if tt.wantAny != "" && !strings.Contains(strings.Join(got, " "), tt.wantAny) {
				t.Errorf("warnings %v do not mention %q", got, tt.wantAny)
			}
		})
	}
}
