package config

import "github.com/rakuv3r/codi/internal/analyzer"

// QualityFiles returns quality file mappings for a language.
// Maps source template file names to destination file names.
func QualityFiles(lang analyzer.Language) map[string]string {
	base := map[string]string{
		"Makefile.tmpl":            "Makefile",
		".gitignore":               ".gitignore",
		"hooks/pre-commit":         "hooks/pre-commit",
		"scripts/install_hooks.sh": "scripts/install_hooks.sh",
	}

	// Add language-specific files
	if lang == analyzer.LanguageGo {
		base[".golangci.yml.tmpl"] = ".golangci.yml"
	}

	return base
}
