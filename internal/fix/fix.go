// Package fix turns recommendations into a concrete, reviewable plan of changes,
// and knows the package-manager command that applies each one. It does not run
// anything; the caller decides whether to preview or apply. npm for now.
package fix

import "github.com/verifisecurity/verifi/internal/reason"

// Action is one change to make.
type Action struct {
	Kind    string   // "upgrade" or "remove"
	Name    string   // package
	From    string   // current version (upgrade)
	To      string   // target version (upgrade)
	Command []string // the package-manager command that applies it
	Reason  string   // why, from the recommendation
}

// Plan is the set of changes for a project, plus the packages left unactioned.
type Plan struct {
	Actions []Action
	Skipped []string // packages with no applicable fix (for example no published fix)
}

// Empty reports whether there is nothing to apply.
func (p Plan) Empty() bool { return len(p.Actions) == 0 }

// Build turns recommendations into a plan. npm commands for now.
func Build(recs []reason.Recommendation) Plan {
	var p Plan
	for _, r := range recs {
		switch r.Action {
		case "upgrade":
			p.Actions = append(p.Actions, Action{
				Kind:    "upgrade",
				Name:    r.Name,
				From:    r.Current,
				To:      r.Target,
				Command: []string{"npm", "install", r.Name + "@" + r.Target},
				Reason:  r.Reason,
			})
		case "remove":
			p.Actions = append(p.Actions, Action{
				Kind:    "remove",
				Name:    r.Name,
				From:    r.Current,
				Command: []string{"npm", "uninstall", r.Name},
				Reason:  r.Reason,
			})
		default:
			p.Skipped = append(p.Skipped, r.Name)
		}
	}
	return p
}
