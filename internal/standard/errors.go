// Package standard manages coding standards.
package standard

import "errors"

var (
	// ErrStandardNotFound is returned when standard file not found.
	ErrStandardNotFound = errors.New("standard not found")
)
