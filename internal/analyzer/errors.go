package analyzer

import "errors"

var (
	// ErrLanguageNotDetected is returned when language cannot be detected.
	ErrLanguageNotDetected = errors.New("language not detected")
)
