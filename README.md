# Scaffo

Scaffo is a powerful yet simple CLI tool for scaffolding new projects from existing local codebases. It automatically detects your source project's name and recursively replaces it with your new project's name—handling file paths, contents, and various casing styles (PascalCase, camelCase, snake_case, kebab-case, etc.).

## Features

- **Zero Config Start**: Point it at a folder and go.
- **Smart Replacement**: Automatically detects the source project name (e.g., `MyOldProject`) and replaces it with the new name (e.g., `NewApp`) everywhere.
- **Case Preservation**: Handles variations like `myOldProject` -> `newApp`, `MY_OLD_PROJECT` -> `NEW_APP`, etc.
- **Interactive Mode**: Easy-to-use terminal UI to select source folders.
- **Configurable**: Use `scaffold.config.json` to ignore specific files or folders.

## Installation

### From Source

```bash
go install github.com/razpinator/scaffo@latest
```

Or clone and build:

```bash
git clone https://github.com/razpinator/scaffo.git
cd scaffo
go build -o scaffo cmd/main.go
```

### Homebrew

```bash
brew install razpinator/tap/scaffo
```

## Usage

### Interactive Mode

Simply run `scaffo` without arguments to launch the interactive UI:

```bash
scaffo
```

Select **Run** to choose a source folder from your current directory or enter a custom path.

### CLI Commands

#### Initialize Configuration

Create a default `scaffold.config.json` for a source project:

```bash
scaffo init --from /path/to/source-project
```

#### Run Scaffolding

Scaffold a new project directly:

```bash
scaffo run --from /path/to/source-project --out ./my-new-project
```

This will:
1. Copy files from `/path/to/source-project` to `./my-new-project`.
2. Rename files and directories matching the source project name.
3. Replace content within files matching the source project name.

## Configuration

Scaffo uses `scaffold.config.json` to control the scaffolding process.

```json
{
  "sourceRoot": ".",
  "ignoreFolders": [".git", "node_modules", "dist"],
  "ignoreFiles": [".DS_Store", "*.log"],
  "staticFiles": ["**/*.png", "**/*.jpg"],
  "variables": {
    "PROJECT_NAME": {
      "type": "string",
      "required": true
    }
  }
}
```

## License

MIT
