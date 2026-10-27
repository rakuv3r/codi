package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
)

// needsRefactoring checks if project needs refactoring based on language.
func needsRefactoring(projectPath string, lang analyzer.Language) bool {
	switch lang {
	case analyzer.LanguageGo:
		// Check for missing Go modules
		if _, err := os.Stat(filepath.Join(projectPath, "go.mod")); os.IsNotExist(err) {
			return true
		}

		// Check for vendor directory (old dependency management)
		if _, err := os.Stat(filepath.Join(projectPath, "vendor")); !os.IsNotExist(err) {
			// vendor exists, might need refactoring to use go modules properly
			return true
		}

	case analyzer.LanguagePython:
		// Check for setup.py without pyproject.toml (old style)
		hasSetupPy := false
		if _, err := os.Stat(filepath.Join(projectPath, "setup.py")); !os.IsNotExist(err) {
			hasSetupPy = true
		}

		hasPyproject := false
		if _, err := os.Stat(filepath.Join(projectPath, "pyproject.toml")); !os.IsNotExist(err) {
			hasPyproject = true
		}

		// If has setup.py but no pyproject.toml, needs refactoring
		if hasSetupPy && !hasPyproject {
			return true
		}
	}

	return false
}

// generateRefactorGuide generates REFACTOR.md.
func generateRefactorGuide(projectPath string, projectInfo *analyzer.ProjectInfo, templatesDir string) error {
	refactorDir := filepath.Join(templatesDir, string(projectInfo.Language), "refactor")
	refactorSrc := filepath.Join(refactorDir, "REFACTOR.md")
	refactorDest := filepath.Join(projectPath, "REFACTOR.md")

	content, err := os.ReadFile(refactorSrc)
	if err != nil {
		return fmt.Errorf("failed to read REFACTOR.md template: %w", err)
	}

	if err := os.WriteFile(refactorDest, content, config.FilePermission); err != nil {
		return fmt.Errorf("failed to write REFACTOR.md: %w", err)
	}

	return nil
}
