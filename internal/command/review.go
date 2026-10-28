package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewReviewCmd creates the review command.
func NewReviewCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Auto-review code and fix issues",
		Long: `Automatically review code using AI and fix issues based on standards.

This command:
1. Reads .codi/review-checklist.md
2. Loads applicable coding standards
3. Generates a review prompt
4. Executes AI tool (Claude Code or Codex) to review and fix code

Examples:
  codi review              # Review and auto-fix issues
  codi review --dry-run    # Generate prompt without executing`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReview(cmd.OutOrStdout(), dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Generate prompt without executing")

	return cmd
}

func runReview(out io.Writer, dryRun bool) error {
	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Generate review prompt
	prompt := generateReviewPrompt(projectPath)

	if dryRun {
		// Print prompt and exit
		_, _ = fmt.Fprintln(out, prompt)
		return nil
	}

	// Execute AI tool
	_, _ = fmt.Fprintln(out, "🔍 Running AI code review...")
	return executeAITool(out, projectPath, prompt)
}

func generateReviewPrompt(projectPath string) string {
	var sb strings.Builder

	// Header - instruct to use /code-review
	sb.WriteString("# Code Review Task\n\n")
	sb.WriteString("Please run the `/code-review` command to review this codebase.\n\n")
	sb.WriteString("After the review completes:\n")
	sb.WriteString("1. Read the findings\n")
	sb.WriteString("2. Fix the issues according to the coding standards below\n")
	sb.WriteString("3. Commit the fixes\n\n")

	// Load review checklist as additional context
	checklistPath := filepath.Join(projectPath, ".codi", "review-checklist.md")
	checklist, err := os.ReadFile(checklistPath)
	if err == nil {
		sb.WriteString("## Additional Review Context\n\n")
		sb.WriteString("Use this checklist as supplementary guidance:\n\n")
		sb.Write(checklist)
		sb.WriteString("\n\n")
	}

	// Load standards reference
	claudePath := filepath.Join(projectPath, ".claude", "CLAUDE.md")
	if claudeContent, err := os.ReadFile(claudePath); err == nil {
		// Extract standards section
		if idx := strings.Index(string(claudeContent), "## 📚 Coding Standards"); idx != -1 {
			sb.WriteString("## Coding Standards Reference\n\n")
			sb.WriteString("When fixing issues, reference these standards:\n\n")
			sb.Write(claudeContent[idx:])
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func executeAITool(out io.Writer, projectPath string, prompt string) error {
	// Detect which AI tool to use
	var cmd *exec.Cmd

	claudePath := filepath.Join(projectPath, ".claude", "CLAUDE.md")
	codexPath := filepath.Join(projectPath, ".codex", "AGENTS.md")

	if _, err := os.Stat(claudePath); err == nil {
		// Use Claude Code
		_, _ = fmt.Fprintln(out, "  Using: Claude Code")
		cmd = exec.Command("claude", prompt)
	} else if _, err := os.Stat(codexPath); err == nil {
		// Use Codex
		_, _ = fmt.Fprintln(out, "  Using: Codex")
		cmd = exec.Command("codex", prompt)
	} else {
		return errors.New("no AI tool detected (missing .claude/CLAUDE.md or .codex/AGENTS.md)")
	}

	cmd.Dir = projectPath
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AI tool execution failed: %w", err)
	}

	_, _ = fmt.Fprintln(out, "\n✅ Review complete!")
	return nil
}
