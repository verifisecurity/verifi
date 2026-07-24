// Package usage detects which packages a project's own source actually imports.
// It is the cheap first half of reachability: a vulnerable direct dependency
// that your code never imports is a strong remove candidate, and one it does
// import is where attention belongs. npm for now (JavaScript and TypeScript).
package usage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// specifier matches the string in require('x'), import ... from 'x', import 'x',
// import('x'), and export ... from 'x'.
var specifier = regexp.MustCompile(`(?:require|from|import)\s*\(?\s*['"]([^'"]+)['"]`)

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true, "coverage": true, ".next": true,
}

var sourceExt = map[string]bool{
	".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
}

// Scan walks a project's own source and returns the set of npm package names it
// imports. Relative imports and Node builtins are ignored; subpaths map to the
// package (lodash/fp becomes lodash, @scope/pkg/x becomes @scope/pkg).
func Scan(dir string) (map[string]bool, error) {
	imports := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries rather than abort the scan
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !sourceExt[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, m := range specifier.FindAllStringSubmatch(string(data), -1) {
			if name := packageName(m[1]); name != "" {
				imports[name] = true
			}
		}
		return nil
	})
	return imports, err
}

// packageName reduces an import specifier to a package name, or "" for a
// relative or absolute path or a Node builtin.
func packageName(spec string) string {
	if spec == "" || strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") || strings.HasPrefix(spec, "node:") {
		return ""
	}
	parts := strings.Split(spec, "/")
	if strings.HasPrefix(spec, "@") {
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return spec
	}
	return parts[0]
}
