package app

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BuildTemplateCommand converts the source project into a reusable template tree.
func BuildTemplateCommand(configPath, sourceRoot, outputPath string) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(outputPath) == "" {
		outputPath = defaultTemplateOut
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

	// Ensure the config file itself is not copied into the template
	cfg.IgnoreFiles = append(cfg.IgnoreFiles, filepath.Base(configPath))

	if err := buildTemplate(cfg, outputPath); err != nil {
		fmt.Println("Error building template:", err)
		return
	}
	fmt.Printf("Template written to %s\n", outputPath)
}

func buildTemplate(cfg *Config, outputPath string) error {
	sourceRoot, err := filepath.Abs(cfg.SourceRoot)
	if err != nil {
		return err
	}
	outputPath, err = filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	outputPath = filepath.Clean(outputPath)
	if err := os.RemoveAll(outputPath); err != nil {
		return err
	}
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return err
	}
	outputPathWithSep := outputPath + string(filepath.Separator)

	scaffoldIgnore := loadScaffoldIgnore(sourceRoot)
	staticGlobs := mergePatterns(defaultStaticGlobs, cfg.StaticFiles)

	templatedCount := 0
	staticCount := 0
	fileCount := 0

	walkErr := filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == outputPath || strings.HasPrefix(path, outputPathWithSep) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if path == sourceRoot {
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
		targetRel := applyRenameRules(rel, cfg.RenameRules)
		targetPath := filepath.Join(outputPath, filepath.FromSlash(targetRel))
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		fileCount++
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		isStatic := matchesPatternList(rel, staticGlobs)
		if !isStatic {
			binary, err := looksBinary(path)
			if err != nil {
				return err
			}
			isStatic = binary
		}
		if isStatic {
			if err := copyFile(path, targetPath, info.Mode()); err != nil {
				return err
			}
			staticCount++
			return nil
		}
		if err := processTemplatedFile(path, targetPath, info.Mode(), cfg.Replacements); err != nil {
			return err
		}
		templatedCount++
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	meta := templateMetadata{
		Token:       cfg.Token,
		Variables:   cfg.Variables,
		StaticFiles: staticGlobs,
		GeneratedAt: time.Now().UTC(),
	}
	if err := writeTemplateMetadata(outputPath, &meta); err != nil {
		return err
	}
	fmt.Printf("Copied %d file(s): %d templated, %d static\n", fileCount, templatedCount, staticCount)
	return nil
}
