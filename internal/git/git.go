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

// CmdCurrentBranch returns the name of the current Git branch
func CmdCurrentBranch() (string, error) {
	return getCurrentBranch()
}

// getCurrentBranch gets the current branch name with fallback
func getCurrentBranch() (string, error) {
	// Try rev-parse first (works for most cases)
	if branch, err := runGitCommand("rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		return strings.TrimSpace(branch), nil
	}

	// Fallback to symbolic-ref (works for repos without commits)
	branch, err := runGitCommand("symbolic-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	// Remove refs/heads/ prefix
	return strings.TrimPrefix(strings.TrimSpace(branch), "refs/heads/"), nil
}

// runGitCommand executes a git command and returns its output
func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// CmdHasLocalChanges checks for uncommitted changes in the local repository
func CmdHasLocalChanges() (bool, error) {
	output, err := runGitCommand("status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to execute git command: %w", err)
	}
	return strings.TrimSpace(output) != "", nil
}

// hasRemote checks if the repository has any remote configured
func hasRemote() (bool, error) {
	output, err := runGitCommand("remote")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

// CmdHasRemoteChanges checks if there are remote changes that need to be pulled
func CmdHasRemoteChanges() (bool, error) {
	if ok, _ := hasRemote(); !ok {
		return false, fmt.Errorf("no remotes found in repository")
	}

	// Fetch latest changes
	cmd := exec.Command("git", "fetch", "origin")
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Get current branch
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return false, err
	}

	// Check for remote changes (try origin/main first, then current branch)
	for _, remoteBranch := range []string{"origin/main", fmt.Sprintf("origin/%s", currentBranch)} {
		output, err := runGitCommand("log", fmt.Sprintf("HEAD..%s", remoteBranch), "--oneline")
		if err == nil {
			return strings.TrimSpace(output) != "", nil
		}
	}

	return false, fmt.Errorf("failed to check remote changes")
}

// CmdHasUnpushedChanges checks if there are unpushed commits on the given branch
func CmdHasUnpushedChanges(branch string) (bool, error) {
	if ok, _ := hasRemote(); !ok {
		return false, nil
	}

	// Try to count commits ahead of remote
	output, err := runGitCommand("rev-list", "--count", fmt.Sprintf("origin/%s..%s", branch, branch))
	if err == nil {
		return strings.TrimSpace(output) != "0", nil
	}

	// If remote branch doesn't exist, check if we have local commits
	remoteBranchOutput, _ := runGitCommand("ls-remote", "--heads", "origin", branch)
	if strings.TrimSpace(remoteBranchOutput) == "" {
		localCount, err := runGitCommand("rev-list", "--count", branch)
		if err != nil {
			return false, fmt.Errorf("failed to check local commits: %w", err)
		}
		return strings.TrimSpace(localCount) != "0", nil
	}

	return false, fmt.Errorf("failed to check unpushed changes: %w", err)
}

// CmdHasRemoteUnfetchedTags checks if there are unfetched tags in the remote repository
func CmdHasRemoteUnfetchedTags() (bool, error) {
	if ok, _ := hasRemote(); !ok {
		return false, fmt.Errorf("no remotes found in repository")
	}

	// Get local tags
	localTagsOutput, err := runGitCommand("tag")
	if err != nil {
		return false, fmt.Errorf("failed to get local tags: %w", err)
	}

	localTagSet := make(map[string]bool)
	for _, tag := range strings.Split(strings.TrimSpace(localTagsOutput), "\n") {
		if tag != "" {
			localTagSet[tag] = true
		}
	}

	// Get remote tags
	remoteOutput, err := runGitCommand("ls-remote", "--tags", "origin")
	if err != nil {
		return false, fmt.Errorf("failed to list remote tags: %w", err)
	}

	// Check for unfetched tags
	for _, line := range strings.Split(strings.TrimSpace(remoteOutput), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 || strings.Contains(parts[1], "^{}") {
			continue
		}
		tagName := strings.TrimPrefix(parts[1], "refs/tags/")
		if !localTagSet[tagName] {
			return true, nil
		}
	}

	return false, nil
}

// CmdGetTag retrieves the latest semver tag from the repository.
// Tags are grouped by creation timestamp, and the highest semver tag from the most
// recent group is returned. An optional prefix can filter tags (e.g., "pkg/x/").
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
