package semver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPrereleaseComparisons specifically tests the pre-release precedence
// rules as defined in the SemVer spec
func TestPrereleaseComparisons(t *testing.T) {
	// From the spec:
	// "1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0"
	versions := []string{
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11",
		"1.0.0-rc.1",
		"1.0.0",
	}

	// Test each pair to ensure proper ordering
	for i := 0; i < len(versions)-1; i++ {
		v1, err := Parse(versions[i])
		require.NoError(t, err, "Failed to parse %s", versions[i])

		v2, err := Parse(versions[i+1])
		require.NoError(t, err, "Failed to parse %s", versions[i+1])

		require.True(t, v1.LessThan(v2), "Expected %s < %s, but got false", versions[i], versions[i+1])
		require.False(t, v2.LessThan(v1), "Expected %s > %s, but got false", versions[i+1], versions[i])
	}

	// Test numeric identifiers having lower precedence than non-numeric identifiers
	v1, _ := Parse("1.0.0-1")
	v2, _ := Parse("1.0.0-alpha")
	require.True(t, v1.LessThan(v2), "Numeric identifier should have lower precedence than non-numeric")

	// Test a larger set of pre-release identifiers having higher precedence when
	// all preceding identifiers are equal
	v1, _ = Parse("1.0.0-alpha")
	v2, _ = Parse("1.0.0-alpha.1")
	require.True(t, v1.LessThan(v2), "Larger set of identifiers should have higher precedence when preceding identifiers are equal")

	// Test different types of pre-release identifiers
	testCases := []struct {
		v1       string
		v2       string
		expected int // -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
	}{
		// Numeric identifiers compared numerically
		{"1.0.0-1", "1.0.0-2", -1},
		{"1.0.0-2", "1.0.0-10", -1}, // 2 < 10 numerically

		// Identifiers with letters compared lexically
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},

		// Identifiers with hyphens
		{"1.0.0-alpha-1", "1.0.0-alpha-2", -1},
		{"1.0.0-alpha-beta", "1.0.0-alpha-gamma", -1},

		// Mixed case
		{"1.0.0-Alpha", "1.0.0-alpha", -1}, // Capital A is before lowercase a in ASCII

		// Length effects - longer ones don't automatically win
		{"1.0.0-alpha.beta.delta", "1.0.0-alpha.beta.gamma", -1}, // delta < gamma
		{"1.0.0-alpha.beta.delta", "1.0.0-alpha.gamma", -1},      // beta.delta < gamma
	}

	for _, tc := range testCases {
		t.Run(tc.v1+" vs "+tc.v2, func(t *testing.T) {
			v1, err := Parse(tc.v1)
			require.NoError(t, err, "Failed to parse %s", tc.v1)

			v2, err := Parse(tc.v2)
			require.NoError(t, err, "Failed to parse %s", tc.v2)

			got := Compare(v1, v2)
			require.Equal(t, tc.expected, got, "Compare(%s, %s) = %v, want %v", tc.v1, tc.v2, got, tc.expected)
		})
	}
}

// TestBuildMetadataHandling ensures build metadata is properly parsed
// and ignored in version precedence
func TestBuildMetadataHandling(t *testing.T) {
	// Build metadata should be ignored in precedence
	v1, _ := Parse("1.0.0+build.1")
	v2, _ := Parse("1.0.0+build.2")

	require.Equal(t, 0, Compare(v1, v2), "Build metadata should be ignored in precedence")

	// Build metadata with pre-release
	v1, _ = Parse("1.0.0-alpha+build.1")
	v2, _ = Parse("1.0.0-alpha+build.2")

	require.Equal(t, 0, Compare(v1, v2), "Build metadata should be ignored in precedence")

	// But pre-release vs non-pre-release still matters
	v1, _ = Parse("1.0.0-alpha+build.1")
	v2, _ = Parse("1.0.0+build.1")

	require.Equal(t, -1, Compare(v1, v2), "Pre-release version should have lower precedence despite build metadata")

	// Test GetMetadata returns the correct values
	v, _ := Parse("1.0.0+build.timestamp.12345")
	metadata := v.GetMetadata()

	require.Len(t, metadata, 3)
	require.Equal(t, "build", metadata[0])
	require.Equal(t, "timestamp", metadata[1])
	require.Equal(t, "12345", metadata[2])
}

// TestInitialDevelopmentVersions tests rules for 0.y.z versions
func TestInitialDevelopmentVersions(t *testing.T) {
	// For initial development versions (0.y.z), normal semver precedence rules apply
	testCases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"0.1.0", "0.2.0", -1},       // Minor version increment
		{"0.1.0", "0.1.1", -1},       // Patch version increment
		{"0.0.1", "0.0.2", -1},       // Patch version increment at lowest level
		{"0.0.0", "0.0.1", -1},       // Zero version vs patch increment
		{"0.1.0-alpha", "0.1.0", -1}, // Pre-release vs release
	}

	for _, tc := range testCases {
		t.Run(tc.v1+" vs "+tc.v2, func(t *testing.T) {
			v1, err := Parse(tc.v1)
			require.NoError(t, err, "Failed to parse %s", tc.v1)

			v2, err := Parse(tc.v2)
			require.NoError(t, err, "Failed to parse %s", tc.v2)

			got := Compare(v1, v2)
			require.Equal(t, tc.expected, got, "Compare(%s, %s) = %v, want %v", tc.v1, tc.v2, got, tc.expected)
		})
	}
}

// TestInvalidVersionCharacters tests the rejection of invalid characters
func TestInvalidVersionCharacters(t *testing.T) {
	invalidVersions := []string{
		// Invalid core
		"1.2.3a", // letter in core
		"1.a.3",  // letter in core
		"a.2.3",  // letter in core

		// Invalid pre-release identifiers
		"1.0.0-alpha@beta",  // @ not allowed
		"1.0.0-alpha.beta!", // ! not allowed
		"1.0.0-alpha beta",  // space not allowed
		"1.0.0-alpha_beta",  // _ not allowed

		// Invalid build metadata identifiers
		"1.0.0+build@123",  // @ not allowed
		"1.0.0+build.123!", // ! not allowed
		"1.0.0+build 123",  // space not allowed
		"1.0.0+build_123",  // _ not allowed
	}

	for _, v := range invalidVersions {
		t.Run(v, func(t *testing.T) {
			_, err := Parse(v)
			require.Error(t, err, "Parse(%q) should fail for invalid version", v)
		})
	}
}

// TestCornerCases tests various edge cases to ensure robustness
func TestCornerCases(t *testing.T) {
	// Large version numbers (within int limits)
	largeVersion := "2147483647.2147483647.2147483647"
	v, err := Parse(largeVersion)
	require.NoError(t, err, "Parse(%q) failed", largeVersion)
	require.Equal(t, 2147483647, v.Major)
	require.Equal(t, 2147483647, v.Minor)
	require.Equal(t, 2147483647, v.Patch)

	// Long identifiers
	longIdentifier := "1.0.0-abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	_, err = Parse(longIdentifier)
	require.NoError(t, err, "Parse(%q) failed", longIdentifier)

	// Multiple pre-release identifiers
	multiPre := "1.0.0-alpha.beta.gamma.delta.epsilon"
	_, err = Parse(multiPre)
	require.NoError(t, err, "Parse(%q) failed", multiPre)

	// Multiple build metadata identifiers
	multiBuild := "1.0.0+build.timestamp.12345.sha.abc123"
	_, err = Parse(multiBuild)
	require.NoError(t, err, "Parse(%q) failed", multiBuild)

	// Both pre-release and build metadata
	combined := "1.0.0-alpha.1+build.timestamp.12345"
	_, err = Parse(combined)
	require.NoError(t, err, "Parse(%q) failed", combined)
}
