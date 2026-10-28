package command

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi"
	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/setup"
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
	// Load standards using setup package
	standards, codiRoot, err := setup.LoadStandards(projectPath, projectInfo, deps)
	if err != nil {
		return fmt.Errorf("failed to load standards: %w", err)
	}

	_, _ = fmt.Fprintf(out, "  ✓ Standards: %d loaded\n", standards.Count())

	// Copy standards to .codi/standards/ using setup package
	if err := setup.CopyStandardsToProject(projectPath, standards, codiRoot); err != nil {
		return fmt.Errorf("failed to copy standards: %w", err)
	}

	// Generate AI tool files using setup package
	if err := setup.GenerateAIToolFiles(projectPath, projectInfo, cfg.AITools, standards, out); err != nil {
		return fmt.Errorf("failed to generate AI tool files: %w", err)
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

func generateAIToolFile(path string, projectInfo *analyzer.ProjectInfo, standards any) error {
	content := buildAIToolContent(projectInfo, standards)
	return os.WriteFile(path, []byte(content), config.FilePermission)
}

func buildAIToolContent(projectInfo *analyzer.ProjectInfo, standards any) string {
	// Convert standards to typed list
	stdList, _ := standards.(*standard.List)

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

	content += "\n---\n\n"
	content += "## 📚 Coding Standards (READ BEFORE CODING)\n\n"
	content += "**AI: Before generating ANY code, read the relevant standard file!**\n\n"

	// Add standards router
	if stdList != nil {
		content += "### Quick Router\n\n"

		// Universal standards (applies to all projects)
		if len(stdList.Universal) > 0 {
			content += "**Universal (All Projects):**\n"
			var contentSb233 strings.Builder
			for _, std := range stdList.Universal {
				if shouldIncludeStandard(&std, projectInfo.ProjectType) {
					fmt.Fprintf(&contentSb233, "- %s → `.codi/standards/%s`\n", formatStandardName(std.Name), std.Path)
				}
			}
			content += contentSb233.String()
			content += "\n"
		}

		// Core standards (language-specific, filtered by project type)
		if len(stdList.Core) > 0 {
			content += "**Core Standards:**\n"
			var contentSb244 strings.Builder
			for _, std := range stdList.Core {
				if shouldIncludeStandard(&std, projectInfo.ProjectType) {
					fmt.Fprintf(&contentSb244, "- %s → `.codi/standards/%s`\n", formatStandardName(std.Name), std.Path)
				}
			}
			content += contentSb244.String()
			content += "\n"
		}

		// Dependency standards (filtered by project type)
		if len(stdList.Dependencies) > 0 {
			content += "**Dependencies:**\n"
			var contentSb255 strings.Builder
			for _, std := range stdList.Dependencies {
				if shouldIncludeStandard(&std, projectInfo.ProjectType) {
					fmt.Fprintf(&contentSb255, "- %s → `.codi/standards/%s`\n", std.Dependency.Path, std.Path)
				}
			}
			content += contentSb255.String()
		}
	}

	return content
}

// formatStandardName converts kebab-case to readable format
func formatStandardName(name string) string {
	words := strings.Split(name, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// shouldIncludeStandard checks if a standard applies to the project type.
func shouldIncludeStandard(std *standard.Standard, projectType analyzer.ProjectType) bool {
	// Empty AppliesTo means applies to all projects
	if len(std.AppliesTo) == 0 {
		return true
	}

	// Check if project type is in AppliesTo list
	projectTypeStr := string(projectType)
	for _, t := range std.AppliesTo {
		if t == projectTypeStr {
			return true
		}
	}

	return false
}

// copyStandardsToProject copies standard files to .codi/standards/.
func copyStandardsToProject(projectPath string, standards *standard.List, codiRoot string) error {
	destBase := filepath.Join(projectPath, ".codi", "standards")

	// Get source filesystem
	var srcFS fs.FS
	if codiRoot == "" {
		// Use embedded standards
		var err error
		srcFS, err = fs.Sub(codi.EmbeddedFS, "standards")
		if err != nil {
			return fmt.Errorf("failed to access embedded standards: %w", err)
		}
	} else {
		// Use disk standards
		srcFS = os.DirFS(filepath.Join(codiRoot, "standards"))
	}

	// Copy all standards
	for _, std := range standards.AllStandards() {
		srcPath := std.Path
		destPath := filepath.Join(destBase, std.Path)

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), config.DirPermission); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", std.Path, err)
		}

		// Read source
		content, err := fs.ReadFile(srcFS, srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", srcPath, err)
		}

		// Write destination
		if err := os.WriteFile(destPath, content, config.FilePermission); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}
	}

	return nil
}
