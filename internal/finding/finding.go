// Package finding holds the Finding contract: one advisory that affects one
// installed package version. Match (internal/osv) produces these; the status
// view consumes them. Every finding is evidence-bound, it records where the
// claim came from. This shape moves to verifi-core once a second repo consumes
// it; for now it lives with the CLI.
package finding

// Finding is a single advisory matched against a single installed package
// version. A package with two advisories yields two findings; the view groups
// them for display.
type Finding struct {
	Purl          string     `json:"purl"`
	Name          string     `json:"name"`
	Version       string     `json:"version"`
	Advisory      string     `json:"advisory"`
	Aliases       []string   `json:"aliases,omitempty"`
	Severity      string     `json:"severity"`
	Summary       string     `json:"summary,omitempty"`
	FixedVersions []string   `json:"fixed_versions"`
	Source        string     `json:"source"`
	Evidence      []Evidence `json:"evidence"`
}

// Evidence is one citation backing a finding: which source made the claim and
// the reference within it.
type Evidence struct {
	Source string `json:"source"`
	Ref    string `json:"ref"`
	Detail string `json:"detail,omitempty"`
}
