// Package command provides CLI commands for codi.
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
	"github.com/rakuv3r/codi/internal/setup"
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
	// Prompt for AI tools
	selectedTools := promptAITools()

	// Setup project using setup package
	_, _ = fmt.Fprintln(out, "\n📝 Setting up project...")
	err := setup.Project(setup.Options{
		ProjectPath:        projectPath,
		ProjectInfo:        projectInfo,
		Dependencies:       deps,
		AITools:            selectedTools,
		CopyStandards:      true,
		CopyReviewTemplate: true,
		AddQualityTools:    true,
		InstallGitHooks:    true,
		PromptForSetup:     true,
		Output:             out,
	})
	if err != nil {
		return fmt.Errorf("failed to setup project: %w", err)
	}

	// Save config with selected AI tools
	cfg := config.NewConfigWithTools(projectInfo, deps, selectedTools)
	if err := config.Save(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	_, _ = fmt.Fprintln(out, "  ✓ .codi/config.yaml updated")
	printAdoptSuccessMessage(out)
	return nil
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

func printAdoptSuccessMessage(out io.Writer) {
	_, _ = fmt.Fprintln(out, "\n✅ Project successfully integrated with codi!")
	_, _ = fmt.Fprintln(out, "\nNext steps:")
	_, _ = fmt.Fprintln(out, "  - Review AI tool files (.claude/CLAUDE.md, .codex/AGENTS.md)")
	_, _ = fmt.Fprintln(out, "  - Review .codi/review-checklist.md (AI automated checks)")
	_, _ = fmt.Fprintln(out, "  - Run 'codi sync' after adding dependencies")
	_, _ = fmt.Fprintln(out, "  - Install Git hooks: ./scripts/install_hooks.sh")
}
