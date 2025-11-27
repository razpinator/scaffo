package tests

import (
	"testing"

	"scaffo/internal/app"
)

func TestMatchIgnoreFolders(t *testing.T) {
	ignoreFolders := []string{"node_modules", ".git"}
	ignoreFiles := []string{"*.log"}
	patterns := []string{"build/**", "**/*.tmp"}
	cases := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{"node_modules", true, true},
		{"src/node_modules", true, true},
		{"src/app/main.go", false, false},
		{"logs/build/app.log", false, true}, // *.log pattern
		{"build/output.bin", false, true},   // build/** pattern
		{"docs/readme.tmp", false, true},    // **/*.tmp pattern
		{".git", true, true},
		{"lib/.git/config", false, true},
	}
	for _, c := range cases {
		got := app.MatchIgnore(c.path, c.isDir, ignoreFolders, ignoreFiles, patterns)
		if got != c.want {
			t.Fatalf("MatchIgnore(%s) = %v want %v", c.path, got, c.want)
		}
	}
}

func TestMatchInclude(t *testing.T) {
	includes := []string{"**/*.go", "README.*", "cmd/**"}
	cases := []struct {
		path string
		want bool
	}{
		{"main.go", true},
		{"README.md", true},
		{"cmd/tool/main.go", true},
		{"assets/style.css", false},
		{"readme.txt", false}, // case-sensitive
	}
	for _, c := range cases {
		got := app.MatchInclude(c.path, includes)
		if got != c.want {
			t.Fatalf("MatchInclude(%s) = %v want %v", c.path, got, c.want)
		}
	}
}
