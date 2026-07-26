// Package gate maps fix evidence to an authorization: given only facts already
// computed upstream (the action, whether the package is a direct dependency,
// whether your code imports it, its scope, whether a published target exists,
// how large the version jump is, and any advisories left unfixed), it decides
// how far verifi may go on its own. It is a pure, deterministic function: the
// same evidence in gives the same decision out. No network, no model, no side
// effects (see docs/adr/0001, stdlib only).
//
// Authorization is a ladder. In this release only two rungs are reachable:
//
//	confirm  verifi may apply the fix once a human confirms it
//	propose  verifi may only describe the fix, never apply it
//
// A third level, unattended (apply with no human in the loop), is defined so the
// set can grow without reshaping this type, but nothing in this release returns
// it: the gate only ever authorises confirm or propose.
package gate

import (
	"fmt"
	"strings"
)

// RungAdvisory is the confidence rung of this release: the advisory is cleared
// by the target version, but the fix's code impact and behaviour are unchecked.
const RungAdvisory = "advisory"

// Authorization levels, from most to least autonomy. Unattended is never returned
// in this release (see the package doc); it is defined so the set can grow
// without changing this type.
const (
	Unattended = "unattended"
	Confirm    = "confirm"
	Propose    = "propose"
)

// Evidence is the set of already-computed facts the gate decides from. Every
// field is produced earlier in the pipeline (candidate, inventory, usage); the
// gate performs no lookups of its own.
type Evidence struct {
	Action       string   `json:"action"`             // "upgrade" | "remove" | "none"
	Direct       bool     `json:"direct"`             // a direct dependency of the project
	Imported     bool     `json:"imported"`           // the project's own code imports it
	Scope        string   `json:"scope"`              // "prod" | "dev"
	TargetExists bool     `json:"targetExists"`       // a published fix version exists
	Distance     string   `json:"distance"`           // "patch" | "minor" | "major" | ""
	Residual     []string `json:"residual,omitempty"` // advisories with no published fix
}

// Decision is the gate's verdict for one fix.
type Decision struct {
	Rung          string   `json:"rung"`          // confidence rung, "advisory" for now
	Authorization string   `json:"authorization"` // "confirm" | "propose"
	Reasons       []string `json:"reasons"`       // why this authorization
	Warnings      []string `json:"warnings,omitempty"`
}

// Evaluate maps evidence to a decision. It is table-driven and total: every
// input lands on exactly one row, and the result is always confirm or propose,
// never unattended.
func Evaluate(e Evidence) Decision {
	d := Decision{Rung: RungAdvisory, Authorization: Propose}

	switch e.Action {
	case "remove":
		// Removing a direct dependency the code never imports clears its
		// advisories with no compatibility surface, so it may be applied on a
		// confirm. Any other removal (transitive, or one the code does import)
		// is not a clean call and stays a proposal.
		if e.Direct && !e.Imported {
			d.Authorization = Confirm
			d.Reasons = append(d.Reasons, "Removing an unused direct dependency clears its advisories with no compatibility surface.")
		} else {
			d.Reasons = append(d.Reasons, "Removal is not a clean call here (the package is transitive or your code imports it); describing it only.")
		}

	case "upgrade":
		switch {
		case !e.TargetExists:
			d.Reasons = append(d.Reasons, "No published version to upgrade to; describing the exposure only.")
		case !e.Direct:
			// A transitive package cannot be fixed by installing it. `npm install
			// pkg@version` adds a top-level entry for something the project never
			// depended on directly, and if the parent's range excludes the new
			// version npm keeps the vulnerable copy nested underneath, so the
			// advisory survives a change that reported success. Forcing a
			// transitive version needs an overrides entry, which is a different
			// change from an install, so this stays a proposal until verifi can
			// make it properly.
			d.Reasons = append(d.Reasons, "Pulled in by another package, so installing it directly would add a top-level pin and might not clear the advisory. Forcing a transitive version needs an overrides entry; describing it only.")
		case e.Distance == "patch" || e.Distance == "minor":
			d.Authorization = Confirm
			d.Reasons = append(d.Reasons, fmt.Sprintf("A %s upgrade to a published version clears the advisory.", e.Distance))
		case e.Distance == "major":
			d.Authorization = Confirm
			d.Reasons = append(d.Reasons, "A major upgrade to a published version clears the advisory.")
			d.Warnings = append(d.Warnings, "Major version bump: the public API may change. Review before applying.")
		default:
			d.Reasons = append(d.Reasons, "Unrecognised version distance; describing the exposure only.")
		}

	default: // "none", or any action with no fix
		d.Reasons = append(d.Reasons, "No fix available; describing the exposure only.")
	}

	if len(e.Residual) > 0 {
		d.Warnings = append(d.Warnings, "Still exposed after this fix: "+strings.Join(e.Residual, ", ")+" has no published fix.")
	}
	return d
}
