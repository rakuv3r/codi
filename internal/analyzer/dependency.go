package analyzer

// Dependency represents a project dependency.
type Dependency struct {
	Path string // Import path (e.g., "github.com/gin-gonic/gin" or "fmt")
}

// DependencyList contains all project dependencies.
type DependencyList struct {
	Dependencies []Dependency // All dependencies (modules and stdlib combined)
}

// Dedupe removes duplicate dependencies based on Path.
func Dedupe(deps []Dependency) []Dependency {
	seen := make(map[string]bool)
	result := make([]Dependency, 0, len(deps))
	for _, dep := range deps {
		if !seen[dep.Path] {
			result = append(result, dep)
			seen[dep.Path] = true
		}
	}
	return result
}
