// Package generator provides project scaffolding generation from templates.
package generator

// AITool represents an AI coding assistant.
type AITool string

// AI tool identifiers.
const (
	// AIToolClaude represents Claude AI assistant.
	AIToolClaude AITool = "claude"
	// AIToolCodex represents OpenAI Codex CLI.
	AIToolCodex AITool = "codex"
)

func (at AITool) String() string {
	return string(at)
}

// AIToolConfig contains AI tool configuration.
type AIToolConfig struct {
	Name       AITool `yaml:"name"`
	OutputPath string `yaml:"output_path"`
	Enabled    bool   `yaml:"enabled"`
}

// AvailableAITools returns all available AI tool configurations.
func AvailableAITools() []AIToolConfig {
	return []AIToolConfig{
		{
			Name:       AIToolClaude,
			OutputPath: ".claude/CLAUDE.md",
		},
		{
			Name:       AIToolCodex,
			OutputPath: ".codex/AGENTS.md",
		},
	}
}

// DefaultAITools returns default AI tool configurations (non-interactive mode).
func DefaultAITools() []AIToolConfig {
	return []AIToolConfig{
		{
			Name:       AIToolClaude,
			Enabled:    true,
			OutputPath: ".claude/CLAUDE.md",
		},
	}
}
