# Coding Conventions

**Analysis Date:** 2026-04-30

## Naming Patterns

**Files:**
- Snake_case for multi-word Go files: `cani_device_types.go`, `struct_to_map.go`
- Test files use `_test.go` suffix co-located with source: `inventory_queries_test.go`
- Subcommand files named after the noun they operate on: `rack.go`, `device.go`, `module.go`

**Functions:**
- Exported functions use PascalCase: `NewInventory()`, `FindLocationByName()`, `SetDeviceStore()`
- Unexported functions use camelCase: `addRack()`, `makeRack()`, `alphaToIndex()`
- Constructor functions prefixed with `New`: `NewExporter()`, `NewLocation()`, `NewDefaultLocation()`
- Command factory functions prefixed with `new...Command`: `newRackAddCommand()`, `newLocationCommand()`

**Variables:**
- camelCase for locals: `statusArg`, `serialArg`, `nameArg`
- PascalCase for exported package-level vars: `Cfg`, `Datastore`
- Constants use PascalCase: `SchemaVersionV1Alpha1`, `DefaultTypesRepo`

**Types:**
- PascalCase with descriptive suffixes: `CaniDeviceType`, `CaniRackType`, `CaniLocationType`
- Interfaces named by capability: `Provider`, `Exporter`, `Importer`
- Options structs suffixed with `Opts`: `ExporterOpts`

## Code Style

**Formatting:**
- `go fmt` (standard Go formatting)
- Run via: `make fmt`

**Linting:**
- `golint` with `-set_exit_status` for `./cmd/...`, `./internal/...`, `./pkg/...`
- `go vet ./...`
- Run via: `make lint` and `make vet`

**License Header:**
- Every `.go` file begins with a full MIT License block comment (C-style `/* */`)
- Copyright: `(C) Copyright YYYY[-YYYY] Hewlett Packard Enterprise Development LP`

## Import Organization

**Order:**
1. Standard library packages (`fmt`, `os`, `path/filepath`, `testing`)
2. External packages (`github.com/google/uuid`, `github.com/spf13/cobra`)
3. Internal project packages (`github.com/Cray-HPE/cani/internal/...`, `github.com/Cray-HPE/cani/pkg/...`)

**Path Aliases:**
- No import aliases used except for side-effect imports using `_`
- Side-effect imports for provider registration in `main.go`:
  ```go
  _ "github.com/Cray-HPE/cani/pkg/provider/csm"      // CSM provider
  _ "github.com/Cray-HPE/cani/pkg/provider/example"  // Example provider
  ```

## Error Handling

**Patterns:**
- Use `fmt.Errorf("context: %w", err)` for error wrapping throughout
- Return early on error — never nest happy-path logic inside error checks
- Error messages start lowercase, describe the failed operation:
  ```go
  return fmt.Errorf("failed to load inventory: %w", err)
  return fmt.Errorf("name resolution failed: %w", err)
  return fmt.Errorf("provider hook failed: %w", err)
  ```
- Bare `return err` when no additional context is needed (e.g., validated args)
- No use of `errors.Wrap` or third-party error libraries — standard `fmt.Errorf` with `%w` only

**Cobra Command Errors:**
- `RunE` functions return errors; cobra prints them
- `Args` validators use cobra's built-in `cobra.ArbitraryArgs` or custom validator functions

## Logging

**Framework:** Standard library `log` package

**Patterns:**
- `log.Printf()` for informational messages during CLI operations
- `log.Println()` for simple status messages
- `fmt.Fprintf(cmd.OutOrStdout(), ...)` for user-facing output that respects cobra's output writer
- `fmt.Fprintln(cmd.ErrOrStderr(), ...)` for error-like messages via cobra's stderr
- No structured logging library (logrus, zap, slog) — keep it simple with stdlib

## Comments

**When to Comment:**
- Package-level doc comments on exported types and functions
- Brief inline comments for non-obvious logic
- Test coverage tables at top of test files (ASCII table format) documenting which functions have happy-path and failure tests

**Test Coverage Table Pattern:**
```go
// | Function       | Happy-path test           | Failure test                  |
// |----------------|---------------------------|-------------------------------|
// | SetDeviceStore | TestSetDeviceStoreJSON     | TestSetDeviceStoreUnsupported |
```

**JSDoc/TSDoc:** N/A (Go project)

## Function Design

**Size:** Functions are typically 20–60 lines. Larger cobra `RunE` handlers may reach 80–100 lines.

**Parameters:** Prefer passing `*cobra.Command` and `[]string` for command handlers. Use option structs (`ExporterOpts`) for complex configuration.

**Return Values:** Standard Go `(result, error)` pattern. Single `error` for void operations.

## Module Design

**Exports:**
- Each package exports its primary types and constructor functions
- Internal helper functions remain unexported (camelCase)
- Provider interface defined in `internal/provider/interface.go`

**Barrel Files:** Not applicable (Go packages export via capitalization)

**Package Organization:**
- `cmd/` — Cobra command definitions, one subdirectory per verb (`add/`, `remove/`, `show/`, `update/`)
- `internal/` — Non-importable packages (config, core, provider registry, utilities)
- `pkg/` — Importable library packages (datastores, devicetypes, providers)

## Provider Registration Pattern

Providers self-register via `init()` functions using a global registry:
```go
// In pkg/provider/csm/register.go
func init() {
    provider.Register("csm", &CSMProvider{})
}
```

Main imports providers as side effects:
```go
_ "github.com/Cray-HPE/cani/pkg/provider/csm"
```

## Configuration Pattern

- Viper for config/env/flag precedence: CLI flags > env vars > config file > defaults
- Config file: `~/.cani/cani.yml`
- Global singleton: `config.Cfg`
- Flags bound to viper with `viper.BindPFlag()`

## Cobra Command Pattern

```go
func NewCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "verb [args]",
        Short: "Brief description",
        Long:  "Detailed description",
        Args:  cobra.ArbitraryArgs, // or custom validator
        RunE:  handlerFunc,
    }
    cmd.AddCommand(subcommand1())
    cmd.Flags().StringP("flag-name", "f", "default", "description")
    return cmd
}
```

---

*Convention analysis: 2026-04-30*
