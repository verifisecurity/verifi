// Package inventory resolves a project's dependency tree into a structured,
// ecosystem-agnostic inventory and emits it as a CycloneDX SBOM. npm is the
// first supported ecosystem. See docs/pipeline.md.
package inventory

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Inventory is the resolved dependency set of a project.
type Inventory struct {
	Ecosystem string    `json:"ecosystem"`
	Root      Root      `json:"root"`
	Packages  []Package `json:"packages"`
}

// Root identifies the project itself.
type Root struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Package is one resolved dependency. Identity is a purl (Package URL).
type Package struct {
	Purl    string `json:"purl"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Direct  bool   `json:"direct"`
	Scope   string `json:"scope"` // "prod" or "dev"
}

// npm package-lock.json (lockfileVersion 2 and 3) shapes we read.
type npmLock struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	LockfileVersion int               `json:"lockfileVersion"`
	Packages        map[string]npmPkg `json:"packages"`
}

type npmPkg struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dev             bool              `json:"dev"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// ParseNpmLock reads a package-lock.json (lockfileVersion 2 or 3) into an Inventory.
func ParseNpmLock(data []byte) (*Inventory, error) {
	var lock npmLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("parse package-lock.json: %w", err)
	}
	if lock.Packages == nil {
		return nil, fmt.Errorf(`unsupported package-lock.json: no "packages" map (need lockfileVersion 2 or 3)`)
	}

	root := lock.Packages[""]
	rootName := firstNonEmpty(root.Name, lock.Name)
	rootVersion := firstNonEmpty(root.Version, lock.Version)

	direct := map[string]bool{}
	for name := range root.Dependencies {
		direct[name] = true
	}
	for name := range root.DevDependencies {
		direct[name] = true
	}

	var pkgs []Package
	for key, p := range lock.Packages {
		if key == "" {
			continue // the root project, not a dependency
		}
		name := npmName(key)
		if name == "" || p.Version == "" {
			continue
		}
		scope := "prod"
		if p.Dev {
			scope = "dev"
		}
		pkgs = append(pkgs, Package{
			Purl:    "pkg:npm/" + name + "@" + p.Version,
			Name:    name,
			Version: p.Version,
			Direct:  direct[name],
			Scope:   scope,
		})
	}

	sort.Slice(pkgs, func(i, j int) bool {
		if pkgs[i].Name != pkgs[j].Name {
			return pkgs[i].Name < pkgs[j].Name
		}
		return pkgs[i].Version < pkgs[j].Version
	})

	return &Inventory{
		Ecosystem: "npm",
		Root:      Root{Name: rootName, Version: rootVersion},
		Packages:  pkgs,
	}, nil
}

// npmName extracts the package name from a package-lock "packages" key such as
// "node_modules/left-pad" or "node_modules/a/node_modules/@scope/b" (nested).
func npmName(key string) string {
	const marker = "node_modules/"
	i := strings.LastIndex(key, marker)
	if i < 0 {
		return ""
	}
	return key[i+len(marker):]
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
