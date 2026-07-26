//go:build real_e2e

// This opt-in test drives the full fix loop against a real npm project with a
// real installed dependency, not the fixture: analyze, check the gate, preview
// (which must write nothing), apply (which actually runs npm), then a re-scan to
// prove the advisory is gone. It writes to a throwaway temp dir OUTSIDE the repo,
// needs npm and network, and is kept off the default path (make check,
// go test ./...) by the real_e2e build tag. Run it explicitly:
//
//	go test -tags real_e2e -run TestFixApply_Real -v .
//
// It uses the checked-in fixture OSV database so the advisory match is stable and
// does not depend on `verifi update`.
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/verifisecurity/verifi/internal/fix"
	"github.com/verifisecurity/verifi/internal/gate"
)

func TestFixApply_Real(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not installed; skipping real apply loop")
	}

	// A real project: lodash 4.17.11, which the fixture OSV database flags as
	// GHSA-35jh-r3h4-6jhm, fixed in 4.17.21. Kept in a temp dir outside the repo.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "package.json"), `{
  "name": "real-apply-app",
  "version": "1.0.0",
  "dependencies": { "lodash": "4.17.11" }
}`)
	writeFile(t, filepath.Join(dir, "index.js"), "const _ = require('lodash');\nmodule.exports = () => _.padStart('x', 3);\n")

	npm(t, dir, "install", "--no-audit", "--no-fund")

	db, err := filepath.Abs(filepath.Join("testdata", "osv"))
	if err != nil {
		t.Fatal(err)
	}

	// The gate must authorize the lodash upgrade at confirm. This is the check
	// that keeps the fix path from writing anything the gate has not cleared.
	res, err := analyze(dir, db, true)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	var gated bool
	for _, g := range evaluateFixes(res) {
		if g.Rec.Name != "lodash" {
			continue
		}
		gated = true
		if g.Decision.Authorization != gate.Confirm {
			t.Fatalf("lodash gate authorization = %q, want %q", g.Decision.Authorization, gate.Confirm)
		}
	}
	if !gated {
		t.Fatal("analyze recommended no fix for lodash")
	}

	// Preview must change nothing on disk.
	if err := runFix([]string{dir, "--db", db, "--offline"}); err != nil {
		t.Fatalf("preview: %v", err)
	}
	if v := installedVersion(t, dir, "lodash"); v != "4.17.11" {
		t.Fatalf("lodash changed during preview: version=%s", v)
	}

	// Applying actually upgrades.
	if err := runFix([]string{dir, "--apply", "--db", db, "--offline"}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if v := installedVersion(t, dir, "lodash"); v == "4.17.11" {
		t.Fatalf("lodash not upgraded after apply: version=%s", v)
	}

	// A re-scan should now find nothing left to fix.
	res, err = analyze(dir, db, true)
	if err != nil {
		t.Fatalf("re-analyze: %v", err)
	}
	if plan := fix.Build(res.recs); !plan.Empty() {
		t.Fatalf("expected no remaining fixes after apply, got %d", len(plan.Actions))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func npm(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("npm", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm %v: %v\n%s", args, err, out)
	}
}

func installedVersion(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "node_modules", name, "package.json"))
	if err != nil {
		t.Fatalf("read installed %s: %v", name, err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatal(err)
	}
	return pkg.Version
}
