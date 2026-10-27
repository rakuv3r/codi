package generator

// ProjectConfig represents user's project configuration choices.
type ProjectConfig struct {
	ProjectName string
	Description string
	Module      string
	Language    string
	Base        string
	EntryPoint  string
	Modules     []string
	AITools     []AIToolConfig
	InPlace     bool
}

// TemplateData is passed to template rendering.
type TemplateData struct {
	ProjectName string
	Description string
	Module      string
	EntryPoint  string
}

// ModulesManifest defines available modules and their configuration.
type ModulesManifest struct {
	Bases   map[string]BaseConfig   `yaml:"bases"`
	Modules map[string]ModuleConfig `yaml:"modules"`
}

// BaseConfig defines a base project type.
type BaseConfig struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Structure   string `yaml:"structure"`
	EntryPoint  string `yaml:"entry_point"`
}

// ModuleConfig defines a module's configuration.
type ModuleConfig struct {
	Category    string   `yaml:"category"`
	Base        string   `yaml:"base"`
	Description string   `yaml:"description"`
	Replaces    []string `yaml:"replaces"`
	Adds        []string `yaml:"adds"`
	GoPackages  []string `yaml:"go_packages"`
	Standards   []string `yaml:"standards"`
}
