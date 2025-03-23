package internal

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const DefaultVersion = "0.0.1"

type SemVerTagError struct {
	NoTags bool
	Tag    string
	Msg    string
}

func (e SemVerTagError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("error parsing semver tag: '%s': %s", e.Tag, e.Msg)
	}
	return fmt.Sprintf("error parsing semver tag: '%s'", e.Tag)
}

var defaultBranches = []string{"main", "master", "develop", "feature", "release", "hotfix", "bugfix", "latest"}

type GitState struct {
}

// CheckLocalChanges checks for uncommitted changes in the local Git repository by running `git status --porcelain` and returns the status.
func (gs *GitState) CheckLocalChanges() (bool, error) {
	// Run git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute git command: %w", err)
	}

	// If output is not empty, there are uncommitted changes
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// CheckRemoteChanges checks if there are changes in the remote repository that are not present in the local repository.
// It fetches the latest changes from the remote and compares the local branch with the tracking branch to detect differences.
func (gs *GitState) CheckRemoteChanges() (bool, error) {
	// Fetch the latest changes from remote
	fetchCmd := exec.Command("git", "fetch", "origin")
	if err := fetchCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Check if there are remote changes not in local
	cmd := exec.Command("git", "log", "HEAD..origin/main", "--oneline")
	output, err := cmd.Output()
	if err != nil {
		// Check if the error is due to invalid HEAD..origin/main reference
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			// Try with default branch
			cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
			branchOutput, branchErr := cmd.Output()
			if branchErr != nil {
				return false, fmt.Errorf("failed to get current branch: %w", branchErr)
			}

			currentBranch := strings.TrimSpace(string(branchOutput))
			cmd = exec.Command("git", "log", fmt.Sprintf("HEAD..origin/%s", currentBranch), "--oneline")
			output, err = cmd.Output()
			if err != nil {
				return false, fmt.Errorf("failed to check remote changes: %w", err)
			}
		} else {
			return false, fmt.Errorf("failed to check remote changes: %w", err)
		}
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// IsDefaultBranch checks if the current Git branch is one of the predefined default branches and returns a boolean and an error if one occurs.
func (gs *GitState) IsDefaultBranch() (string, bool, error) {
	// Try the normal approach first
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()

	// If the command fails, try the fallback method
	if err != nil {
		// Try using symbolic-ref instead which works for repos without commits
		fallbackCmd := exec.Command("git", "symbolic-ref", "HEAD")
		fallbackOutput, fallbackErr := fallbackCmd.Output()

		if fallbackErr != nil {
			return "", false, fmt.Errorf("failed to get current branch: %w", fallbackErr)
		}

		// Remove the refs/heads/ prefix from the output
		branchRef := strings.TrimSpace(string(fallbackOutput))
		b := strings.TrimPrefix(branchRef, "refs/heads/")
		return b, slices.Contains(defaultBranches, b), nil
	}

	b := strings.TrimSpace(string(output))
	return b, slices.Contains(defaultBranches, b), nil
}

// getLatestGitTag retrieves the latest Git tag from the current repository.
// Returns the tag as a string, a boolean indicating initialization state, and an error if unsuccessful.
func getLatestGitTag() (string, error) {
	// Run the git command to get the latest tag
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.CombinedOutput()

	// If the command failed, check if it's because there are no tags
	if err != nil {
		// Convert output to string for error checking
		errOutput := string(output)
		if strings.Contains(errOutput, "No names found") ||
			strings.Contains(errOutput, "No tags") ||
			strings.Contains(errOutput, "fatal: No names found") {
			return "", SemVerTagError{NoTags: true}
		}
		return "", fmt.Errorf("error getting git tag: %v - %s", err, string(output))
	}

	// Trim the output to remove newlines and whitespace
	tag := strings.TrimSpace(string(output))
	return tag, nil
}

// SetGitTag creates a new Git tag with the specified name and returns an error if the process fails or the tag could not be created.
func (gs *GitState) SetGitTag(tag string) error {
	cmd := exec.Command("git", "tag", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error setting git tag: %v - %s", err, string(output))
	}
	return nil
}

// PushGitTag pushes the specified Git tag to the origin remote repository. It returns an error if the command execution fails.
func (gs *GitState) PushGitTag(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pushing git tag: %v - %s", err, string(output))
	}
	return nil
}

// GetCurrentVersion retrieves the current version state from Git tags.
// Returns the current version as a semver.Version and an error if unsuccessful.
func (gs *GitState) GetCurrentVersion() (*semver.Version, error) {
	tag, err := getLatestGitTag()
	if err != nil {
		return nil, err
	}

	// Remove 'v' prefix if present
	tag = strings.TrimPrefix(tag, "v")

	// Parse the version string
	version, err := semver.NewVersion(tag)
	if err != nil {
		return nil, SemVerTagError{Tag: tag, Msg: err.Error()}
	}

	return version, nil
}
