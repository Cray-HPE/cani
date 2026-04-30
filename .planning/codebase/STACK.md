# Technology Stack

**Analysis Date:** 2026-04-30

## Languages

**Primary:**
- Go 1.25 - All core application code (`cmd/`, `internal/`, `pkg/`, `main.go`)

**Secondary:**
- Shell (Bash) - BDD/integration tests via ShellSpec (`spec/`)
- Python - Documentation tooling only (`requirements.txt` — mkdocs, tavern)
- YAML - Configuration, device type definitions (`pkg/devicetypes/`)

## Runtime

**Environment:**
- Go 1.25 (specified in `go.mod` and `.github/workflows/unit_tests.yml`)

**Package Manager:**
- Go modules
- Lockfile: `go.sum` present
- Vendoring: `vendor/` directory present (committed)

## Frameworks

**Core:**
- `github.com/spf13/cobra` v1.10.1 - CLI command framework (`cmd/`)
- `github.com/spf13/viper` v1.21.0 - Configuration management (`internal/config/`)

**Testing:**
- Go standard `testing` package - Unit tests
- ShellSpec - BDD functional/integration tests (`spec/`)
- Tavern (Python) - Edge-case/API tests (`requirements.txt`)

**Build/Dev:**
- GNU Make - Build orchestration (`Makefile`)
- `oapi-codegen` v2.5.1 - OpenAPI client generation (`pkg/nautobot/`)
- Swagger Codegen - SLS/HSM/HPCM client generation (via `bin/swagger-codegen-cli.jar`)

## Key Dependencies

**Critical:**
- `github.com/spf13/cobra` v1.10.1 - Entire CLI structure
- `github.com/spf13/viper` v1.21.0 - Config file and flag binding
- `github.com/oapi-codegen/runtime` v1.1.2 - Generated Nautobot API client runtime
- `github.com/google/uuid` v1.6.0 - Hardware inventory UUID generation
- `gopkg.in/yaml.v3` v3.0.1 - YAML config and device type parsing
- `gopkg.in/ini.v1` v1.67.1 - INI file parsing (HPCM config import)

**Infrastructure:**
- `github.com/spf13/afero` v1.15.0 - Filesystem abstraction (indirect, via viper)
- `github.com/fsnotify/fsnotify` v1.9.0 - Config file watching (indirect, via viper)
- `crypto/tls` (stdlib) - TLS client for CSM API connections

## Configuration

**Environment:**
- Config file: YAML format, managed via viper (`internal/config/config.go`)
- Singleton pattern: `config.Cfg` global variable
- Config fields: `providers`, `datastore`, `debug`, `strict`, `types_dirs`, `types_repos`
- Provider-specific config nested under `providers.<name>` keys

**Build:**
- `Makefile` - Primary build configuration
- Version: derived from `git describe --tags`
- Cross-compilation: `GOOS`/`GOARCH` environment variables
- ldflags: `-s -w` (stripped, no debug)
- Binary output: `bin/cani`

## Platform Requirements

**Development:**
- Go 1.25+
- GNU Make
- Git (for version tagging)
- ShellSpec (for BDD tests, installed via `make spec-setup`)
- Docker (for CSM simulator and Nautobot integration tests)
- Python 3 + pip (for docs only)

**Production:**
- Single static binary (`bin/cani`)
- Targets: Linux (amd64, arm64), macOS (arm64)
- RPM packaging supported (`cani.spec`)

## Build Commands

```bash
make all          # fmt → vet → compile
make bin          # Compile binary to bin/
make install      # go install
make clean        # Remove build artifacts
make fmt          # Format Go source
make vet          # Run go vet
make lint         # Static analysis
make utest        # Unit tests
make ftest        # Functional (ShellSpec) tests
make itest        # Integration tests
make tidy         # go mod tidy + vendor
make tools        # Install code-generation tools
make nautobot_client  # Regenerate Nautobot OpenAPI client
```

---

*Stack analysis: 2026-04-30*
