// Package plan is the contract between scan and fix. `verifi scan` writes a
// plan: every fix it found, with the evidence behind it and the gate's verdict,
// each entry marked apply=false. A human opens it and flips apply to true on the
// ones they want. `verifi fix` reads it back and applies only those.
//
// That file is the whole human loop, and it is why fix needs no preview mode:
// the plan is the preview, and it is a real artifact you can read, diff, and
// keep, rather than terminal output that disappears.
//
// Nothing here decides anything. The schema and its reader and writer live
// together so the two commands cannot drift, and both sides import this rather
// than reaching into each other.
package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/verifisecurity/verifi/internal/gate"
)

// Schema is the plan format version. fix refuses a plan it does not understand
// rather than guessing at fields that may have moved.
const Schema = 1

// Name is the plan's filename inside a project's directory.
const Name = "candidates.json"

// Plan is one scan's output: what was found, what it was matched against, and
// the fix for each vulnerable package.
type Plan struct {
	Schema     int         `json:"schema"`
	Generated  time.Time   `json:"generated"`
	Verifi     string      `json:"verifi"`    // CLI version that wrote it
	Workspace  string      `json:"workspace"` // absolute path to the project
	Ecosystem  string      `json:"ecosystem"`
	Advisories Advisories  `json:"advisories"`
	Candidates []Candidate `json:"candidates"`
}

// Advisories records which advisory data the scan matched against, so a plan
// says what it was computed from and a stale one can be recognised as stale.
type Advisories struct {
	Source    string     `json:"source"` // "cache" | "db"
	Path      string     `json:"path"`
	FetchedAt *time.Time `json:"fetchedAt,omitempty"` // cache only; a --db directory is not age-checked
	Count     int        `json:"count,omitempty"`
}

// Candidate is one package's fix, and the decision about it.
//
// Apply is deliberately the first field: it is the only one a human is meant to
// edit, so it reads first in every entry. Everything after it is the evidence
// for that decision, written by scan and not meant to be hand-edited.
type Candidate struct {
	Apply bool `json:"apply"`

	Package     string        `json:"package"`
	Current     string        `json:"current"`
	Action      string        `json:"action"` // upgrade | remove | none
	Target      string        `json:"target,omitempty"`
	Command     []string      `json:"command,omitempty"` // the package-manager command fix would run
	Distance    string        `json:"distance,omitempty"`
	Clears      []string      `json:"clears,omitempty"`
	Residual    []string      `json:"residual,omitempty"`
	Used        string        `json:"used"`
	Reason      string        `json:"reason"`
	Limitations []string      `json:"limitations,omitempty"`
	Gate        gate.Decision `json:"gate"`
	Evidence    gate.Evidence `json:"evidence"`
}

// Actionable reports whether this candidate has a fix that could be run at all,
// regardless of whether anyone marked it.
func (c Candidate) Actionable() bool { return c.Action != "none" && len(c.Command) > 0 }

// Selected is what fix acts on: marked by a human, and actually runnable.
func (p Plan) Selected() []Candidate {
	var out []Candidate
	for _, c := range p.Candidates {
		if c.Apply && c.Actionable() {
			out = append(out, c)
		}
	}
	return out
}

// Dir is where a project's plan lives: one directory per workspace under the
// verifi home, keyed by a digest of its absolute path. Plans stay out of the
// scanned repo so scanning cannot dirty someone's working tree.
func Dir(home, workspace string) string {
	abs, err := filepath.Abs(workspace)
	if err != nil {
		abs = workspace
	}
	return filepath.Join(home, "projects", projectID(abs))
}

// Path is the plan file for a workspace.
func Path(home, workspace string) string { return filepath.Join(Dir(home, workspace), Name) }

// Write saves a plan, creating its directory. The file is rewritten by every
// scan: it is the current plan, not a record. The append-only history is
// separate.
func Write(path string, p Plan) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Read loads a plan and rejects one written by a schema this build does not
// understand, rather than silently misreading fields that may have moved.
func Read(path string) (Plan, error) {
	var p Plan
	data, err := os.ReadFile(path)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("%s is not a readable plan: %w", path, err)
	}
	if p.Schema != Schema {
		return p, fmt.Errorf("%s was written for plan schema %d, this build reads %d: re-run `verifi scan`",
			path, p.Schema, Schema)
	}
	return p, nil
}
