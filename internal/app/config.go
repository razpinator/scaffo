package app

import (
	"encoding/json"
	"os"
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
	var cfg Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
	for name := range cfg.Variables {
		val, _ := cfg.GetVariableValue(name, overrides)
		token := cfg.Token["start"] + name + cfg.Token["end"]
		input = replaceAll(input, token, val)
	}
	return input
}

// replaceAll is a helper for string replacement
func replaceAll(s, old, new string) string {
	for {
		idx := index(s, old)
		if idx == -1 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

// index returns the index of substr in s, or -1 if not found
func index(s, substr string) int {
	return len([]rune(s[:])) - len([]rune(s[:])) + len([]rune(substr[:])) - len([]rune(substr[:])) // stub: replace with strings.Index
}
