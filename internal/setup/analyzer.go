package setup

import (
	"fmt"
	"path/filepath"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
)

// AnalyzeDependencies analyzes project dependencies based on language.
func AnalyzeDependencies(language, projectPath string) (*analyzer.DependencyList, error) {
	lang := analyzer.Language(language)

	switch lang {
	case analyzer.LanguageGo:
		goAnalyzer := analyzer.NewGoAnalyzer()
		deps, err := goAnalyzer.Analyze(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze project: %w", err)
		}
		return deps, nil
	case analyzer.LanguagePython:
		pythonAnalyzer := analyzer.NewPythonAnalyzer()
		deps, err := pythonAnalyzer.Analyze(projectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze project: %w", err)
		}
		return deps, nil
	default:
		return &analyzer.DependencyList{}, nil
	}
}

// RebuildProjectInfo rebuilds ProjectInfo from config and dependencies.
func RebuildProjectInfo(cfg *config.Config, projectPath string, deps *analyzer.DependencyList) *analyzer.ProjectInfo {
	projectInfo := &analyzer.ProjectInfo{
		Name:           filepath.Base(projectPath),
		Path:           projectPath,
		Language:       cfg.Language,
		ProjectType:    inferProjectType(deps), // Re-infer based on new deps
		HTTPFramework:  cfg.Project.HTTPFramework,
		Database:       cfg.Project.Database,
		DatabaseDriver: cfg.Project.DatabaseDriver,
		Features:       cfg.Project.Features,
	}

	return projectInfo
}

// inferProjectType infers project type from dependencies.
func inferProjectType(deps *analyzer.DependencyList) analyzer.ProjectType {
	// Check for HTTP frameworks
	for _, dep := range deps.Dependencies {
		switch dep.Path {
		case "github.com/gin-gonic/gin",
			"github.com/labstack/echo/v4",
			"github.com/gofiber/fiber/v2",
			"fastapi",
			"flask":
			return analyzer.ProjectTypeAPI
		case "github.com/spf13/cobra",
			"github.com/urfave/cli/v2":
			return analyzer.ProjectTypeCLI
		}
	}

	return analyzer.ProjectTypeCLI // Default
}
