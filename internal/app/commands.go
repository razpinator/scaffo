package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"
)

const (
	defaultConfigPath    = "scaffold.config.json"
	defaultTemplateOut   = "./template-out"
	defaultGenerateOut   = "./new-app"
	templateMetadataFile = ".scaffo-template.json"
)

var (
	defaultIgnoreFolders = []string{".git", "node_modules", ".vscode", "dist", "coverage", "logs"}
	defaultIgnoreFiles   = []string{"package-lock.json", "yarn.lock", "*.log"}
	defaultStaticGlobs   = []string{"**/*.png", "**/*.jpg", "**/*.jpeg", "**/*.gif", "**/*.ico", "**/*.webp", "**/*.svg", "**/*.ttf", "**/*.otf", "**/*.woff", "**/*.woff2", "**/*.pdf", "**/*.zip", "**/*.tar", "**/*.gz", "**/*.mp3", "**/*.mp4"}
)

type templateMetadata struct {
	Token       map[string]string   `json:"token"`
	Variables   map[string]Variable `json:"variables"`
	StaticFiles []string            `json:"staticFiles"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

// MatchIgnore checks if a file/folder should be ignored based on config ignore patterns and .scaffoldignore.
func MatchIgnore(path string, isDir bool, ignoreFolders, ignoreFiles []string, scaffoldIgnorePatterns []string) bool {
	path = filepath.ToSlash(path)
	base := filepath.Base(path)
	if isDir {
		for _, pat := range ignoreFolders {
			if matchGlob(base, pat) || matchGlob(path, pat) {
				return true
			}
		}
	} else {
		for _, pat := range ignoreFiles {
			if matchGlob(base, pat) || matchGlob(path, pat) {
				return true
			}
		}
		// Check if any parent directory matches ignoreFolders
		parts := strings.Split(path, "/")
		for i := 0; i < len(parts)-1; i++ {
			subPath := strings.Join(parts[:i+1], "/")
			baseName := parts[i]
			for _, pat := range ignoreFolders {
				if matchGlob(baseName, pat) || matchGlob(subPath, pat) {
					return true
				}
			}
		}
	}
	for _, pat := range scaffoldIgnorePatterns {
		if matchGlob(path, pat) || matchGlob(base, pat) {
			return true
		}
	}
	return false
}

// MatchInclude returns true when a path matches an explicit include glob.
func MatchInclude(path string, includePatterns []string) bool {
	path = filepath.ToSlash(path)
	for _, pat := range includePatterns {
		if matchGlob(path, pat) {
			return true
		}
	}
	return false
}

func matchGlob(path, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)
	matched, err := doublestar.Match(pattern, path)
	if err != nil {
		return strings.Contains(path, pattern)
	}
	return matched
}

func loadScaffoldIgnore(sourceRoot string) []string {
	filePath := filepath.Join(sourceRoot, ".scaffoldignore")
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, filepath.ToSlash(line))
	}
	return patterns
}

func resolveConfigPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return defaultConfigPath
	}
	return path
}

// InitCommand scans the source root, suggests ignore patterns, and writes a starter config.
func InitCommand(configPath, sourceRoot string) {
	if strings.TrimSpace(sourceRoot) == "" {
		sourceRoot = "."
	}
	configPath = resolveConfigPath(configPath)

	fmt.Printf("Initializing scaffold config using source root %s\n", sourceRoot)
	entries, err := os.ReadDir(sourceRoot)
	if err != nil {
		fmt.Println("Error reading source root:", err)
		return
	}
	fmt.Printf("Discovered %d item(s) in %s\n", len(entries), sourceRoot)

	scaffoldIgnore := loadScaffoldIgnore(sourceRoot)
	if len(scaffoldIgnore) > 0 {
		fmt.Printf("Loaded %d pattern(s) from .scaffoldignore\n", len(scaffoldIgnore))
	}

	variables := map[string]Variable{
		"PROJECT_NAME": {
			Type:        "string",
			Required:    true,
			Description: "Human-readable project name",
		},
		"PROJECT_SLUG": {
			Type:        "string",
			Required:    true,
			Description: "kebab-case slug for folder names",
			From:        "PROJECT_NAME",
			Transform:   "slug-kebab",
		},
	}

	cfg := Config{
		SourceRoot:    sourceRoot,
		TemplateRoot:  defaultTemplateOut,
		Token:         map[string]string{"start": "{{", "end": "}}"},
		IgnoreFolders: defaultIgnoreFolders,
		IgnoreFiles:   defaultIgnoreFiles,
		StaticFiles:   defaultStaticGlobs,
		Variables:     variables,
	}

	if err := cfg.Save(configPath); err != nil {
		fmt.Println("Error writing config file:", err)
		return
	}
	fmt.Printf("Config file written to %s\n", configPath)
}

// AnalyzeCommand summarizes the config and reports basic file counts.
func AnalyzeCommand(configPath string) {
	configPath = resolveConfigPath(configPath)
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	fmt.Println("Config summary:")
	fmt.Printf("  SourceRoot: %s\n", cfg.SourceRoot)
	fmt.Printf("  TemplateRoot: %s\n", cfg.TemplateRoot)
	fmt.Printf("  IgnoreFolders (%d): %v\n", len(cfg.IgnoreFolders), cfg.IgnoreFolders)
	fmt.Printf("  IgnoreFiles (%d): %v\n", len(cfg.IgnoreFiles), cfg.IgnoreFiles)
	fmt.Printf("  StaticFiles (%d)\n", len(cfg.StaticFiles))
	fmt.Printf("  Variables (%d)\n", len(cfg.Variables))
	for name, variable := range cfg.Variables {
		fmt.Printf("    - %s (type=%s, required=%t)\n", name, variable.Type, variable.Required)
	}

	scaffoldIgnore := loadScaffoldIgnore(cfg.SourceRoot)
	var included, skipped int
	_ = filepath.WalkDir(cfg.SourceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == cfg.SourceRoot {
			return nil
		}
		rel, _ := filepath.Rel(cfg.SourceRoot, path)
		if MatchIgnore(rel, d.IsDir(), cfg.IgnoreFolders, cfg.IgnoreFiles, scaffoldIgnore) {
			skipped++
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			included++
		}
		return nil
	})
	fmt.Printf("  Files to template: %d (skipped %d)\n", included, skipped)
}

// BuildTemplateCommand converts the source project into a reusable template tree.
func BuildTemplateCommand(configPath, outputPath string) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(outputPath) == "" {
		outputPath = defaultTemplateOut
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
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

// BuildAndGenerateCommand chains template building immediately followed by project generation.
func BuildAndGenerateCommand(configPath, templatePath, outPath string, copyConfig bool) {
	configPath = resolveConfigPath(configPath)
	if strings.TrimSpace(templatePath) == "" {
		templatePath = defaultTemplateOut
	}
	if strings.TrimSpace(outPath) == "" {
		outPath = defaultGenerateOut
	}

	fmt.Println("Running build-template then generate in a single step...")
	BuildTemplateCommand(configPath, templatePath)
	GenerateCommand(templatePath, outPath, copyConfig, configPath)
}

func applyRenameRules(rel string, rules []RenameRule) string {
	if len(rules) == 0 {
		return rel
	}
	for _, rule := range rules {
		from := strings.TrimPrefix(filepath.ToSlash(rule.From), "./")
		if from == "" {
			continue
		}
		to := filepath.ToSlash(rule.To)
		if rel == from {
			return to
		}
		if strings.HasPrefix(rel, from+"/") {
			suffix := strings.TrimPrefix(rel, from)
			return to + suffix
		}
	}
	return rel
}

func matchesPatternList(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, pat := range patterns {
		if matchGlob(path, pat) {
			return true
		}
	}
	return false
}

func mergePatterns(base, overrides []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, list := range [][]string{base, overrides} {
		for _, pat := range list {
			pat = strings.TrimSpace(pat)
			if pat == "" {
				continue
			}
			if _, ok := seen[pat]; ok {
				continue
			}
			seen[pat] = struct{}{}
			result = append(result, pat)
		}
	}
	return result
}

func copyFile(src, dest string, perm fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func processTemplatedFile(src, dest string, perm fs.FileMode, replacements []Replacement) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	content := string(data)
	for _, repl := range replacements {
		if repl.Find == "" {
			continue
		}
		content = strings.ReplaceAll(content, repl.Find, repl.ReplaceWith)
	}
	return os.WriteFile(dest, []byte(content), perm)
}

func looksBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	buf := make([]byte, 1024)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}

func collectVariableValues(vars map[string]Variable) (map[string]string, error) {
	values := make(map[string]string, len(vars))
	reader := bufio.NewReader(os.Stdin)
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		v := vars[name]
		if envVal, ok := os.LookupEnv("SCAFFO_" + name); ok && strings.TrimSpace(envVal) != "" {
			values[name] = envVal
			continue
		}
		if v.From != "" {
			continue
		}
		prompt := v.Description
		if prompt == "" {
			prompt = "Enter value"
		}
		defaultHint := ""
		if v.Default != "" {
			defaultHint = " [" + v.Default + "]"
		}
		for {
			fmt.Printf("%s (%s)%s: ", name, prompt, defaultHint)
			text, err := reader.ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}
			text = strings.TrimSpace(text)
			if text == "" {
				text = v.Default
			}
			if text == "" && v.Required {
				fmt.Println("Value required.")
				continue
			}
			values[name] = text
			break
		}
	}

	for _, name := range keys {
		if _, ok := values[name]; ok {
			continue
		}
		v := vars[name]
		if v.From != "" {
			if source, ok := values[v.From]; ok && source != "" {
				values[name] = applyTransform(source, v.Transform)
				continue
			}
		}
		if v.Default != "" {
			values[name] = v.Default
			continue
		}
		if v.Required {
			return nil, fmt.Errorf("missing value for %s", name)
		}
		values[name] = ""
	}

	return values, nil
}

func applyTransform(input, transform string) string {
	switch strings.ToLower(transform) {
	case "", "identity":
		return input
	case "slug-kebab":
		return slugify(input, '-')
	case "slug-snake":
		return slugify(input, '_')
	case "upper":
		return strings.ToUpper(input)
	case "lower":
		return strings.ToLower(input)
	case "title":
		return titleCase(input)
	default:
		return input
	}
}

func slugify(input string, sep rune) string {
	var b strings.Builder
	lastWasSep := true
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastWasSep = false
			continue
		}
		if !lastWasSep {
			b.WriteRune(sep)
			lastWasSep = true
		}
	}
	res := b.String()
	return strings.Trim(res, string(sep))
}

func titleCase(input string) string {
	words := strings.Fields(strings.ToLower(input))
	for i, word := range words {
		if word == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(word)
		if r == utf8.RuneError && size == 0 {
			continue
		}
		words[i] = string(unicode.ToUpper(r)) + word[size:]
	}
	return strings.Join(words, " ")
}

func replaceTokens(input string, values map[string]string, start, end string) string {
	for name, value := range values {
		token := start + name + end
		input = strings.ReplaceAll(input, token, value)
	}
	return input
}

func defaultTokenDelims(token map[string]string) (string, string) {
	start := "{{"
	end := "}}"
	if token != nil {
		if v := strings.TrimSpace(token["start"]); v != "" {
			start = v
		}
		if v := strings.TrimSpace(token["end"]); v != "" {
			end = v
		}
	}
	return start, end
}

func loadTemplateMetadata(path string) (*templateMetadata, error) {
	f, err := os.Open(filepath.Join(path, templateMetadataFile))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var meta templateMetadata
	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return nil, err
	}
	if meta.Token == nil {
		meta.Token = map[string]string{"start": "{{", "end": "}}"}
	}
	if meta.Variables == nil {
		meta.Variables = map[string]Variable{}
	}
	return &meta, nil
}

func writeTemplateMetadata(path string, meta *templateMetadata) error {
	if meta == nil {
		return errors.New("template metadata is nil")
	}
	f, err := os.Create(filepath.Join(path, templateMetadataFile))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}
