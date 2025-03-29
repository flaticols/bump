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

// Note: In a real implementation, you would implement all methods of GitState
// in TestableGitState and write tests for each. This is a simplified version
// to demonstrate the approach.
