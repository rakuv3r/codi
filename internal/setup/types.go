// Package setup provides shared project setup logic for init, adopt, and sync commands.
package setup

import (
	"io"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/generator"
)

// SetupOptions configures project setup behavior.
type SetupOptions struct {
	// Required fields
	ProjectPath  string
	ProjectInfo  *analyzer.ProjectInfo
	Dependencies *analyzer.DependencyList
	AITools      []generator.AIToolConfig

	// Behavior flags - control what setup steps to perform
	CopyStandards      bool // Copy standards to .codi/standards/
	CopyReviewTemplate bool // Copy review-checklist.md to .codi/
	AddQualityTools    bool // Add missing quality tool files (Makefile, linters)
	InstallGitHooks    bool // Install git hooks to .git/hooks/
	PromptForSetup     bool // Prompt user to run make setup

	// Output writer for progress messages
	Output io.Writer
}
