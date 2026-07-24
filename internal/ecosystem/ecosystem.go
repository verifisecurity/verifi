// Package ecosystem is the source of truth for which package ecosystems the CLI
// supports. The docs render from it, so "supported ecosystems" cannot drift:
// flipping Supported to true here is what puts an ecosystem on the README.
package ecosystem

// Ecosystem is a package ecosystem and the manifest the CLI reads for it.
type Ecosystem struct {
	Name      string
	Manifest  string
	Supported bool // false = planned, not built yet
}

// All is every ecosystem, supported and planned.
var All = []Ecosystem{
	{"npm", "package-lock.json", true},
	{"PyPI", "poetry.lock, requirements.txt", false},
	{"Go", "go.mod, go.sum", false},
	{"Maven", "pom.xml", false},
	{"Cargo", "Cargo.lock", false},
}

// Supported returns the ecosystems that are built and usable today.
func Supported() []Ecosystem {
	var s []Ecosystem
	for _, e := range All {
		if e.Supported {
			s = append(s, e)
		}
	}
	return s
}

// HasPlanned reports whether any ecosystem is not yet supported.
func HasPlanned() bool {
	for _, e := range All {
		if !e.Supported {
			return true
		}
	}
	return false
}
