package init

// Template definitions for root package files

const initTemplate = `package {{.PackageName}}

import (
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/commands"
	"github.com/spf13/cobra"
)

// instance is the singleton provider instance
var instance *{{.StructName}}

func init() {
	instance = New()
	provider.Register("{{.Slug}}", instance)
}

// NewProviderCmd returns provider-specific CLI commands.
// This is called for each base command (import, add, show, etc.) to allow
// the provider to customize or extend the command.
func (p *{{.StructName}}) NewProviderCmd(base *cobra.Command) (*cobra.Command, error) {
	// Switch on the base command name to provide customizations
	switch base.Name() {
	case "import":
		return commands.NewImportCommand(base)

	case "export":
		return commands.NewExportCommand(base)

	case "show":
		return commands.NewShowCommand(base)

	case "add":
		return commands.NewAddCommand(base)

	case "remove":
		return commands.NewRemoveCommand(base)

	case "update":
		return commands.NewUpdateCommand(base)

	default:
		// No customization for this command
		return base, nil
	}
}
`

const providerTemplate = `package {{.PackageName}}

// {{.StructName}} implements the provider.Provider interface
type {{.StructName}} struct {
	Options *ImportOptions
}

// New creates a new {{.StructName}} provider instance
func New() *{{.StructName}} {
	return &{{.StructName}}{}
}

func (p *{{.StructName}}) Slug() string {
	return "{{.Slug}}"
}
`

const optionsTemplate = `package {{.PackageName}}

import "github.com/spf13/cobra"

// Options holds the provider's configuration options.
// These are written to the config file with YAML comments preserved.
// TODO: Add provider-specific configuration fields
type Options struct {
	// Example fields with YAML tags for config file serialization:
	// URL   string ` + "`" + `yaml:"url" head_comment:"Base URL for the API"` + "`" + `
	// Token string ` + "`" + `yaml:"token" head_comment:"API authentication token"` + "`" + `
}

// ImportOptions holds import-specific configuration.
// These options correspond to CLI flags for the import command.
// TODO: Add import-specific configuration fields
type ImportOptions struct {
	// Example:
	// Source string ` + "`" + `yaml:"source" head_comment:"Source file or URL to import from"` + "`" + `
}

// ExportOptions holds export-specific configuration.
// These options correspond to CLI flags for the export command.
// TODO: Add export-specific configuration fields
type ExportOptions struct {
	// Example:
	// Format string ` + "`" + `yaml:"format" head_comment:"Output format (json, yaml, csv)"` + "`" + `
}

// --- HasOptions interface implementation ---

// GetDefaultOptions returns the default configuration options for this provider.
// These are auto-populated in the config file if they do not exist.
func (p *{{.StructName}}) GetDefaultOptions() map[string]any {
	// TODO: Return default configuration values
	return map[string]any{
		// Example:
		// "url":   "https://api.example.com",
		// "token": "",
	}
}

// GetOptionsStruct returns the configuration struct for comment extraction.
// This enables YAML serialization with preserved field ordering and comments.
func (p *{{.StructName}}) GetOptionsStruct() interface{} {
	return &Options{}
}

// --- HasImportOptions interface implementation ---

// GetImportOptionsStruct returns the import options struct for reflection.
func (p *{{.StructName}}) GetImportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetImportDefaults returns default import configuration options.
func (p *{{.StructName}}) GetImportDefaults() map[string]any {
	// TODO: Return default import values
	return map[string]any{
		// Example:
		// "source": "",
	}
}

// BindImportFlags binds CLI flags to Viper for the import command.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *{{.StructName}}) BindImportFlags(cmd *cobra.Command) error {
	// TODO: Bind import-related flags
	// Example:
	// viper.BindPFlag("{{.Slug}}.import.source", cmd.Flags().Lookup("source"))
	return nil
}

// --- HasExportOptions interface implementation ---

// GetExportOptionsStruct returns the export options struct for reflection.
func (p *{{.StructName}}) GetExportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetExportDefaults returns default export configuration options.
func (p *{{.StructName}}) GetExportDefaults() map[string]any {
	// TODO: Return default export values
	return map[string]any{
		// Example:
		// "format": "json",
	}
}

// BindExportFlags binds CLI flags to Viper for the export command.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *{{.StructName}}) BindExportFlags(cmd *cobra.Command) error {
	// TODO: Bind export-related flags
	// Example:
	// viper.BindPFlag("{{.Slug}}.export.format", cmd.Flags().Lookup("format"))
	return nil
}
`

const importWrapperTemplate = `package {{.PackageName}}

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/import"
	"github.com/spf13/cobra"
)

// Import syncs the local CANI inventory from an external system.
// This is the "Extract" step in ETL.
func (p *{{.StructName}}) Import(ctx context.Context, cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return import_.Import(cmd, args, inventory)
}
`

const exportWrapperTemplate = `package {{.PackageName}}

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/export"
	"github.com/spf13/cobra"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.
func (p *{{.StructName}}) Export(ctx context.Context, cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored
	return export.Export(*inventory)
}
`

const transformWrapperTemplate = `package {{.PackageName}}

import (
	"context"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/transform"
)

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables.
func (p *{{.StructName}}) Transform(ctx context.Context, existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return transform.Transform(existing)
}
`
