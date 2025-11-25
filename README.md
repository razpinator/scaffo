# Scaffo

Scaffo is an experimental CLI for analyzing an existing codebase, inferring template variables, and producing reusable project templates. It wraps a small Bubble Tea UI as well as direct subcommands so you can bootstrap, inspect, and generate projects from a single tool.

## Prerequisites

- Go 1.21 or newer
- macOS or Linux shell (the provided `builder.sh` assumes a Unix-like environment)

## Installation

```bash
# From the repository root
bash builder.sh   # builds ./cmd and places the binary at /usr/local/bin/scaffo
```

If you prefer to keep the binary locally, run `go build -o scaffo ./cmd` and invoke it via `./scaffo`.

## Usage

```bash
scaffo init --config scaffold.config.json --from /path/to/project
scaffo analyze --config scaffold.config.json
scaffo build-template --config scaffold.config.json --output ./template-out
scaffo generate --template ./template-out --out ./new-app
```

You can also launch the Bubble Tea UI by running `go run ./cmd` and selecting a command interactively.

## Test The Functionality

The following sequence exercises the main code paths end-to-end. Run these commands from the repo root:

```bash
# 1. Verify dependencies and run unit tests (currently stubbed)
go test ./...

# 2. Build the CLI
 go build -o scaffo ./cmd

# 3. Initialize a config pointed at a sample project root
 ./scaffo init --config scaffold.config.json --from .

# 4. Analyze the generated config
 ./scaffo analyze --config scaffold.config.json

# 5. Produce a template skeleton (writes to ./template-out by default)
 ./scaffo build-template --config scaffold.config.json --output ./template-out

# 6. Generate a new project from the template
 ./scaffo generate --template ./template-out --out ./new-app
```

Adjust the `--from`, `--template`, and `--out` flags to point at real projects once you move beyond the sample data. Each step prints diagnostic output so you can confirm that the functionality behaves as expected.
