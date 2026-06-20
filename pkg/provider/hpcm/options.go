package hpcm

import "github.com/Cray-HPE/cani/internal/cli"

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
type ImportOptions struct {
	NodeJsonFile string `yaml:"node_json_file" head_comment:"Path to HPCM node JSON file"`
	CmConfigFile string `yaml:"cm_config_file" head_comment:"Path to HPCM cm.config file"`
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
func (p *Hpcm) GetDefaultOptions() map[string]any {
	// TODO: Return default configuration values
	return map[string]any{
		// Example:
		// "url":   "https://api.example.com",
		// "token": "",
	}
}

// GetOptionsStruct returns the configuration struct for comment extraction.
// This enables YAML serialization with preserved field ordering and comments.
func (p *Hpcm) GetOptionsStruct() interface{} {
	return &Options{}
}

// --- HasImportOptions interface implementation ---

// GetImportOptionsStruct returns the import options struct for reflection.
func (p *Hpcm) GetImportOptionsStruct() interface{} {
	return &ImportOptions{}
}

// GetImportDefaults returns default import configuration options.
func (p *Hpcm) GetImportDefaults() map[string]any {
	return map[string]any{
		"node_json_file": "",
		"cm_config_file": "",
	}
}

// BindImportFlags registers the import command's CLI flags.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *Hpcm) BindImportFlags(cmd *cli.Command) error {
	// TODO: Bind import-related flags.
	// Read the flag directly, e.g. cmd.Flags().GetString("source").
	return nil
}

// --- HasExportOptions interface implementation ---

// GetExportOptionsStruct returns the export options struct for reflection.
func (p *Hpcm) GetExportOptionsStruct() interface{} {
	return map[string]any{}
}

// GetExportDefaults returns default export configuration options.
func (p *Hpcm) GetExportDefaults() map[string]any {
	// TODO: Return default export values
	return map[string]any{
		// Example:
		// "format": "json",
	}
}

// BindExportFlags registers the export command's CLI flags.
// This enables precedence: CLI flags > env vars > config file > defaults.
func (p *Hpcm) BindExportFlags(cmd *cli.Command) error {
	// TODO: Bind export-related flags.
	// Read the flag directly, e.g. cmd.Flags().GetString("format").
	return nil
}
