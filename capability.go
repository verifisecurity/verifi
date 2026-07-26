package main

import (
	"encoding/json"
	"fmt"
)

// capabilityHandlers binds each capability (internal/capability, the source of
// truth) to the function that runs it. Each runs the same core path as the
// matching CLI command, in process, never a shell. Every capability in the
// registry must have a handler here, enforced by TestMcpServerMatchesCapabilities.
var capabilityHandlers = map[string]func(args json.RawMessage) (string, error){
	"scan_workspace":    scanWorkspace,
	"list_dependencies": listDependencies,
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
