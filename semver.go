package main

import "github.com/Masterminds/semver/v3"

func parseTag(tag string) (*semver.Version, error) {
	ver, err := semver.NewVersion(tag)
	if err != nil {
		return nil, err
	}

	return ver, nil
}
