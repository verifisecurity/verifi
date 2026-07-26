//go:build real_e2e

// This opt-in test drives the whole loop against a real npm project with a real
// installed dependency, not the fixture: scan writes a plan, a fix with nothing
// marked changes nothing, marking one entry and running fix actually invokes npm,
// and a re-scan proves the advisory is gone. It writes to a throwaway temp dir
// OUTSIDE the repo, needs npm and network, and is kept off the default path
// (make check, go test ./...) by the real_e2e build tag. Run it explicitly:
//
//	go test -tags real_e2e -run TestFixLoop_Real -v .
//
// It uses the checked-in fixture OSV database so the advisory match is stable and
// does not depend on a download.
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/verifisecurity/verifi/internal/gate"
	"github.com/verifisecurity/verifi/internal/plan"
)

func TestFixLoop_Real(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not installed; skipping the real fix loop")
	}
	home := isolate(t)

	// A real project: lodash 4.17.11, which the fixture OSV database flags as
	// GHSA-35jh-r3h4-6jhm, fixed in 4.17.21. Kept outside the repo.
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
	scan := func() plan.Plan {
		t.Helper()
		capture(t, func() error { return runScan([]string{dir, "--db", db, "--offline"}) })
		p, err := plan.Read(plan.Path(home, dir))
		if err != nil {
			t.Fatalf("read plan: %v", err)
		}
		return p
	}

	// The scan gates the lodash upgrade at confirm, and marks nothing.
	p := scan()
	var lodash int = -1
	for i, c := range p.Candidates {
		if c.Package == "lodash" {
			lodash = i
		}
	}
	if lodash < 0 {
		t.Fatal("scan found no fix for lodash")
	}
	if got := p.Candidates[lodash].Gate.Authorization; got != gate.Confirm {
		t.Fatalf("lodash authorization = %q, want %q", got, gate.Confirm)
	}

	// A fix with nothing marked must not touch the project.
	capture(t, func() error { return runFix([]string{dir}) })
	if v := installedVersion(t, dir, "lodash"); v != "4.17.11" {
		t.Fatalf("fix changed lodash with nothing marked: version=%s", v)
	}

	// Mark it the way a human would, by editing the file.
	p.Candidates[lodash].Apply = true
	if err := plan.Write(plan.Path(home, dir), p); err != nil {
		t.Fatal(err)
	}

	// Now fix actually runs npm.
	capture(t, func() error { return runFix([]string{dir}) })
	if v := installedVersion(t, dir, "lodash"); v == "4.17.11" {
		t.Fatalf("lodash not upgraded after a marked fix: version=%s", v)
	}

	// A re-scan finds nothing left to fix.
	for _, c := range scan().Candidates {
		if c.Actionable() {
			t.Errorf("expected nothing left to fix, still have %s", c.Package)
		}
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
