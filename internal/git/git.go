package git

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/flaticols/bump/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

// openRepository opens the git repository from the current working directory
func openRepository() (*git.Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	repo, err := git.PlainOpen(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	return repo, nil
}

// CmdCurrentBranch returns the name of the current Git branch
func CmdCurrentBranch() (string, error) {
	repo, err := openRepository()
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		// Fallback: try to get branch from symbolic ref for repos without commits
		return getCurrentBranchFallback()
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	// Detached HEAD - return short hash
	return head.Hash().String()[:7], nil
}

// getCurrentBranchFallback gets the current branch name using exec fallback
func getCurrentBranchFallback() (string, error) {
	branch, err := runGitCommand("symbolic-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
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
	repo, err := openRepository()
	if err != nil {
		return false, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return !status.IsClean(), nil
}

// hasRemote checks if the repository has any remote configured
func hasRemote() (bool, error) {
	repo, err := openRepository()
	if err != nil {
		return false, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return false, err
	}

	return len(remotes) > 0, nil
}

// CmdHasRemoteChanges checks if there are remote changes that need to be pulled
func CmdHasRemoteChanges() (bool, error) {
	if ok, _ := hasRemote(); !ok {
		return false, fmt.Errorf("no remotes found in repository")
	}

	// Fetch latest changes using exec.Command as go-git fetch can be complex
	// with authentication setup
	cmd := exec.Command("git", "fetch", "origin")
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Get current branch using go-git
	currentBranch, err := CmdCurrentBranch()
	if err != nil {
		return false, err
	}

	// Check for remote changes using exec.Command for log comparison
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

	// Use exec.Command for rev-list comparison as it's more reliable
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

	repo, err := openRepository()
	if err != nil {
		return false, err
	}

	// Get local tags using go-git
	localTagSet := make(map[string]bool)
	tags, err := repo.Tags()
	if err != nil {
		return false, fmt.Errorf("failed to get local tags: %w", err)
	}
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		localTagSet[tagName] = true
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to iterate tags: %w", err)
	}

	// Get remote tags using exec.Command as it's more reliable with auth
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
	// Use exec.Command for for-each-ref as go-git doesn't provide easy access
	// to tag creation dates in the same format
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
	repo, err := openRepository()
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	_, err = repo.CreateTag(tag, head.Hash(), nil)
	if err != nil {
		return fmt.Errorf("error setting git tag: %w", err)
	}
	return nil
}

// CmdPushTag pushes the specified Git tag to the origin remote repository
func CmdPushTag(tag string) error {
	// Use exec.Command for push as go-git push requires complex auth setup
	cmd := exec.Command("git", "push", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pushing git tag: %v - %s", err, string(output))
	}
	return nil
}

// CmdRemoveTag removes a git tag from the local repository
func CmdRemoveTag(tag string) error {
	repo, err := openRepository()
	if err != nil {
		return err
	}

	err = repo.DeleteTag(tag)
	if err != nil {
		return fmt.Errorf("error removing local git tag: %w", err)
	}
	return nil
}

// CmdRemoveRemoteTag deletes a git tag from the remote repository
func CmdRemoveRemoteTag(tag string) error {
	// Use exec.Command for remote delete as go-git push requires complex auth setup
	cmd := exec.Command("git", "push", "--delete", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing remote git tag: %v - %s", err, string(output))
	}
	return nil
}
