package analyzer

import "slices"

// Language represents a programming language.
type Language string

// Supported programming languages.
const (
	// LanguageGo represents Go programming language.
	LanguageGo        Language = "go"
	LanguagePython    Language = "python"
	LanguageUniversal Language = "universal"
)

func (l Language) String() string {
	return string(l)
}

// IsValid checks if the language is supported.
func (l Language) IsValid() bool {
	return l == LanguageGo || l == LanguagePython || l == LanguageUniversal
}

// ProjectType represents the type of project.
type ProjectType string

// Project types.
const (
	// ProjectTypeAPI represents REST/GraphQL/gRPC API projects.
	ProjectTypeAPI ProjectType = "api" // REST/GraphQL/gRPC API
	ProjectTypeCLI ProjectType = "cli" // Command-line tool
)

func (pt ProjectType) String() string {
	return string(pt)
}

// ProjectInfo contains project metadata.
type ProjectInfo struct {
	Name           string      // Project name
	Path           string      // Project root path
	Language       Language    // Programming language
	ProjectType    ProjectType // Project type
	Description    string      // Project description
	Module         string      // Go module path or Python package name
	EntryPoint     string      // Entry point for CLI/API (used in templates)
	HTTPFramework  string      // HTTP framework (gin/echo/fiber)
	Database       string      // Database (postgres/mysql/sqlite)
	DatabaseDriver string      // Database driver (gorm)
	Features       []string    // Features (jwt/redis/swagger)
}

// HasFeature checks if project has a specific feature.
func (pi *ProjectInfo) HasFeature(feature string) bool {
	return slices.Contains(pi.Features, feature)
}
