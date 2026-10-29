// Package setup provides shared project setup logic for init, adopt, and sync commands.
package setup

import (
	"io"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/generator"
)

// Options configures project setup behavior.
type Options struct {
	Output             io.Writer
	ProjectInfo        *analyzer.ProjectInfo
	Dependencies       *analyzer.DependencyList
	ProjectPath        string
	AITools            []generator.AIToolConfig
	CopyStandards      bool
	CopyReviewTemplate bool
	AddQualityTools    bool
	InstallGitHooks    bool
	PromptForSetup     bool
}
