package semver

import (
	"strconv"
	"strings"
	"unicode"
)

// Parse parses a version string into a Version object.
// Returns an error if the version string is not valid SemVer.
func Parse(version string) (Version, error) {
	return parse(version, false)
}

// ParseWithVPrefix parses a version string into a Version object,
// optionally accepting a 'v' prefix (e.g., v1.2.3).
// Returns an error if the version string is not valid SemVer.
func ParseWithVPrefix(version string) (Version, error) {
	return parse(version, true)
}

// parse is the internal implementation that handles the actual parsing
// with an option to allow 'v' prefix
func parse(version string, allowVPrefix bool) (Version, error) {
	if version == "" {
		return Version{}, ErrEmptyVersion
	}

	// Remove 'v' prefix if present and allowed
	if allowVPrefix && len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}

	// Split by build metadata delimiter ('+')
	versionAndMeta := strings.SplitN(version, "+", 2)

	var metaPart string
	versionPart := versionAndMeta[0]
	if len(versionAndMeta) > 1 {
		metaPart = versionAndMeta[1]
		// Check for empty metadata immediately
		if metaPart == "" {
			return Version{}, ErrEmptyIdentifier
		}
	}

	// Split by prerelease delimiter ('-')
	versionAndPrerelease := strings.SplitN(versionPart, "-", 2)

	var prereleasePart string
	corePart := versionAndPrerelease[0]
	if len(versionAndPrerelease) > 1 {
		prereleasePart = versionAndPrerelease[1]
		// Check for empty prerelease immediately
		if prereleasePart == "" {
			return Version{}, ErrEmptyIdentifier
		}
	}

	// Parse core version (major.minor.patch)
	coreNums := strings.Split(corePart, ".")
	if len(coreNums) != 3 {
		return Version{}, ErrMalformedCore
	}

	major, err := parseVersionNumber(coreNums[0])
	if err != nil {
		return Version{}, err
	}

	minor, err := parseVersionNumber(coreNums[1])
	if err != nil {
		return Version{}, err
	}

	patch, err := parseVersionNumber(coreNums[2])
	if err != nil {
		return Version{}, err
	}

	// Parse prerelease part
	var prerelease []string
	if prereleasePart != "" {
		prerelease = strings.Split(prereleasePart, ".")
		for _, id := range prerelease {
			if err := validateIdentifier(id, true); err != nil {
				return Version{}, err
			}
		}
	}

	// Parse metadata part
	var metadata []string
	if metaPart != "" {
		metadata = strings.Split(metaPart, ".")
		for _, id := range metadata {
			if err := validateIdentifier(id, false); err != nil {
				return Version{}, err
			}
		}
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Metadata:   metadata,
	}, nil
}

// parseVersionNumber parses a version number component with SemVer rules.
// Rules: must be a non-negative integer without leading zeros.
func parseVersionNumber(s string) (int, error) {
	if s == "" {
		return 0, ErrEmptyVersionComponent
	}

	// Check for leading zeros in multi-digit numbers
	if len(s) > 1 && s[0] == '0' {
		return 0, ErrLeadingZeros
	}

	// Ensure it's a valid integer
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrNonDigitComponent
		}
	}

	num, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	if num < 0 {
		return 0, ErrNegativeComponent
	}

	return num, nil
}

// validateIdentifier checks if an identifier is valid according to SemVer rules.
// isPrerelease indicates if this is a prerelease identifier (which has special rules).
func validateIdentifier(id string, isPrerelease bool) error {
	if id == "" {
		return ErrEmptyIdentifier
	}

	// Check if it's numeric
	isNumeric := true
	for _, c := range id {
		if !unicode.IsDigit(c) {
			isNumeric = false
			break
		}
	}

	// Numeric identifier with leading zeros is invalid
	if isNumeric && isPrerelease && len(id) > 1 && id[0] == '0' {
		return ErrLeadingZeroesIdentifier
	}

	// All characters must be alphanumeric or hyphen
	for _, c := range id {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) && c != '-' {
			return ErrInvalidIdentifierChars
		}
	}

	return nil
}

// IsValid returns true if the version string is valid SemVer.
func IsValid(version string) bool {
	_, err := Parse(version)
	return err == nil
}

// IsValidWithVPrefix returns true if the version string is valid SemVer,
// optionally accepting a 'v' prefix.
func IsValidWithVPrefix(version string) bool {
	_, err := ParseWithVPrefix(version)
	return err == nil
}

// New creates a new Version with the given components.
func New(major, minor, patch int, prerelease, metadata []string) Version {
	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Metadata:   metadata,
	}
}
