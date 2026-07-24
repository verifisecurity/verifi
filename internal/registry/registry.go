// Package registry looks up which versions of a package actually exist on its
// ecosystem's registry. OSV records the version an advisory is "fixed" in, but
// that can be a version not yet published, so a fix candidate must be checked
// against the registry before we recommend it. npm for now.
package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// npmRegistry is the public npm registry. Its per-package document lists every
// published version.
var npmRegistry = "https://registry.npmjs.org"

// NpmVersions returns the set of published versions for an npm package.
func NpmVersions(name string) (map[string]bool, error) {
	// Scoped names (@scope/pkg) need the slash percent-encoded in the path.
	url := npmRegistry + "/" + strings.Replace(name, "/", "%2F", 1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Abbreviated metadata is a much smaller payload than the full document.
	req.Header.Set("Accept", "application/vnd.npm.install-v1+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry lookup %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry lookup %s: HTTP %d", name, resp.StatusCode)
	}

	var doc struct {
		Versions map[string]json.RawMessage `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("registry lookup %s: %w", name, err)
	}
	set := make(map[string]bool, len(doc.Versions))
	for v := range doc.Versions {
		set[v] = true
	}
	return set, nil
}
