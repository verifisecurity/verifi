package osv

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZip(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entries := map[string]string{
		"GHSA-aaa.json":     `{"id":"GHSA-aaa"}`,
		"sub/GHSA-bbb.json": `{"id":"GHSA-bbb"}`, // nested, should flatten
		"README.md":         "not an advisory",   // non-json, should skip
	}
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "npm")
	n, err := extractZip(bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest)
	if err != nil {
		t.Fatalf("extractZip: %v", err)
	}
	if n != 2 {
		t.Errorf("wrote %d, want 2 json advisories", n)
	}
	for _, want := range []string{"GHSA-aaa.json", "GHSA-bbb.json"} {
		if _, err := os.Stat(filepath.Join(dest, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dest, "README.md")); err == nil {
		t.Error("README.md should have been skipped")
	}
}

func TestExtractZip_ReplacesExisting(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "npm")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(dest, "stale.json")
	if err := os.WriteFile(stale, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("fresh.json")
	w.Write([]byte(`{"id":"fresh"}`))
	zw.Close()

	if _, err := extractZip(bytes.NewReader(buf.Bytes()), int64(buf.Len()), dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("stale advisory should have been removed on refresh")
	}
}
