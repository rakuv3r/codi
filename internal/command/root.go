package command

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codi",
		Short: "Dependency-driven coding standards automation for AI",
		Long: `codi automatically generates and syncs coding standards based on project dependencies.

It detects your tech stack and creates AI tool configuration files with relevant standards,
making AI coding assistants follow your project's conventions automatically.`,
	}

	// Add subcommands
	cmd.AddCommand(NewAdoptCmd())
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewSyncCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewViewCmd())

	return cmd
}

// Execute runs the root command.
func Execute() error {
	return NewRootCmd().Execute()
}
