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
	prompt, err := generateReviewPrompt(projectPath)
	if err != nil {
		return err
	}

	if dryRun {
		// Print prompt and exit
		_, _ = fmt.Fprintln(out, prompt)
		return nil
	}

	// Execute AI tool
	_, _ = fmt.Fprintln(out, "🔍 Running AI code review...")
	return executeAITool(out, projectPath, prompt)
}

func generateReviewPrompt(projectPath string) (string, error) {
	var sb strings.Builder

	// Header - instruct /code-review to use our standards
	sb.WriteString("/code-review using the following standards:\n\n")

	// Load review checklist as primary criteria - REQUIRED
	checklistPath := filepath.Join(projectPath, ".codi", "review-checklist.md")
	checklist, err := os.ReadFile(checklistPath)
	if err != nil {
		return "", fmt.Errorf("review checklist not found: %s (run 'codi sync' or 'codi adopt')", checklistPath)
	}
	sb.WriteString("## Review Checklist (Check All Items)\n\n")
	sb.Write(checklist)
	sb.WriteString("\n\n")

	// Load standards reference - REQUIRED
	claudePath := filepath.Join(projectPath, ".claude", "CLAUDE.md")
	claudeContent, err := os.ReadFile(claudePath)
	if err != nil {
		return "", fmt.Errorf("CLAUDE.md not found: %s (run 'codi sync' or 'codi adopt')", claudePath)
	}

	// Extract standards section - search from end to avoid false matches in examples
	content := string(claudeContent)
	marker := "## 📚 Coding Standards"
	idx := strings.LastIndex(content, marker)
	if idx == -1 {
		return "", errors.New("coding standards section not found in CLAUDE.md (missing '## 📚 Coding Standards' header)")
	}
	sb.WriteString("## Coding Standards (Reference When Fixing)\n\n")
	sb.WriteString(content[idx:])
	sb.WriteString("\n")

	sb.WriteString("\n## Review Requirements\n\n")
	sb.WriteString("1. Check all checklist items\n")
	sb.WriteString("2. Fix issues according to standards\n")
	sb.WriteString("3. Run tests and lint to verify fixes\n")
	sb.WriteString("4. Commit fixes with clear messages\n")

	return sb.String(), nil
}

func executeAITool(out io.Writer, projectPath string, prompt string) error {
	// Detect which AI tool to use
	var cmd *exec.Cmd

	claudePath := filepath.Join(projectPath, ".claude", "CLAUDE.md")
	codexPath := filepath.Join(projectPath, ".codex", "AGENTS.md")

	if _, err := os.Stat(claudePath); err == nil {
		// Use Claude Code - -p flag for non-interactive, prompt as positional arg
		_, _ = fmt.Fprintln(out, "  Using: Claude Code")
		cmd = exec.Command("claude", "-p", "--", prompt)
	} else if _, err := os.Stat(codexPath); err == nil {
		// Use Codex - -p flag for non-interactive, prompt as positional arg
		_, _ = fmt.Fprintln(out, "  Using: Codex")
		cmd = exec.Command("codex", "-p", "--", prompt)
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
