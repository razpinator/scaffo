package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GenerateCommand materializes a new project from a built template.
func GenerateCommand(templatePath, outPath string, copyConfig bool, configPath string) {
	if strings.TrimSpace(templatePath) == "" {
		templatePath = defaultTemplateOut
	}
	if strings.TrimSpace(outPath) == "" {
		outPath = defaultGenerateOut
	}

	var err error
	templatePath, err = filepath.Abs(templatePath)
	if err != nil {
		fmt.Println("Error resolving template path:", err)
		return
	}
	outPath, err = filepath.Abs(outPath)
	if err != nil {
		fmt.Println("Error resolving output path:", err)
		return
	}

	meta, err := loadTemplateMetadata(templatePath)
	if err != nil {
		fmt.Printf("Template metadata missing or invalid: %v\n", err)
		fmt.Println("Ensure you ran build-template before generate.")
		return
	}

	values, err := collectVariableValues(meta.Variables)
	if err != nil {
		fmt.Println("Error collecting variable values:", err)
		return
	}

	// If a "name" variable is provided, use it as the directory name for the new project
	var nameVar string
	if val, ok := values["name"]; ok {
		nameVar = val
	} else if val, ok := values["projectName"]; ok {
		nameVar = val
	} else if val, ok := values["PROJECT_NAME"]; ok {
		nameVar = val
	}

	if strings.TrimSpace(nameVar) != "" {
		outPath = filepath.Join(filepath.Dir(outPath), strings.TrimSpace(nameVar))
	}

	if err := generateProject(templatePath, outPath, meta, values); err != nil {
		fmt.Println("Error generating project:", err)
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

func generateProject(templatePath, outPath string, meta *templateMetadata, values map[string]string) error {
	if _, err := os.Stat(outPath); err == nil {
		return fmt.Errorf("output path %s already exists", outPath)
	}
	if err := os.MkdirAll(outPath, 0o755); err != nil {
		return err
	}

	start, end := defaultTokenDelims(meta.Token)
	var templated, static int
	walkErr := filepath.WalkDir(templatePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == templatePath {
			return nil
		}
		rel, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == templateMetadataFile {
			return nil
		}
		resolvedRel := replaceTokens(rel, values, start, end)
		targetPath := filepath.Join(outPath, filepath.FromSlash(resolvedRel))
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if matchesPatternList(rel, meta.StaticFiles) {
			if err := copyFile(path, targetPath, info.Mode()); err != nil {
				return err
			}
			static++
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := replaceTokens(string(data), values, start, end)
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
