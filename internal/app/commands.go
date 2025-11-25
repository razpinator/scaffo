package app

import (
	"fmt"
)

func InitCommand(configPath, sourceRoot string) {
	// InitCommand scans the source root, suggests ignore patterns, variables, and writes config file
	fmt.Println("Initializing scaffold config...")
	// 1. Scan sourceRoot for folders/files
	// 2. Suggest ignore patterns and candidate variables
	// 3. Write scaffold.config.yaml/json/toml
	// Stub: Just print what would be done
	fmt.Printf("Scan: %s\nWrite config: %s\n", sourceRoot, configPath)
}

func AnalyzeCommand(configPath string) {
	// AnalyzeCommand prints folder/file counts, candidate variables, and conflicts
	fmt.Println("Analyzing project for variables and structure...")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Printf("SourceRoot: %s\nTemplateRoot: %s\n", cfg.SourceRoot, cfg.TemplateRoot)
	fmt.Printf("IgnoreFolders: %v\nIgnoreFiles: %v\n", cfg.IgnoreFolders, cfg.IgnoreFiles)
	fmt.Printf("Variables: %v\n", cfg.Variables)
	// Stub: Print config summary
}

func BuildTemplateCommand(configPath, outputPath string) {
	// BuildTemplateCommand executes conversion logic, variable injection, renaming, copying
	fmt.Println("Building template skeleton...")
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Printf("Would build template from %s to %s\n", cfg.SourceRoot, outputPath)
	// Stub: Print what would be done
}

func GenerateCommand(templatePath, outPath string) {
	// GenerateCommand prompts for variables, copies files, applies replacements
	fmt.Println("Generating new project from template...")
	fmt.Printf("Would generate new project from %s to %s\n", templatePath, outPath)
	// Stub: Print what would be done
}
