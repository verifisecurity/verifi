// Package candidate turns findings into concrete fix candidates: for each
// vulnerable package, the nearest single upgrade that clears the advisories we
// can clear. This is raw-OSV depth. It answers "which version to move to and
// how far a jump it is", not yet "is that version genuinely safe for your
// code"; the fix tree and impact analysis refine that later.
package candidate

import (
	"sort"

	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/semver"
)

// Candidate is one recommended fix for one package.
type Candidate struct {
	Purl     string   `json:"purl"`
	Name     string   `json:"name"`
	Current  string   `json:"current"`
	Action   string   `json:"action"` // "upgrade" or "none"
	Target   string   `json:"target,omitempty"`
	Clears   []string `json:"clears"`             // advisory ids the target resolves
	Residual []string `json:"residual,omitempty"` // advisories with no published fix
	Distance string   `json:"distance,omitempty"` // major | minor | patch
}

// Compute produces one candidate per vulnerable package. Deterministic: sorted
// by package name, advisory lists sorted.
func Compute(findings []finding.Finding) []Candidate {
	byPkg := map[string][]finding.Finding{}
	var order []string
	for _, f := range findings {
		if _, ok := byPkg[f.Name]; !ok {
			order = append(order, f.Name)
		}
		byPkg[f.Name] = append(byPkg[f.Name], f)
	}
	sort.Strings(order)

	var out []Candidate
	for _, name := range order {
		fs := byPkg[name]
		current := fs[0].Version
		var clears, residual []string
		target := ""
		for _, f := range fs {
			fix := nearestFix(current, f.FixedVersions)
			if fix == "" {
				residual = append(residual, f.Advisory)
				continue
			}
			clears = append(clears, f.Advisory)
			// The package target must clear every advisory, so take the
			// furthest of the per-advisory nearest fixes.
			if target == "" || semver.Compare(fix, target) > 0 {
				target = fix
			}
		}
		sort.Strings(clears)
		sort.Strings(residual)

		c := Candidate{Purl: fs[0].Purl, Name: name, Current: current, Clears: clears, Residual: residual}
		if target != "" {
			c.Action = "upgrade"
			c.Target = target
			c.Distance = semver.Bump(current, target)
		} else {
			c.Action = "none"
		}
		out = append(out, c)
	}
	return out
}

// nearestFix returns the smallest fixed version greater than current, or "" if
// none of the fixes move forward (for example an advisory with no fix).
func nearestFix(current string, fixes []string) string {
	best := ""
	for _, f := range fixes {
		if semver.Compare(f, current) <= 0 {
			continue
		}
		if best == "" || semver.Compare(f, best) < 0 {
			best = f
		}
	}
	return best
}
