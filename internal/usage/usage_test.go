package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageName(t *testing.T) {
	cases := map[string]string{
		"lodash":         "lodash",
		"lodash/fp":      "lodash",
		"@scope/pkg":     "@scope/pkg",
		"@scope/pkg/sub": "@scope/pkg",
		"./local":        "",
		"../up":          "",
		"/abs":           "",
		"node:fs":        "",
		"":               "",
	}
	for in, want := range cases {
		if got := packageName(in); got != want {
			t.Errorf("packageName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestScan(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.js"), `const x = require('lodash'); import y from '@scope/pkg';`)
	write(t, filepath.Join(dir, "c.ts"), `import {a} from 'axios'; import './local';`)
	// node_modules must be skipped.
	write(t, filepath.Join(dir, "node_modules", "foo", "b.js"), `require('should-be-ignored')`)

	got, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	for _, want := range []string{"lodash", "@scope/pkg", "axios"} {
		if !got[want] {
			t.Errorf("missing import %q, got %v", want, got)
		}
	}
	if got["should-be-ignored"] {
		t.Error("node_modules should be skipped")
	}
	if got["local"] {
		t.Error("relative import should be ignored")
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
