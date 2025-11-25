package app

import (
	"fmt"
	"os"
)

func Execute() {
	selected, err := RunUI()
	if err != nil {
		fmt.Println("Error running Bubble Tea UI:", err)
		os.Exit(1)
	}
	switch selected {
	case "init":
		InitCommand("scaffold.config.json", "./")
	case "analyze":
		AnalyzeCommand("scaffold.config.json")
	case "build-template":
		BuildTemplateCommand("scaffold.config.json", "./template-out")
	case "generate":
		GenerateCommand("./template-out", "./new-app")
	default:
		fmt.Println("Goodbye!")
	}
}
