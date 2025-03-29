package semver

import (
	"strconv"
)

// Compare compares two versions v1 and v2:
//
//	-1 if v1 < v2
//	 0 if v1 = v2
//	+1 if v1 > v2
//
// Build metadata is ignored when comparing versions.
func Compare(v1, v2 Version) int {
	// Compare major version
	if v1.Major < v2.Major {
		return -1
	}
	if v1.Major > v2.Major {
		return 1
	}

	// Compare minor version
	if v1.Minor < v2.Minor {
		return -1
	}
	if v1.Minor > v2.Minor {
		return 1
	}

	// Compare patch version
	if v1.Patch < v2.Patch {
		return -1
	}
	if v1.Patch > v2.Patch {
		return 1
	}

	// At this point, the core version is the same
	// If one has prerelease and the other doesn't, the one without prerelease is greater
	if len(v1.Prerelease) == 0 && len(v2.Prerelease) > 0 {
		return 1
	}
	if len(v1.Prerelease) > 0 && len(v2.Prerelease) == 0 {
		return -1
	}

	// If both have prerelease, compare them
	return comparePrerelease(v1.Prerelease, v2.Prerelease)
}

// comparePrerelease compares two prerelease arrays according to SemVer rules.
func comparePrerelease(pre1, pre2 []string) int {
	// Compare identifiers one by one
	minLen := len(pre1)
	if len(pre2) < minLen {
		minLen = len(pre2)
	}

	for i := 0; i < minLen; i++ {
		id1 := pre1[i]
		id2 := pre2[i]

		// If both are numeric, compare them as integers
		num1, err1 := strconv.Atoi(id1)
		num2, err2 := strconv.Atoi(id2)
		if err1 == nil && err2 == nil {
			if num1 < num2 {
				return -1
			}
			if num1 > num2 {
				return 1
			}
			// Continue to next identifier if they're equal
			continue
		}

		// If one is numeric and the other is not, numeric is smaller
		if err1 == nil && err2 != nil {
			return -1
		}
		if err1 != nil && err2 == nil {
			return 1
		}

		// If both are non-numeric, compare them lexically
		if id1 < id2 {
			return -1
		}
		if id1 > id2 {
			return 1
		}
		// Continue to next identifier if they're equal
	}

	// If we get here, one prerelease array is a prefix of the other
	// The shorter array is smaller
	if len(pre1) < len(pre2) {
		return -1
	}
	if len(pre1) > len(pre2) {
		return 1
	}

	// They're the same
	return 0
}

// Equal returns true if v1 and v2 are equal.
// Build metadata is ignored when comparing versions.
func (ver Version) Equal(v2 Version) bool {
	return Compare(ver, v2) == 0
}

// GreaterThan returns true if v1 is greater than v2.
func (ver Version) GreaterThan(v2 Version) bool {
	return Compare(ver, v2) > 0
}

// LessThan returns true if v1 is less than v2.
func (ver Version) LessThan(v2 Version) bool {
	return Compare(ver, v2) < 0
}

// GreaterThanOrEqual returns true if v1 is greater than or equal to v2.
func (ver Version) GreaterThanOrEqual(v2 Version) bool {
	return Compare(ver, v2) >= 0
}

// LessThanOrEqual returns true if v1 is less than or equal to v2.
func (ver Version) LessThanOrEqual(v2 Version) bool {
	return Compare(ver, v2) <= 0
}
