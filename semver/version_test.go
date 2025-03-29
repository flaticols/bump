package semver

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input       string
		wantErr     bool
		expectedErr error
		expected    Version
	}{
		{"1.2.3", false, nil, Version{1, 2, 3, nil, nil}},
		{"1.2.3-beta", false, nil, Version{1, 2, 3, []string{"beta"}, nil}},
		{"1.2.3+build", false, nil, Version{1, 2, 3, nil, []string{"build"}}},
		{"1.2.3-beta+build", false, nil, Version{1, 2, 3, []string{"beta"}, []string{"build"}}},
		{"1.2.3-beta.1+build.123", false, nil, Version{1, 2, 3, []string{"beta", "1"}, []string{"build", "123"}}},
		{"78", true, ErrMalformedCore, Version{}},
		{"1.2", true, ErrMalformedCore, Version{}},
		{"1.2.3.4", true, ErrMalformedCore, Version{}},
		{"1.2.3-", true, ErrEmptyIdentifier, Version{}},
		{"1.2.3+", true, ErrEmptyIdentifier, Version{}},
		{"01.2.3", true, ErrLeadingZeros, Version{}},
		{"1.02.3", true, ErrLeadingZeros, Version{}},
		{".02.", true, ErrEmptyVersionComponent, Version{}},
		{"1.2.03", true, ErrLeadingZeros, Version{}},
		{"1.2.3-01", true, ErrLeadingZeroesIdentifier, Version{}},
		{"1.2.3-beta!1", true, ErrInvalidIdentifierChars, Version{}},
		{"", true, ErrEmptyVersion, Version{}},
		{"v1.2.3", true, ErrNonDigitComponent, Version{}}, // Regular Parse should reject 'v' prefix
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.True(t, errors.Is(err, tt.expectedErr))
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.Major, got.Major, "Major version mismatch")
			require.Equal(t, tt.expected.Minor, got.Minor, "Minor version mismatch")
			require.Equal(t, tt.expected.Patch, got.Patch, "Patch version mismatch")
			require.True(t, slices.Equal(got.Prerelease, tt.expected.Prerelease),
				"Prerelease mismatch, got %v, want %v", got.Prerelease, tt.expected.Prerelease)
			require.True(t, slices.Equal(got.Metadata, tt.expected.Metadata),
				"Metadata mismatch, got %v, want %v", got.Metadata, tt.expected.Metadata)
		})
	}
}

func TestParseWithVPrefix(t *testing.T) {
	tests := []struct {
		input       string
		wantErr     bool
		expectedErr error
		expected    Version
	}{
		{"v1.2.3", false, nil, Version{1, 2, 3, nil, nil}},
		{"v1.2.3-beta", false, nil, Version{1, 2, 3, []string{"beta"}, nil}},
		{"v1.2.3+build", false, nil, Version{1, 2, 3, nil, []string{"build"}}},
		{"v1.2.3-beta+build", false, nil, Version{1, 2, 3, []string{"beta"}, []string{"build"}}},
		{"v1.2.3-beta.1+build.123", false, nil, Version{1, 2, 3, []string{"beta", "1"}, []string{"build", "123"}}},
		{"1.2.3", false, nil, Version{1, 2, 3, nil, nil}}, // Should also work without 'v'
		{"v78", true, ErrMalformedCore, Version{}},
		{"v1.2", true, ErrMalformedCore, Version{}},
		{"v1.2.3.4", true, ErrMalformedCore, Version{}},
		{"v1.2.3-", true, ErrEmptyIdentifier, Version{}},
		{"v1.2.3+", true, ErrEmptyIdentifier, Version{}},
		{"v01.2.3", true, ErrLeadingZeros, Version{}},
		{"v1.02.3", true, ErrLeadingZeros, Version{}},
		{"v.02.", true, ErrEmptyVersionComponent, Version{}},
		{"v1.2.03", true, ErrLeadingZeros, Version{}},
		{"v1.2.3-01", true, ErrLeadingZeroesIdentifier, Version{}},
		{"v1.2.3-beta!1", true, ErrInvalidIdentifierChars, Version{}},
		{"v", true, ErrMalformedCore, Version{}},
		{"", true, ErrEmptyVersion, Version{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseWithVPrefix(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.True(t, errors.Is(err, tt.expectedErr))
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected.Major, got.Major, "Major version mismatch")
			require.Equal(t, tt.expected.Minor, got.Minor, "Minor version mismatch")
			require.Equal(t, tt.expected.Patch, got.Patch, "Patch version mismatch")
			require.True(t, slices.Equal(got.Prerelease, tt.expected.Prerelease),
				"Prerelease mismatch, got %v, want %v", got.Prerelease, tt.expected.Prerelease)
			require.True(t, slices.Equal(got.Metadata, tt.expected.Metadata),
				"Metadata mismatch, got %v, want %v", got.Metadata, tt.expected.Metadata)
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta", "1.0.0-alpha.1", 1},
		{"1.0.0+build", "1.0.0", 0},
		{"1.0.0", "1.0.0+build", 0},
		{"1.0.0+build.1", "1.0.0+build.2", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			v1, err := Parse(tt.v1)
			require.NoError(t, err, "Failed to parse v1 %s", tt.v1)

			v2, err := Parse(tt.v2)
			require.NoError(t, err, "Failed to parse v2 %s", tt.v2)

			got := Compare(v1, v2)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestCompareWithVPrefix(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v2.0.0", -1},
		{"v2.0.0", "v1.0.0", 1},
		{"v1.0.0-alpha", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0-alpha", 1},
		{"v1.0.0+build", "v1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},             // Mixed with and without v prefix
		{"1.0.0", "v1.0.0", 0},             // Mixed with and without v prefix
		{"v1.2.3", "1.2.3", 0},             // Mixed with and without v prefix
		{"v1.0.0-alpha", "1.0.0-alpha", 0}, // Mixed with and without v prefix
	}

	for _, tt := range tests {
		t.Run(tt.v1+" vs "+tt.v2, func(t *testing.T) {
			v1, err := ParseWithVPrefix(tt.v1)
			require.NoError(t, err, "Failed to parse v1 %s", tt.v1)

			v2, err := ParseWithVPrefix(tt.v2)
			require.NoError(t, err, "Failed to parse v2 %s", tt.v2)

			got := Compare(v1, v2)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestConstraint(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		expected   bool
	}{
		{"= 1.0.0", "1.0.0", true},
		{"= 1.0.0", "1.0.1", false},
		{"!= 1.0.0", "1.0.0", false},
		{"!= 1.0.0", "1.0.1", true},
		{"> 1.0.0", "1.0.1", true},
		{"> 1.0.0", "1.0.0", false},
		{"> 1.0.0", "0.9.9", false},
		{"< 1.0.0", "0.9.9", true},
		{"< 1.0.0", "1.0.0", false},
		{"< 1.0.0", "1.0.1", false},
		{">= 1.0.0", "1.0.0", true},
		{">= 1.0.0", "1.0.1", true},
		{">= 1.0.0", "0.9.9", false},
		{"<= 1.0.0", "0.9.9", true},
		{"<= 1.0.0", "1.0.0", true},
		{"<= 1.0.0", "1.0.1", false},
		{"~1.0.0", "1.0.1", true},
		{"~1.0.0", "1.1.0", false},
		{"~1.0", "1.0.0", true},
		{"~1.0", "1.1.0", true},
		{"~1.0", "2.0.0", false},
		{"^1.0.0", "1.0.1", true},
		{"^1.0.0", "1.1.0", true},
		{"^1.0.0", "2.0.0", false},
		{"^0.1.0", "0.1.1", true},
		{"^0.1.0", "0.2.0", false},
		{"^0.0.1", "0.0.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+" vs "+tt.version, func(t *testing.T) {
			c, err := ParseConstraint(tt.constraint)
			require.NoError(t, err, "ParseConstraint() failed")

			v, err := Parse(tt.version)
			require.NoError(t, err, "Failed to parse version %s", tt.version)

			got := c.Check(v)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestIncrementVersions(t *testing.T) {
	v, err := Parse("1.2.3-beta.1+build.123")
	require.NoError(t, err, "Failed to parse version")

	major := v.IncrementMajor()
	require.Equal(t, "2.0.0", major.String())

	minor := v.IncrementMinor()
	require.Equal(t, "1.3.0", minor.String())

	patch := v.IncrementPatch()
	require.Equal(t, "1.2.4", patch.String())
}

func TestSetFunctions(t *testing.T) {
	v, err := Parse("1.2.3+build.123")
	require.NoError(t, err, "Failed to parse version")

	// Test SetPrerelease
	v2 := SetPrerelease(v, []string{"alpha", "1"})
	require.Equal(t, []string{"alpha", "1"}, v2.Prerelease)

	// Test SetMetadata
	v3 := SetMetadata(v, []string{"commit", "abc123"})
	require.Equal(t, []string{"commit", "abc123"}, v3.Metadata)

	// Test SetPrereleaseMap
	prereleaseMap := map[string]string{
		"feature": "x",
		"build":   "123",
	}
	v4 := SetPrereleaseMap(v, prereleaseMap)
	// Since map iteration order is non-deterministic, check length and contents differently
	require.Equal(t, 4, len(v4.Prerelease), "SetPrereleaseMap() resulted in prerelease of wrong length")
	gotMap := v4.GetPrereleaseMap()
	for k, val := range prereleaseMap {
		require.Equal(t, val, gotMap[k], "SetPrereleaseMap() map doesn't contain %s: %s", k, val)
	}

	// Test SetMetadataMap
	metadataMap := map[string]string{
		"commit":    "abc123",
		"timestamp": "1617293965",
	}
	v5 := SetMetadataMap(v, metadataMap)
	require.Equal(t, 4, len(v5.Metadata), "SetMetadataMap() resulted in metadata of wrong length")
	gotMetaMap := v5.GetMetadataMap()
	for k, val := range metadataMap {
		require.Equal(t, val, gotMetaMap[k], "SetMetadataMap() map doesn't contain %s: %s", k, val)
	}
}

func TestVersionStringv(t *testing.T) {
	v, err := Parse("1.2.3-beta.1+build.123")
	require.NoError(t, err)
	require.Equal(t, "v1.2.3-beta.1+build.123", v.Stringv())
	require.Equal(t, "1.2.3-beta.1+build.123", v.String())
}

func TestErrorsPropagation(t *testing.T) {
	// Test that invalid versions in constraints propagate errors correctly
	_, err := ParseConstraint(">= 1.x.0")
	require.Error(t, err, "ParseConstraint() with invalid version should return error")

	_, err = ParseConstraintSet(">= 1.0.0, < x.0.0")
	require.Error(t, err, "ParseConstraintSet() with invalid version should return error")
}

func TestParsePerformance(t *testing.T) {
	// This isn't a true benchmark, but it's a simple way to ensure we're not regressing
	// on performance in regular tests

	validVersions := []string{
		"0.0.0",
		"1.2.3",
		"10.20.30",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
		"1.0.0",
		"1.0.0+build.1",
		"1.0.0+build.1.2.3",
		"1.0.0+build.1.2.3.4.5.6.7.8.9",
		"1.0.0-alpha+beta",
		"1.0.0-alpha.1+beta.2",
		"1.0.0-alpha.1+beta.2.3.4.5.6.7.8.9",
	}

	for i := 0; i < 100; i++ {
		for _, v := range validVersions {
			_, err := Parse(v)
			require.NoError(t, err, "Parse(%q) returned error: %v", v, err)
		}
	}
}

// Test the IsValid and IsValidWithVPrefix functions
func TestIsValid(t *testing.T) {
	validVersions := []string{
		"0.0.0",
		"1.2.3",
		"10.20.30",
		"1.0.0-alpha",
		"1.0.0+build",
	}

	invalidVersions := []string{
		"",
		"1",
		"1.2",
		"v1.2.3",
		"1.2.3.4",
		"01.2.3",
		"1.02.3",
		"1.2.03",
	}

	for _, v := range validVersions {
		t.Run(v, func(t *testing.T) {
			require.True(t, IsValid(v), "IsValid(%q) should be true", v)
		})
	}

	for _, v := range invalidVersions {
		t.Run(v, func(t *testing.T) {
			require.False(t, IsValid(v), "IsValid(%q) should be false", v)
		})
	}
}

func TestIsValidWithVPrefix(t *testing.T) {
	validVersions := []string{
		"0.0.0",
		"1.2.3",
		"v1.2.3",
		"v10.20.30",
		"v1.0.0-alpha",
		"v1.0.0+build",
	}

	invalidVersions := []string{
		"",
		"1",
		"v1",
		"1.2",
		"v1.2",
		"1.2.3.4",
		"v1.2.3.4",
		"01.2.3",
		"v01.2.3",
		"1.02.3",
		"v1.02.3",
	}

	for _, v := range validVersions {
		t.Run(v, func(t *testing.T) {
			require.True(t, IsValidWithVPrefix(v), "IsValidWithVPrefix(%q) should be true", v)
		})
	}

	for _, v := range invalidVersions {
		t.Run(v, func(t *testing.T) {
			require.False(t, IsValidWithVPrefix(v), "IsValidWithVPrefix(%q) should be false", v)
		})
	}
}
