package standard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakuv3r/codi/internal/analyzer"
)

// Manager manages coding standards.
type Manager struct {
	officialRoot string // Path to built-in standards
	projectRoot  string // Path to project root
}

// NewManager creates a new standards manager.
func NewManager(officialRoot, projectRoot string) *Manager {
	return &Manager{
		officialRoot: officialRoot,
		projectRoot:  projectRoot,
	}
}

// LoadForProject loads standards for a project.
func (m *Manager) LoadForProject(info *analyzer.ProjectInfo, deps *analyzer.DependencyList) (*List, error) {
	list := &List{
		Universal:    []Standard{},
		Core:         []Standard{},
		Dependencies: []Standard{},
	}

	// Load universal standards
	universal, err := m.loadUniversal()
	if err != nil {
		return nil, err
	}
	list.Universal = universal

	// Load core language standards
	core, err := m.loadCore(info.Language)
	if err != nil {
		return nil, err
	}
	list.Core = core

	// Load all dependency standards
	for _, dep := range deps.Dependencies {
		std, err := m.loadStandard(info.Language, dep.Path, &dep)
		if err != nil {
			continue // Skip if standard not found
		}
		list.Dependencies = append(list.Dependencies, *std)
	}

	return list, nil
}

// loadUniversal loads universal cross-language standards.
func (m *Manager) loadUniversal() ([]Standard, error) {
	return m.loadDirectory(filepath.Join(m.officialRoot, "universal"))
}

// loadCore loads core language standards.
func (m *Manager) loadCore(lang analyzer.Language) ([]Standard, error) {
	langDir := m.langDir(lang)
	corePath := filepath.Join(m.officialRoot, langDir, "core")
	return m.loadDirectory(corePath)
}

// loadDirectory loads all standards from a directory.
func (m *Manager) loadDirectory(dirPath string) ([]Standard, error) {
	var standards []Standard

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		relPath, _ := filepath.Rel(m.officialRoot, fullPath)
		name := strings.TrimSuffix(entry.Name(), ".md")

		standards = append(standards, Standard{
			Name:       name,
			Path:       relPath,
			FullPath:   fullPath,
			Source:     SourceOfficial,
			Dependency: nil,
		})
	}

	return standards, nil
}

// loadStandard loads a single standard by dependency path.
func (m *Manager) loadStandard(lang analyzer.Language, depPath string, dep *analyzer.Dependency) (*Standard, error) {
	// Resolve standard path using zero-mapping
	stdPath := m.resolvePath(lang, depPath)

	// Check local override first
	localPath := filepath.Join(m.projectRoot, ".codi", "standards", stdPath)
	if fileExists(localPath) {
		relPath, _ := filepath.Rel(filepath.Join(m.projectRoot, ".codi", "standards"), localPath)
		return &Standard{
			Name:       filepath.Base(depPath),
			Path:       relPath,
			FullPath:   localPath,
			Source:     SourceLocal,
			Dependency: dep,
		}, nil
	}

	// Fallback to official
	officialPath := filepath.Join(m.officialRoot, stdPath)
	if fileExists(officialPath) {
		relPath, _ := filepath.Rel(m.officialRoot, officialPath)
		return &Standard{
			Name:       filepath.Base(depPath),
			Path:       relPath,
			FullPath:   officialPath,
			Source:     SourceOfficial,
			Dependency: dep,
		}, nil
	}

	return nil, ErrStandardNotFound
}

// resolvePath resolves dependency path to standard file path (zero-mapping).
func (m *Manager) resolvePath(lang analyzer.Language, depPath string) string {
	// Zero-mapping: dependency path = file path
	// e.g., "github.com/gin-gonic/gin" -> "go/github.com/gin-gonic/gin.md"
	//       "log/slog" -> "go/log/slog.md"

	// Security: clean path to prevent directory traversal
	depPath = filepath.Clean(depPath)
	if strings.Contains(depPath, "..") {
		// Reject paths with parent directory references
		return ""
	}

	langDir := m.langDir(lang)
	return filepath.Join(langDir, depPath+".md")
}

// langDir returns directory name for language.
func (m *Manager) langDir(lang analyzer.Language) string {
	switch lang {
	case analyzer.LanguageGo:
		return "go"
	case analyzer.LanguagePython:
		return "python"
	default:
		return lang.String()
	}
}

// fileExists checks if file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
