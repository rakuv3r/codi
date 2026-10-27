package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi/internal/config"
)

// NewStatusCmd creates the status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		Long:  `Display current project configuration and standards status.`,
		RunE:  runStatus,
	}

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	projectPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config
	cfg, err := config.Load(projectPath)
	if err != nil {
		_, _ = fmt.Fprintln(out, "❌ Project not integrated with codi")
		_, _ = fmt.Fprintln(out, "\nRun 'codi adopt' to get started")
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Display status
	_, _ = fmt.Fprintln(out, "📊 Project Status")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "Language:     %s\n", cfg.Language)
	_, _ = fmt.Fprintf(out, "Project Type: %s\n", cfg.Project.Type)
	if cfg.Project.HTTPFramework != "" {
		_, _ = fmt.Fprintf(out, "HTTP:         %s\n", cfg.Project.HTTPFramework)
	}
	if cfg.Project.Database != "" {
		_, _ = fmt.Fprintf(out, "Database:     %s\n", cfg.Project.Database)
	}
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Dependencies:")
	_, _ = fmt.Fprintf(out, "  Total:      %d\n", len(cfg.Dependencies.Dependencies))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "AI Tools:")
	for _, tool := range cfg.AITools {
		status := "✓"
		if !tool.Enabled {
			status = "✗"
		}
		_, _ = fmt.Fprintf(out, "  %s %s (%s)\n", status, tool.Name, tool.OutputPath)
	}
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "Last Sync:    %s\n", cfg.LastSync.Format("2006-01-02 15:04:05"))
	_, _ = fmt.Fprintln(out)

	// Check AI tool files
	_, _ = fmt.Fprintln(out, "AI Tool Files:")
	for _, tool := range cfg.AITools {
		if !tool.Enabled {
			continue
		}

		toolPath := filepath.Join(projectPath, tool.OutputPath)
		if info, err := os.Stat(toolPath); err == nil {
			_, _ = fmt.Fprintf(out, "  %s: %d lines\n", tool.OutputPath, countLines(toolPath))
			_, _ = fmt.Fprintf(out, "              Updated %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		} else {
			_, _ = fmt.Fprintf(out, "  %s: missing\n", tool.OutputPath)
		}
	}

	return nil
}

func countLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	return lines
}
