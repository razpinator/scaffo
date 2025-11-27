package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// InitCommand scans the source root, suggests ignore patterns, and writes a starter config.
func InitCommand(configPath, sourceRoot string) {
	if strings.TrimSpace(sourceRoot) == "" {
		sourceRoot = "."
	}
	configPath = resolveConfigPath(configPath)

	configSourceRoot := sourceRoot
	// Make sourceRoot relative to the config file location
	if absConfig, err := filepath.Abs(configPath); err == nil {
		if absSource, err := filepath.Abs(sourceRoot); err == nil {
			configDir := filepath.Dir(absConfig)
			if rel, err := filepath.Rel(configDir, absSource); err == nil {
				configSourceRoot = rel
			}
		}
	}

	fmt.Printf("Initializing scaffold config using source root %s\n", sourceRoot)
	entries, err := os.ReadDir(sourceRoot)
	if err != nil {
		fmt.Println("Error reading source root:", err)
		return
	}
	fmt.Printf("Discovered %d item(s) in %s\n", len(entries), sourceRoot)

	detectedName := detectProjectName(sourceRoot)
	var replacements []Replacement
	var renameRules []RenameRule

	if detectedName != "" {
		fmt.Printf("Detected project name: %s\n", detectedName)
		replacements = []Replacement{
			{Find: detectedName, ReplaceWith: "{{PROJECT_NAME}}"},
		}
		renameRules = []RenameRule{
			{From: detectedName, To: "{{PROJECT_NAME}}"},
		}
	}

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
		SourceRoot:    configSourceRoot,
		TemplateRoot:  defaultTemplateOut,
		Token:         map[string]string{"start": "{{", "end": "}}"},
		IgnoreFolders: defaultIgnoreFolders,
		IgnoreFiles:   defaultIgnoreFiles,
		StaticFiles:   defaultStaticGlobs,
		Variables:     variables,
		Replacements:  replacements,
		RenameRules:   renameRules,
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Println("Error writing config file:", err)
		return
	}
	fmt.Printf("Config file written to %s\n", configPath)
}

func detectProjectName(root string) string {
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if found != "" {
			return fs.SkipAll
		}
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root {
				name := d.Name()
				for _, ignore := range defaultIgnoreFolders {
					if name == ignore {
						return fs.SkipDir
					}
				}
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".csproj") {
			found = strings.TrimSuffix(d.Name(), ".csproj")
			return fs.SkipAll
		}
		return nil
	})
	return found
}
