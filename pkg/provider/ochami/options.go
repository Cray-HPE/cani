package ochami

import "github.com/spf13/cobra"

// Options holds the provider's configuration options.
// These are written to the config file with YAML comments preserved.
// TODO: Add provider-specific configuration fields
type Options struct {
	// Example fields with YAML tags for config file serialization:
	// URL   string `yaml:"url" head_comment:"Base URL for the API"`
	// Token string `yaml:"token" head_comment:"API authentication token"`
}

// ImportOptions holds import-specific configuration.
// These options correspond to CLI flags for the import command.
// TODO: Add import-specific configuration fields
type ImportOptions struct {
	// Example:
	// Source string `yaml:"source" head_comment:"Source file or URL to import from"`
}

// ExportOptions holds export-specific configuration.
// These options correspond to CLI flags for the export command.
// TODO: Add export-specific configuration fields
type ExportOptions struct {
	// Example:
	// Format string `yaml:"format" head_comment:"Output format (json, yaml, csv)"`
}

// --- HasOptions interface implementation ---

// GetDefaultOptions returns the default configuration options for this provider.
// These are auto-populated in the config file if they do not exist.
func (p *Ochami) GetDefaultOptions() map[string]any {
	// TODO: Return default configuration values
	return map[string]any{
		// Example:
		// "url":   "https://api.example.com",
		// "token": "",
	}
}

// GetOptionsStruct returns the configuration struct for comment extraction.
// This enables YAML serialization with preserved field ordering and comments.
func (p *Ochami) GetOptionsStruct() interface{} {
	return &Options{}
}

// --- HasImportOptions interface implementation ---

// GetImportOptionsStruct returns the import options struct for reflection.
func (p *Ochami) GetImportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetImportDefaults returns default import configuration options.
func (p *Ochami) GetImportDefaults() map[string]any {
	// TODO: Return default import values
	return map[string]any{
		// Example:
		// "source": "",
	}
}

// BindImportFlags binds CLI flags to Viper for the import command.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *Ochami) BindImportFlags(cmd *cobra.Command) error {
	// TODO: Bind import-related flags
	// Example:
	// viper.BindPFlag("ochami.import.source", cmd.Flags().Lookup("source"))
	return nil
}

// --- HasExportOptions interface implementation ---

// GetExportOptionsStruct returns the export options struct for reflection.
func (p *Ochami) GetExportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetExportDefaults returns default export configuration options.
func (p *Ochami) GetExportDefaults() map[string]any {
	// TODO: Return default export values
	return map[string]any{
		// Example:
		// "format": "json",
	}
}

// BindExportFlags binds CLI flags to Viper for the export command.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *Ochami) BindExportFlags(cmd *cobra.Command) error {
	// TODO: Bind export-related flags
	// Example:
	// viper.BindPFlag("ochami.export.format", cmd.Flags().Lookup("format"))
	return nil
}
