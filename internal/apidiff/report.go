package apidiff

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// WriteText writes a human-readable report of API changes to w.
func (r *Report) WriteText(w io.Writer, printCompatible bool) {
	if len(r.Changes) == 0 {
		fmt.Fprintln(w, "No API changes detected.")
		return
	}

	// Group changes by package
	byPkg := make(map[string][]Change)
	for _, c := range r.Changes {
		byPkg[c.Package] = append(byPkg[c.Package], c)
	}

	// Sort package names for deterministic output
	pkgs := make([]string, 0, len(byPkg))
	for pkg := range byPkg {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	for _, pkg := range pkgs {
		changes := byPkg[pkg]
		var incompatible, compatible []string
		for _, c := range changes {
			if c.Kind == Incompatible {
				incompatible = append(incompatible, c.Message)
			} else {
				compatible = append(compatible, c.Message)
			}
		}

		if len(incompatible) == 0 && (!printCompatible || len(compatible) == 0) {
			continue
		}

		fmt.Fprintln(w, pkg)
		if len(incompatible) > 0 {
			fmt.Fprintln(w, "  Incompatible changes:")
			for _, msg := range incompatible {
				fmt.Fprintf(w, "    - %s\n", msg)
			}
		}
		if printCompatible && len(compatible) > 0 {
			fmt.Fprintln(w, "  Compatible changes:")
			for _, msg := range compatible {
				fmt.Fprintf(w, "    + %s\n", msg)
			}
		}
		fmt.Fprintln(w)
	}
}

// WriteJSON writes the report as a JSON object to w.
func (r *Report) WriteJSON(w io.Writer) error {
	type jsonChange struct {
		Package string `json:"package"`
		Message string `json:"message"`
		Kind    string `json:"kind"`
	}

	type jsonReport struct {
		Changes       []jsonChange `json:"changes"`
		Summary       string       `json:"summary"`
		SuggestedBump string       `json:"suggested_bump"`
		HasBreaking   bool         `json:"has_breaking"`
	}

	changes := make([]jsonChange, 0, len(r.Changes))
	for _, c := range r.Changes {
		kind := "compatible"
		if c.Kind == Incompatible {
			kind = "incompatible"
		}
		changes = append(changes, jsonChange{
			Package: c.Package,
			Message: c.Message,
			Kind:    kind,
		})
	}

	return json.NewEncoder(w).Encode(jsonReport{
		Changes:       changes,
		Summary:       r.Summary(),
		SuggestedBump: r.SuggestedBump(),
		HasBreaking:   r.HasBreaking(),
	})
}

// Summary returns a one-line summary string.
func (r *Report) Summary() string {
	var breaking, additions int
	for _, c := range r.Changes {
		if c.Kind == Incompatible {
			breaking++
		} else {
			additions++
		}
	}

	var parts []string
	if breaking > 0 {
		parts = append(parts, fmt.Sprintf("%d breaking", breaking))
	}
	if additions > 0 {
		parts = append(parts, fmt.Sprintf("%d compatible", additions))
	}
	if len(parts) == 0 {
		return "no API changes"
	}
	return strings.Join(parts, ", ")
}
