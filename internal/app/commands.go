package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultConfigPath  = "scaffold.config.json"
	defaultTemplateOut = "./template-out"
	defaultGenerateOut = "./new-app"
)

var (
	defaultIgnoreFolders = []string{".git", "node_modules", ".vscode", "dist", "coverage", "logs"}
	defaultIgnoreFiles   = []string{"package-lock.json", "yarn.lock", "*.log"}
)

// MatchIgnore checks if a file/folder should be ignored based on config ignore patterns and .scaffoldignore.
func MatchIgnore(path string, ignoreFolders, ignoreFiles []string, scaffoldIgnorePatterns []string) bool {
	base := filepath.Base(path)
	for _, pat := range ignoreFolders {
		if matchGlob(base, pat) {
			return true
		}
	}
	for _, pat := range ignoreFiles {
		if matchGlob(base, pat) {
			return true
		}
	}
	for _, pat := range scaffoldIgnorePatterns {
		if matchGlob(path, pat) {
			return true
		}
	}
	return false
}

// MatchInclude returns true when a path matches an explicit include glob.
func MatchInclude(path string, includePatterns []string) bool {
	for _, pat := range includePatterns {
		if matchGlob(path, pat) {
			return true
		}
	}
	return false
}

func matchGlob(path, pattern string) bool {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return strings.Contains(path, pattern)
	}
	return matched
}

func loadScaffoldIgnore(sourceRoot string) []string {
	filePath := filepath.Join(sourceRoot, ".scaffoldignore")
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func resolveConfigPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return defaultConfigPath
	}
	return path
}

// InitCommand scans the source root, suggests ignore patterns, and writes a starter config.
func InitCommand(configPath, sourceRoot string) {
	if strings.TrimSpace(sourceRoot) == "" {
		sourceRoot = "."
	}
	configPath = resolveConfigPath(configPath)

	fmt.Printf("Initializing scaffold config using source root %s\n", sourceRoot)
	entries, err := os.ReadDir(sourceRoot)
	if err != nil {
		fmt.Println("Error reading source root:", err)
		return
	}
	fmt.Printf("Discovered %d item(s) in %s\n", len(entries), sourceRoot)

	scaffoldIgnore := loadScaffoldIgnore(sourceRoot)
	if len(scaffoldIgnore) > 0 {
		fmt.Printf("Loaded %d pattern(s) from .scaffoldignore\n", len(scaffoldIgnore))
	}

	variables := map[string]Variable{
		"PROJECT_NAME": {
			Type:        "string",
			Required:    true,
			Description: "Human-readable project name",
		},
		"PROJECT_SLUG": {
			Type:        "string",
			Required:    true,
			Description: "kebab-case slug for folder names",
			From:        "PROJECT_NAME",
			Transform:   "slug-kebab",
		},
	}

	cfg := Config{
		SourceRoot:    sourceRoot,
		TemplateRoot:  defaultTemplateOut,
		Token:         map[string]string{"start": "{{", "end": "}}"},
		IgnoreFolders: defaultIgnoreFolders,
		IgnoreFiles:   defaultIgnoreFiles,
		Variables:     variables,
	}

	if err := writeConfig(configPath, &cfg); err != nil {
		fmt.Println("Error writing config file:", err)
		return
	}
	fmt.Printf("Config file written to %s\n", configPath)
}

// AnalyzeCommand prints a terse summary of the resolved config file.
func AnalyzeCommand(configPath string) {
	configPath = resolveConfigPath(configPath)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	fmt.Println("Config summary:")
	fmt.Printf("  SourceRoot: %s\n", cfg.SourceRoot)
	fmt.Printf("  TemplateRoot: %s\n", cfg.TemplateRoot)
	fmt.Printf("  IgnoreFolders (%d): %v\n", len(cfg.IgnoreFolders), cfg.IgnoreFolders)
	fmt.Printf("  IgnoreFiles (%d): %v\n", len(cfg.IgnoreFiles), cfg.IgnoreFiles)
	fmt.Printf("  Variables (%d)\n", len(cfg.Variables))
	for name, variable := range cfg.Variables {
		fmt.Printf("    - %s (type=%s, required=%t)\n", name, variable.Type, variable.Required)
	}
}

// BuildTemplateCommand is a placeholder that reports the planned template build.
func BuildTemplateCommand(configPath, outputPath string) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(outputPath) == "" {
		outputPath = defaultTemplateOut
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	fmt.Printf("Would build template from %s into %s\n", cfg.SourceRoot, outputPath)
}

// GenerateCommand is a placeholder that reports the generation parameters.
func GenerateCommand(templatePath, outPath string) {
	if strings.TrimSpace(templatePath) == "" {
		templatePath = defaultTemplateOut
	}
	if strings.TrimSpace(outPath) == "" {
		outPath = defaultGenerateOut
	}

	fmt.Printf("Would generate new project from %s into %s\n", templatePath, outPath)
}

func writeConfig(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}
