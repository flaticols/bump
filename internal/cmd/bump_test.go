package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGetIncPart tests the getIncPart function
func TestGetIncPart(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected semVerPart
	}{
		{
			name:     "No arguments defaults to patch",
			args:     []string{},
			expected: patch,
		},
		{
			name:     "Major argument",
			args:     []string{"major"},
			expected: major,
		},
		{
			name:     "Minor argument",
			args:     []string{"minor"},
			expected: minor,
		},
		{
			name:     "Patch argument",
			args:     []string{"patch"},
			expected: patch,
		},
		{
			name:     "Package name only defaults to patch",
			args:     []string{"pkg/x"},
			expected: patch,
		},
		{
			name:     "Major with package name",
			args:     []string{"major", "pkg/x"},
			expected: major,
		},
		{
			name:     "Minor with package name",
			args:     []string{"minor", "services/api"},
			expected: minor,
		},
		{
			name:     "Patch with package name",
			args:     []string{"patch", "libs/utils"},
			expected: patch,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getIncPart(tc.args)
			require.Equal(t, tc.expected, result)
		})
	}
}

// TestGetPackageName tests the getPackageName function
func TestGetPackageName(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "No arguments",
			args:     []string{},
			expected: "",
		},
		{
			name:     "Major only, no package",
			args:     []string{"major"},
			expected: "",
		},
		{
			name:     "Minor only, no package",
			args:     []string{"minor"},
			expected: "",
		},
		{
			name:     "Patch only, no package",
			args:     []string{"patch"},
			expected: "",
		},
		{
			name:     "Package name only",
			args:     []string{"pkg/x"},
			expected: "pkg/x",
		},
		{
			name:     "Package name with underscores",
			args:     []string{"pkg_x"},
			expected: "pkg_x",
		},
		{
			name:     "Package name with nested path",
			args:     []string{"pkg/x/foo"},
			expected: "pkg/x/foo",
		},
		{
			name:     "Major with package name",
			args:     []string{"major", "pkg/x"},
			expected: "pkg/x",
		},
		{
			name:     "Minor with package name",
			args:     []string{"minor", "services/api"},
			expected: "services/api",
		},
		{
			name:     "Patch with package name",
			args:     []string{"patch", "libs/utils"},
			expected: "libs/utils",
		},
		{
			name:     "Service prefix pattern",
			args:     []string{"services/api"},
			expected: "services/api",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getPackageName(tc.args)
			require.Equal(t, tc.expected, result)
		})
	}
}
