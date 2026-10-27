package command

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

// addMissingQualityTools checks and adds missing quality tool files.
func addMissingQualityTools(out io.Writer, projectPath string, projectInfo *analyzer.ProjectInfo, templatesDir string) error {
	var qualityDir string
	var useEmbed bool

	if templatesDir == "" {
		// Use embedded templates
		useEmbed = true
		qualityDir = filepath.Join("templates", string(projectInfo.Language), "base", "quality")
	} else {
		// Use disk templates
		useEmbed = false
		qualityDir = filepath.Join(templatesDir, string(projectInfo.Language), "base", "quality")
	}

	// Read all files in quality template directory
	var files []fs.DirEntry
	var err error

	if useEmbed {
		files, err = fs.ReadDir(codi.EmbeddedFS, qualityDir)
	} else {
		files, err = os.ReadDir(qualityDir)
	}

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
			if err := copyQualityFile(qualityDir, projectPath, srcFile, dstFile, projectInfo, useEmbed); err != nil {
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
func copyQualityFile(srcDir, dstDir, srcFile, dstFile string, projectInfo *analyzer.ProjectInfo, useEmbed bool) error {
	srcPath := filepath.Join(srcDir, srcFile)
	dstPath := filepath.Join(dstDir, dstFile)

	var data []byte
	var err error

	if useEmbed {
		data, err = fs.ReadFile(codi.EmbeddedFS, srcPath)
	} else {
		data, err = os.ReadFile(srcPath)
	}

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

	if err := os.WriteFile(dstPath, data, config.FilePermission); err != nil {
		return fmt.Errorf("failed to write %s: %w", dstPath, err)
	}

	return nil
}
