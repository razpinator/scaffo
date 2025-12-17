package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunCommand scaffolds a project directly from source to output without an intermediate build step.
func RunCommand(configPath, sourceRoot, outPath string, copyConfig bool) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(outPath) == "" {
		outPath = defaultGenerateOut
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Override source root if provided
	if strings.TrimSpace(sourceRoot) != "" {
		cfg.SourceRoot = sourceRoot
	}

	// Resolve SourceRoot
	sourceRoot, err = filepath.Abs(cfg.SourceRoot)
	if err != nil {
		fmt.Println("Error resolving source root:", err)
		return
	}

	// Collect variables
	values, err := collectVariableValues(cfg.Variables)
	if err != nil {
		fmt.Println("Error collecting variable values:", err)
		return
	}

	// Determine output path based on name variable if present
	var nameVar string
	if val, ok := values["name"]; ok {
		nameVar = val
	} else if val, ok := values["projectName"]; ok {
		nameVar = val
	} else if val, ok := values["PROJECT_NAME"]; ok {
		nameVar = val
	}

	if strings.TrimSpace(nameVar) != "" {
		// If outPath was just "." or default, append the name
		// If outPath ends with the name already, don't append?
		// The original logic was: outPath = filepath.Join(filepath.Dir(outPath), nameVar)
		// But that assumes outPath was a placeholder.
		// Let's stick to the original logic if it makes sense.
		// Original: outPath = filepath.Join(filepath.Dir(outPath), strings.TrimSpace(nameVar))
		// This effectively replaces the last segment of outPath with nameVar.
		// If user said --out ./my-new-app, and name is "cool-app", it becomes ./cool-app.
		// That seems a bit aggressive if the user explicitly set --out.
		// But let's keep it consistent with previous behavior for now, or maybe improve it.
		// If the user didn't specify --out (it's default), then we definitely want to use the name.
		// If the user DID specify --out, maybe we should respect it?
		// The original code did:
		// if strings.TrimSpace(outPath) == "" { outPath = defaultGenerateOut }
		// ...
		// if nameVar != "" { outPath = filepath.Join(filepath.Dir(outPath), nameVar) }
		// This overrides the last part of outPath.
		// I'll keep it for compatibility/behavior preservation.
		outPath = filepath.Join(filepath.Dir(outPath), strings.TrimSpace(nameVar))
	}

	outPath, err = filepath.Abs(outPath)
	if err != nil {
		fmt.Println("Error resolving output path:", err)
		return
	}

	if _, err := os.Stat(outPath); err == nil {
		fmt.Printf("Output path %s already exists\n", outPath)
		return
	}

	fmt.Printf("Scaffolding from %s to %s...\n", sourceRoot, outPath)

	// Generate variations for automatic replacement
	sourceName := filepath.Base(sourceRoot)
	targetName := filepath.Base(outPath)

	fmt.Printf("Detecting variations: %s -> %s\n", sourceName, targetName)

	sourceVars := generateVariations(sourceName)
	targetVars := generateVariations(targetName)

	for key, srcVal := range sourceVars {
		if tgtVal, ok := targetVars[key]; ok {
			if srcVal == tgtVal {
				continue
			}
			// Add to Replacements
			cfg.Replacements = append(cfg.Replacements, Replacement{
				Find:        srcVal,
				ReplaceWith: tgtVal,
			})
			// Add to RenameRules
			cfg.RenameRules = append(cfg.RenameRules, RenameRule{
				From: srcVal,
				To:   tgtVal,
			})
		}
	}

	// Sort replacements by length of Find string (descending) to avoid partial matches
	sort.Slice(cfg.Replacements, func(i, j int) bool {
		return len(cfg.Replacements[i].Find) > len(cfg.Replacements[j].Find)
	})
	sort.Slice(cfg.RenameRules, func(i, j int) bool {
		return len(cfg.RenameRules[i].From) > len(cfg.RenameRules[j].From)
	})

	if err := scaffoldProject(cfg, sourceRoot, outPath, values); err != nil {
		fmt.Println("Error scaffolding project:", err)
		return
	}

	if copyConfig && configPath != "" {
		src, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("Warning: Could not read config file to copy: %v\n", err)
		} else {
			dstPath := filepath.Join(outPath, filepath.Base(configPath))
			if err := os.WriteFile(dstPath, src, 0644); err != nil {
				fmt.Printf("Warning: Could not write config file: %v\n", err)
			} else {
				fmt.Printf("Copied config file to %s\n", dstPath)
			}
		}
	}

	fmt.Printf("Project generated at %s\n", outPath)
}

func scaffoldProject(cfg *Config, sourceRoot, outPath string, values map[string]string) error {
	if err := os.MkdirAll(outPath, 0o755); err != nil {
		return err
	}

	scaffoldIgnore := loadScaffoldIgnore(sourceRoot)
	start, end := defaultTokenDelims(cfg.Token)

	// Ensure we don't copy the config file itself if it's in sourceRoot
	// (Though usually config is outside or we want to ignore it)
	// cfg.IgnoreFiles = append(cfg.IgnoreFiles, filepath.Base(configPath)) // We don't have configPath here easily, but it's fine.

	var templated, static int
	walkErr := filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sourceRoot {
			return nil
		}

		// Skip the output directory if it's inside sourceRoot (to avoid infinite recursion)
		if path == outPath || strings.HasPrefix(path, outPath+string(filepath.Separator)) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if MatchIgnore(rel, d.IsDir(), cfg.IgnoreFolders, cfg.IgnoreFiles, scaffoldIgnore) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// 1. Apply RenameRules
		renamedRel := applyRenameRules(rel, cfg.RenameRules)

		// 2. Apply Token Replacement to path
		resolvedRel := replaceTokens(renamedRel, values, start, end)

		targetPath := filepath.Join(outPath, filepath.FromSlash(resolvedRel))

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		// Ensure parent dir exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		// Check if static
		if matchesPatternList(rel, cfg.StaticFiles) {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if err := copyFile(path, targetPath, info.Mode()); err != nil {
				return err
			}
			static++
			return nil
		}

		// Templated file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(data)

		// Apply Config.Replacements first (e.g. "Banker" -> "{{PROJECT_NAME}}")
		for _, repl := range cfg.Replacements {
			if repl.Find == "" {
				continue
			}
			content = strings.ReplaceAll(content, repl.Find, repl.ReplaceWith)
		}

		content = replaceTokens(content, values, start, end)

		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, []byte(content), info.Mode()); err != nil {
			return err
		}
		templated++
		return nil
	})

	if walkErr != nil {
		return walkErr
	}
	fmt.Printf("Created %d templated file(s) and %d static asset(s)\n", templated, static)
	return nil
}
