// Package apidiff compares Go API surfaces between two git refs using temporary
// worktrees, ensuring the current working tree is never modified.
// It loads type information via golang.org/x/tools/go/packages and delegates
// the actual API comparison to golang.org/x/tools/go/analysis (go/types).
package apidiff

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/flaticols/bump/internal/git"
	"golang.org/x/tools/go/packages"
)

// ChangeKind classifies an API change.
type ChangeKind int

const (
	Compatible   ChangeKind = iota // Backwards-compatible addition
	Incompatible                   // Breaking change
)

// Change represents a single API change in a package.
type Change struct {
	Package    string
	Message    string
	Kind       ChangeKind
}

// Report holds the result of an API comparison.
type Report struct {
	Changes []Change
}

// HasBreaking returns true if any incompatible change was found.
func (r *Report) HasBreaking() bool {
	for _, c := range r.Changes {
		if c.Kind == Incompatible {
			return true
		}
	}
	return false
}

// HasAdditions returns true if any compatible addition was found.
func (r *Report) HasAdditions() bool {
	for _, c := range r.Changes {
		if c.Kind == Compatible {
			return true
		}
	}
	return false
}

// SuggestedBump returns the recommended version bump level based on changes.
func (r *Report) SuggestedBump() string {
	if r.HasBreaking() {
		return "major"
	}
	if r.HasAdditions() {
		return "minor"
	}
	return "patch"
}

// Compare compares the public Go API between oldRef and newRef.
// Both refs are checked out in temporary worktrees; the current working tree is untouched.
// If newRef is empty, HEAD (current state) is used via a worktree of HEAD.
func Compare(oldRef, newRef string) (*Report, error) {
	oldWT, err := git.CreateTempWorktree(oldRef)
	if err != nil {
		return nil, fmt.Errorf("worktree for %s: %w", oldRef, err)
	}
	defer oldWT.Remove()

	var newPath string
	if newRef == "" || newRef == "HEAD" {
		newWT, err := git.CreateTempWorktree("HEAD")
		if err != nil {
			return nil, fmt.Errorf("worktree for HEAD: %w", err)
		}
		defer newWT.Remove()
		newPath = newWT.Path
	} else {
		newWT, err := git.CreateTempWorktree(newRef)
		if err != nil {
			return nil, fmt.Errorf("worktree for %s: %w", newRef, err)
		}
		defer newWT.Remove()
		newPath = newWT.Path
	}

	oldPkgs, err := loadTypedPackages(oldWT.Path)
	if err != nil {
		return nil, fmt.Errorf("load packages from %s: %w", oldRef, err)
	}

	newPkgs, err := loadTypedPackages(newPath)
	if err != nil {
		return nil, fmt.Errorf("load packages from %s: %w", newRef, err)
	}

	return comparePackages(oldPkgs, newPkgs), nil
}

func loadTypedPackages(dir string) (map[string]*types.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedName | packages.NeedImports,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	result := make(map[string]*types.Package)
	for _, pkg := range pkgs {
		if pkg.Types == nil {
			continue
		}
		if isInternal(pkg.PkgPath) {
			continue
		}
		result[pkg.PkgPath] = pkg.Types
	}
	return result, nil
}

func isInternal(path string) bool {
	return strings.Contains(path, "/internal") || strings.Contains(path, "/internal/")
}

func comparePackages(oldPkgs, newPkgs map[string]*types.Package) *Report {
	report := &Report{}

	// Removed packages (breaking)
	for path := range oldPkgs {
		if _, ok := newPkgs[path]; !ok {
			report.Changes = append(report.Changes, Change{
				Package: path,
				Message: "package removed",
				Kind:    Incompatible,
			})
		}
	}

	// Added packages (compatible)
	for path := range newPkgs {
		if _, ok := oldPkgs[path]; !ok {
			report.Changes = append(report.Changes, Change{
				Package: path,
				Message: "package added",
				Kind:    Compatible,
			})
		}
	}

	// Changed packages — compare exported names
	for path, oldPkg := range oldPkgs {
		newPkg, ok := newPkgs[path]
		if !ok {
			continue
		}
		compareScope(report, path, oldPkg.Scope(), newPkg.Scope())
	}

	return report
}

// compareScope compares exported names between two package scopes.
func compareScope(report *Report, pkgPath string, oldScope, newScope *types.Scope) {
	// Check for removed or changed exports
	for _, name := range oldScope.Names() {
		oldObj := oldScope.Lookup(name)
		if !oldObj.Exported() {
			continue
		}

		newObj := newScope.Lookup(name)
		if newObj == nil {
			report.Changes = append(report.Changes, Change{
				Package: pkgPath,
				Message: fmt.Sprintf("removed: %s", name),
				Kind:    Incompatible,
			})
			continue
		}

		if !typesCompatible(oldObj, newObj) {
			report.Changes = append(report.Changes, Change{
				Package: pkgPath,
				Message: fmt.Sprintf("changed: %s", name),
				Kind:    Incompatible,
			})
		}
	}

	// Check for new exports
	for _, name := range newScope.Names() {
		newObj := newScope.Lookup(name)
		if !newObj.Exported() {
			continue
		}
		if oldScope.Lookup(name) == nil {
			report.Changes = append(report.Changes, Change{
				Package: pkgPath,
				Message: fmt.Sprintf("added: %s", name),
				Kind:    Compatible,
			})
		}
	}
}

// typesCompatible checks if two objects are type-compatible.
func typesCompatible(old, new types.Object) bool {
	// Different object kinds are incompatible
	if fmt.Sprintf("%T", old) != fmt.Sprintf("%T", new) {
		return false
	}
	return types.Identical(old.Type(), new.Type())
}
