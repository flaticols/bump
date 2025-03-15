package main

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

var defaultBranches = []string{"main", "master", "develop", "feature", "release", "hotfix", "bugfix", "latest"}

func checkLocalChanges() (bool, error) {
	// Run git status --porcelain
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to execute git command: %w", err)
	}

	// If output is not empty, there are uncommitted changes
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func checkRemoteChanges() (bool, error) {
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

func isDefaultBranch() (bool, error) {
	// Get the current branch name
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get current branch: %w", err)
	}

	b := strings.TrimSpace(string(output))
	return slices.Contains(defaultBranches, b), nil
}

// GetLatestGitTag returns the latest Git tag or a default value if no tags exist
func getLatestGitTag() (string, bool, error) {
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
			return "", true, nil
		}
		return "", false, fmt.Errorf("error getting git tag: %v - %s", err, string(output))
	}

	// Trim the output to remove newlines and whitespace
	tag := strings.TrimSpace(string(output))
	return tag, false, nil
}

func setGitTag(tag string) error {
	cmd := exec.Command("git", "tag", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error setting git tag: %v - %s", err, string(output))
	}
	return nil
}

func pushGitTag(tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pushing git tag: %v - %s", err, string(output))
	}
	return nil
}
