package app

import (
	"fmt"
	"os"
	"strings"
)

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
		StaticFiles:   defaultStaticGlobs,
		Variables:     variables,
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Println("Error writing config file:", err)
		return
	}
	fmt.Printf("Config file written to %s\n", configPath)
}
