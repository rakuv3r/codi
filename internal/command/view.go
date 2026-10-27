package command

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rakuv3r/codi/internal/config"
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
	var language string

	// 1. Determine language
	if len(args) == 0 {
		// Interactive: select language
		choice := promptSelect("Language?", []string{"Go", "Python", "Universal"})
		language = strings.ToLower(choice)
	} else {
		// Direct: use argument
		language = strings.ToLower(args[0])
		if language != languageGo && language != languagePython && language != languageUniversal {
			return fmt.Errorf("invalid language: %s (use: go, python, universal)", args[0])
		}
	}

	// 2. Get standards for language
	codiRoot := getCodiRoot()
	if codiRoot == "" {
		return errors.New("view command requires codi source directory")
	}

	standardsRoot := filepath.Join(codiRoot, "standards")
	standards := collectStandardsForLanguage(standardsRoot, language)

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
	return displayStandard(out, standards[selectedIdx].Path, standardsRoot)
}

// StandardItem represents a coding standard.
type StandardItem struct {
	Name string
	Path string
}

func collectStandardsForLanguage(standardsRoot, language string) []StandardItem {
	var standards []StandardItem

	switch language {
	case languageUniversal:
		standards = collectCategoryStandards(standardsRoot, "universal", false)

	case languageGo:
		standards = append(standards, collectCategoryStandards(standardsRoot, "go/core", false)...)
		standards = append(standards, collectCategoryStandards(standardsRoot, "go/log", false)...)
		standards = append(standards, collectCategoryStandards(standardsRoot, "go/testing", false)...)
		standards = append(standards, collectCategoryStandards(standardsRoot, "go/github.com", true)...)
		standards = append(standards, collectCategoryStandards(standardsRoot, "go/gorm.io", true)...)

	case languagePython:
		standards = append(standards, collectCategoryStandards(standardsRoot, "python/core", false)...)
		// Add Python library standards from python/*.md
		pythonRoot := filepath.Join(standardsRoot, "python")
		entries, err := os.ReadDir(pythonRoot)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == config.ReadmeFilename {
					continue
				}
				name := strings.TrimSuffix(entry.Name(), ".md")
				standards = append(standards, StandardItem{
					Name: name,
					Path: filepath.Join(pythonRoot, entry.Name()),
				})
			}
		}
	}

	return standards
}

func collectCategoryStandards(standardsRoot, dir string, recurse bool) []StandardItem {
	var standards []StandardItem
	dirPath := filepath.Join(standardsRoot, dir)

	if recurse {
		_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") || info.Name() == config.ReadmeFilename {
				return nil //nolint:nilerr
			}
			name := strings.TrimSuffix(info.Name(), ".md")
			standards = append(standards, StandardItem{
				Name: name,
				Path: path,
			})
			return nil
		})
	} else {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return standards
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") || entry.Name() == config.ReadmeFilename {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ".md")
			standards = append(standards, StandardItem{
				Name: name,
				Path: filepath.Join(dirPath, entry.Name()),
			})
		}
	}

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

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

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

	relPath, _ := filepath.Rel(standardsRoot, path)
	_, _ = fmt.Fprintf(out, "📄 %s\n", relPath)
	_, _ = fmt.Fprintln(out, strings.Repeat("─", 60))
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, string(content))

	return nil
}
