package command

import "errors"

// Sentinel errors for cmd package.
var (
	ErrInvalidLanguage   = errors.New("invalid language")
	ErrStandardsNotFound = errors.New("standards directory not found")
)
