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

// Exists reports whether a specific version of a package is published on its
// registry. Compute only recommends versions for which this returns true. A nil
// predicate means "do not filter" (offline: trust OSV's fixed versions).
type Exists func(name, version string) bool

// Compute produces one candidate per vulnerable package. Deterministic: sorted
// by package name, advisory lists sorted. A fixed version that does not exist
// on the registry (per exists) is skipped, so an advisory whose only fix is
// unpublished falls to residual instead of recommending a phantom version.
func Compute(findings []finding.Finding, exists Exists) []Candidate {
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
			fix := nearestFix(name, current, f.FixedVersions, exists)
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

// nearestFix returns the smallest fixed version greater than current that also
// exists on the registry, or "" if none qualify (no fix, or the only fixes are
// unpublished).
func nearestFix(name, current string, fixes []string, exists Exists) string {
	best := ""
	for _, f := range fixes {
		if semver.Compare(f, current) <= 0 {
			continue
		}
		if exists != nil && !exists(name, f) {
			continue
		}
		if best == "" || semver.Compare(f, best) < 0 {
			best = f
		}
	}
	return best
}
