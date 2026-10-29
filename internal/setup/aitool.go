package setup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/generator"
	"github.com/rakuv3r/codi/internal/standard"
)

// GenerateAITools generates AI tool configuration files (.claude/CLAUDE.md, etc).
func GenerateAITools(projectPath string, projectInfo *analyzer.ProjectInfo, aiTools []generator.AIToolConfig, standards *standard.List, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "\n📝 Updating AI tool files...")

	// Generate AI tool files based on config
	for _, tool := range aiTools {
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
		if err := generateAIToolFile(toolPath, projectInfo, standards); err != nil {
			return fmt.Errorf("failed to generate %s: %w", tool.OutputPath, err)
		}

		_, _ = fmt.Fprintf(out, "  ✓ %s updated\n", tool.OutputPath)
	}

	return nil
}

func generateAIToolFile(path string, projectInfo *analyzer.ProjectInfo, standards *standard.List) error {
	content := buildAIToolContent(projectInfo, standards)
	return os.WriteFile(path, []byte(content), config.FilePermission)
}

func buildAIToolContent(projectInfo *analyzer.ProjectInfo, standards *standard.List) string {
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
	if standards != nil {
		content += "### Quick Router\n\n"

		// Universal standards (applies to all projects)
		if len(standards.Universal) > 0 {
			content += "**Universal (All Projects):**\n"
			var sb strings.Builder
			for _, std := range standards.Universal {
				if shouldIncludeStandard(&std, projectInfo.ProjectType) {
					fmt.Fprintf(&sb, "- %s → `.codi/standards/%s`\n", formatStandardName(std.Name), std.Path)
				}
			}
			content += sb.String()
			content += "\n"
		}

		// Core standards (language-specific, filtered by project type)
		if len(standards.Core) > 0 {
			content += "**Core Standards:**\n"
			var sb strings.Builder
			for _, std := range standards.Core {
				if shouldIncludeStandard(&std, projectInfo.ProjectType) {
					fmt.Fprintf(&sb, "- %s → `.codi/standards/%s`\n", formatStandardName(std.Name), std.Path)
				}
			}
			content += sb.String()
			content += "\n"
		}

		// Dependency standards (filtered by project type)
		if len(standards.Dependencies) > 0 {
			content += "**Dependencies:**\n"
			var sb strings.Builder
			for _, std := range standards.Dependencies {
				if shouldIncludeStandard(&std, projectInfo.ProjectType) {
					fmt.Fprintf(&sb, "- %s → `.codi/standards/%s`\n", std.Dependency.Path, std.Path)
				}
			}
			content += sb.String()
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
