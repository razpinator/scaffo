package app

import (
	"fmt"
	"strings"
)

// RunCommand chains template building immediately followed by project generation.
func RunCommand(configPath, sourceRoot, templatePath, outPath string, copyConfig bool) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(templatePath) == "" {
		templatePath = defaultTemplateOut
	}
	if strings.TrimSpace(outPath) == "" {
		outPath = defaultGenerateOut
	}

	fmt.Println("Running build-template then generate in a single step...")
	BuildTemplateCommand(configPath, sourceRoot, templatePath)
	GenerateCommand(templatePath, outPath, copyConfig, configPath)
}
