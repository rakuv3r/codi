// Package analyzer provides project analysis functionality including language detection and dependency extraction.
package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// languageDetector defines how to detect a language.
type languageDetector struct {
	Language Language
	Files    []string // Files to look for
}

var detectors = []languageDetector{
	{
		Language: LanguageGo,
		Files:    []string{"go.mod"},
	},
	{
		Language: LanguagePython,
		Files:    []string{"pyproject.toml"},
	},
}

// DetectLanguage detects programming language from project directory.
func DetectLanguage(projectPath string) (Language, error) {
	for _, detector := range detectors {
		for _, file := range detector.Files {
			if fileExists(filepath.Join(projectPath, file)) {
				return detector.Language, nil
			}
		}
	}

	return "", newLanguageNotDetectedError()
}

// newLanguageNotDetectedError creates a detailed error message.
func newLanguageNotDetectedError() error {
	var sb strings.Builder
	sb.WriteString("no supported project detected\n\nLooking for:\n")
	for _, detector := range detectors {
		// Capitalize first letter
		langName := string(detector.Language)
		if len(langName) > 0 {
			langName = strings.ToUpper(langName[:1]) + langName[1:]
		}
		fmt.Fprintf(&sb, "  - %s: %s\n", langName, strings.Join(detector.Files, ", "))
	}
	sb.WriteString("\nIs this the correct directory?")
	return fmt.Errorf("%w: %s", ErrLanguageNotDetected, sb.String())
}

// fileExists checks if file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
