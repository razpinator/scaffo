package tests

import (
	"os"
	"path/filepath"
	"testing"

	"scaffo/internal/app"
)

func TestBuildTemplateAndGenerateIntegration(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.txt"), []byte("Project: {{PROJECT_NAME}} Slug: {{PROJECT_SLUG}}"), 0o644); err != nil {
		t.Fatalf("write templated file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "logo.png"), []byte{0x89, 0x50, 0x4e}, 0o644); err != nil { // pseudo-binary header
		t.Fatalf("write static file: %v", err)
	}
	configPath := filepath.Join(root, "scaffold.config.yaml")
	cfg := &app.Config{
		SourceRoot:    root,
		TemplateRoot:  filepath.Join(root, "template-out"),
		IgnoreFolders: []string{},
		IgnoreFiles:   []string{},
		StaticFiles:   []string{"**/*.png"},
		Token:         map[string]string{"start": "{{", "end": "}}"},
		Variables: map[string]app.Variable{
			"PROJECT_NAME": {Type: "string", Required: true, Default: "Fallback Name"},
			"PROJECT_SLUG": {Type: "string", Required: true, From: "PROJECT_NAME", Transform: "slug-kebab"},
		},
	}
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("save config: %v", err)
	}
	app.BuildTemplateCommand(configPath, cfg.TemplateRoot)
	metaPath := filepath.Join(cfg.TemplateRoot, ".scaffo-template.json")
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("template metadata missing: %v", err)
	}
	t.Setenv("SCAFFO_PROJECT_NAME", "Generated App")
	outPath := filepath.Join(root, "generated")
	app.GenerateCommand(cfg.TemplateRoot, outPath)
	generatedReadme := filepath.Join(outPath, "README.txt")
	data, err := os.ReadFile(generatedReadme)
	if err != nil {
		t.Fatalf("read generated README: %v", err)
	}
	if string(data) != "Project: Generated App Slug: generated-app" {
		t.Fatalf("token replacement failed in generated project: %s", string(data))
	}
	if _, err := os.Stat(filepath.Join(outPath, "logo.png")); err != nil {
		t.Fatalf("static file not generated: %v", err)
	}
}
