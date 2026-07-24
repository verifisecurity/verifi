// Package semver compares npm-style semantic versions with the standard
// library only. It is deliberately small: enough to evaluate OSV version
// ranges (is this version affected, which fixed version is nearest), not a full
// semver range grammar. See https://semver.org for the precedence rules.
package semver

import (
	"strconv"
	"strings"
)

// Compare returns -1, 0, or 1 as version a is less than, equal to, or greater
// than b. It parses an optional leading "v", a MAJOR.MINOR.PATCH core (missing
// parts count as 0), an optional "-prerelease", and ignores "+build" metadata.
// A release outranks any of its prereleases, and prerelease identifiers compare
// per the spec: numeric identifiers numerically, others lexically, numeric lower
// than non-numeric.
func Compare(a, b string) int {
	ac, ap := split(a)
	bc, bp := split(b)
	for i := 0; i < 3; i++ {
		if ac[i] != bc[i] {
			if ac[i] < bc[i] {
				return -1
			}
			return 1
		}
	}
	return comparePre(ap, bp)
}

func split(v string) ([3]int, string) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	pre := ""
	if i := strings.IndexByte(v, '-'); i >= 0 {
		pre = v[i+1:]
		v = v[:i]
	}
	var core [3]int
	for i, part := range strings.SplitN(v, ".", 3) {
		n, _ := strconv.Atoi(strings.TrimSpace(part))
		core[i] = n
	}
	return core, pre
}

func comparePre(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" { // a release outranks any prerelease
		return 1
	}
	if b == "" {
		return -1
	}
	ai, bi := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(ai) && i < len(bi); i++ {
		if c := compareIdent(ai[i], bi[i]); c != 0 {
			return c
		}
	}
	switch { // a larger set of identifiers wins when the shared ones are equal
	case len(ai) < len(bi):
		return -1
	case len(ai) > len(bi):
		return 1
	}
	return 0
}

func compareIdent(a, b string) int {
	an, aerr := strconv.Atoi(a)
	bn, berr := strconv.Atoi(b)
	switch {
	case aerr == nil && berr == nil:
		return cmpInt(an, bn)
	case aerr == nil: // numeric identifiers rank below non-numeric
		return -1
	case berr == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
