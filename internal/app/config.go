package app

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Variable struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
	From        string `json:"from,omitempty"`
	Transform   string `json:"transform,omitempty"`
}

type Replacement struct {
	Find        string `json:"find"`
	ReplaceWith string `json:"replaceWith"`
}

type RenameRule struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Hook struct {
	Command string `json:"command"`
	Cwd     string `json:"cwd"`
}

type Config struct {
	SourceRoot    string              `json:"sourceRoot"`
	TemplateRoot  string              `json:"templateRoot"`
	Token         map[string]string   `json:"token"`
	IgnoreFolders []string            `json:"ignoreFolders"`
	IgnoreFiles   []string            `json:"ignoreFiles"`
	StaticFiles   []string            `json:"staticFiles"`
	Variables     map[string]Variable `json:"variables"`
	Replacements  []Replacement       `json:"replacements"`
	RenameRules   []RenameRule        `json:"renameRules"`
	Hooks         map[string][]Hook   `json:"hooks"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg Config
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	cfg.applyDefaults()
	return &cfg, nil
}

func (cfg *Config) Save(path string) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	var (
		data []byte
		err  error
	)
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
	default:
		data, err = json.MarshalIndent(cfg, "", "  ")
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// GetVariableValue returns the value for a variable, using override or default if not set
func (cfg *Config) GetVariableValue(name string, overrides map[string]string) (string, bool) {
	v, ok := cfg.Variables[name]
	if !ok {
		return "", false
	}
	// Use override if provided
	if val, ok := overrides[name]; ok {
		return val, true
	}
	// Use default if available
	if v.Default != "" {
		return v.Default, true
	}
	return "", !v.Required
}

// ApplyVariableReplacements replaces all variable tokens in a string
func (cfg *Config) ApplyVariableReplacements(input string, overrides map[string]string) string {
	start, end := cfg.tokenDelimiters()
	for name := range cfg.Variables {
		val, _ := cfg.GetVariableValue(name, overrides)
		token := start + name + end
		input = strings.ReplaceAll(input, token, val)
	}
	return input
}

func (cfg *Config) tokenDelimiters() (string, string) {
	start := "{{"
	end := "}}"
	if cfg.Token != nil {
		if v := strings.TrimSpace(cfg.Token["start"]); v != "" {
			start = v
		}
		if v := strings.TrimSpace(cfg.Token["end"]); v != "" {
			end = v
		}
	}
	return start, end
}

func (cfg *Config) applyDefaults() {
	if cfg.Token == nil {
		cfg.Token = map[string]string{"start": "{{", "end": "}}"}
	}
	if strings.TrimSpace(cfg.SourceRoot) == "" {
		cfg.SourceRoot = "."
	}
	if strings.TrimSpace(cfg.TemplateRoot) == "" {
		cfg.TemplateRoot = defaultTemplateOut
	}
	if len(cfg.IgnoreFolders) == 0 {
		cfg.IgnoreFolders = append([]string{}, defaultIgnoreFolders...)
	}
	if len(cfg.IgnoreFiles) == 0 {
		cfg.IgnoreFiles = append([]string{}, defaultIgnoreFiles...)
	}
	if len(cfg.StaticFiles) == 0 {
		cfg.StaticFiles = append([]string{}, defaultStaticGlobs...)
	}
	if cfg.Variables == nil {
		cfg.Variables = map[string]Variable{}
	}
}
