// Package server provides a full SemVer 2.0.0 implementation for parsing,
// comparing, and manipulating semantic versions.
package server

import (
	"fmt"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Prerelease []string
	Metadata   []string
	Major      int
	Minor      int
	Patch      int
}

// String returns the string representation without v prefix.
func (ver Version) String() string {
	return ver.format(false)
}

// Stringv returns the string representation with v prefix.
func (ver Version) Stringv() string {
	return ver.format(true)
}

func (ver Version) format(withV bool) string {
	var b strings.Builder
	if withV {
		b.WriteByte('v')
	}
	fmt.Fprintf(&b, "%d.%d.%d", ver.Major, ver.Minor, ver.Patch)
	if len(ver.Prerelease) > 0 {
		b.WriteByte('-')
		b.WriteString(strings.Join(ver.Prerelease, "."))
	}
	if len(ver.Metadata) > 0 {
		b.WriteByte('+')
		b.WriteString(strings.Join(ver.Metadata, "."))
	}
	return b.String()
}

// IncrementMajor returns a new Version with major incremented, minor and patch reset.
func (ver Version) IncrementMajor() Version {
	return Version{Major: ver.Major + 1}
}

// IncrementMinor returns a new Version with minor incremented, patch reset.
func (ver Version) IncrementMinor() Version {
	return Version{Major: ver.Major, Minor: ver.Minor + 1}
}

// IncrementPatch returns a new Version with patch incremented.
func (ver Version) IncrementPatch() Version {
	return Version{Major: ver.Major, Minor: ver.Minor, Patch: ver.Patch + 1}
}

// GetPrerelease returns prerelease identifiers.
func (ver Version) GetPrerelease() []string { return ver.Prerelease }

// GetMetadata returns metadata identifiers.
func (ver Version) GetMetadata() []string { return ver.Metadata }

// SetPrerelease returns a copy with the given prerelease.
func SetPrerelease(v Version, pre []string) Version {
	v.Prerelease = pre
	return v
}

// SetMetadata returns a copy with the given metadata.
func SetMetadata(v Version, meta []string) Version {
	v.Metadata = meta
	return v
}
