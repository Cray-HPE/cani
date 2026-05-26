# Provider Scaffolding

The `init` command generates a complete provider skeleton with stubbed implementations for all required and optional interfaces.

## Generate A Provider

```shell
# Generate a new provider in the default location (pkg/provider/<name>/)
cani init mycloud

# Generate in a custom directory
cani init mycloud --output ./custom/path

# Overwrite an existing directory
cani init mycloud --force
```

## Generated Files

The scaffold creates a provider package with:

| File | Purpose |
|------|---------|
| `init.go` | Provider constructor and registration |
| `provider.go` | `Provider` interface implementation (`Transform`, `NewProviderCmd`, `Slug`) |
| `import.go` | `Importer` interface implementation |
| `export.go` | `Exporter` interface implementation |
| `options.go` | Configuration structs (`Options`, `ImportOptions`, `ExportOptions`) |
| `transform/transform.go` | Transform logic |
| `commands/` | Provider-specific CLI command definitions |

Each method includes TODO comments indicating what needs to be implemented.

## Provider Interface

At minimum, a provider must implement:

```go
type Provider interface {
    Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error)
    NewProviderCmd(base *cobra.Command) (*cobra.Command, error)
    Slug() string
}
```

Optional interfaces for import, export, and configuration:

| Interface | Purpose |
|-----------|---------|
| `Importer` | Import data from external sources |
| `Exporter` | Export data to external targets |
| `HasOptions` | Expose default configuration options |
| `HasImportOptions` | Import-specific CLI flags and config |
| `HasExportOptions` | Export-specific CLI flags and config |
