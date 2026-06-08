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
