package standard

import "github.com/rakuv3r/codi/internal/analyzer"

// Source represents where a standard comes from.
type Source string

// Standard sources.
const (
	// SourceOfficial represents official codi standards.
	SourceOfficial Source = "official" // Built-in standards
	SourceLocal    Source = "local"    // Project-local overrides
)

// Standard represents a coding standard document.
type Standard struct {
	Dependency *analyzer.Dependency
	Name       string
	Path       string
	FullPath   string
	Source     Source
	Language   analyzer.Language
}

// List categorizes standards by type.
type List struct {
	Universal    []Standard // Universal cross-language standards
	Core         []Standard // Core language principles
	Dependencies []Standard // Dependency-related standards (stdlib and third-party)
}

// AllStandards returns all standards combined.
func (sl *List) AllStandards() []Standard {
	totalLen := len(sl.Universal) + len(sl.Core) + len(sl.Dependencies)
	all := make([]Standard, 0, totalLen)
	all = append(all, sl.Universal...)
	all = append(all, sl.Core...)
	all = append(all, sl.Dependencies...)
	return all
}

// Count returns total number of standards.
func (sl *List) Count() int {
	return len(sl.Universal) + len(sl.Core) + len(sl.Dependencies)
}
