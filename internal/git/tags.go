package git

import (
	"sort"
	"time"
)

// GitTag represents a Git tag with its creation timestamp
type GitTag struct {
	Name      string
	CreatedAt time.Time
}

// GetLatestGitTag returns the most recently created tag from the repository
// When multiple tags have the same timestamp, it returns only the latest one created
func GetLatestGitTag(repoPath string) (string, error) {
	// Get all tags with their creation timestamps
	tags, err := getAllTagsWithTimestamps(repoPath)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	// Sort tags by creation time (newest first)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].CreatedAt.After(tags[j].CreatedAt)
	})

	// Return the most recent tag
	return tags[0].Name, nil
}

// getAllTagsWithTimestamps gets all tags with their creation timestamps
func getAllTagsWithTimestamps(repoPath string) ([]GitTag, error) {
	// Execute git command to get tags with timestamps
	// Example: git for-each-ref --sort=-creatordate --format='%(refname:short) %(creatordate:iso8601)' refs/tags

	// Parse the output and create GitTag objects

	// When multiple tags have the same timestamp, the git command will return them
	// in reverse creation order, with the most recently created first

	// Implementation details would depend on how you execute git commands in your codebase

	// This is a placeholder - implement based on your actual git command execution

	return []GitTag{}, nil
}
