package command

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi"
	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/config"
	"github.com/rakuv3r/codi/internal/setup"
)

const (
	separatorWidth = 60
)

// NewViewCmd creates the view command.
func NewViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [language]",
		Short: "View coding standards",
		Long: `View coding standards interactively.

Without arguments: select language first, then select standard
With language: show standards for that language

Examples:
  codi view              # Interactive: select language, then standard
  codi view go           # Show Go standards
  codi view python       # Show Python standards`,
		RunE: runView,
	}

	return cmd
}

func runView(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// 1. Determine language
	language, err := determineLanguage(args)
	if err != nil {
		return err
	}

	// 2. Get standards for language
	standards, err := loadStandards(language)
	if err != nil {
		return err
	}

	if len(standards) == 0 {
		return fmt.Errorf("no standards found for %s", language)
	}

	// 3. Display list and let user select
	displayStandardsList(out, standards, language)
	selectedIdx := promptUserSelection(len(standards))

	if selectedIdx < 0 {
		return nil
	}

	// 4. Display selected standard
	_, _ = fmt.Fprintln(out)
	return displaySelectedStandard(out, standards[selectedIdx].Path)
}

func determineLanguage(args []string) (string, error) {
	if len(args) == 0 {
		// Interactive: select language
		choice := promptSelect("Language?", []string{"Go", "Python", "Universal"})
		return strings.ToLower(choice), nil
	}

	// Direct: use argument
	language := strings.ToLower(args[0])
	if language != string(analyzer.LanguageGo) && language != string(analyzer.LanguagePython) && language != string(analyzer.LanguageUniversal) {
		return "", fmt.Errorf("invalid language: %s (use: go, python, universal)", args[0])
	}
	return language, nil
}

func loadStandards(language string) ([]StandardItem, error) {
	codiRoot := setup.GetCodiRoot()

	if codiRoot == "" {
		// Use embedded standards
		return collectStandardsFromEmbed(language)
	}

	// Use disk standards
	standardsRoot := filepath.Join(codiRoot, "standards")
	return collectStandardsForLanguage(standardsRoot, language), nil
}

func displaySelectedStandard(out io.Writer, path string) error {
	codiRoot := setup.GetCodiRoot()

	if codiRoot == "" {
		return displayStandardFromEmbed(out, path)
	}

	standardsRoot := filepath.Join(codiRoot, "standards")
	return displayStandard(out, path, standardsRoot)
}

// StandardItem represents a coding standard file entry.
// Name is the display name shown to users (includes subdirectory path).
// Path is the file path (absolute for disk, relative for embed).
type StandardItem struct {
	Name string
	Path string
}

func collectStandardsForLanguage(standardsRoot, language string) []StandardItem {
	langDir := filepath.Join(standardsRoot, language)
	return scanDirectory(langDir, langDir)
}

// scanDirectory scans a language directory and returns all standards.
// Automatically detects whether to scan recursively based on directory structure.
// langDir is the directory to scan, baseDir is used to compute relative display names.
func scanDirectory(langDir, baseDir string) []StandardItem {
	var standards []StandardItem

	entries, err := os.ReadDir(langDir)
	if err != nil {
		return standards
	}

	// Scan direct .md files in language root (e.g., python/fastapi.md)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != config.ReadmeFilename {
			name := strings.TrimSuffix(entry.Name(), ".md")
			standards = append(standards, StandardItem{
				Name: name,
				Path: filepath.Join(langDir, entry.Name()),
			})
		}
	}

	// Scan subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subDir := entry.Name()
		subPath := filepath.Join(langDir, subDir)

		if shouldScanRecursively(subDir) {
			standards = append(standards, scanRecursiveWithPrefix(subPath, baseDir, subDir)...)
		} else {
			standards = append(standards, scanFlatWithPrefix(subPath, baseDir, subDir)...)
		}
	}

	return standards
}

// shouldScanRecursively determines if a directory should be scanned recursively.
// Dependency directories (containing dots like github.com, gorm.io) are scanned recursively.
func shouldScanRecursively(dirName string) bool {
	return strings.Contains(dirName, ".")
}

// scanFlatWithPrefix scans a directory for .md files (non-recursive) and adds prefix to names.
func scanFlatWithPrefix(dir, _ /* baseDir */, prefix string) []StandardItem {
	var standards []StandardItem

	entries, err := os.ReadDir(dir)
	if err != nil {
		return standards
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == config.ReadmeFilename {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		standards = append(standards, StandardItem{
			Name: prefix + "/" + name,
			Path: filepath.Join(dir, entry.Name()),
		})
	}

	return standards
}

// scanRecursiveWithPrefix scans a directory recursively for .md files and adds prefix to names.
func scanRecursiveWithPrefix(dir, baseDir, prefix string) []StandardItem {
	var standards []StandardItem

	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") || info.Name() == config.ReadmeFilename {
			return nil //nolint:nilerr
		}

		// Get relative path from base directory
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			// Fallback: use just filename with prefix
			relPath = filepath.Join(prefix, info.Name())
		}
		// Remove .md extension
		name := strings.TrimSuffix(relPath, ".md")

		standards = append(standards, StandardItem{
			Name: name,
			Path: path,
		})
		return nil
	})

	return standards
}

func displayStandardsList(out io.Writer, standards []StandardItem, language string) {
	_, _ = fmt.Fprintf(out, "📚 %s Standards\n\n", strings.ToUpper(language[:1])+language[1:])

	for i, std := range standards {
		_, _ = fmt.Fprintf(out, "  %2d. %s\n", i+1, std.Name)
	}
}

func promptUserSelection(count int) int {
	fmt.Printf("\nSelect (1-%d, 0 to cancel): ", count)

	if !inputScanner.Scan() {
		return -1 // EOF or read error
	}
	input := strings.TrimSpace(inputScanner.Text())

	if input == "0" || input == "" {
		return -1
	}

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > count {
		return -1
	}

	return num - 1
}

func displayStandard(out io.Writer, path, standardsRoot string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read standard: %w", err)
	}

	relPath, err := filepath.Rel(standardsRoot, path)
	if err != nil {
		relPath = path // Fallback to absolute path
	}
	_, _ = fmt.Fprintf(out, "📄 %s\n", relPath)
	_, _ = fmt.Fprintln(out, strings.Repeat("─", separatorWidth))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, string(content))

	return nil
}

func collectStandardsFromEmbed(language string) ([]StandardItem, error) {
	standardsFS, err := fs.Sub(codi.EmbeddedFS, "standards")
	if err != nil {
		return nil, fmt.Errorf("failed to access embedded standards: %w", err)
	}

	return scanFS(standardsFS, language)
}

// scanFS scans a language directory from fs.FS and returns all standards.
func scanFS(fsys fs.FS, language string) ([]StandardItem, error) {
	var standards []StandardItem

	entries, err := fs.ReadDir(fsys, language)
	if err != nil {
		return nil, fmt.Errorf("failed to read language dir: %w", err)
	}

	// Scan direct .md files in language root
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != config.ReadmeFilename {
			name := strings.TrimSuffix(entry.Name(), ".md")
			standards = append(standards, StandardItem{
				Name: name,
				Path: filepath.Join(language, entry.Name()),
			})
		}
	}

	// Scan subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subDir := entry.Name()
		relPath := filepath.Join(language, subDir)

		if shouldScanRecursively(subDir) {
			standards = append(standards, scanRecursiveFSWithPrefix(fsys, relPath, subDir)...)
		} else {
			standards = append(standards, scanFlatFSWithPrefix(fsys, relPath, subDir)...)
		}
	}

	return standards, nil
}

// scanFlatFSWithPrefix scans a directory for .md files from fs.FS (non-recursive) with prefix.
func scanFlatFSWithPrefix(fsys fs.FS, dir, prefix string) []StandardItem {
	var standards []StandardItem

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return standards
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == config.ReadmeFilename {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		standards = append(standards, StandardItem{
			Name: prefix + "/" + name,
			Path: filepath.Join(dir, entry.Name()),
		})
	}

	return standards
}

// scanRecursiveFSWithPrefix scans a directory recursively for .md files from fs.FS with prefix.
func scanRecursiveFSWithPrefix(fsys fs.FS, dir, prefix string) []StandardItem {
	var standards []StandardItem

	_ = fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") || d.Name() == config.ReadmeFilename {
			return nil //nolint:nilerr
		}

		// Get relative path from directory prefix
		relPath := strings.TrimPrefix(path, dir)
		relPath = strings.TrimPrefix(relPath, "/")
		// Remove .md extension
		name := strings.TrimSuffix(relPath, ".md")

		standards = append(standards, StandardItem{
			Name: prefix + "/" + name,
			Path: path,
		})
		return nil
	})

	return standards
}

func displayStandardFromEmbed(out io.Writer, path string) error {
	standardsFS, err := fs.Sub(codi.EmbeddedFS, "standards")
	if err != nil {
		return fmt.Errorf("failed to access embedded standards: %w", err)
	}

	content, err := fs.ReadFile(standardsFS, path)
	if err != nil {
		return fmt.Errorf("failed to read standard: %w", err)
	}

	_, _ = fmt.Fprintf(out, "📄 %s\n", path)
	_, _ = fmt.Fprintln(out, strings.Repeat("─", separatorWidth))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, string(content))

	return nil
}
