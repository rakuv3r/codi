package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// InstallHooks installs git hooks to .git/hooks/ directory.
// If prompt is true, asks user for confirmation before running make setup.
func InstallHooks(out io.Writer, projectPath string, prompt bool) error {
	if !prompt {
		// No interactive prompt, just inform user
		_, _ = fmt.Fprintln(out, "\n💡 Run './scripts/install_hooks.sh' to install git hooks")
		return nil
	}

	// Prompt user to run make setup
	_, _ = fmt.Fprintln(out, "\nRun 'make setup' now to install development tools and Git hooks? [y/n] ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		// Non-fatal, just skip the prompt
		_, _ = fmt.Fprintln(out, "\n💡 Skipping make setup. Run 'make setup' manually to install tools and hooks.")
		return nil //nolint:nilerr
	}

	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		_, _ = fmt.Fprintln(out, "\nRunning make setup...")

		cmd := exec.Command("make", "setup")
		cmd.Dir = projectPath
		cmd.Stdout = out
		cmd.Stderr = out

		if err := cmd.Run(); err != nil {
			_, _ = fmt.Fprintf(out, "⚠️  make setup failed: %v\n", err)
			_, _ = fmt.Fprintln(out, "   You can run 'make setup' manually later")
			return nil // Non-fatal
		}

		_, _ = fmt.Fprintln(out, "✅ Setup complete!")
	}

	return nil
}
