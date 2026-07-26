package osv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Meta records when an ecosystem's OSV database was last fetched, so the tool
// can tell the user how current their advisories are. Silent staleness is the
// real risk for a security scan, not a lack of live querying: a cache that is
// weeks old can miss a day-old critical without any signal.
type Meta struct {
	Ecosystem string    `json:"ecosystem"`
	FetchedAt time.Time `json:"fetched_at"`
	Count     int       `json:"count"`
}

// metaPath is a sibling of the ecosystem's advisory directory, not a file
// inside it: Fetch wipes and rewrites that directory, and Load reads every
// .json in it, so the stamp has to live one level up to survive both.
func metaPath(cacheRoot, ecosystem string) string {
	return filepath.Join(cacheRoot, ecosystem+".meta.json")
}

// WriteMeta stamps the cache for an ecosystem with the current time and count.
func WriteMeta(cacheRoot, ecosystem string, count int) error {
	if err := os.MkdirAll(cacheRoot, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(Meta{
		Ecosystem: ecosystem,
		FetchedAt: time.Now().UTC(),
		Count:     count,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath(cacheRoot, ecosystem), data, 0o644)
}

// ReadMeta reads the stamp for an ecosystem. It errors if none was written.
func ReadMeta(cacheRoot, ecosystem string) (Meta, error) {
	var m Meta
	data, err := os.ReadFile(metaPath(cacheRoot, ecosystem))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

// Age reports how long since the database was fetched.
func (m Meta) Age() time.Duration { return time.Since(m.FetchedAt) }

// Stale reports whether the database is older than threshold.
func (m Meta) Stale(threshold time.Duration) bool { return m.Age() > threshold }
