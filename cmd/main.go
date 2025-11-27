package main

import (
	"flag"
	"fmt"
	"os"

	"scaffo/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
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
	case "analyze":
		var configPath string
		fs := flag.NewFlagSet("analyze", flag.ExitOnError)
		fs.StringVar(&configPath, "config", "scaffold.config.json", "Path to config file")
		mustParse(fs, args)
		app.AnalyzeCommand(configPath)
	case "build-template":
		var configPath, sourceRoot, outputPath string
		fs := flag.NewFlagSet("build-template", flag.ExitOnError)
		fs.StringVar(&configPath, "config", "scaffold.config.json", "Path to config file")
		fs.StringVar(&sourceRoot, "from", "", "Source project root (default: .)")
		fs.StringVar(&outputPath, "output", "", "Output path for template artifacts")
		mustParse(fs, args)
		app.BuildTemplateCommand(configPath, sourceRoot, outputPath)
	case "generate":
		var templatePath, outPath string
		fs := flag.NewFlagSet("generate", flag.ExitOnError)
		fs.StringVar(&templatePath, "template", "", "Path to template directory")
		fs.StringVar(&outPath, "out", "", "Destination for generated project")
		mustParse(fs, args)
		app.GenerateCommand(templatePath, outPath, false, "")
	case "run":
		var configPath, sourceRoot, templatePath, outPath string
		var copyConfig bool
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		fs.StringVar(&configPath, "config", "scaffold.config.json", "Path to config file")
		fs.StringVar(&sourceRoot, "from", "", "Source project root (default: .)")
		fs.StringVar(&templatePath, "output", "", "Intermediary template output path")
		fs.StringVar(&outPath, "out", "", "Destination for generated project")
		fs.BoolVar(&copyConfig, "copy-config", false, "Copy scaffold.config.json to the generated project")
		mustParse(fs, args)
		app.RunCommand(configPath, sourceRoot, templatePath, outPath, copyConfig)
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
	fmt.Println("  analyze --config <path>")
	fmt.Println("  build-template --config <path> [--output <dir>]")
	fmt.Println("  generate --template <dir> --out <dir>")
	fmt.Println("  run --config <path> [--output <tmp>] --out <dir>")
}
