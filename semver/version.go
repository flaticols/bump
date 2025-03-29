package semver

import (
	"fmt"
	"strings"
)

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease []string
	Metadata   []string
}

// String returns the string representation of the Version.
func (ver Version) String() string {
	return ver.string(false)
}

func (ver Version) Stringv() string {
	return ver.string(true)
}

func (ver Version) string(withPrefix bool) string {
	var buf strings.Builder

	if withPrefix {
		buf.WriteString("v")
	}

	// Write core version
	fmt.Fprintf(&buf, "%d.%d.%d", ver.Major, ver.Minor, ver.Patch)

	// Write prerelease if exists
	if len(ver.Prerelease) > 0 {
		buf.WriteByte('-')
		buf.WriteString(strings.Join(ver.Prerelease, "."))
	}

	// Write metadata if exists
	if len(ver.Metadata) > 0 {
		buf.WriteByte('+')
		buf.WriteString(strings.Join(ver.Metadata, "."))
	}

	return buf.String()
}

// IncrementMajor returns a new Version with incremented major version
// and resets minor and patch to 0.
func (ver Version) IncrementMajor() Version {
	return Version{
		Major:      ver.Major + 1,
		Minor:      0,
		Patch:      0,
		Prerelease: []string{},
		Metadata:   []string{},
	}
}

// IncrementMinor returns a new Version with incremented minor version
// and resets patch to 0.
func (ver Version) IncrementMinor() Version {
	return Version{
		Major:      ver.Major,
		Minor:      ver.Minor + 1,
		Patch:      0,
		Prerelease: []string{},
		Metadata:   []string{},
	}
}

// IncrementPatch returns a new Version with incremented patch version.
func (ver Version) IncrementPatch() Version {
	return Version{
		Major:      ver.Major,
		Minor:      ver.Minor,
		Patch:      ver.Patch + 1,
		Prerelease: []string{},
		Metadata:   []string{},
	}
}

// GetPrerelease returns the prerelease identifiers.
func (ver Version) GetPrerelease() []string {
	return ver.Prerelease
}

// GetMetadata returns the build metadata identifiers.
func (ver Version) GetMetadata() []string {
	return ver.Metadata
}

// GetMetadataMap returns the build metadata identifiers as a map.
// Assumes identifiers are alternating key-value pairs.
// If there's an odd number of identifiers, the last one gets an empty string value.
// This is a convention and not part of the SemVer spec.
func (ver Version) GetMetadataMap() map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(ver.Metadata); i += 2 {
		if i+1 < len(ver.Metadata) {
			result[ver.Metadata[i]] = ver.Metadata[i+1]
		} else {
			result[ver.Metadata[i]] = ""
		}
	}
	return result
}

// GetPrereleaseMap returns the prerelease identifiers as a map.
// Assumes identifiers are alternating key-value pairs.
// If there's an odd number of identifiers, the last one gets an empty string value.
// This is a convention and not part of the SemVer spec.
func (ver Version) GetPrereleaseMap() map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(ver.Prerelease); i += 2 {
		if i+1 < len(ver.Prerelease) {
			result[ver.Prerelease[i]] = ver.Prerelease[i+1]
		} else {
			result[ver.Prerelease[i]] = ""
		}
	}
	return result
}

// Package-level functions to replace pointer receivers

// SetPrerelease creates a new Version with the given prerelease identifiers.
func SetPrerelease(v Version, prerelease []string) Version {
	v.Prerelease = prerelease
	return v
}

// SetMetadata creates a new Version with the given metadata identifiers.
func SetMetadata(v Version, metadata []string) Version {
	v.Metadata = metadata
	return v
}

// SetPrereleaseMap creates a new Version with prerelease identifiers from a map.
// Each key-value pair is added as two separate identifiers.
// For example, {"feature": "x", "build": "123"} becomes ["feature", "x", "build", "123"].
// This is a convention and not part of the SemVer spec.
func SetPrereleaseMap(v Version, prerelease map[string]string) Version {
	v.Prerelease = make([]string, 0, len(prerelease)*2)
	for k, val := range prerelease {
		v.Prerelease = append(v.Prerelease, k, val)
	}
	return v
}

// SetMetadataMap creates a new Version with metadata identifiers from a map.
// Each key-value pair is added as two separate identifiers.
// For example, {"commit": "abc123", "timestamp": "1234567890"} becomes
// ["commit", "abc123", "timestamp", "1234567890"].
// This is a convention and not part of the SemVer spec.
func SetMetadataMap(v Version, metadata map[string]string) Version {
	v.Metadata = make([]string, 0, len(metadata)*2)
	for k, val := range metadata {
		v.Metadata = append(v.Metadata, k, val)
	}
	return v
}
