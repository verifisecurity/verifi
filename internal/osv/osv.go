// Package osv matches an inventory against a local OSV advisory database,
// offline. Nothing leaves the machine: the database is a directory of OSV JSON
// files (fetched separately) and matching is a lookup, not analysis. See the
// OSV schema at https://ossf.github.io/osv-schema/.
package osv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/verifisecurity/verifi/internal/finding"
	"github.com/verifisecurity/verifi/internal/inventory"
	"github.com/verifisecurity/verifi/internal/semver"
)

// Advisory is the subset of the OSV schema we read.
type Advisory struct {
	ID       string     `json:"id"`
	Aliases  []string   `json:"aliases"`
	Summary  string     `json:"summary"`
	Severity []Severity `json:"severity"`
	Affected []Affected `json:"affected"`
}

type Severity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

type Affected struct {
	Package          Package        `json:"package"`
	Ranges           []Range        `json:"ranges"`
	DatabaseSpecific map[string]any `json:"database_specific"`
}

type Package struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
}

type Range struct {
	Type   string  `json:"type"`
	Events []Event `json:"events"`
}

type Event struct {
	Introduced   string `json:"introduced"`
	Fixed        string `json:"fixed"`
	LastAffected string `json:"last_affected"`
}

// DB is an in-memory OSV index keyed by ecosystem and package name.
type DB struct {
	byKey map[string][]Advisory
}

func key(ecosystem, name string) string { return ecosystem + "/" + name }

// Load reads every .json OSV advisory under dir into an index.
func Load(dir string) (*DB, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read osv database %s: %w", dir, err)
	}
	db := &DB{byKey: map[string][]Advisory{}}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		var adv Advisory
		if err := json.Unmarshal(data, &adv); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		for _, aff := range adv.Affected {
			k := key(aff.Package.Ecosystem, aff.Package.Name)
			db.byKey[k] = append(db.byKey[k], adv)
		}
	}
	return db, nil
}

// Match returns one finding per advisory that affects an installed package
// version. Deterministic: sorted by package name then advisory id.
func (db *DB) Match(inv *inventory.Inventory) []finding.Finding {
	var out []finding.Finding
	for _, p := range inv.Packages {
		seen := map[string]bool{}
		for _, adv := range db.byKey[key(inv.Ecosystem, p.Name)] {
			if seen[adv.ID] {
				continue
			}
			fixed, ok := matchAdvisory(adv, inv.Ecosystem, p.Name, p.Version)
			if !ok {
				continue
			}
			seen[adv.ID] = true
			out = append(out, finding.Finding{
				Purl:          p.Purl,
				Name:          p.Name,
				Version:       p.Version,
				Advisory:      adv.ID,
				Aliases:       adv.Aliases,
				Severity:      severity(adv),
				Summary:       adv.Summary,
				FixedVersions: fixed,
				Source:        "OSV",
				Evidence: []finding.Evidence{{
					Source: "OSV",
					Ref:    adv.ID,
					Detail: fmt.Sprintf("%s@%s is within an affected range", p.Name, p.Version),
				}},
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Advisory < out[j].Advisory
	})
	return out
}

// matchAdvisory reports whether version is affected by adv, and the fixed
// versions the advisory records for the matching package.
func matchAdvisory(adv Advisory, ecosystem, name, version string) ([]string, bool) {
	affected := false
	var fixed []string
	for _, aff := range adv.Affected {
		if aff.Package.Ecosystem != ecosystem || aff.Package.Name != name {
			continue
		}
		for _, r := range aff.Ranges {
			if inRange(version, r.Events) {
				affected = true
			}
			for _, e := range r.Events {
				if e.Fixed != "" {
					fixed = append(fixed, e.Fixed)
				}
			}
		}
	}
	if !affected {
		return nil, false
	}
	return dedupeSorted(fixed), true
}

// inRange evaluates OSV SEMVER range events, which are ordered ascending: an
// "introduced" opens the affected window, a "fixed" or "last_affected" closes it.
func inRange(version string, events []Event) bool {
	affected := false
	for _, e := range events {
		switch {
		case e.Introduced != "" && semver.Compare(version, e.Introduced) >= 0:
			affected = true
		case e.Fixed != "" && semver.Compare(version, e.Fixed) >= 0:
			affected = false
		case e.LastAffected != "" && semver.Compare(version, e.LastAffected) > 0:
			affected = false
		}
	}
	return affected
}

func severity(adv Advisory) string {
	for _, aff := range adv.Affected {
		if s, ok := aff.DatabaseSpecific["severity"].(string); ok && s != "" {
			return strings.ToUpper(s)
		}
	}
	if len(adv.Severity) > 0 {
		return adv.Severity[0].Score
	}
	return "UNKNOWN"
}

func dedupeSorted(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return semver.Compare(out[i], out[j]) < 0 })
	return out
}
