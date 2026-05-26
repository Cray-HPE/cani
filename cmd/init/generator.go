package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// templateData holds the data passed to all templates
type templateData struct {
	PackageName string // lowercase package name (e.g., "mycloud")
	StructName  string // PascalCase struct name (e.g., "Mycloud")
	Slug        string // provider slug (e.g., "mycloud")
}

// generateProvider creates the provider scaffold in the target directory
func generateProvider(name, targetDir string) error {
	// Create target directory and subdirectories
	subdirs := []string{"commands", "export", "import", "transform"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(targetDir, subdir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", subdir, err)
		}
	}

	// Prepare template data
	data := templateData{
		PackageName: name,
		StructName:  toPascalCase(name),
		Slug:        name,
	}

	// Generate root package files
	rootFiles := []struct {
		filename string
		tmpl     string
	}{
		{"init.go", initTemplate},
		{"provider.go", providerTemplate},
		{"options.go", optionsTemplate},
		{"import.go", importWrapperTemplate},
		{"export.go", exportWrapperTemplate},
		{"transform.go", transformWrapperTemplate},
	}

	for _, f := range rootFiles {
		if err := generateFile(filepath.Join(targetDir, f.filename), f.tmpl, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.filename, err)
		}
	}

	// Generate subpackage files
	subpkgFiles := []struct {
		subdir   string
		filename string
		tmpl     string
	}{
		{"commands", "commands.go", commandsSubpkgTemplate},
		{"export", "export.go", exportSubpkgTemplate},
		{"import", "import.go", importSubpkgTemplate},
		{"transform", "transform.go", transformSubpkgTemplate},
	}

	for _, f := range subpkgFiles {
		path := filepath.Join(targetDir, f.subdir, f.filename)
		if err := generateFile(path, f.tmpl, data); err != nil {
			return fmt.Errorf("failed to generate %s/%s: %w", f.subdir, f.filename, err)
		}
	}

	return nil
}

// generateFile creates a single file from a template
func generateFile(path, tmplContent string, data templateData) error {
	tmpl, err := template.New(filepath.Base(path)).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// toPascalCase converts a snake_case or lowercase string to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

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
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/import"
	"github.com/spf13/cobra"
)

// Import syncs the local CANI inventory from an external system.
// This is the "Extract" step in ETL.
func (p *{{.StructName}}) Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	return import_.Import(cmd, args, inventory)
}
`

const exportWrapperTemplate = `package {{.PackageName}}

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/export"
	"github.com/spf13/cobra"
)

// Export syncs the local CANI inventory to an external system.
// This is the "Load" step in ETL.
func (p *{{.StructName}}) Export(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored
	return export.Export(*inventory)
}
`

const transformWrapperTemplate = `package {{.PackageName}}

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/{{.PackageName}}/transform"
)

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables.
func (p *{{.StructName}}) Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	return transform.Transform(existing)
}
`

// Template definitions for subpackage files

const commandsSubpkgTemplate = `package commands

import "github.com/spf13/cobra"

func NewImportCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add import-specific flags or subcommands
	// Example:
	// base.Flags().String("source", "", "Source file to import")
	return &cobra.Command{}, nil
}

func NewExportCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add export-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewShowCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add show-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewAddCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add add-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewRemoveCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add remove-specific flags or subcommands
	return &cobra.Command{}, nil
}

func NewUpdateCommand(base *cobra.Command) (*cobra.Command, error) {
	// TODO: Add update-specific flags or subcommands
	return &cobra.Command{}, nil
}
`

const exportSubpkgTemplate = `package export

import "github.com/Cray-HPE/cani/pkg/devicetypes"

func Export(existing devicetypes.Inventory) error {
	// Common patterns:
	//   - Compare local inventory with external system
	//   - Create/update/delete resources in external system
	//   - Report what was created, updated, skipped, or errored
	return nil
}
`

const importSubpkgTemplate = `package import_

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

func Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	// Common patterns:
	//   - Parse files or query APIs to get data
	//   - Store data in provider struct for later processing in Transform()
	//   - Report what was imported, skipped, or errored
	return nil
}
`

const transformSubpkgTemplate = `package transform

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// Transform converts imported data into CANI's inventory format.
// Returns a TransformResult containing devices, racks, and cables (nil maps indicate not applicable).
func Transform(existing devicetypes.Inventory) (*devicetypes.TransformResult, error) {
	// TODO: Implement Transform
	// Common patterns:
	//   - Convert queued data from Extract to CaniDeviceType
	//   - Check existing inventory for duplicates
	//   - Set parent-child relationships
	//
	// Might accept queues or other data structures as parameters, which were collected during Import
	return &devicetypes.TransformResult{}, nil
}
`
