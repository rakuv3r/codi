package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/standard"
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
	return &analyzer.ProjectInfo{
		Name:           filepath.Base(projectPath),
		Path:           projectPath,
		Language:       cfg.Language,
		ProjectType:    inferProjectType(deps), // Re-infer based on new deps
		HTTPFramework:  cfg.Project.HTTPFramework,
		Database:       cfg.Project.Database,
		DatabaseDriver: cfg.Project.DatabaseDriver,
		Features:       cfg.Project.Features,
	}
}

func regenerateProject(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList, cfg *config.Config) error {
	// Load standards
	codiRoot, err := getCodiRoot()
	if err != nil {
		return fmt.Errorf("failed to locate codi root: %w", err)
	}

	stdMgr := standard.NewManager(filepath.Join(codiRoot, "standards"), projectPath)
	standards, err := stdMgr.LoadForProject(projectInfo, deps)
	if err != nil {
		return fmt.Errorf("failed to load standards: %w", err)
	}

	_, _ = fmt.Fprintf(out, "  ✓ Standards: %d loaded\n", standards.Count())
	_, _ = fmt.Fprintln(out, "\n📝 Updating AI tool files...")

	// Generate AI tool files based on config
	for _, tool := range cfg.AITools {
		if !tool.Enabled {
			continue
		}

		// Create parent directory
		toolDir := filepath.Dir(filepath.Join(projectPath, tool.OutputPath))
		if err := os.MkdirAll(toolDir, config.DirPermission); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", tool.Name, err)
		}

		// Generate file
		toolPath := filepath.Join(projectPath, tool.OutputPath)
		if err := generateAIToolFile(toolPath, projectInfo, deps, standards); err != nil {
			return fmt.Errorf("failed to generate %s: %w", tool.OutputPath, err)
		}

		_, _ = fmt.Fprintf(out, "  ✓ %s updated\n", tool.OutputPath)
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

func generateAIToolFile(path string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList, standards any) error {
	content := fmt.Sprintf("# %s\n\n", projectInfo.Name)
	content += "## Project Information\n\n"
	content += fmt.Sprintf("- **Language**: %s\n", projectInfo.Language)
	content += fmt.Sprintf("- **Type**: %s\n", projectInfo.ProjectType)
	if projectInfo.HTTPFramework != "" {
		content += fmt.Sprintf("- **HTTP Framework**: %s\n", projectInfo.HTTPFramework)
	}
	if projectInfo.Database != "" {
		content += fmt.Sprintf("- **Database**: %s\n", projectInfo.Database)
	}
	content += "\n## Dependencies\n\n"
	content += fmt.Sprintf("Total: %d\n\n", len(deps.Dependencies))
	var contentSb strings.Builder
	for _, dep := range deps.Dependencies {
		fmt.Fprintf(&contentSb, "- %s\n", dep.Path)
	}
	content += contentSb.String()
	content += "\n---\n\n"
	content += "## Coding Standards\n\n"
	content += "This project follows the standards below. All code generation and modifications should adhere to these standards.\n\n"

	// Add standards references
	if standards != nil {
		content += "See codi standards directory for detailed coding standards.\n"
	}

	return os.WriteFile(path, []byte(content), config.FilePermission)
}
