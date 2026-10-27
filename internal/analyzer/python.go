package analyzer

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// PythonAnalyzer analyzes Python projects.
type PythonAnalyzer struct{}

// NewPythonAnalyzer creates a new Python analyzer.
func NewPythonAnalyzer() *PythonAnalyzer {
	return &PythonAnalyzer{}
}

// Language returns Python.
func (a *PythonAnalyzer) Language() Language {
	return LanguagePython
}

// Analyze analyzes Python project dependencies.
func (a *PythonAnalyzer) Analyze(projectPath string) (*DependencyList, error) {
	var allDeps []Dependency

	// Parse pyproject.toml
	pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		modules, err := a.parsePyprojectToml(pyprojectPath)
		if err != nil {
			return nil, err
		}
		allDeps = append(allDeps, modules...)
	} else {
		return nil, errors.New("pyproject.toml not found: only modern Python projects with pyproject.toml are supported")
	}

	// Scan .py files for imports
	imports, err := a.scanImports(projectPath)
	if err != nil {
		return nil, err
	}
	allDeps = append(allDeps, imports...)

	// Dedupe
	allDeps = Dedupe(allDeps)

	return &DependencyList{
		Dependencies: allDeps,
	}, nil
}

// pyProjectConfig represents pyproject.toml structure.
type pyProjectConfig struct {
	Project struct {
		Dependencies []string `toml:"dependencies"`
	} `toml:"project"`
}

// parsePyprojectToml parses pyproject.toml for dependencies.
func (a *PythonAnalyzer) parsePyprojectToml(path string) ([]Dependency, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}

	var config pyProjectConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}

	seen := make(map[string]bool)
	deps := make([]Dependency, 0, len(config.Project.Dependencies))
	for _, depStr := range config.Project.Dependencies {
		dep := a.parseRequirementLine(depStr)
		if dep.Path != "" && !seen[dep.Path] {
			deps = append(deps, dep)
			seen[dep.Path] = true
		}
	}

	return deps, nil
}

// parseRequirementLine parses a single requirement line.
func (a *PythonAnalyzer) parseRequirementLine(line string) Dependency {
	// Handle common formats:
	// package==1.0.0
	// package>=1.0.0
	// package~=1.0.0
	// package[extra]==1.0.0
	// git+https://...

	// Skip git URLs
	if strings.HasPrefix(line, "git+") {
		return Dependency{}
	}

	// Remove extras [extra]
	if idx := strings.Index(line, "["); idx != -1 {
		line = line[:idx]
	}

	// Split by version specifier
	for _, sep := range []string{"==", ">=", "<=", "~=", ">", "<", "!="} {
		if idx := strings.Index(line, sep); idx != -1 {
			pkg := strings.TrimSpace(line[:idx])
			return Dependency{
				Path: pkg,
			}
		}
	}

	// No version specifier
	return Dependency{
		Path: strings.TrimSpace(line),
	}
}

// scanImports scans Python files for import statements.
func (a *PythonAnalyzer) scanImports(projectPath string) ([]Dependency, error) {
	stdlibSet := make(map[string]bool)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip venv and site-packages directories
		if info.IsDir() {
			name := info.Name()
			if name == "venv" || name == "site-packages" || name == "__pycache__" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .py files
		if !strings.HasSuffix(path, ".py") {
			return nil
		}

		// Parse imports from file
		imports, _ := a.parseImports(path)
		// Skip files with parse errors

		for _, imp := range imports {
			stdlibSet[imp] = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert set to slice
	var stdlib []Dependency
	for pkg := range stdlibSet {
		stdlib = append(stdlib, Dependency{
			Path: pkg,
		})
	}

	return stdlib, nil
}

// parseImports extracts import statements from a Python file.
func (a *PythonAnalyzer) parseImports(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var imports []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// import module
		if strings.HasPrefix(line, "import ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				module := strings.Split(parts[1], ".")[0] // Get top-level module
				module = strings.TrimSuffix(module, ",")
				imports = append(imports, module)
			}
		}

		// from module import ...
		if strings.HasPrefix(line, "from ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				module := strings.Split(parts[1], ".")[0]
				imports = append(imports, module)
			}
		}
	}

	return imports, scanner.Err()
}
