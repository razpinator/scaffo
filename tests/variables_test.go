package tests

import (
	"testing"

	"github.com/razpinator/scaffo/internal/app"
)

func TestVariableTransformIndirect(t *testing.T) {
	vars := map[string]app.Variable{
		"PROJECT_NAME":  {Type: "string", Required: true, Default: "My Cool App"},
		"PROJECT_SLUG":  {Type: "string", Required: true, From: "PROJECT_NAME", Transform: "slug-kebab"},
		"PROJECT_UPPER": {Type: "string", Required: true, From: "PROJECT_NAME", Transform: "upper"},
	}
	cfg := &app.Config{Variables: vars}
	overrides := map[string]string{"PROJECT_NAME": "Example Project"}
	slug := "example-project" // expected slug
	upper := "EXAMPLE PROJECT"
	text := "Name={{PROJECT_NAME}} Slug={{PROJECT_SLUG}} Upper={{PROJECT_UPPER}}"
	replaced := cfg.ApplyVariableReplacements(text, map[string]string{
		"PROJECT_NAME":  overrides["PROJECT_NAME"],
		"PROJECT_SLUG":  slug,
		"PROJECT_UPPER": upper,
	})
	want := "Name=Example Project Slug=example-project Upper=EXAMPLE PROJECT"
	if replaced != want {
		t.Fatalf("transform indirect replacement failed: got %s want %s", replaced, want)
	}
}

func TestApplyVariableReplacementsEmptyOverrides(t *testing.T) {
	cfg := &app.Config{Variables: map[string]app.Variable{"A": {Type: "string", Default: "x"}, "B": {Type: "string", Default: "y"}}}
	res := cfg.ApplyVariableReplacements("{{A}}/{{B}}", nil)
	if res != "x/y" {
		t.Fatalf("expected defaults used, got %s", res)
	}
}
