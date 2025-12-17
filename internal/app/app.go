package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Execute runs the Bubble Tea menu and dispatches to the matching command.
func Execute() {
	for {
		selected, arg, err := RunUI()
		if err != nil {
			fmt.Println("Error running Bubble Tea UI:", err)
			os.Exit(1)
		}

		if selected == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		configPath := "scaffold.config.json"
		sourceRoot := "./"

		switch selected {
		case "init":
			InitCommand(configPath, sourceRoot)
			pause()
		case "run":
			if arg != "" {
				if arg == "Other (enter path)" {
					fmt.Print("Enter source path: ")
					reader := bufio.NewReader(os.Stdin)
					text, _ := reader.ReadString('\n')
					sourceRoot = strings.TrimSpace(text)
				} else {
					sourceRoot = arg
				}
			}
			// RunCommand handles default outPath and prompting for variables
			RunCommand(configPath, sourceRoot, "", false)
			pause()
		default:
			// Should not happen if RunUI returns valid commands or quit
			fmt.Println("Goodbye!")
			return
		}
	}
}

func pause() {
	fmt.Println("\nPress Enter to return to menu...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
