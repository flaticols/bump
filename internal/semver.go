package internal

import (
	"github.com/flaticols/bump/semver"
)

// parseTag parses a semantic version string from a tag and returns the version or an error if parsing fails.
func parseTag(tag string) (semver.Version, error) {
	ver, err := semver.Parse(tag)
	if err != nil {
		return semver.Version{}, err
	}

	return ver, nil
}
