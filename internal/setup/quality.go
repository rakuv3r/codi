package setup

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rakuv3r/codi"
	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/generator"
)

// projectInfoToTemplateData converts ProjectInfo to generator.TemplateData
func projectInfoToTemplateData(info *analyzer.ProjectInfo) generator.TemplateData {
	return generator.TemplateData{
		ProjectName: info.Name,
		Description: info.Description,
		Module:      info.Module,
		EntryPoint:  info.EntryPoint,
	}
}

// InstallTools checks and adds missing quality tool files.
func InstallTools(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo) error {
	_, _ = fmt.Fprintln(out, "\n📋 Checking quality tools...")

	qualityFiles := config.QualityFiles(projectInfo.Language)

	var added []string
	for srcFile, dstFile := range qualityFiles {
		dstPath := filepath.Join(projectPath, dstFile)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			_, _ = fmt.Fprintf(out, "  + Adding %s\n", dstFile)
			if err := copyQualityFile(projectPath, srcFile, dstFile, projectInfo); err != nil {
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
func copyQualityFile(projectPath, srcFile, dstFile string, projectInfo *analyzer.ProjectInfo) error {
	qualityDir := filepath.Join("templates", string(projectInfo.Language), "base", "quality")
	srcPath := filepath.Join(qualityDir, srcFile)
	dstPath := filepath.Join(projectPath, dstFile)

	// Read from embedded FS
	data, err := fs.ReadFile(codi.EmbeddedFS, srcPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", srcPath, err)
	}

	// Check if file needs template processing
	if strings.HasSuffix(srcFile, ".tmpl") {
		// Process as Go template
		tmpl, err := template.New(srcFile).Parse(string(data))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", srcFile, err)
		}

		// Convert ProjectInfo to template data
		templateData := projectInfoToTemplateData(projectInfo)

		var buf strings.Builder
		if err := tmpl.Execute(&buf, templateData); err != nil {
			return fmt.Errorf("failed to execute template %s: %w", srcFile, err)
		}
		data = []byte(buf.String())
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(dstPath), config.DirPermission); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dstPath, err)
	}

	if err := os.WriteFile(dstPath, data, config.FilePermission); err != nil {
		return fmt.Errorf("failed to write %s: %w", dstPath, err)
	}

	return nil
}
