package app

import (
	"fmt"
)

func InitCommand(configPath, sourceRoot string) {
	fmt.Println("Initializing scaffold config...")
	// TODO: Scan sourceRoot, suggest ignore patterns, variables, and write config file
}

func AnalyzeCommand(configPath string) {
	fmt.Println("Analyzing project for variables and structure...")
	// TODO: Print folder/file counts, candidate variables, conflicts
}

func BuildTemplateCommand(configPath, outputPath string) {
	fmt.Println("Building template skeleton...")
	// TODO: Execute conversion logic, variable injection, renaming, copying
}

func GenerateCommand(templatePath, outPath string) {
	fmt.Println("Generating new project from template...")
	// TODO: Prompt for variables, copy files, apply replacements
}
