package internal

import "github.com/Masterminds/semver/v3"

// parseTag parses a semantic version string from a tag and returns the version or an error if parsing fails.
func parseTag(tag string) (*semver.Version, error) {
	ver, err := semver.NewVersion(tag)
	if err != nil {
		return nil, err
	}

	return ver, nil
}
