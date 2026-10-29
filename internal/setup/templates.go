package setup

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rakuv3r/codi"
	"github.com/rakuv3r/codi/internal/config"
)

// CopyReviewTemplate copies review-checklist.md to .codi/ directory.
func CopyReviewTemplate(out io.Writer, codiRoot string, projectPath string) error {
	codiDir := filepath.Join(projectPath, config.CodiDirName)

	// Create .codi directory if needed
	if err := os.MkdirAll(codiDir, config.DirPermission); err != nil {
		return fmt.Errorf("failed to create .codi directory: %w", err)
	}

	var checklistData []byte
	var err error

	if codiRoot == "" {
		// Use embedded files
		checklistData, err = fs.ReadFile(codi.EmbeddedFS, filepath.Join("standards", config.ReviewChecklistFilename))
		if err != nil {
			return fmt.Errorf("failed to read embedded checklist: %w", err)
		}
	} else {
		// Use disk files
		checklistSrc := filepath.Join(codiRoot, config.StandardsDirName, config.ReviewChecklistFilename)
		checklistData, err = os.ReadFile(checklistSrc)
		if err != nil {
			return fmt.Errorf("failed to read checklist from disk: %w", err)
		}
	}

	checklistDest := filepath.Join(codiDir, config.ReviewChecklistFilename)
	if err := os.WriteFile(checklistDest, checklistData, config.ConfigPermission); err != nil {
		return fmt.Errorf("failed to write checklist: %w", err)
	}

	_, _ = fmt.Fprintln(out, "  ✓ .codi/review-checklist.md created")
	return nil
}
