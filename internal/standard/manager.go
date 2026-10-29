package standard

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/rakuv3r/codi/internal/analyzer"
)

// Manager manages coding standards.
type Manager struct {
	fsys         fs.FS
	officialRoot string
	projectRoot  string
}

// NewManager creates a new standards manager using disk.
func NewManager(officialRoot, projectRoot string) *Manager {
	return &Manager{
		officialRoot: officialRoot,
		projectRoot:  projectRoot,
		fsys:         os.DirFS(officialRoot),
	}
}

// NewManagerWithFS creates a new standards manager using a custom FS.
func NewManagerWithFS(fsys fs.FS, projectRoot string) *Manager {
	return &Manager{
		fsys:        fsys,
		projectRoot: projectRoot,
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
	return m.loadDirectory("universal")
}

// loadCore loads core language standards.
func (m *Manager) loadCore(lang analyzer.Language) ([]Standard, error) {
	langDir := m.langDir(lang)
	corePath := filepath.Join(langDir, "core")
	return m.loadDirectory(corePath)
}

// loadDirectory loads all standards from a directory.
func (m *Manager) loadDirectory(dirPath string) ([]Standard, error) {
	var standards []Standard

	entries, err := fs.ReadDir(m.fsys, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ".md")

		// Read file content to parse frontmatter
		content, err := fs.ReadFile(m.fsys, fullPath)
		if err != nil {
			continue // Skip if cannot read
		}

		metadata := parseFrontmatter(content)

		standards = append(standards, Standard{
			Name:       name,
			Path:       fullPath,
			FullPath:   fullPath,
			Source:     SourceOfficial,
			Dependency: nil,
			Category:   metadata.Category,
			AppliesTo:  metadata.AppliesTo,
		})
	}

	return standards, nil
}

// loadStandard loads a single standard by dependency path.
func (m *Manager) loadStandard(lang analyzer.Language, depPath string, dep *analyzer.Dependency) (*Standard, error) {
	// Resolve standard path using zero-mapping
	stdPath := m.resolvePath(lang, depPath)
	if stdPath == "" {
		return nil, fmt.Errorf("invalid dependency path: %s", depPath)
	}

	// Check local override first
	localPath := filepath.Join(m.projectRoot, ".codi", "standards", stdPath)
	if _, err := os.Stat(localPath); err == nil {
		relPath, err := filepath.Rel(filepath.Join(m.projectRoot, ".codi", "standards"), localPath)
		if err != nil {
			relPath = stdPath // Fallback to standard path
		}
		return &Standard{
			Name:       filepath.Base(depPath),
			Path:       relPath,
			FullPath:   localPath,
			Source:     SourceLocal,
			Dependency: dep,
		}, nil
	}

	// Check in fsys (embedded or disk)
	if _, err := fs.Stat(m.fsys, stdPath); err == nil {
		return &Standard{
			Name:       filepath.Base(depPath),
			Path:       stdPath,
			FullPath:   stdPath,
			Source:     SourceOfficial,
			Dependency: dep,
		}, nil
	}

	return nil, fmt.Errorf("standard not found: %s", stdPath)
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

// frontmatterMetadata holds parsed frontmatter fields.
type frontmatterMetadata struct {
	Language  string   `yaml:"language"`
	Category  string   `yaml:"category"`
	AppliesTo []string `yaml:"applies_to"`
}

// parseFrontmatter extracts metadata from YAML frontmatter.
func parseFrontmatter(content []byte) frontmatterMetadata {
	metadata := frontmatterMetadata{
		AppliesTo: []string{}, // Default: empty means applies to all
	}

	// Check for frontmatter: must start with "---\n"
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return metadata
	}

	// Find the closing "---"
	endIdx := bytes.Index(content[4:], []byte("\n---\n"))
	if endIdx == -1 {
		return metadata
	}

	// Extract YAML content between delimiters
	yamlContent := content[4 : 4+endIdx]

	// Parse YAML
	_ = yaml.Unmarshal(yamlContent, &metadata)

	return metadata
}
