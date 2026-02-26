// Package gomod detects Go multi-module workspaces.
package gomod

import (
	"os"
	"path/filepath"
	"strings"
)

// Modules returns module directory paths from go.work.
// Returns nil if no go.work exists or the workspace has fewer than 2 modules.
// The root module is represented as ".".
func Modules() ([]string, error) {
	data, err := os.ReadFile("go.work")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	dirs := parseUseDirectives(string(data))

	// Verify each directory has a go.mod.
	var modules []string
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			modules = append(modules, dir)
		}
	}

	if len(modules) < 2 {
		return nil, nil
	}
	return modules, nil
}

// parseUseDirectives extracts directory paths from go.work use directives.
func parseUseDirectives(content string) []string {
	var dirs []string
	inBlock := false

	for _, line := range strings.Split(content, "\n") {
		line = stripComment(strings.TrimSpace(line))
		if line == "" {
			continue
		}

		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			dirs = append(dirs, filepath.Clean(line))
			continue
		}

		if rest, ok := strings.CutPrefix(line, "use "); ok {
			rest = strings.TrimSpace(rest)
			if rest == "(" {
				inBlock = true
				continue
			}
			dirs = append(dirs, filepath.Clean(rest))
		}
	}

	return dirs
}

func stripComment(s string) string {
	if i := strings.Index(s, "//"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
