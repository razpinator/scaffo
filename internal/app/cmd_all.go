package app

import (
	"fmt"
	"strings"
)

// BuildAndGenerateCommand chains template building immediately followed by project generation.
func BuildAndGenerateCommand(configPath, templatePath, outPath string, copyConfig bool) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(templatePath) == "" {
		templatePath = defaultTemplateOut
	}
	if strings.TrimSpace(outPath) == "" {
		outPath = defaultGenerateOut
	}

	fmt.Println("Running build-template then generate in a single step...")
	BuildTemplateCommand(configPath, templatePath)
	GenerateCommand(templatePath, outPath, copyConfig, configPath)
}
