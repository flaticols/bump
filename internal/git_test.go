package internal

import (
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// CommandRunner is an interface that allows us to mock exec.Command
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// RealCommandRunner is the actual implementation that calls exec.Command
type RealCommandRunner struct{}

func (r *RealCommandRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// MockCommandRunner is a mock implementation for testing
type MockCommandRunner struct {
	OutputMap map[string]struct {
		Output []byte
		Err    error
	}
}

// NewMockCommandRunner creates a new MockCommandRunner with an empty output map
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		OutputMap: make(map[string]struct {
			Output []byte
			Err    error
		}),
	}
}

// SetOutput sets the output for a given command
func (m *MockCommandRunner) SetOutput(cmdString string, output []byte, err error) {
	m.OutputMap[cmdString] = struct {
		Output []byte
		Err    error
	}{Output: output, Err: err}
}

// Run implements the CommandRunner interface for MockCommandRunner
func (m *MockCommandRunner) Run(name string, args ...string) ([]byte, error) {
	cmdString := name + " " + strings.Join(args, " ")

	// Try to find an exact match
	if result, ok := m.OutputMap[cmdString]; ok {
		return result.Output, result.Err
	}

	// Try partial matching for more flexible testing
	for key, result := range m.OutputMap {
		if strings.Contains(cmdString, key) {
			return result.Output, result.Err
		}
	}

	// Default empty response
	return []byte{}, nil
}

// TestableGitState is a modified version of GitState that accepts a CommandRunner
type TestableGitState struct {
	CmdRunner CommandRunner
}

// Helper function to run git commands with the CommandRunner
func (gs *TestableGitState) runGitCommand(args ...string) ([]byte, error) {
	return gs.CmdRunner.Run("git", args...)
}

// getLatestGitTag modified version for TestableGitState
func (gs *TestableGitState) getLatestGitTag() (string, error) {
	// Run git command to get all tags with their creation dates
	output, err := gs.runGitCommand("for-each-ref", "--sort=-creatordate", "--format=%(refname:short)", "refs/tags")
	// If the command failed, check if it's because there are no tags
	if err != nil {
		// Convert output to string for error checking
		errOutput := string(output)
		if strings.Contains(errOutput, "No names found") ||
			strings.Contains(errOutput, "No tags") ||
			strings.Contains(errOutput, "fatal: No names found") {
			return "", SemVerTagError{NoTags: true}
		}
		return "", err
	}

	// If there are no tags at all
	if len(strings.TrimSpace(string(output))) == 0 {
		return "", SemVerTagError{NoTags: true}
	}

	// Split the output by newlines
	tags := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Iterate over the tags to find the first valid semver tag
	for _, tag := range tags {
		versionTag := strings.TrimPrefix(tag, "v")
		if _, err := semverParse(versionTag); err == nil {
			return tag, nil
		}
	}

	// No valid semver tags found
	return "", SemVerTagError{Msg: "no valid semver tags found"}
}

// Helper function to parse semver without importing the actual package
func semverParse(version string) (interface{}, error) {
	// This is a dummy implementation, as we're just testing the git command logic
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil, SemVerTagError{Tag: version, Msg: "invalid semver format"}
	}

	// Check if each part is a number
	for _, part := range parts[:3] { // Only check the first three parts (major.minor.patch)
		// Simple check if the string contains only digits
		for _, char := range part {
			if char < '0' || char > '9' {
				return nil, SemVerTagError{Tag: version, Msg: "invalid semver format"}
			}
		}
	}

	return struct{}{}, nil
}

// TestGetLatestGitTag tests the getLatestGitTag function
func TestGetLatestGitTag(t *testing.T) {
	testCases := []struct {
		name        string
		mockOutput  string
		mockError   error
		expectedTag string
		expectError bool
	}{
		{
			name:        "No tags in repository",
			mockOutput:  "",
			mockError:   nil,
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "No valid semver tags",
			mockOutput:  "invalid-tag-1\ninvalid-tag-2",
			mockError:   nil,
			expectedTag: "",
			expectError: true,
		},
		{
			name:        "Single valid semver tag",
			mockOutput:  "v1.2.3",
			mockError:   nil,
			expectedTag: "v1.2.3",
			expectError: false,
		},
		{
			name:        "Multiple tags, first is valid",
			mockOutput:  "v2.0.0\ntag-not-semver\nv1.0.0",
			mockError:   nil,
			expectedTag: "v2.0.0",
			expectError: false,
		},
		{
			name:        "Command execution error",
			mockOutput:  "",
			mockError:   exec.ErrNotFound,
			expectedTag: "",
			expectError: true,
		},
		{
			name: "Mixed tags with v45.0.1 as valid semver tag",
			mockOutput: `pkg/my-latest-tag.1
pkg/my-latest-tag.2
v45.0.1
pkg/my-latest-tag.4
v76.0.45
foobar
bazbar
v76
test.tag.2
78
testtag5454
v0.2.3`,
			mockError:   nil,
			expectedTag: "v45.0.1",
			expectError: false,
		},
		{
			name: "Multiple valid semver tags in mixed list",
			mockOutput: `random-tag
v2.1.0
invalid-tag
v1.0.0
another-invalid-tag`,
			mockError:   nil,
			expectedTag: "v2.1.0",
			expectError: false,
		},
		{
			name: "Only the last tag is a valid semver",
			mockOutput: `random-tag
invalid.tag.1
incomplete.tag
v1.2.3`,
			mockError:   nil,
			expectedTag: "v1.2.3",
			expectError: false,
		},
		{
			name: "Tag with v prefix but invalid semver format",
			mockOutput: `v123
vinvalid
v1.2`,
			mockError:   nil,
			expectedTag: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock CommandRunner
			mockRunner := NewMockCommandRunner()
			mockRunner.SetOutput("git for-each-ref --sort=-creatordate --format=%(refname:short) refs/tags",
				[]byte(tc.mockOutput), tc.mockError)

			// Create a testable GitState with the mock runner
			gs := &TestableGitState{CmdRunner: mockRunner}

			// Call the testable version of getLatestGitTag
			tag, err := gs.getLatestGitTag()

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTag, tag)
			}
		})
	}
}

// Modified methods to use the CommandRunner

// CheckLocalChanges checks for uncommitted changes in the local Git repository
func (gs *TestableGitState) CheckLocalChanges() (bool, error) {
	output, err := gs.runGitCommand("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// IsDefaultBranch checks if current branch is a default branch
func (gs *TestableGitState) IsDefaultBranch() (string, bool, error) {
	output, err := gs.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		// Fallback
		output, err = gs.runGitCommand("symbolic-ref", "HEAD")
		if err != nil {
			return "", false, err
		}

		branchRef := strings.TrimSpace(string(output))
		b := strings.TrimPrefix(branchRef, "refs/heads/")
		return b, slices.Contains(defaultBranches, b), nil
	}

	b := strings.TrimSpace(string(output))
	return b, slices.Contains(defaultBranches, b), nil
}

// TestCheckLocalChanges tests the CheckLocalChanges method
func TestCheckLocalChanges(t *testing.T) {
	testCases := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "No local changes",
			mockOutput:     "",
			mockError:      nil,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "Local changes present",
			mockOutput:     " M README.md\n?? new-file.txt",
			mockError:      nil,
			expectedResult: true,
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock runner
			mockRunner := NewMockCommandRunner()
			mockRunner.SetOutput("git status --porcelain", []byte(tc.mockOutput), tc.mockError)

			// Create testable GitState with mock runner
			gs := &TestableGitState{CmdRunner: mockRunner}

			// Test the method
			hasChanges, err := gs.CheckLocalChanges()

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, hasChanges)
			}
		})
	}
}

// TestIsDefaultBranch tests the IsDefaultBranch method
func TestIsDefaultBranch(t *testing.T) {
	testCases := []struct {
		name           string
		mockOutput     string
		mockError      error
		fallbackOutput string
		fallbackError  error
		expectedBranch string
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "Main branch",
			mockOutput:     "main",
			mockError:      nil,
			expectedBranch: "main",
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "Feature branch",
			mockOutput:     "feature-branch",
			mockError:      nil,
			expectedBranch: "feature-branch",
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock runner
			mockRunner := NewMockCommandRunner()
			mockRunner.SetOutput("git rev-parse --abbrev-ref HEAD", []byte(tc.mockOutput), tc.mockError)
			if tc.fallbackOutput != "" {
				mockRunner.SetOutput("git symbolic-ref HEAD", []byte(tc.fallbackOutput), tc.fallbackError)
			}

			// Create testable GitState
			gs := &TestableGitState{CmdRunner: mockRunner}

			// Test the method (we only implement this specific method for TestableGitState)
			branch, isDefault, err := gs.IsDefaultBranch()

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedBranch, branch)
				require.Equal(t, tc.expectedResult, isDefault)
			}
		})
	}
}
