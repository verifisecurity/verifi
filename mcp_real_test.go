//go:build real_e2e

// This opt-in test drives the full gated apply loop against a real npm project
// with a real installed dependency, not the fixture: propose_fix, then apply_fix
// with confirm=true (which actually runs npm), then a re-scan to prove the
// advisory is gone. It writes to a throwaway temp dir OUTSIDE the repo, needs npm
// and network, and is kept off the default path (make check, go test ./...) by
// the real_e2e build tag. Run it explicitly:
//
//	go test -tags real_e2e -run TestMcpApply_Real -v .
//
// It uses the checked-in fixture OSV database so the advisory match is stable and
// does not depend on `verifi update`.
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMcpApply_Real(t *testing.T) {
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
	writeFile(t, filepath.Join(dir, "index.js"), `const _ = require('lodash');\nmodule.exports = () => _.padStart('x', 3);\n`)

	npm(t, dir, "install", "--no-audit", "--no-fund")

	db, err := filepath.Abs(filepath.Join("testdata", "osv"))
	if err != nil {
		t.Fatal(err)
	}
	call := func(name, args string) string {
		reqs := []string{
			`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"` + name + `","arguments":` + args + `}}`,
		}
		var out strings.Builder
		if err := newMcpServer().Serve(strings.NewReader(strings.Join(reqs, "\n")+"\n"), &out); err != nil {
			t.Fatalf("serve %s: %v", name, err)
		}
		return lastToolText(t, out.String())
	}

	base := `{"workspace":"` + dir + `","db":"` + db + `","offline":true`

	// propose_fix should recommend the lodash upgrade and gate it at confirm.
	prop := call("propose_fix", base+`}`)
	if !strings.Contains(prop, `"package": "lodash"`) || !strings.Contains(prop, `"authorization": "confirm"`) {
		t.Fatalf("propose_fix did not gate lodash at confirm:\n%s", prop)
	}

	// apply_fix without confirm must change nothing.
	if got := call("apply_fix", base+`,"confirm":false}`); !strings.Contains(got, `"refused"`) || strings.Contains(got, `"applied"`) {
		t.Fatalf("apply_fix without confirm should refuse and apply nothing:\n%s", got)
	}
	if v := installedVersion(t, dir, "lodash"); v != "4.17.11" {
		t.Fatalf("lodash changed without confirm: version=%s", v)
	}

	// apply_fix with confirm actually upgrades.
	if got := call("apply_fix", base+`,"confirm":true}`); !strings.Contains(got, `"applied"`) {
		t.Fatalf("apply_fix with confirm should apply lodash:\n%s", got)
	}
	if v := installedVersion(t, dir, "lodash"); v == "4.17.11" {
		t.Fatalf("lodash not upgraded after confirmed apply: version=%s", v)
	}

	// A re-scan should now find nothing to fix.
	if got := call("propose_fix", base+`}`); strings.TrimSpace(got) != "[]" {
		t.Fatalf("expected no remaining fixes after apply, got:\n%s", got)
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
		t.Fatalf("npm %s: %v\n%s", strings.Join(args, " "), err, out)
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

// lastToolText pulls the text of the last tool result out of a transcript.
func lastToolText(t *testing.T, transcript string) string {
	t.Helper()
	var text string
	for _, line := range strings.Split(strings.TrimSpace(transcript), "\n") {
		if line == "" {
			continue
		}
		var resp struct {
			Result struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}
		if len(resp.Result.Content) > 0 {
			text = resp.Result.Content[0].Text
		}
	}
	if text == "" {
		t.Fatalf("no tool text in transcript:\n%s", transcript)
	}
	return text
}
