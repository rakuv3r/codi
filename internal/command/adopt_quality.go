package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
)

// addMissingQualityTools checks and adds missing quality tool files.
func addMissingQualityTools(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, templatesDir string) error {
	qualityDir := filepath.Join(templatesDir, string(projectInfo.Language), "base", "quality")

	// Read all files in quality template directory
	files, err := os.ReadDir(qualityDir)
	if err != nil {
		return fmt.Errorf("failed to read quality templates: %w", err)
	}

	added := []string{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		srcFile := file.Name()
		dstFile := strings.TrimSuffix(srcFile, ".tmpl")

		// Check if file exists in project
		dstPath := filepath.Join(projectPath, dstFile)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			_, _ = fmt.Fprintf(out, "  + Adding %s\n", dstFile)
			if err := copyQualityFile(qualityDir, projectPath, srcFile, dstFile, projectInfo); err != nil {
				return err
			}
			added = append(added, dstFile)
		} else {
			_, _ = fmt.Fprintf(out, "  ✓ %s exists\n", dstFile)
		}
	}

	if len(added) > 0 {
		_, _ = fmt.Fprintf(out, "  ✓ Added: %s\n", strings.Join(added, ", "))
	}

	return nil
}

// copyQualityFile copies a quality tool file, processing as template if needed.
func copyQualityFile(srcDir, dstDir, srcFile, dstFile string, projectInfo *analyzer.ProjectInfo) error {
	srcPath := filepath.Join(srcDir, srcFile)
	dstPath := filepath.Join(dstDir, dstFile)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	// Check if file needs template processing
	if strings.HasSuffix(srcFile, ".tmpl") {
		// Process as template
		// For now, simple string replacement
		content := string(data)
		content = strings.ReplaceAll(content, "{{.ProjectName}}", projectInfo.Name)
		content = strings.ReplaceAll(content, "{{.Language}}", string(projectInfo.Language))
		data = []byte(content)
	}

	if err := os.WriteFile(dstPath, data, config.FilePermission); err != nil {
		return fmt.Errorf("failed to write %s: %w", dstPath, err)
	}

	return nil
}
