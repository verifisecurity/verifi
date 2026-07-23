package inventory

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "regenerate golden files")

func TestParseNpmLock_Table(t *testing.T) {
	const lock = `{"name":"app","version":"1.0.0","lockfileVersion":3,"packages":{
		"":{"name":"app","version":"1.0.0","dependencies":{"left-pad":"1.3.0"},"devDependencies":{"is-odd":"3.0.1"}},
		"node_modules/left-pad":{"version":"1.3.0"},
		"node_modules/is-odd":{"version":"3.0.1","dev":true},
		"node_modules/is-number":{"version":"6.0.0","dev":true}}}`

	inv, err := ParseNpmLock([]byte(lock))
	if err != nil {
		t.Fatalf("ParseNpmLock: %v", err)
	}
	if inv.Ecosystem != "npm" {
		t.Errorf("ecosystem = %q, want npm", inv.Ecosystem)
	}
	if inv.Root.Name != "app" || inv.Root.Version != "1.0.0" {
		t.Errorf("root = %+v, want app@1.0.0", inv.Root)
	}
	if len(inv.Packages) != 3 {
		t.Fatalf("packages = %d, want 3", len(inv.Packages))
	}

	wantDirect := map[string]bool{"left-pad": true, "is-odd": true, "is-number": false}
	wantScope := map[string]string{"left-pad": "prod", "is-odd": "dev", "is-number": "dev"}
	for _, p := range inv.Packages {
		if p.Direct != wantDirect[p.Name] {
			t.Errorf("%s direct = %v, want %v", p.Name, p.Direct, wantDirect[p.Name])
		}
		if p.Scope != wantScope[p.Name] {
			t.Errorf("%s scope = %q, want %q", p.Name, p.Scope, wantScope[p.Name])
		}
		want := "pkg:npm/" + p.Name + "@" + p.Version
		if p.Purl != want {
			t.Errorf("%s purl = %q, want %q", p.Name, p.Purl, want)
		}
	}
}

func TestParseNpmLock_Errors(t *testing.T) {
	if _, err := ParseNpmLock([]byte("not json")); err == nil {
		t.Error("expected error for invalid json")
	}
	if _, err := ParseNpmLock([]byte(`{"lockfileVersion":1}`)); err == nil {
		t.Error("expected error for missing packages map")
	}
}

func TestNpmName(t *testing.T) {
	cases := map[string]string{
		"node_modules/left-pad":                   "left-pad",
		"node_modules/@scope/pkg":                 "@scope/pkg",
		"node_modules/a/node_modules/b":           "b",
		"node_modules/a/node_modules/@scope/deep": "@scope/deep",
		"": "",
	}
	for in, want := range cases {
		if got := npmName(in); got != want {
			t.Errorf("npmName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestInventory_Golden is the fixture-based end-to-end check: parse a real
// package-lock.json and assert the inventory and SBOM against golden files.
// Regenerate with: go test ./... -update
func TestInventory_Golden(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "npm", "simple")
	data, err := os.ReadFile(filepath.Join(dir, "package-lock.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	inv, err := ParseNpmLock(data)
	if err != nil {
		t.Fatalf("ParseNpmLock: %v", err)
	}

	invJSON, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	sbom, err := inv.ToCycloneDX()
	if err != nil {
		t.Fatal(err)
	}

	checkGolden(t, filepath.Join(dir, "expected.inventory.json"), invJSON)
	checkGolden(t, filepath.Join(dir, "expected.sbom.json"), sbom)
}

func checkGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run: go test ./... -update): %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Errorf("%s mismatch\n--- got ---\n%s\n--- want ---\n%s", filepath.Base(path), got, want)
	}
}
