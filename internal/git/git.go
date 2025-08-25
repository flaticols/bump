package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/flaticols/bump/semver"
)

const DefaultVersion = "0.0.1"

type (
	SemVerTagError struct {
		Tag    string
		Msg    string
		NoTags bool
	}
)

func (e SemVerTagError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("error parsing semver tag: '%s': %s", e.Tag, e.Msg)
	}
	return fmt.Sprintf("error parsing semver tag: '%s'", e.Tag)
}

// CmdCurrentBranch checks if the current Git branch is one of the predefined default branches.
// Returns a boolean and an error if one occurs.
func CmdCurrentBranch() (string, error) {
	// Try the normal approach first
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	// If the command fails, try the fallback method
	if err != nil {
		// Try using symbolic-ref instead which works for repos without commits
		fallbackCmd := exec.Command("git", "symbolic-ref", "HEAD")
		fallbackOutput, fallbackErr := fallbackCmd.Output()
		if fallbackErr != nil {
			return "", fmt.Errorf("failed to get current branch: %w", fallbackErr)
		}
		// Remove the refs/heads/ prefix from the output
		branchRef := strings.TrimSpace(string(fallbackOutput))
		b := strings.TrimPrefix(branchRef, "refs/heads/")
		return b, nil
	}
	b := strings.TrimSpace(string(output))
	return b, nil
}

// CmdHasLocalChanges checks for uncommitted changes in the local Git repository by running `git status --porcelain` and returns the status.
func CmdHasLocalChanges() (bool, error) {
	// Run git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute git command: %w", err)
	}

	// If output is not empty, there are uncommitted changes
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// CmdHasRemoteChanges checks if there are changes in the remote repository that are not present in the local repository.
// It fetches the latest changes from the remote and compares the local branch with the tracking branch to detect differences.
func CmdHasRemoteChanges() (bool, error) {
	// First check if remotes exist
	remoteCmd := exec.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()

	// If no remotes exist
	if err != nil || len(strings.TrimSpace(string(remoteOutput))) == 0 {
		return false, fmt.Errorf("no remotes found in repository")
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

// CmdHasUnpushedChanges checks if there are commits in the local branch that haven't been pushed to the remote.
func CmdHasUnpushedChanges(branch string) (bool, error) {
	remoteCmd := exec.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()

	if err != nil || len(strings.TrimSpace(string(remoteOutput))) == 0 {
		return false, nil
	}

	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("origin/%s..%s", branch, branch))
	output, err := cmd.Output()
	if err != nil {
		checkRemoteBranchCmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
		remoteBranchOutput, _ := checkRemoteBranchCmd.Output()

		if len(strings.TrimSpace(string(remoteBranchOutput))) == 0 {
			checkLocalCommitsCmd := exec.Command("git", "rev-list", "--count", branch)
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

// CmdHasRemoteUnfetchedTags checks if there are tags in the remote repository that haven't been fetched locally.
// Returns true if unfetched tags exist, false otherwise, and an error if the process fails.
func CmdHasRemoteUnfetchedTags() (bool, error) {
	remoteCmd := exec.Command("git", "remote")
	remoteOutput, err := remoteCmd.Output()
	if err != nil || len(strings.TrimSpace(string(remoteOutput))) == 0 {
		return false, fmt.Errorf("no remotes found in repository")
	}

	localTagsCmd := exec.Command("git", "tag")
	localTagsOutput, err := localTagsCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get local tags: %w", err)
	}
	localTags := strings.Split(strings.TrimSpace(string(localTagsOutput)), "\n")
	localTagSet := make(map[string]bool)
	for _, tag := range localTags {
		if tag != "" {
			localTagSet[tag] = true
		}
	}

	lsRemoteCmd := exec.Command("git", "ls-remote", "--tags", "origin")
	lsRemoteOutput, err := lsRemoteCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list remote tags: %w", err)
	}

	for _, line := range strings.Split(strings.TrimSpace(string(lsRemoteOutput)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		refPath := parts[1]
		if strings.Contains(refPath, "^{}") {
			continue
		}
		tagName := strings.TrimPrefix(refPath, "refs/tags/")
		if !localTagSet[tagName] {
			return true, nil
		}
	}

	return false, nil
}

// CmdGetTag retrieves the latest Git tag from the current repository.
// Returns the tag as a semver.Version and an error if unsuccessful.
// CmdGetTag retrieves the latest Git tag from the current repository by grouping tags by creation timestamp.
// It returns the highest semver tag from the most recent group of tags that share the same creation timestamp.
func CmdGetTag(prefix string) (semver.Version, error) {
	cmd := exec.Command("git", "for-each-ref", "--sort=-creatordate", "--format=%(refname:short) %(creatordate:iso-strict)", "refs/tags")
	output, err := cmd.CombinedOutput()
	if err != nil {
		errOutput := string(output)
		if strings.Contains(errOutput, "No names found") ||
			strings.Contains(errOutput, "No tags") ||
			strings.Contains(errOutput, "fatal: No names found") {
			return semver.Version{}, SemVerTagError{NoTags: true}
		}
		return semver.Version{}, fmt.Errorf("error getting git tags: %v - %s", err, string(output))
	}

	trimmedOutput := strings.TrimSpace(string(output))
	if len(trimmedOutput) == 0 {
		return semver.Version{}, SemVerTagError{NoTags: true, Msg: "no tags found"}
	}

	lines := strings.Split(trimmedOutput, "\n")

	// Iterate over groups of same timestamp, return first group that has a valid tag for the given prefix
	var currentTimestamp string
	var groupTags []string
	flushGroup := func() (semver.Version, bool) {
		var validVersions []semver.Version
		for _, tag := range groupTags {
			candidate := tag
			if prefix != "" {
				if !strings.HasPrefix(candidate, prefix) {
					continue
				}
				candidate = strings.TrimPrefix(candidate, prefix)
			}
			if ver, ok := semver.IsValid(candidate); ok {
				validVersions = append(validVersions, ver)
			}
		}
		if len(validVersions) == 0 {
			return semver.Version{}, false
		}
		sort.Slice(validVersions, func(i, j int) bool {
			return semver.Compare(validVersions[i], validVersions[j]) > 0
		})
		return validVersions[0], true
	}

	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		tag := parts[0]
		timestamp := parts[1]
		if currentTimestamp == "" {
			currentTimestamp = timestamp
			groupTags = []string{tag}
			continue
		}
		if timestamp == currentTimestamp {
			groupTags = append(groupTags, tag)
			continue
		}
		// new timestamp encountered: process previous group
		if ver, ok := flushGroup(); ok {
			return ver, nil
		}
		// reset for new group
		currentTimestamp = timestamp
		groupTags = []string{tag}
	}
	// process last group
	if ver, ok := flushGroup(); ok {
		return ver, nil
	}

	// No tags found for the given prefix
	return semver.Version{}, SemVerTagError{NoTags: true, Msg: "no tags found for the given prefix"}
}

// CmdCreateTag creates a new Git tag with the specified name
func CmdCreateTag(tag string) error {
	cmd := exec.Command("git", "tag", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error setting git tag: %v - %s", err, string(output))
	}
	return nil
}

// CmdPushTag pushes the specified Git tag to the origin remote repository
func CmdPushTag(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pushing git tag: %v - %s", err, string(output))
	}
	return nil
}

// CmdRemoveTag removes a git tag from the local repository
func CmdRemoveTag(tag string) error {
	cmd := exec.Command("git", "tag", "-d", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing local git tag: %v - %s", err, string(output))
	}
	return nil
}

// CmdRemoveRemoteTag deletes a git tag from the remote repository
func CmdRemoveRemoteTag(tag string) error {
	cmd := exec.Command("git", "push", "--delete", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing remote git tag: %v - %s", err, string(output))
	}
	return nil
}
