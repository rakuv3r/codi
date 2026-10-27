package analyzer

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// GoAnalyzer analyzes Go projects.
type GoAnalyzer struct{}

// NewGoAnalyzer creates a new Go analyzer.
func NewGoAnalyzer() *GoAnalyzer {
	return &GoAnalyzer{}
}

// Language returns Go.
func (a *GoAnalyzer) Language() Language {
	return LanguageGo
}

// Analyze analyzes Go project dependencies.
func (a *GoAnalyzer) Analyze(projectPath string) (*DependencyList, error) {
	var allDeps []Dependency

	// Parse go.mod for modules
	modules, err := a.parseGoMod(filepath.Join(projectPath, "go.mod"))
	if err != nil {
		return nil, err
	}
	allDeps = append(allDeps, modules...)

	// Scan .go files for imports
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

// parseGoMod parses go.mod file.
func (a *GoAnalyzer) parseGoMod(goModPath string) ([]Dependency, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	deps := make([]Dependency, 0, len(modFile.Require))
	for _, req := range modFile.Require {
		if req.Indirect {
			continue // Skip indirect dependencies
		}
		deps = append(deps, Dependency{
			Path: req.Mod.Path,
		})
	}

	return deps, nil
}

// extractImports extracts import paths from a Go file using AST parser.
func (a *GoAnalyzer) extractImports(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	var imports []string
	for _, imp := range node.Imports {
		// imp.Path.Value is quoted: "path/to/package"
		path := strings.Trim(imp.Path.Value, `"`)
		if path != "" && path != "C" {
			imports = append(imports, path)
		}
	}

	return imports, nil
}

func (a *GoAnalyzer) scanImports(projectPath string) ([]Dependency, error) {
	importSet := make(map[string]bool)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}

		// Skip vendor and hidden directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files (excluding _test.go)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Scan file for imports
		imports, err := a.extractImports(path)
		if err != nil {
			return err
		}

		for _, imp := range imports {
			importSet[imp] = true
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk project: %w", err)
	}

	// Convert set to list
	deps := make([]Dependency, 0, len(importSet))
	for imp := range importSet {
		deps = append(deps, Dependency{
			Path: imp,
		})
	}

	return deps, nil
}
