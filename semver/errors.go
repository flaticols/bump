package semver

import "errors"

var (
	ErrInvalidVersion = errors.New("invalid version")
	ErrEmptyVersion   = errors.New("version string empty")
	ErrMalformedCore  = errors.New("core version must be major.minor.patch")

	ErrLeadingZeros      = errors.New("version component cannot have leading zeros")
	ErrNonDigitComponent = errors.New("version component must contain only digits")
	ErrNegativeComponent = errors.New("version component must be non-negative")

	ErrEmptyIdentifier         = errors.New("identifier cannot be empty")
	ErrLeadingZeroesIdentifier = errors.New("numeric identifier cannot have leading zeros")
	ErrInvalidIdentifierChars  = errors.New("identifier can only contain alphanumeric characters and hyphens")

	ErrInvalidConstraint         = errors.New("invalid constraint")
	ErrUnknownConstraintOperator = errors.New("unknown constraint operator")

	ErrEmptyVersionComponent = errors.New("version component cannot be empty")
)
