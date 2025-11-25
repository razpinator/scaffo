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
