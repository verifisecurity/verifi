// Package reason explains a fix candidate: why it clears the advisory, how far
// the jump is, and, just as important, what has not been checked. The text is
// templated from structured facts, deterministic and reproducible, never
// generated. Confidence is stated honestly:
//
//   - advisory:   the advisory is cleared by this version. Code impact and
//     behaviour are not checked. This is the raw-OSV baseline, below structural.
//   - structural: no symbol the project uses changed (needs impact analysis).
//   - behavioural: the project's own tests pass on the new version (needs a run).
//
// See ADR 0005. Only behavioural earns an unattended fix.
package reason

import (
	"fmt"
	"strings"

	"github.com/verifisecurity/verifi/internal/candidate"
)

// Recommendation is the reasoned verdict for one package's fix.
type Recommendation struct {
	Purl        string   `json:"purl"`
	Name        string   `json:"name"`
	Action      string   `json:"action"` // upgrade | none
	Target      string   `json:"target,omitempty"`
	Confidence  string   `json:"confidence"` // advisory | structural | behavioural
	Reason      string   `json:"reason"`
	Limitations []string `json:"limitations"`
	Evidence    []string `json:"evidence"`
}

// Explain turns candidates into reasoned recommendations.
func Explain(cands []candidate.Candidate) []Recommendation {
	out := make([]Recommendation, 0, len(cands))
	for _, c := range cands {
		r := Recommendation{
			Purl:       c.Purl,
			Name:       c.Name,
			Action:     c.Action,
			Target:     c.Target,
			Confidence: "advisory",
			Evidence:   evidence(c),
		}
		if c.Action != "upgrade" {
			r.Reason = fmt.Sprintf("No published version clears %s.", join(c.Residual))
			r.Limitations = []string{
				"No safe version to upgrade to yet.",
				"Consider removing or replacing the package if it is not essential.",
			}
			out = append(out, r)
			continue
		}
		r.Reason = fmt.Sprintf("Clears %s (fixed in %s per OSV), a %s bump from %s.",
			join(c.Clears), c.Target, c.Distance, c.Current)
		r.Limitations = limitations(c)
		out = append(out, r)
	}
	return out
}

func limitations(c candidate.Candidate) []string {
	var lim []string
	switch c.Distance {
	case "major":
		lim = append(lim, "Major version bump: the public API may change, review before applying.")
	case "minor":
		lim = append(lim, "Minor bump: adds features, breaking changes are possible but uncommon.")
	}
	lim = append(lim,
		"Code impact not checked: not yet verified whether your code uses a changed part of the package.",
		"Behaviour not verified: run your tests after upgrading.")
	if len(c.Residual) > 0 {
		lim = append(lim, fmt.Sprintf("Still exposed to %s: no published fix.", join(c.Residual)))
	}
	return lim
}

func evidence(c candidate.Candidate) []string {
	ev := []string{"OSV"}
	ev = append(ev, c.Clears...)
	ev = append(ev, c.Residual...)
	return ev
}

func join(ids []string) string {
	if len(ids) == 0 {
		return "none"
	}
	return strings.Join(ids, ", ")
}
