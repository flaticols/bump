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
func (gs *GitState) CheckRemoteChanges(allowNoRemotes bool) (bool, error) {
	// First check if remotes exist
	remoteCmd := exec.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()

	// If no remotes exist
	if err != nil || len(strings.TrimSpace(string(remoteOutput))) == 0 {
		if !allowNoRemotes {
			return false, fmt.Errorf("no remotes found in repository")
		}
		// If we don't want to error on no remotes, just return no changes
		return false, nil
	}

	// Fetch the latest changes from remote
	fetchCmd := exec.Command("git", "fetch", "origin")
	if err := fetchCmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Get current branch
	branchCmd := exec.Command("git", "symbolic-ref", "HEAD")
	branchOutput, branchErr := branchCmd.Output()

	// If we can't get the branch, try the fallback
	if branchErr != nil {
		fallbackCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		fallbackOutput, fallbackErr := fallbackCmd.Output()
		if fallbackErr != nil {
			return false, fmt.Errorf("failed to get current branch: %w", fallbackErr)
		}
		branchOutput = fallbackOutput
	}

	currentBranch := strings.TrimSpace(string(branchOutput))
	currentBranch = strings.TrimPrefix(currentBranch, "refs/heads/")

	// Check if there are remote changes not in local, first try with origin/main
	cmd := exec.Command("git", "log", "HEAD..origin/main", "--oneline")
	output, err := cmd.Output()
	if err != nil {
		// Try with current branch
		cmd = exec.Command("git", "log", fmt.Sprintf("HEAD..origin/%s", currentBranch), "--oneline")
		output, err = cmd.Output()
		if err != nil {
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

func (gs *GitState) HasUnpushedChanges(currentBranch string) (bool, error) {
	remoteCmd := exec.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()

	if err != nil || len(strings.TrimSpace(string(remoteOutput))) == 0 {
		return false, nil
	}

	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("origin/%s..%s", currentBranch, currentBranch))
	output, err := cmd.Output()

	if err != nil {
		checkRemoteBranchCmd := exec.Command("git", "ls-remote", "--heads", "origin", currentBranch)
		remoteBranchOutput, _ := checkRemoteBranchCmd.Output()

		if len(strings.TrimSpace(string(remoteBranchOutput))) == 0 {
			checkLocalCommitsCmd := exec.Command("git", "rev-list", "--count", currentBranch)
			localCommitsOutput, localErr := checkLocalCommitsCmd.Output()
			if localErr != nil {
				return false, fmt.Errorf("failed to check local commits: %w", localErr)
			}

			count := strings.TrimSpace(string(localCommitsOutput))
			return count != "0", nil
		}
		return false, fmt.Errorf("failed to check unpushed changes: %w", err)
	}
	count := strings.TrimSpace(string(output))
	return count != "0", nil
}

// getLatestGitTag retrieves the latest Git tag from the current repository.
// Returns the tag as a string, a boolean indicating initialization state, and an error if unsuccessful.
func getLatestGitTag() (string, error) {
	// Run git command to get all tags
	cmd := exec.Command("git", "tag")
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
		return "", fmt.Errorf("error getting git tags: %v - %s", err, string(output))
	}

	// If there are no tags at all
	if len(strings.TrimSpace(string(output))) == 0 {
		return "", SemVerTagError{NoTags: true}
	}

	// Split the output by newlines
	tags := strings.Split(strings.TrimSpace(string(output)), "\n")

	var highestSemver *semver.Version
	var highestTag string

	// Find the highest semver tag
	for _, tag := range tags {
		v, err := semver.NewVersion(tag)
		if err == nil {
			// This is a valid semver tag
			if highestSemver == nil || v.GreaterThan(highestSemver) {
				highestSemver = v
				highestTag = tag
			}
		}
	}

	if highestSemver == nil {
		// No valid semver tags found
		return "", SemVerTagError{Msg: "not valid semver tags found"}
	}

	return highestTag, nil
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

func (gs *GitState) RemoveLocalGitTag(tag string) error {
	cmd := exec.Command("git", "tag", "-d", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing local git tag: %v - %s", err, string(output))
	}
	return nil
}

// removeRemoteGitTag deletes a git tag from the remote repository
func (gs *GitState) RemoveRemoteGitTag(tag string) error {
	cmd := exec.Command("git", "push", "--delete", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing remote git tag: %v - %s", err, string(output))
	}
	return nil
}
