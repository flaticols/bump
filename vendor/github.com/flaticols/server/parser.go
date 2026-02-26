package server

import (
	"strconv"
	"strings"
	"unicode"
)

// Parse parses a semver string, optionally with a v prefix.
func Parse(version string) (Version, error) {
	if version == "" {
		return Version{}, ErrEmptyVersion
	}
	if version[0] == 'v' {
		version = version[1:]
	}

	metaParts := strings.SplitN(version, "+", 2)
	var metaStr string
	versionPart := metaParts[0]
	if len(metaParts) > 1 {
		metaStr = metaParts[1]
		if metaStr == "" {
			return Version{}, ErrEmptyIdentifier
		}
	}

	preParts := strings.SplitN(versionPart, "-", 2)
	var preStr string
	core := preParts[0]
	if len(preParts) > 1 {
		preStr = preParts[1]
		if preStr == "" {
			return Version{}, ErrEmptyIdentifier
		}
	}

	nums := strings.Split(core, ".")
	if len(nums) != 3 {
		return Version{}, ErrMalformedCore
	}

	maj, err := parseNum(nums[0])
	if err != nil {
		return Version{}, err
	}
	min, err := parseNum(nums[1])
	if err != nil {
		return Version{}, err
	}
	pat, err := parseNum(nums[2])
	if err != nil {
		return Version{}, err
	}

	var prerelease []string
	if preStr != "" {
		prerelease = strings.Split(preStr, ".")
		for _, id := range prerelease {
			if err := validateID(id, true); err != nil {
				return Version{}, err
			}
		}
	}

	var metadata []string
	if metaStr != "" {
		metadata = strings.Split(metaStr, ".")
		for _, id := range metadata {
			if err := validateID(id, false); err != nil {
				return Version{}, err
			}
		}
	}

	return Version{Major: maj, Minor: min, Patch: pat, Prerelease: prerelease, Metadata: metadata}, nil
}

// ParseStrict parses a semver string without allowing v prefix.
func ParseStrict(version string) (Version, error) {
	if version == "" {
		return Version{}, ErrEmptyVersion
	}
	if version[0] == 'v' {
		return Version{}, ErrInvalidVPrefix
	}
	return Parse(version)
}

// IsValid returns the parsed version and whether parsing succeeded.
func IsValid(version string) (Version, bool) {
	v, err := Parse(version)
	return v, err == nil
}

// New constructs a Version directly.
func New(major, minor, patch int, prerelease, metadata []string) Version {
	return Version{Major: major, Minor: minor, Patch: patch, Prerelease: prerelease, Metadata: metadata}
}

func parseNum(s string) (int, error) {
	if s == "" {
		return 0, ErrEmptyVersionComponent
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, ErrLeadingZeros
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrNonDigitComponent
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, ErrNegativeComponent
	}
	return n, nil
}

func validateID(id string, isPrerelease bool) error {
	if id == "" {
		return ErrEmptyIdentifier
	}
	isNum := true
	for _, c := range id {
		if !unicode.IsDigit(c) {
			isNum = false
			break
		}
	}
	if isNum && isPrerelease && len(id) > 1 && id[0] == '0' {
		return ErrLeadingZeroesIdentifier
	}
	for _, c := range id {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) && c != '-' {
			return ErrInvalidIdentifierChars
		}
	}
	return nil
}
