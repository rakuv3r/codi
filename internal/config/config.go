// Package config handles codi configuration file management.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/rakuv3r/codi/internal/analyzer"
	"github.com/rakuv3r/codi/internal/generator"
)

// File and directory constants.
const (
	CodiDirName             = ".codi"
	ClaudeDirName           = ".claude"
	StandardsDirName        = "standards"
	ConfigFilename          = "config.yaml"
	ClaudeMDFilename        = "CLAUDE.md"
	ReviewChecklistFilename = "review-checklist.md"
	ReadmeFilename          = "README.md"
)

// Permission constants.
const (
	DirPermission    = 0755 // Directories: owner rwx, group r-x, others r-x
	FilePermission   = 0644 // Public files: owner rw-, group r--, others r--
	ConfigPermission = 0600 // Config files: owner rw-, group ---, others ---
)

// Config represents codi configuration.
type Config struct {
	LastSync     time.Time                `yaml:"last_sync"`
	Language     analyzer.Language        `yaml:"language"`
	Project      ProjectConfig            `yaml:"project"`
	Dependencies DependencyConfig         `yaml:"dependencies"`
	AITools      []generator.AIToolConfig `yaml:"ai_tools"`
}

// ProjectConfig contains project-specific configuration.
type ProjectConfig struct {
	Type           analyzer.ProjectType `yaml:"type"`
	HTTPFramework  string               `yaml:"http_framework,omitempty"`
	Database       string               `yaml:"database,omitempty"`
	DatabaseDriver string               `yaml:"database_driver,omitempty"`
	Features       []string             `yaml:"features,omitempty"`
}

// DependencyConfig contains dependency configuration.
type DependencyConfig struct {
	Dependencies []string `yaml:"dependencies"`
}

// Load loads configuration from file.
func Load(projectPath string) (*Config, error) {
	configPath := filepath.Join(projectPath, CodiDirName, ConfigFilename)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Save saves configuration to file.
func Save(projectPath string, cfg *Config) error {
	codiDir := filepath.Join(projectPath, CodiDirName)
	if err := os.MkdirAll(codiDir, DirPermission); err != nil {
		return fmt.Errorf("failed to create .codi directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(codiDir, ConfigFilename)
	if err := os.WriteFile(configPath, data, ConfigPermission); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// NewConfig creates a new configuration.
func NewConfig(info *analyzer.ProjectInfo, deps *analyzer.DependencyList) *Config {
	return NewConfigWithTools(info, deps, generator.DefaultAITools())
}

// NewConfigWithTools creates a new configuration with custom AI tools.
func NewConfigWithTools(info *analyzer.ProjectInfo, deps *analyzer.DependencyList, aiTools []generator.AIToolConfig) *Config {
	cfg := &Config{
		Language: info.Language,
		Project: ProjectConfig{
			Type:           info.ProjectType,
			HTTPFramework:  info.HTTPFramework,
			Database:       info.Database,
			DatabaseDriver: info.DatabaseDriver,
			Features:       info.Features,
		},
		Dependencies: DependencyConfig{
			Dependencies: make([]string, 0, len(deps.Dependencies)),
		},
		AITools:  aiTools,
		LastSync: time.Now(),
	}

	// Convert dependencies to string paths
	for _, dep := range deps.Dependencies {
		cfg.Dependencies.Dependencies = append(cfg.Dependencies.Dependencies, dep.Path)
	}

	return cfg
}
