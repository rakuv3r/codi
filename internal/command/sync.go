package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/setup"
)

// NewSyncCmd creates the sync command.
func NewSyncCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync AI tool files after dependency changes",
		Long: `Re-analyze dependencies and update AI tool files.

Run this after:
- Adding/removing dependencies (go get, pip install)
- Changing project structure
- Updating codi standards`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd, args, force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force regenerate all AI tool files")

	return cmd
}

func runSync(cmd *cobra.Command, _ []string, force bool) error {
	out := cmd.OutOrStdout()

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	_, _ = fmt.Fprintln(out, "🔄 Syncing dependencies...")

	// Load existing config
	cfg, err := config.Load(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load config (run 'codi adopt' first): %w", err)
	}

	// Re-analyze dependencies
	deps, err := analyzeDependencies(cfg.Language, projectPath)
	if err != nil {
		return err
	}

	// Check for changes or force regeneration
	needsRegenerate := force || dependenciesChanged(out, cfg, deps) || aiToolFilesMissing(projectPath, cfg)

	if !needsRegenerate {
		_, _ = fmt.Fprintln(out, "  ✓ No changes detected")
		return nil
	}

	// Rebuild and regenerate
	projectInfo := rebuildProjectInfo(cfg, projectPath, deps)
	return regenerateProject(out, projectPath, projectInfo, deps, cfg)
}

func analyzeDependencies(lang analyzer.Language, projectPath string) (*analyzer.DependencyList, error) {
	switch lang {
	case analyzer.LanguageGo:
		goAnalyzer := analyzer.NewGoAnalyzer()
		deps, err := goAnalyzer.Analyze(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze project: %w", err)
		}
		return deps, nil
	case analyzer.LanguagePython:
		pythonAnalyzer := analyzer.NewPythonAnalyzer()
		deps, err := pythonAnalyzer.Analyze(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze project: %w", err)
		}
		return deps, nil
	default:
		return nil, ErrInvalidLanguage
	}
}

func dependenciesChanged(out io.Writer, cfg *config.Config, deps *analyzer.DependencyList) bool {
	oldDeps := len(cfg.Dependencies.Dependencies)
	newDeps := len(deps.Dependencies)

	if oldDeps == newDeps {
		return false
	}

	_, _ = fmt.Fprintf(out, "  Changes: dependencies %d→%d\n", oldDeps, newDeps)
	return true
}

// aiToolFilesMissing checks if any enabled AI tool files are missing.
func aiToolFilesMissing(projectPath string, cfg *config.Config) bool {
	for _, tool := range cfg.AITools {
		if !tool.Enabled {
			continue
		}

		toolPath := filepath.Join(projectPath, tool.OutputPath)
		if _, err := os.Stat(toolPath); os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func rebuildProjectInfo(cfg *config.Config, projectPath string, deps *analyzer.DependencyList) *analyzer.ProjectInfo {
	return setup.RebuildProjectInfo(cfg, projectPath, deps)
}

func regenerateProject(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList, cfg *config.Config) error {
	// Setup project using setup package (sync mode: only refresh standards and AI files)
	err := setup.Project(setup.Options{
		ProjectPath:        projectPath,
		ProjectInfo:        projectInfo,
		Dependencies:       deps,
		AITools:            cfg.AITools,
		CopyStandards:      true,  // Refresh standards
		CopyReviewTemplate: false, // Don't overwrite
		AddQualityTools:    false, // Don't modify
		InstallGitHooks:    false, // Don't reinstall
		PromptForSetup:     false, // Don't prompt
		Output:             out,
	})
	if err != nil {
		return fmt.Errorf("failed to sync project: %w", err)
	}

	// Update config (preserve selected AI tools)
	newCfg := config.NewConfigWithTools(projectInfo, deps, cfg.AITools)
	if err := config.Save(projectPath, newCfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	_, _ = fmt.Fprintln(out, "  ✓ .codi/config.yaml updated")
	_, _ = fmt.Fprintln(out, "\n✅ Sync complete!")

	return nil
}
