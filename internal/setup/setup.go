package setup

import (
	"fmt"
)

// Project orchestrates all project setup steps based on options.
// This is the main entry point for init, adopt, and sync commands.
func Project(opts Options) error {
	// Phase 1: Load standards
	standards, codiRoot, err := LoadStandards(opts.ProjectPath, opts.ProjectInfo, opts.Dependencies)
	if err != nil {
		return fmt.Errorf("failed to load standards: %w", err)
	}

	// Phase 2: Copy standards to .codi/standards/
	if opts.CopyStandards {
		if err := CopyStandards(opts.ProjectPath, standards, codiRoot); err != nil {
			return fmt.Errorf("failed to copy standards: %w", err)
		}
	}

	// Phase 3: Generate AI tool files
	if err := GenerateAITools(opts.ProjectPath, opts.ProjectInfo, opts.AITools, standards, opts.Output); err != nil {
		return fmt.Errorf("failed to generate AI tool files: %w", err)
	}

	// Phase 4: Copy review templates
	if opts.CopyReviewTemplate {
		if err := CopyReviewTemplate(opts.Output, codiRoot, opts.ProjectPath); err != nil {
			return fmt.Errorf("failed to copy review templates: %w", err)
		}
	}

	// Phase 5: Add quality tools
	if opts.AddQualityTools {
		if err := InstallTools(opts.Output, opts.ProjectPath, opts.ProjectInfo); err != nil {
			return fmt.Errorf("failed to add quality tools: %w", err)
		}
	}

	// Phase 6: Install git hooks
	if opts.InstallGitHooks {
		if err := InstallHooks(opts.Output, opts.ProjectPath, opts.PromptForSetup); err != nil {
			return fmt.Errorf("failed to install git hooks: %w", err)
		}
	}

	return nil
}
