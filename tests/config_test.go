package tests

import (
	"os"
	"path/filepath"
	"testing"

	"scaffo/internal/app"
)

func TestConfigSaveAndLoadJSON(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "scaffold.config.json")
	cfg := &app.Config{ // deliberately sparse to test defaults
		Variables: map[string]app.Variable{
			"PROJECT_NAME": {Type: "string", Required: true, Default: "Example", Description: "Project name"},
		},
	}
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := app.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.SourceRoot != "." {
		t.Fatalf("expected SourceRoot '.', got %s", loaded.SourceRoot)
	}
	if len(loaded.IgnoreFolders) == 0 || len(loaded.IgnoreFiles) == 0 || len(loaded.StaticFiles) == 0 {
		t.Fatalf("expected defaults for ignore/static lists to be populated")
	}
	start, end := loaded.Token["start"], loaded.Token["end"]
	if start == "" || end == "" {
		t.Fatalf("expected token delimiters to be set")
	}
}

func TestConfigSaveAndLoadYAML(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "scaffold.config.yaml")
	cfg := &app.Config{SourceRoot: "src", TemplateRoot: "tmpl", Token: map[string]string{"start": "[[", "end": "]]"}}
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := app.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.SourceRoot != "src" || loaded.TemplateRoot != "tmpl" {
		t.Fatalf("expected persisted values; got %+v", loaded)
	}
	if loaded.Token["start"] != "[[" || loaded.Token["end"] != "]]" {
		t.Fatalf("custom token delimiters not preserved")
	}
}

func TestGetVariableValueAndReplacement(t *testing.T) {
	cfg := &app.Config{Variables: map[string]app.Variable{
		"NAME": {Type: "string", Required: true, Default: "Alpha"},
		"DESC": {Type: "string", Required: false, Default: "A sample"},
	}}
	overrides := map[string]string{"NAME": "Bravo"}
	val, ok := cfg.GetVariableValue("NAME", overrides)
	if !ok || val != "Bravo" {
		t.Fatalf("override failed, got %s, ok=%v", val, ok)
	}
	val, ok = cfg.GetVariableValue("DESC", overrides)
	if !ok || val != "A sample" {
		t.Fatalf("default fallback failed, got %s", val)
	}
	text := "Project {{NAME}} -- {{DESC}}"
	replaced := cfg.ApplyVariableReplacements(text, overrides)
	if replaced != "Project Bravo -- A sample" {
		t.Fatalf("replacement mismatch: %s", replaced)
	}
}

func TestApplyVariableReplacementsCustomTokens(t *testing.T) {
	cfg := &app.Config{Token: map[string]string{"start": "<<", "end": ">>"}, Variables: map[string]app.Variable{"X": {Type: "string", Required: true, Default: "42"}}}
	text := "Value: <<X>>"
	replaced := cfg.ApplyVariableReplacements(text, nil)
	if replaced != "Value: 42" {
		t.Fatalf("expected custom token replacement, got %s", replaced)
	}
}

func TestConfigSaveNil(t *testing.T) {
	var cfg *app.Config
	err := cfg.Save("/nope/config.json")
	if err == nil {
		t.Fatalf("expected error saving nil config")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := app.LoadConfig("/path/does/not/exist.json")
	if err == nil {
		t.Fatalf("expected error loading missing file")
	}
}

func TestConfigRoundTripDirectoryCreation(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nested", "scaffold.config.json")
	cfg := &app.Config{}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed with nested dirs: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
}
