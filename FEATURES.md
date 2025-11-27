# Scaffolding Template CLI: Tactical Feature List

---

## 1. Core Concepts & Config

- **Config file support** (YAML/JSON/TOML)
  - `scaffold.config.json` at project root
  - Optional CLI flag `--config`
- **Source & output roots**
  - `sourceRoot`: path to the “real” project
  - `templateRoot`: where processed scaffold files will be written
- **Variable definitions**
  - Named placeholders with:
    - Default value
    - Type (`string`, `int`, `bool`, `enum`, `list`)
    - Description
    - Required / optional
- **Token / placeholder syntax**
  - e.g. `{{PROJECT_NAME}}` or `${PROJECT_NAME}`
  - Configurable token delimiters for different templating engines
- **Profiles / presets**
  - Named profiles inside config (web-api, spa, cli) with different ignore/include rules and variable sets

---

## 2. Input/Output Selection

- **Folder include/exclude (ignore lists)**
  - Glob patterns:
    - `ignoreFolders`: `.git`, `.idea`, `.vscode`, `node_modules`, `bin`, `obj`, `dist`, `coverage`, `logs`, `tmp`, `.DS_Store`, etc.
    - `ignoreFiles`: `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `*.log`, compiled artifacts
    - `.scaffoldignore` file similar to `.gitignore`
- **Explicit include patterns**
  - `include`: allow whitelisting certain file patterns (`src/**/*.ts`, `config/**`, `public/**`, etc.)
  - Include beats ignore when explicitly specified
- **Template vs copy-as-is**
  - `templateFiles`: files where variables should be rendered (search & replace)
  - `staticFiles`: copied byte-for-byte (images, binaries, fonts, etc.)
  - Auto-detect using extension + config (e.g. treat `.png`, `.jpg`, `.ico`, `.ttf`, `.woff`, `.pdf` as static by default)
- **Root structure mapping**
  - Optional ability to “flatten” or adjust structure:
    - Move: `src/TemplateApp/**` → `src/{{PROJECT_SLUG}}/**`
    - Rename: `TemplateApp.sln` → `{{PROJECT_NAME}}.sln`

---

## 3. Template Conversion Logic

- **Variable injection in file content**
  - Replace literals (e.g. old project name, company name, base namespace) with placeholders
  - Support multiple patterns:
    - `TemplateApp` → `{{PROJECT_NAME}}`
    - `template-app` → `{{PROJECT_SLUG}}`
    - `com.company.template` → `{{NAMESPACE}}`
- **File/folder name templating**
  - Allow placeholders in path names:
    - `src/TemplateApp.Api` → `src/{{PROJECT_NAME}}.Api`
    - `packages/template-core` → `packages/{{PROJECT_SLUG}}-core`
- **Multi-pass replace**
  - Handle scenarios where the same token appears in content + path + config files reliably
- **Binary-safety**
  - Content templating only applied to files detected as text
  - Use MIME/type or simple heuristic to skip binaries
- **Language/Framework aware strategies (optional but powerful)**
  - Pluggable “analyzers”:
    - Node: `package.json` (name, version, scripts)
    - .NET: `*.csproj`, `*.sln` (AssemblyName, RootNamespace)
    - Go: `go.mod` (module name)
    - Python: `pyproject.toml`, `setup.py`
    - Java: `pom.xml`, `build.gradle`
  - Each analyzer suggests variables and patterns to replace

---

## 4. Variable Discovery & Management

- **Auto-detect candidate variables**
  - Scan project for repeated tokens like project name, org name, domain, etc.
  - Suggest to user: “I found: TemplateApp, com.company, template-app – map them to variables?”
- **Variable kinds**
  - `valueFromEnv`: load default from environment variables
  - `valueFromCommand`: prompt user when generating
  - `computed`: derived variables (slug, upper-case, camelCase)
- **Transformers**
  - Built-in transforms for same logical variable:
    - `{{PROJECT_NAME}}`, `{{project_name}}`, `{{PROJECT_NAME_UPPER}}`, `{{PROJECT_NAME_SLUG}}`
  - Declarative patterns in config
- **Interactive “analyze” mode**
  - `scaffold analyze`:
    - Shows discovered variables
    - Shows example occurrences
    - Lets user confirm / override variable names

---

## 5. CLI Commands & UX

- **init command**
  - `scaffold init --from /path/to/app`
  - Creates default `scaffold.config.json` with:
    - Suggested ignore patterns
    - Empty or auto-detected variable list
    - Default token delimiters
- **analyze command**
  - `scaffold analyze`
  - Prints:
    - Folder/file counts
    - Ignored vs included paths
    - Candidate variables & frequency
    - Potential conflicts (e.g. binary file marked for templating)
- **build-template command**
  - `scaffold build-template --output templates/myapp`
  - Executes all conversions and produces the reusable template skeleton
- **generate (consume template)**
  - `scaffold generate --template templates/myapp --out my-new-app`
  - Prompts for variable values (or reads from flags / config file)
  - Shows summary of actions (created files, replaced variables)
- **Dry-run mode**
  - `--dry-run`: Print what would be converted, renamed, or skipped
  - Optional `--diff` to show example diff on a sample file
- **Verbose logging / debug**
  - `-v`, `-vv` flags:
    - Show which variables applied to which files
    - Show unresolved variables or missed patterns
- **Overwrite strategy**
  - Flags:
    - `--force`: overwrite without asking
    - `--no-overwrite`: skip existing
    - `--backup`: rename existing files to `*.bak`

---

## 6. Plugin & Hook System

- **Pre-/post-conversion hooks (for building template)**
  - Pre: clean up extra stuff from original (delete build artifacts, lock files, secrets)
  - Post: validate generated template (ensure no hard-coded original names remain)
- **Pre-/post-generation hooks (for scaffolding new projects)**
  - Run commands after project generation (`npm install`, `dotnet restore`, `go mod tidy`, etc.)
  - Per-profile overrides: SPA vs API vs CLI
- **Custom script hooks**
  - Shell, Node, Python hooks (e.g. `hooks/postGenerate: "node scripts/post-generate.js"`)

---

## 7. Safety, Security & Secrets

- **Secret detection**
  - Basic secret scanning (API keys, tokens, passwords)
  - Warn if these would go into template (`.env`, `.env.local`, `config/secrets.*`, etc.)
  - Config option to auto-exclude known secret files
- **License and legal**
  - Template license override (replace LICENSE file with your own or template-ready license)
  - Config to strip proprietary content if needed
- **Audit mode**
  - `scaffold audit`:
    - Check for unreplaced original names
    - Embedded absolute paths (`/Users/tehseen/...`)
    - Hard-coded emails/domains

---

## 8. Testing & Validation

- **Self-check command**
  - `scaffold validate-template`:
    - Generate a test project into temp directory using sample values
    - Verify:
      - No unresolved placeholders remain (`{{...}}`)
      - Required variables all used at least once
- **Snapshot tests**
  - Option to compare generated skeleton against known-good snapshot
  - Helps you maintain template stability across upgrades

---

## 9. Example Config Shape (Concrete Sketch)

Just to make all of the above tangible:

```json
{
  "sourceRoot": "./my-existing-app",
  "templateRoot": "./templates/my-existing-app",
  "token": {
    "start": "{{",
    "end": "}}"
  },
  "ignoreFolders": [
    ".git",
    "node_modules",
    ".vscode",
    "dist",
    "coverage",
    "logs"
  ],
  "ignoreFiles": [
    "package-lock.json",
    "yarn.lock",
    "*.log"
  ],
  "staticFiles": [
    "public/assets/**/*",
    "**/*.png",
    "**/*.jpg",
    "**/*.ico",
    "**/*.woff",
    "**/*.woff2"
  ],
  "variables": {
    "PROJECT_NAME": {
      "type": "string",
      "required": true,
      "description": "Human-readable project name"
    },
    "PROJECT_SLUG": {
      "type": "string",
      "required": true,
      "description": "kebab-case slug for folder names",
      "from": "PROJECT_NAME",
      "transform": "slug-kebab"
    },
    "ORG_NAME": {
      "type": "string",
      "required": false,
      "default": "My Org",
      "description": "Organization or company name"
    }
  },
  "replacements": [
    {
      "find": "TemplateApp",
      "replaceWith": "{{PROJECT_NAME}}"
    },
    {
      "find": "template-app",
      "replaceWith": "{{PROJECT_SLUG}}"
    },
    {
      "find": "Acme Corp",
      "replaceWith": "{{ORG_NAME}}"
    }
  ],
  "renameRules": [
    {
      "from": "TemplateApp.sln",
      "to": "{{PROJECT_NAME}}.sln"
    },
    {
      "from": "src/TemplateApp.Api",
      "to": "src/{{PROJECT_NAME}}.Api"
    }
  ],
  "hooks": {
    "postGenerate": [
      {
        "command": "npm install",
        "cwd": "{{TARGET_DIR}}"
      }
    ]
  }
}
```
