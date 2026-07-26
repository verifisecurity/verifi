package osv

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMeta_RoundTrip(t *testing.T) {
	root := t.TempDir()
	if err := WriteMeta(root, "npm", 42); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	got, err := ReadMeta(root, "npm")
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if got.Ecosystem != "npm" || got.Count != 42 {
		t.Errorf("got %+v, want ecosystem npm count 42", got)
	}
	if got.Age() > time.Minute {
		t.Errorf("age %v, want a freshly written stamp to read as recent", got.Age())
	}
}

// The stamp must be a sibling of the advisory directory, so that Fetch wiping
// that directory and Load scanning it never touch the stamp.
func TestMeta_StoredBesideAdvisoryDir(t *testing.T) {
	root := t.TempDir()
	if err := WriteMeta(root, "npm", 1); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "npm.meta.json")); err != nil {
		t.Errorf("stamp not at <root>/npm.meta.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "npm", "meta.json")); !os.IsNotExist(err) {
		t.Errorf("stamp must not live inside the advisory dir")
	}
}

func TestMeta_Stale(t *testing.T) {
	old := Meta{Ecosystem: "npm", FetchedAt: time.Now().Add(-10 * 24 * time.Hour)}
	fresh := Meta{Ecosystem: "npm", FetchedAt: time.Now().Add(-1 * time.Hour)}
	week := 7 * 24 * time.Hour
	if !old.Stale(week) {
		t.Error("a 10-day-old cache should be stale against a 7-day threshold")
	}
	if fresh.Stale(week) {
		t.Error("a 1-hour-old cache should not be stale against a 7-day threshold")
	}
}

func TestReadMeta_MissingErrors(t *testing.T) {
	if _, err := ReadMeta(t.TempDir(), "npm"); err == nil {
		t.Error("ReadMeta should error when no stamp was written")
	}
}
