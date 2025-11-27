package app

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// AnalyzeCommand summarizes the config and reports basic file counts.
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
	fmt.Printf("  StaticFiles (%d)\n", len(cfg.StaticFiles))
	fmt.Printf("  Variables (%d)\n", len(cfg.Variables))
	for name, variable := range cfg.Variables {
		fmt.Printf("    - %s (type=%s, required=%t)\n", name, variable.Type, variable.Required)
	}

	scaffoldIgnore := loadScaffoldIgnore(cfg.SourceRoot)
	var included, skipped int
	_ = filepath.WalkDir(cfg.SourceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == cfg.SourceRoot {
			return nil
		}
		rel, _ := filepath.Rel(cfg.SourceRoot, path)
		if MatchIgnore(rel, d.IsDir(), cfg.IgnoreFolders, cfg.IgnoreFiles, scaffoldIgnore) {
			skipped++
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			included++
		}
		return nil
	})
	fmt.Printf("  Files to template: %d (skipped %d)\n", included, skipped)
}
