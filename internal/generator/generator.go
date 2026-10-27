// Package generator provides project scaffolding generation from templates.
package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
)

// Permission constants.
const (
	DirPermission  = 0755 // Directories: owner rwx, group r-x, others r-x
	FilePermission = 0644 // Public files: owner rw-, group r--, others r--
)

// Generator generates project scaffolding from templates.
type Generator struct {
	fsys         fs.FS
	templatesDir string
	useEmbed     bool
}

// New creates a new generator using disk.
func New(templatesDir string) (*Generator, error) {
	return &Generator{
		templatesDir: templatesDir,
		useEmbed:     false,
	}, nil
}

// NewWithFS creates a new generator using fs.FS.
func NewWithFS(fsys fs.FS) *Generator {
	return &Generator{
		fsys:     fsys,
		useEmbed: true,
	}
}

// Generate creates a new project from configuration.
func (g *Generator) Generate(cfg *ProjectConfig) error {
	var projectDir string

	if cfg.InPlace {
		// Initialize in current directory
		projectDir = "."
	} else {
		// Create new directory
		projectDir = cfg.ProjectName

		// Create project directory
		if err := os.MkdirAll(projectDir, DirPermission); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	// 1. Copy quality tools (always)
	if err := g.copyQualityTools(projectDir, cfg); err != nil {
		return fmt.Errorf("failed to copy quality tools: %w", err)
	}

	// 2. Copy base structure
	if err := g.copyBase(projectDir, cfg); err != nil {
		return fmt.Errorf("failed to copy base: %w", err)
	}

	// 3. Apply modules
	for _, moduleName := range cfg.Modules {
		if err := g.applyModule(projectDir, moduleName, cfg); err != nil {
			return fmt.Errorf("failed to apply module %s: %w", moduleName, err)
		}
	}

	// 4. Language-specific initialization
	if cfg.Language == "go" {
		// Generate go.mod
		if err := g.generateGoMod(projectDir, cfg); err != nil {
			return fmt.Errorf("failed to generate go.mod: %w", err)
		}
	}
	// Python uses pyproject.toml from templates, no generation needed

	// 5. Generate AI tool files
	for _, tool := range cfg.AITools {
		if !tool.Enabled {
			continue
		}

		// Create parent directory
		toolDir := filepath.Dir(filepath.Join(projectDir, tool.OutputPath))
		if err := os.MkdirAll(toolDir, DirPermission); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", tool.Name, err)
		}

		// Generate file
		toolPath := filepath.Join(projectDir, tool.OutputPath)
		if err := g.generateAIToolFile(toolPath, cfg); err != nil {
			return fmt.Errorf("failed to generate %s: %w", tool.OutputPath, err)
		}
	}

	// 6. Generate .codi/config.yaml (minimal config)
	if err := g.generateCodiYaml(projectDir, cfg); err != nil {
		return fmt.Errorf("failed to generate config.yaml: %w", err)
	}

	return nil
}

// copyQualityTools copies the quality tools (Makefile, golangci, etc.)
func (g *Generator) copyQualityTools(projectDir string, cfg *ProjectConfig) error {
	qualityDir := filepath.Join(cfg.Language, "base", "quality")
	return g.copyTemplateDir(qualityDir, projectDir, cfg)
}

// copyBase copies the base project structure.
func (g *Generator) copyBase(projectDir string, cfg *ProjectConfig) error {
	baseDir := filepath.Join(cfg.Language, "base", cfg.Base)
	return g.copyTemplateDir(baseDir, projectDir, cfg)
}

// applyModule applies a module to the project.
func (g *Generator) applyModule(projectDir, moduleName string, cfg *ProjectConfig) error {
	moduleDir := filepath.Join(cfg.Language, "module", moduleName)

	// Check if module directory exists
	if g.useEmbed {
		if _, err := fs.Stat(g.fsys, moduleDir); err != nil {
			return fmt.Errorf("module %s not found", moduleName)
		}
	} else {
		fullPath := filepath.Join(g.templatesDir, moduleDir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("module %s not found", moduleName)
		}
	}

	return g.copyTemplateDir(moduleDir, projectDir, cfg)
}

// copyTemplateDir recursively copies and renders templates from src to dst.
func (g *Generator) copyTemplateDir(src, dst string, cfg *ProjectConfig) error {
	if g.useEmbed {
		return g.copyTemplateDirEmbed(src, dst, cfg)
	}
	return g.copyTemplateDirDisk(src, dst, cfg)
}

// copyTemplateDirDisk copies from disk using filepath.Walk
func (g *Generator) copyTemplateDirDisk(src, dst string, cfg *ProjectConfig) error {
	fullSrc := filepath.Join(g.templatesDir, src)
	return filepath.Walk(fullSrc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(fullSrc, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, DirPermission)
		}

		// Render and copy file
		return g.renderTemplateFileDisk(path, dstPath, cfg)
	})
}

// copyTemplateDirEmbed copies from embed.FS using fs.WalkDir
func (g *Generator) copyTemplateDirEmbed(src, dst string, cfg *ProjectConfig) error {
	return fs.WalkDir(g.fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, DirPermission)
		}

		// Render and copy file
		return g.renderTemplateFileEmbed(path, dstPath, cfg)
	})
}

// renderTemplateFileDisk renders a template file from disk and writes it to destination.
func (g *Generator) renderTemplateFileDisk(src, dst string, cfg *ProjectConfig) error {
	// Read template
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return g.renderTemplateContent(src, dst, content, cfg)
}

// renderTemplateFileEmbed renders a template file from embed.FS and writes it to destination.
func (g *Generator) renderTemplateFileEmbed(src, dst string, cfg *ProjectConfig) error {
	// Read template
	content, err := fs.ReadFile(g.fsys, src)
	if err != nil {
		return err
	}

	return g.renderTemplateContent(src, dst, content, cfg)
}

// renderTemplateContent renders template content and writes to destination.
func (g *Generator) renderTemplateContent(src, dst string, content []byte, cfg *ProjectConfig) error {
	// Check if it's a template file (ends with .tmpl)
	if filepath.Ext(src) == ".tmpl" {
		// Remove .tmpl extension from destination
		dst = dst[:len(dst)-5]

		// Parse and execute template
		tmpl, err := template.New("file").Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", src, err)
		}

		// Create destination file
		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		// Render template
		data := TemplateData{
			ProjectName: cfg.ProjectName,
			Description: cfg.Description,
			Module:      cfg.Module,
			EntryPoint:  cfg.EntryPoint,
		}

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", src, err)
		}
	} else {
		// Copy file as-is
		if err := os.WriteFile(dst, content, FilePermission); err != nil {
			return err
		}
	}

	return nil
}

// generateGoMod creates the go.mod file.
func (g *Generator) generateGoMod(projectDir string, cfg *ProjectConfig) error {
	goModPath := filepath.Join(projectDir, "go.mod")

	goVersion, err := detectGoVersion()
	if err != nil {
		return fmt.Errorf("failed to detect Go version: %w", err)
	}

	content := fmt.Sprintf("module %s\n\ngo %s\n", cfg.Module, goVersion)

	return os.WriteFile(goModPath, []byte(content), FilePermission)
}

// detectGoVersion detects the Go version used to build codi.
func detectGoVersion() (string, error) {
	// runtime.Version() returns "go1.23.0"
	version := runtime.Version()

	// Parse "go1.23.0" -> "1.23"
	re := regexp.MustCompile(`go(\d+\.\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) >= 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("could not parse Go version from %q", version)
}

// generateAIToolFile creates an AI tool configuration file.
func (g *Generator) generateAIToolFile(path string, cfg *ProjectConfig) error {
	content := fmt.Sprintf("# Project: %s\n\n", cfg.ProjectName)
	content += cfg.Description + "\n\n"
	content += "## Quick Start\n\n"
	content += "```bash\n"
	content += "make setup    # Install development tools\n"
	content += "make build    # Build the application\n"
	content += "make test     # Run tests\n"
	content += "make run      # Run the application\n"
	content += "```\n\n"
	content += "## Standards\n\n"
	content += "Run `make setup` to install dependencies, then run `codi sync` to load coding standards.\n"

	return os.WriteFile(path, []byte(content), FilePermission)
}

func (g *Generator) generateCodiYaml(projectDir string, cfg *ProjectConfig) error {
	codiDir := filepath.Join(projectDir, ".codi")
	if err := os.MkdirAll(codiDir, DirPermission); err != nil {
		return err
	}

	configPath := filepath.Join(codiDir, "config.yaml")

	// Use RFC3339 format for timestamp
	now := "2000-01-01T00:00:00Z" // Placeholder, will be updated by sync

	var sb strings.Builder
	fmt.Fprintf(&sb, `# codi configuration
language: %s
project_type: %s

dependencies:
  modules: []
  stdlib: []

ai_tools:
`, cfg.Language, cfg.Base)

	// Add AI tools from user selection
	for _, tool := range cfg.AITools {
		fmt.Fprintf(&sb, `  - name: %s
    enabled: %t
    output_path: %s
`, tool.Name, tool.Enabled, tool.OutputPath)
	}

	fmt.Fprintf(&sb, `
last_sync: %s
`, now)

	return os.WriteFile(configPath, []byte(sb.String()), FilePermission)
}
