package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rakuv3r/codi"
	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/standard"
)

// LoadStandards loads applicable standards for the project.
// Returns standards list and codi root path.
func LoadStandards(projectPath string, projectInfo *analyzer.ProjectInfo, deps *analyzer.DependencyList) (*standard.List, string, error) {
	codiRoot := GetCodiRoot()

	var stdMgr *standard.Manager
	if codiRoot == "" {
		// Use embedded standards
		standardsFS, err := fs.Sub(codi.EmbeddedFS, "standards")
		if err != nil {
			return nil, "", fmt.Errorf("failed to access embedded standards: %w", err)
		}
		stdMgr = standard.NewManagerWithFS(standardsFS, projectPath)
	} else {
		// Use disk standards
		stdMgr = standard.NewManager(filepath.Join(codiRoot, "standards"), projectPath)
	}

	standards, err := stdMgr.LoadForProject(projectInfo, deps)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load standards: %w", err)
	}

	return standards, codiRoot, nil
}

// CopyStandardsToProject copies standards to .codi/standards/ directory.
func CopyStandardsToProject(projectPath string, standards *standard.List, codiRoot string) error {
	destBase := filepath.Join(projectPath, ".codi", "standards")

	// Get source filesystem
	var srcFS fs.FS
	if codiRoot == "" {
		// Use embedded standards
		var err error
		srcFS, err = fs.Sub(codi.EmbeddedFS, "standards")
		if err != nil {
			return fmt.Errorf("failed to access embedded standards: %w", err)
		}
	} else {
		// Use disk standards
		srcFS = os.DirFS(filepath.Join(codiRoot, "standards"))
	}

	// Copy all standards
	for _, std := range standards.AllStandards() {
		srcPath := std.Path
		destPath := filepath.Join(destBase, std.Path)

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), config.DirPermission); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", std.Path, err)
		}

		// Read source
		content, err := fs.ReadFile(srcFS, srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", srcPath, err)
		}

		// Write destination
		if err := os.WriteFile(destPath, content, config.FilePermission); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}
	}

	return nil
}

// GetCodiRoot finds the codi installation root directory.
func GetCodiRoot() string {
	// Try to get exe path
	exe, err := os.Executable()
	if err != nil {
		// If we can't get exe path, try current directory
		cwd, _ := os.Getwd()
		if _, statErr := os.Stat(filepath.Join(cwd, "standards")); statErr == nil {
			return cwd
		}
		// Return empty string to signal use of embed
		return ""
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "" // Use embed
	}

	// Assume: bin/codi is in codi root
	// e.g., /path/to/codi/bin/codi -> /path/to/codi
	root := filepath.Dir(filepath.Dir(exe))

	// Verify standards/ exists
	standardsPath := filepath.Join(root, "standards")
	if _, err := os.Stat(standardsPath); err != nil {
		// Fallback: try current working directory (for development)
		cwd, _ := os.Getwd()
		if _, statErr := os.Stat(filepath.Join(cwd, "standards")); statErr == nil {
			return cwd
		}
		// Return empty string to signal use of embed
		return ""
	}

	return root
}
