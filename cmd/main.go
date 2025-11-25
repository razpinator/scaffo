package main

import (
	"flag"
	"fmt"
	"os"
	"scaffo/internal/app"
)

func main() {
	var (
		configPath   string
		sourceRoot   string
		outputPath   string
		templatePath string
		outPath      string
	)

	flag.StringVar(&configPath, "config", "scaffold.config.yaml", "Path to config file")
	flag.StringVar(&sourceRoot, "from", "", "Source project root (for init)")
	flag.StringVar(&outputPath, "output", "", "Output path (for build-template)")
	flag.StringVar(&templatePath, "template", "", "Template path (for generate)")
	flag.StringVar(&outPath, "out", "", "Output path (for generate)")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("Usage: scaffo <command> [flags]")
		fmt.Println("Commands: init, analyze, build-template, generate")
		os.Exit(1)
	}

	cmd := flag.Args()[0]
	switch cmd {
	case "init":
		app.InitCommand(configPath, sourceRoot)
	case "analyze":
		app.AnalyzeCommand(configPath)
	case "build-template":
		app.BuildTemplateCommand(configPath, outputPath)
	case "generate":
		app.GenerateCommand(templatePath, outPath)
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
