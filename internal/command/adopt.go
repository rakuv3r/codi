// Package command provides CLI commands for codi.
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
	"github.com/rakuv3r/codi/internal/standard"
)

// NewAdoptCmd creates the adopt command.
func NewAdoptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "adopt",
		Short: "Integrate codi into existing project",
		Long: `Analyze existing project and generate AI tool files with coding standards.

This command:
1. Detects project language
2. Analyzes dependencies
3. Loads relevant coding standards
4. Generates AI tool files
5. Creates .codi/config.yaml configuration`,
		RunE: runAdopt,
	}
}

func runAdopt(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	_, _ = fmt.Fprintln(out, "🔍 Analyzing project...")

	// Detect and analyze
	lang, deps, err := detectAndAnalyze(out, projectPath)
	if err != nil {
		return err
	}

	// Build project info
	projectInfo := &analyzer.ProjectInfo{
		Name:        filepath.Base(projectPath),
		Path:        projectPath,
		Language:    lang,
		ProjectType: inferProjectType(deps),
		Description: fmt.Sprintf("A %s %s project", lang, inferProjectType(deps)),
	}

	// Set entry point based on project type
	if projectInfo.ProjectType == analyzer.ProjectTypeAPI {
		projectInfo.EntryPoint = "server"
	} else {
		projectInfo.EntryPoint = "."
	}

	// Extract module name for Go projects
	if lang == analyzer.LanguageGo {
		if modFile := filepath.Join(projectPath, "go.mod"); true {
			if data, err := os.ReadFile(modFile); err == nil {
				// Simple extraction: find "module xxx" line
				for _, line := range strings.Split(string(data), "\n") {
					if strings.HasPrefix(strings.TrimSpace(line), "module ") {
						projectInfo.Module = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "module "))
						break
					}
				}
			}
		}
	}

	_, _ = fmt.Fprintf(out, "  ✓ Project type: %s\n", projectInfo.ProjectType)

	// Generate files
	return generateProjectFiles(out, projectPath, projectInfo, deps)
}

func detectAndAnalyze(out io.Writer, projectPath string) (analyzer.Language, *analyzer.DependencyList, error) {
	// Detect language
	lang, err := analyzer.DetectLanguage(projectPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to detect language: %w", err)
	}
	_, _ = fmt.Fprintf(out, "  ✓ Language: %s\n", lang)

	// Analyze dependencies
	var deps *analyzer.DependencyList
	switch lang {
	case analyzer.LanguageGo:
		goAnalyzer := analyzer.NewGoAnalyzer()
		deps, err = goAnalyzer.Analyze(projectPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to analyze Go project: %w", err)
		}
	case analyzer.LanguagePython:
		pythonAnalyzer := analyzer.NewPythonAnalyzer()
		deps, err = pythonAnalyzer.Analyze(projectPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to analyze Python project: %w", err)
		}
	default:
		return "", nil, ErrInvalidLanguage
	}

	_, _ = fmt.Fprintf(out, "  ✓ Dependencies: %d total\n", len(deps.Dependencies))
	return lang, deps, nil
}

func generateProjectFiles(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList) error {
	codiRoot := getCodiRoot()

	// Load standards
	var standards any
	var err error
	if codiRoot == "" {
		// Use embedded standards
		standards, err = loadStandardsFromEmbed(out, projectPath, projectInfo, deps)
	} else {
		// Use disk standards
		standards, err = loadStandardsForAdopt(out, codiRoot, projectPath, projectInfo, deps)
	}
	if err != nil {
		return err
	}

	// Find templates directory (or use embed)
	var templatesDir string
	if codiRoot == "" {
		// Use embedded templates - pass special marker
		templatesDir = "" // Signal to use embed
	} else {
		templatesDir, err = findTemplatesDirForAdopt()
		if err != nil {
			return err
		}
	}

	// Generate all files
	_, _ = fmt.Fprintln(out, "\n📝 Generating CLAUDE.md...")
	if err := generateAdoptFiles(out, projectPath, projectInfo, deps, standards, templatesDir, codiRoot); err != nil {
		return err
	}

	printAdoptSuccessMessage(out)
	return nil
}

func loadStandardsFromEmbed(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList) (any, error) {
	// Create sub-FS for standards directory
	standardsFS, err := fs.Sub(codi.EmbeddedFS, "standards")
	if err != nil {
		return nil, fmt.Errorf("failed to access embedded standards: %w", err)
	}

	stdMgr := standard.NewManagerWithFS(standardsFS, projectPath)
	standards, err := stdMgr.LoadForProject(projectInfo, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to load standards: %w", err)
	}
	_, _ = fmt.Fprintf(out, "  ✓ Standards: %d loaded\n", standards.Count())
	return standards, nil
}

func loadStandardsForAdopt(out io.Writer, codiRoot, projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList) (any, error) {
	stdMgr := standard.NewManager(filepath.Join(codiRoot, "standards"), projectPath)
	standards, err := stdMgr.LoadForProject(projectInfo, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to load standards: %w", err)
	}
	_, _ = fmt.Fprintf(out, "  ✓ Standards: %d loaded\n", standards.Count())
	return standards, nil
}

func findTemplatesDirForAdopt() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	templatesDir := filepath.Join(exeDir, "..", "templates")

	// If not found, try current directory (development mode)
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		templatesDir = "templates"
	}
	return templatesDir, nil
}

func generateAdoptFiles(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList, standards any, templatesDir, codiRoot string) error {
	// Check and add missing quality tools
	_, _ = fmt.Fprintln(out, "\n🔧 Checking quality tools...")
	if err := addMissingQualityTools(out, projectPath, projectInfo, templatesDir); err != nil {
		return fmt.Errorf("failed to add quality tools: %w", err)
	}

	// Generate REFACTOR.md if needed
	if needsRefactoring(projectPath, projectInfo.Language) {
		_, _ = fmt.Fprintln(out, "\n📋 Generating REFACTOR.md...")
		if err := generateRefactorGuide(projectPath, projectInfo, templatesDir); err != nil {
			return fmt.Errorf("failed to generate REFACTOR.md: %w", err)
		}
		_, _ = fmt.Fprintln(out, "  ✓ REFACTOR.md created")
	}

	// Copy review templates
	copyReviewTemplates(out, codiRoot, projectPath)

	// Prompt for AI tools
	_, _ = fmt.Fprintln(out, "\n🤖 Select AI tools:")
	selectedTools := promptAITools()

	// Generate AI tool files
	_, _ = fmt.Fprintln(out, "\n📝 Generating AI tool files...")
	for _, tool := range selectedTools {
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

		_, _ = fmt.Fprintf(out, "  ✓ %s created\n", tool.OutputPath)
	}

	// Save config
	cfg := config.NewConfigWithTools(projectInfo, deps, selectedTools)
	if err := config.Save(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	_, _ = fmt.Fprintln(out, "  ✓ .codi/config.yaml created")

	return nil
}

func printAdoptSuccessMessage(out io.Writer) {
	_, _ = fmt.Fprintln(out, "\n✅ Project successfully integrated with codi!")
	_, _ = fmt.Fprintln(out, "\nNext steps:")
	_, _ = fmt.Fprintln(out, "  - Review AI tool files (.claude/CLAUDE.md, .codex/AGENTS.md)")
	_, _ = fmt.Fprintln(out, "  - Review .claude/review-checklist.md (AI automated checks)")
	_, _ = fmt.Fprintln(out, "  - Review .claude/review-manual.md (human review guide)")
	_, _ = fmt.Fprintln(out, "  - Run 'codi sync' after adding dependencies")
	_, _ = fmt.Fprintln(out, "  - Install Git hooks: ./scripts/install_hooks.sh")
}

func copyReviewTemplates(out io.Writer, codiRoot, projectPath string) {
	claudeDir := filepath.Join(projectPath, config.ClaudeDirName)

	// Create .claude directory if needed
	if err := os.MkdirAll(claudeDir, config.DirPermission); err != nil {
		return
	}

	var checklistData, manualData []byte
	var err error

	if codiRoot == "" {
		// Use embedded files
		checklistData, err = fs.ReadFile(codi.EmbeddedFS, filepath.Join("standards", config.ReviewChecklistFilename))
		if err == nil {
			checklistDest := filepath.Join(claudeDir, config.ReviewChecklistFilename)
			if writeErr := os.WriteFile(checklistDest, checklistData, config.ConfigPermission); writeErr == nil {
				_, _ = fmt.Fprintln(out, "  ✓ .claude/review-checklist.md created")
			}
		}

		manualData, err = fs.ReadFile(codi.EmbeddedFS, filepath.Join("standards", config.ReviewManualFilename))
		if err == nil {
			manualDest := filepath.Join(claudeDir, config.ReviewManualFilename)
			if writeErr := os.WriteFile(manualDest, manualData, config.ConfigPermission); writeErr == nil {
				_, _ = fmt.Fprintln(out, "  ✓ .claude/review-manual.md created")
			}
		}
	} else {
		// Use disk files
		checklistSrc := filepath.Join(codiRoot, config.StandardsDirName, config.ReviewChecklistFilename)
		checklistDest := filepath.Join(claudeDir, config.ReviewChecklistFilename)
		if data, err := os.ReadFile(checklistSrc); err == nil {
			if err := os.WriteFile(checklistDest, data, config.ConfigPermission); err == nil {
				_, _ = fmt.Fprintln(out, "  ✓ .claude/review-checklist.md created")
			}
		}

		manualSrc := filepath.Join(codiRoot, config.StandardsDirName, config.ReviewManualFilename)
		manualDest := filepath.Join(claudeDir, config.ReviewManualFilename)
		if data, err := os.ReadFile(manualSrc); err == nil {
			if err := os.WriteFile(manualDest, data, config.ConfigPermission); err == nil {
				_, _ = fmt.Fprintln(out, "  ✓ .claude/review-manual.md created")
			}
		}
	}
}

// inferProjectType infers project type from dependencies.
func inferProjectType(deps *analyzer.DependencyList) analyzer.ProjectType {
	// Check for HTTP frameworks
	for _, dep := range deps.Dependencies {
		switch dep.Path {
		case "github.com/gin-gonic/gin",
			"github.com/labstack/echo/v4",
			"github.com/gofiber/fiber/v2":
			return analyzer.ProjectTypeAPI
		case "github.com/spf13/cobra",
			"github.com/urfave/cli/v2":
			return analyzer.ProjectTypeCLI
		}
	}

	// Check for common CLI indicators in stdlib
	hasCobra := false
	for _, dep := range deps.Dependencies {
		if strings.Contains(dep.Path, "cobra") || strings.Contains(dep.Path, "cli") {
			hasCobra = true
			break
		}
	}
	if hasCobra {
		return analyzer.ProjectTypeCLI
	}

	// Check for HTTP/server indicators in stdlib
	hasHTTP := false
	for _, dep := range deps.Dependencies {
		if dep.Path == "net/http" {
			hasHTTP = true
			break
		}
	}
	if hasHTTP {
		return analyzer.ProjectTypeAPI
	}

	// Default: assume CLI for minimal projects
	// (most learning projects and small tools are CLIs)
	return analyzer.ProjectTypeCLI
}

// getCodiRoot returns the codi installation root directory.
// getCodiRoot returns the codi installation root directory.
// Returns empty string if not found (caller should use embed).
func getCodiRoot() string {
	// Get executable path
	exe, err := os.Executable()
	if err != nil {
		// If we can't get exe path, try current directory
		cwd, _ := os.Getwd()
		if _, statErr := os.Stat(filepath.Join(cwd, "standards")); statErr == nil {
			return cwd
		}
		// Return empty string to signal use of embed
		return ""
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "" // Use embed
	}

	// Assume: bin/codi is in codi root
	// e.g., /path/to/codi/bin/codi -> /path/to/codi
	root := filepath.Dir(filepath.Dir(exe))

	// Verify standards/ exists
	standardsPath := filepath.Join(root, "standards")
	if _, err := os.Stat(standardsPath); err != nil {
		// Fallback: try current working directory (for development)
		cwd, _ := os.Getwd()
		if _, statErr := os.Stat(filepath.Join(cwd, "standards")); statErr == nil {
			return cwd
		}
		// Return empty string to signal use of embed
		return ""
	}

	return root
}
