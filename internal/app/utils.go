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
