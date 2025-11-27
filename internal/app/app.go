package app

import (
	"fmt"
	"os"
)

// Execute runs the Bubble Tea menu and dispatches to the matching command.
func Execute() {
	selected, err := RunUI()
	if err != nil {
		fmt.Println("Error running Bubble Tea UI:", err)
		os.Exit(1)
	}

	configPath := "scaffold.config.json"
	sourceRoot := "./"

	switch selected {
	case "init":
		InitCommand(configPath, sourceRoot)
	case "analyze":
		AnalyzeCommand(configPath)
	case "build-template":
		BuildTemplateCommand(configPath, "./template-out")
	case "generate":
		GenerateCommand("./template-out", "./new-app", false, "")
	case "build-generate":
		BuildAndGenerateCommand(configPath, "./template-out", "./new-app", false)
	default:
		fmt.Println("Goodbye!")
	}
}
