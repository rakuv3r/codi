package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi/internal/generator"
)

const (
	projectTypeAPI    = "api"
	projectTypeCLI    = "cli"
	languageGo        = "go"
	languagePython    = "python"
	languageUniversal = "universal"
	templatesPath     = "templates"
)

// NewInitCmd creates the init command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new project",
		Long: `Create a new project with the specified structure and modules.

Use "." to initialize in the current directory (for existing git repos).
Use a project name to create a new directory.`,
		Args: cobra.ExactArgs(1),
		RunE: runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	projectName, inPlace := determineProjectName(out, args[0])

	// Check if directory is empty when using "."
	if inPlace {
		if err := checkDirectoryEmpty("."); err != nil {
			return err
		}
	}

	cfg := collectProjectConfig(out, projectName, inPlace)

	if !confirmConfiguration(out, cfg) {
		_, _ = fmt.Fprintln(out, "Canceled.")
		return nil
	}

	return generateProject(out, cfg, projectName)
}

// checkDirectoryEmpty checks if directory is empty or safe to initialize.
func checkDirectoryEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Ignore hidden files and directories
	var visibleFiles []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files (starting with .)
		if strings.HasPrefix(name, ".") {
			continue
		}
		visibleFiles = append(visibleFiles, name)
	}

	// If there are visible files, warn user
	if len(visibleFiles) > 0 {
		return fmt.Errorf("directory is not empty (found: %s)\n\nUse a new directory or remove these files first", strings.Join(visibleFiles, ", "))
	}

	return nil
}

func determineProjectName(out io.Writer, arg string) (string, bool) {
	inPlace := arg == "."
	projectName := arg

	if inPlace {
		cwd, err := os.Getwd()
		if err != nil {
			return arg, inPlace
		}
		projectName = filepath.Base(cwd)
		_, _ = fmt.Fprintf(out, "Initializing project in current directory: %s\n\n", projectName)
	} else {
		_, _ = fmt.Fprintf(out, "Creating project: %s\n\n", projectName)
	}

	return projectName, inPlace
}

func collectProjectConfig(out io.Writer, projectName string, inPlace bool) *generator.ProjectConfig {
	cfg := &generator.ProjectConfig{
		ProjectName: projectName,
		InPlace:     inPlace,
	}

	// 1. Project description
	cfg.Description = prompt("Project description:")

	// 2. Language selection
	_, _ = fmt.Fprintln(out)
	languageDisplay := promptSelect("Language?", []string{"Go", "Python"})
	cfg.Language = strings.ToLower(languageDisplay)

	// 3. Project type
	_, _ = fmt.Fprintln(out)
	projectType := promptSelect("Project type?", []string{"CLI", "API"})
	cfg.Base = strings.ToLower(projectType)

	// 4. Language-specific configuration
	if cfg.Language == languageGo {
		_, _ = fmt.Fprintln(out)
		defaultModule := "github.com/username/" + projectName
		cfg.Module = promptWithDefault("Go module path:", defaultModule)
	}

	// Set entry point
	if cfg.Base == projectTypeAPI {
		cfg.EntryPoint = "server"
	} else {
		cfg.EntryPoint = "."
	}

	// 5. Select modules
	cfg.Modules = []string{}
	_, _ = fmt.Fprintln(out)

	switch cfg.Language {
	case languageGo:
		selectGoModules(cfg)
	case languagePython:
		selectPythonModules(cfg)
	}

	// 6. Select AI tools
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "🤖 Select AI tools:")
	cfg.AITools = promptAITools()

	return cfg
}

func confirmConfiguration(out io.Writer, cfg *generator.ProjectConfig) bool {
	_, _ = fmt.Fprintln(out, "\n--- Configuration ---")
	_, _ = fmt.Fprintf(out, "Name:        %s\n", cfg.ProjectName)
	_, _ = fmt.Fprintf(out, "Language:    %s\n", cfg.Language)
	_, _ = fmt.Fprintf(out, "Type:        %s\n", cfg.Base)
	if cfg.Module != "" {
		_, _ = fmt.Fprintf(out, "Module:      %s\n", cfg.Module)
	}
	_, _ = fmt.Fprintf(out, "Modules:     %v\n", cfg.Modules)
	_, _ = fmt.Fprintln(out)

	return promptYesNo("Generate project?")
}

func generateProject(out io.Writer, cfg *generator.ProjectConfig, projectName string) error {
	_, _ = fmt.Fprintln(out, "\nGenerating project...")

	templatesDir := findTemplatesDir()

	gen, err := generator.New(templatesDir)
	if err != nil {
		return fmt.Errorf("failed to create generator: %w", err)
	}

	if err := gen.Generate(cfg); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	printSuccessMessage(out, projectName)
	return nil
}

func findTemplatesDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return templatesPath
	}

	exeDir := filepath.Dir(exePath)
	templatesDir := filepath.Join(exeDir, "..", templatesPath)

	if _, statErr := os.Stat(templatesDir); os.IsNotExist(statErr) {
		return templatesPath
	}

	return templatesDir
}

func printSuccessMessage(out io.Writer, projectName string) {
	_, _ = fmt.Fprintf(out, "\n✓ Project created: %s\n\n", projectName)
	_, _ = fmt.Fprintln(out, "Next steps:")
	_, _ = fmt.Fprintf(out, "  cd %s\n", projectName)
	_, _ = fmt.Fprintln(out, "  make setup")
	_, _ = fmt.Fprintln(out, "  make build")
	_, _ = fmt.Fprintln(out, "  make run")
}

func selectGoModules(cfg *generator.ProjectConfig) {
	if cfg.Base == projectTypeAPI {
		// Framework (required for API)
		framework := promptSelect("Web framework?", []string{"Gin"})
		cfg.Modules = append(cfg.Modules, strings.ToLower(framework))

		// Database
		if promptYesNo("Need database?") {
			_ = promptSelect("Database?", []string{"PostgreSQL (database/sql)"})
			cfg.Modules = append(cfg.Modules, "database-sql")
		}
	} else if promptYesNo("Use Cobra CLI framework?") {
		// CLI framework
		cfg.Modules = append(cfg.Modules, "cobra")
	}

	// Testing (optional for both)
	if promptYesNo("Add testing framework (testify)?") {
		cfg.Modules = append(cfg.Modules, "testify")
	}
}

func selectPythonModules(cfg *generator.ProjectConfig) {
	if cfg.Base == projectTypeAPI {
		// Framework (required for API)
		framework := promptSelect("Web framework?", []string{"FastAPI"})
		cfg.Modules = append(cfg.Modules, strings.ToLower(framework))

		// Database
		if promptYesNo("Need database (SQLAlchemy)?") {
			cfg.Modules = append(cfg.Modules, "sqlalchemy")
		}
	}
	// Python doesn't have CLI module yet

	// Testing (optional for both)
	if promptYesNo("Add testing framework (pytest)?") {
		cfg.Modules = append(cfg.Modules, "pytest")
	}
}

// Helper functions for interactive prompts

var scanner = bufio.NewScanner(os.Stdin)

func prompt(question string) string {
	fmt.Printf("%s ", question)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func promptWithDefault(question, defaultValue string) string {
	fmt.Printf("%s [%s] ", question, defaultValue)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())
	if answer == "" {
		return defaultValue
	}
	return answer
}

func promptSelect(question string, options []string) string {
	fmt.Println(question)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Print("→ ")

	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	// Parse selection
	var selection int
	_, _ = fmt.Sscanf(input, "%d", &selection)

	if selection < 1 || selection > len(options) {
		return options[0]
	}

	return options[selection-1]
}

func promptYesNo(question string) bool {
	fmt.Printf("%s [y/n] ", question)
	scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}
