package main

import (
	"fmt"
	"time"

	"github.com/verifisecurity/verifi/internal/osv"
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

// freshnessLine is the informational cache-age line for the status view, or ""
// when there is no stamp (a user-supplied --db, or a cache from before stamping).
func freshnessLine(m *osv.Meta) string {
	if m == nil {
		return ""
	}
	line := "OSV data " + ageDesc(m.Age())
	if m.Stale(osvStaleAfter) {
		line += ", stale: run `verifi update`"
	}
	return line
}

// staleWarning is the stderr nudge printed before an action runs on stale data,
// or "" when the cache is fresh or unstamped.
func staleWarning(m *osv.Meta) string {
	if m == nil || !m.Stale(osvStaleAfter) {
		return ""
	}
	return fmt.Sprintf("verifi: OSV data is %s, run `verifi update` for current advisories", ageDesc(m.Age()))
}
