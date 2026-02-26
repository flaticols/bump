package server

import "strconv"

// Compare returns -1, 0, or +1 comparing v1 to v2. Metadata is ignored.
func Compare(v1, v2 Version) int {
	if v1.Major != v2.Major {
		return cmpInt(v1.Major, v2.Major)
	}
	if v1.Minor != v2.Minor {
		return cmpInt(v1.Minor, v2.Minor)
	}
	if v1.Patch != v2.Patch {
		return cmpInt(v1.Patch, v2.Patch)
	}
	if len(v1.Prerelease) == 0 && len(v2.Prerelease) > 0 {
		return 1
	}
	if len(v1.Prerelease) > 0 && len(v2.Prerelease) == 0 {
		return -1
	}
	return comparePre(v1.Prerelease, v2.Prerelease)
}

func comparePre(a, b []string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		n1, e1 := strconv.Atoi(a[i])
		n2, e2 := strconv.Atoi(b[i])
		if e1 == nil && e2 == nil {
			if c := cmpInt(n1, n2); c != 0 {
				return c
			}
			continue
		}
		if e1 == nil {
			return -1
		}
		if e2 == nil {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return cmpInt(len(a), len(b))
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// Equal returns true if ver == v2 (ignoring metadata).
func (ver Version) Equal(v2 Version) bool { return Compare(ver, v2) == 0 }

// GreaterThan returns true if ver > v2.
func (ver Version) GreaterThan(v2 Version) bool { return Compare(ver, v2) > 0 }

// LessThan returns true if ver < v2.
func (ver Version) LessThan(v2 Version) bool { return Compare(ver, v2) < 0 }

// GreaterThanOrEqual returns true if ver >= v2.
func (ver Version) GreaterThanOrEqual(v2 Version) bool { return Compare(ver, v2) >= 0 }

// LessThanOrEqual returns true if ver <= v2.
func (ver Version) LessThanOrEqual(v2 Version) bool { return Compare(ver, v2) <= 0 }
