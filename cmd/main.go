package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/razpinator/scaffo/internal/app"
)

const Version = "0.0.4"

func main() {
	if len(os.Args) < 2 {
		app.Execute()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		var configPath, sourceRoot string
		fs := flag.NewFlagSet("init", flag.ExitOnError)
		fs.StringVar(&configPath, "config", "scaffold.config.json", "Path to config file")
		fs.StringVar(&sourceRoot, "from", "", "Source project root")
		mustParse(fs, args)
		app.InitCommand(configPath, sourceRoot)
	case "run":
		var configPath, sourceRoot, outPath string
		var copyConfig bool
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		fs.StringVar(&configPath, "config", "scaffold.config.json", "Path to config file")
		fs.StringVar(&sourceRoot, "from", "", "Source project root (default: .)")
		fs.StringVar(&outPath, "out", "", "Destination for generated project")
		fs.BoolVar(&copyConfig, "copy-config", false, "Copy scaffold.config.json to the generated project")
		mustParse(fs, args)
		app.RunCommand(configPath, sourceRoot, outPath, copyConfig)
	case "version", "--version", "-v":
		fmt.Printf("scaffo version %s\n", Version)
	default:
		fmt.Println("Unknown command:", cmd)
		printUsage()
		os.Exit(1)
	}
}

func mustParse(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println("Usage: scaffo <command> [flags]")
	fmt.Println("Commands:")
	fmt.Println("  init --config <path> --from <source>")
	fmt.Println("  run --config <path> --from <source> --out <dir>")
	fmt.Println("  version")
}
