package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/verifisecurity/verifi/internal/fix"
	"github.com/verifisecurity/verifi/internal/gate"
	"github.com/verifisecurity/verifi/internal/reason"
)

// capabilityHandlers binds each capability (internal/capability, the source of
// truth) to the function that runs it. Each runs the same core path as the
// matching CLI command, in process, never a shell. Every capability in the
// registry must have a handler here, enforced by TestMcpServerMatchesCapabilities.
var capabilityHandlers = map[string]func(args json.RawMessage) (string, error){
	"scan_workspace":    scanWorkspace,
	"list_dependencies": listDependencies,
	"propose_fix":       proposeFix,
	"apply_fix":         applyFix,
}

// scanWorkspace runs the same read-only pipeline as `verifi status`, returning
// the findings as JSON.
func scanWorkspace(args json.RawMessage) (string, error) {
	var in struct {
		Workspace string `json:"workspace"`
		DB        string `json:"db"`
		Offline   bool   `json:"offline"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}
	if in.Workspace == "" {
		in.Workspace = "."
	}
	res, err := analyze(in.Workspace, in.DB, in.Offline)
	if err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(res.findings, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// listDependencies runs the same resolution as `verifi inspect`, returning the
// inventory JSON or a CycloneDX SBOM.
func listDependencies(args json.RawMessage) (string, error) {
	var in struct {
		Workspace string `json:"workspace"`
		SBOM      bool   `json:"sbom"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}
	if in.Workspace == "" {
		in.Workspace = "."
	}
	inv, err := loadInventory(in.Workspace)
	if err != nil {
		return "", err
	}
	if in.SBOM {
		out, err := inv.ToCycloneDX()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
	out, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// proposal is one package's fix as reported by propose_fix: the plan (what to
// run), the evidence, and the gate decision. It writes nothing.
type proposal struct {
	Package  string        `json:"package"`
	Action   string        `json:"action"` // upgrade | remove | none
	From     string        `json:"from,omitempty"`
	To       string        `json:"to,omitempty"`
	Command  []string      `json:"command,omitempty"`
	Reason   string        `json:"reason"`
	Gate     gate.Decision `json:"gate"`
	Evidence gate.Evidence `json:"evidence"`
}

// proposeFix runs the same read-only analysis as `verifi fix` (preview) and, for
// every vulnerable package, returns the plan, the evidence, and the gate
// decision. It never writes; applying is apply_fix's job, behind the gate.
func proposeFix(args json.RawMessage) (string, error) {
	var in struct {
		Workspace string `json:"workspace"`
		DB        string `json:"db"`
		Offline   bool   `json:"offline"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}
	if in.Workspace == "" {
		in.Workspace = "."
	}
	res, err := analyze(in.Workspace, in.DB, in.Offline)
	if err != nil {
		return "", err
	}

	// The plan gives the concrete command, from, and to per actionable package.
	actByName := map[string]fix.Action{}
	for _, a := range fix.Build(res.recs).Actions {
		actByName[a.Name] = a
	}

	gfs := evaluateFixes(res)
	out := make([]proposal, 0, len(gfs))
	for _, g := range gfs {
		p := proposal{
			Package:  g.Rec.Name,
			Action:   g.Rec.Action,
			Reason:   g.Rec.Reason,
			Gate:     g.Decision,
			Evidence: g.Evidence,
		}
		if a, ok := actByName[g.Rec.Name]; ok {
			p.From, p.To, p.Command = a.From, a.To, a.Command
		}
		out = append(out, p)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// refusal records a package apply_fix declined to apply, and why.
type refusal struct {
	Package string `json:"package"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
}

// applyResult is what apply_fix reports back.
type applyResult struct {
	Applied []string  `json:"applied,omitempty"`
	Refused []refusal `json:"refused,omitempty"`
	Output  string    `json:"output,omitempty"`
	Note    string    `json:"note,omitempty"`
}

// applyFix applies the recommended fix, but only where the gate authorizes an
// apply (authorization "confirm") AND the caller passed confirm=true. A fix the
// gate can only propose is never applied, whatever confirm says, and confirm
// must be the human's, relayed by the model, not chosen by it. It runs the same
// fix core as `verifi fix --apply` (fix.Build + the shared apply path).
func applyFix(args json.RawMessage) (string, error) {
	var in struct {
		Workspace string `json:"workspace"`
		DB        string `json:"db"`
		Offline   bool   `json:"offline"`
		Package   string `json:"package"`
		Confirm   bool   `json:"confirm"`
	}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &in); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}
	if in.Workspace == "" {
		in.Workspace = "."
	}
	res, err := analyze(in.Workspace, in.DB, in.Offline)
	if err != nil {
		return "", err
	}

	var eligible []reason.Recommendation
	var refused []refusal
	matched := false
	for _, g := range evaluateFixes(res) {
		if in.Package != "" && g.Rec.Name != in.Package {
			continue
		}
		matched = true
		switch {
		case g.Rec.Action == "none":
			refused = append(refused, refusal{g.Rec.Name, g.Rec.Action, "no fix to apply"})
		case g.Decision.Authorization != gate.Confirm:
			refused = append(refused, refusal{g.Rec.Name, g.Rec.Action,
				"gate authorization is " + g.Decision.Authorization + "; this fix can only be proposed, not applied"})
		case !in.Confirm:
			refused = append(refused, refusal{g.Rec.Name, g.Rec.Action,
				"needs explicit human confirmation; call again with confirm=true"})
		default:
			eligible = append(eligible, g.Rec)
		}
	}
	if in.Package != "" && !matched {
		return "", fmt.Errorf("no recommended fix for package %q", in.Package)
	}

	result := applyResult{Refused: refused}
	plan := fix.Build(eligible)
	if !plan.Empty() {
		var buf bytes.Buffer
		if err := applyPlanTo(&buf, &buf, in.Workspace, plan); err != nil {
			return "", err
		}
		for _, a := range plan.Actions {
			result.Applied = append(result.Applied, a.Name)
		}
		result.Output = buf.String()
	}
	if len(result.Applied) == 0 {
		if !in.Confirm {
			result.Note = "Nothing applied. Set confirm=true to apply the fixes the gate authorizes."
		} else {
			result.Note = "Nothing applied; no fix was both gate-authorized and confirmed."
		}
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
