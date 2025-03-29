# SemVer Package

A Go package for working with Semantic Versioning 2.0.0 (SemVer) specifications.

## Features

- Full compliance with SemVer 2.0.0 specification
- High-performance manual parsing (no regex)
- Centralized error definitions
- Immutable Version type with value receivers
- Increment patch, minor, and major versions
- Set/Get prerelease and metadata fields
- Set/Get metadata as map[string]string
- Constraint matching with multiple operators

## Key Types

### Version

The core type representing a semantic version:

```go
type Version struct {
    Major      int
    Minor      int
    Patch      int
    Prerelease []string
    Metadata   []string
}
```

### Constraint and ConstraintSet

Types for version constraints:

```go
type Constraint struct {
    Operator Operator
    Version  Version
}

type ConstraintSet []Constraint
```

## Usage Examples

### Parsing

```go
v, err := Parse("1.2.3-beta.1+build.123")
if err != nil {
    // handle error
}
```

### Creating

```go
v := New(1, 2, 3, []string{"beta", "1"}, []string{"build", "123"})
```

### Comparing

```go
if v1.LessThan(v2) {
    // v1 is less than v2
}

// Or using the Compare function
result := Compare(v1, v2) // -1, 0, or 1
```

### Incrementing

```go
v2 := v.IncrementMajor() // Returns a new Version object
v3 := v.IncrementMinor()
v4 := v.IncrementPatch()
```

### Modifying (Immutable)

All modification operations return a new Version:

```go
// Setting prerelease 
v2 := SetPrerelease(v, []string{"rc", "1"})

// Setting metadata
v3 := SetMetadata(v, []string{"commit", "abc123"})

// Using map convenience functions
prereleaseMap := map[string]string{"feature": "x", "build": "123"}
v4 := SetPrereleaseMap(v, prereleaseMap)

metadataMap := map[string]string{"commit": "abc123", "timestamp": "123456"}
v5 := SetMetadataMap(v, metadataMap)
```

### Working with Constraints

```go
// Parse constraint
c, err := ParseConstraint(">= 1.0.0")
if err != nil {
    // handle error
}

// Check if version meets constraint
if c.Check(v) {
    // v meets the constraint
}

// Parse constraint set (multiple constraints)
cs, err := ParseConstraintSet(">= 1.0.0, < 2.0.0")
if err != nil {
    // handle error
}

// Check if version meets all constraints
if cs.Check(v) {
    // v meets all constraints
}
```

### Supported Constraint Operators

- Equal (`=`)
- Not Equal (`!=`)
- Greater Than (`>`)
- Less Than (`<`)
- Greater Than or Equal (`>=`)
- Less Than or Equal (`<=`)
- Tilde (`~`) - Allows patch level changes if minor version is specified, minor level changes if only major is specified
- Caret (`^`) - Allows changes that do not modify the left-most non-zero digit

## Implementation Notes

- The Version type is immutable - all modification operations return a new Version
- All Version methods use value receivers for efficiency and thread safety
- For performance reasons, the package uses manual parsing instead of regular expressions
- Builds metadata does not affect precedence (as per SemVer spec)
