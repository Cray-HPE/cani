/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */

package init

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
)

// NewCommand creates the "init" command for generating provider scaffolds
func NewCommand() *cli.Command {
	var outputDir string
	var force bool

	cmd := &cli.Command{
		Use:   "init <provider-name>",
		Short: "Generate a new provider scaffold",
		Long: `Generate a new provider scaffold with stubbed implementations.

This creates a new provider package with all required Provider interface methods
and optional interface stubs (Loader, HasOptions, HasImportOptions, HasExportOptions,
CableTransformer) with TODO comments.

Example:
  cani init mycloud
  cani init mycloud --output ./custom/path
  cani init mycloud --force  # Overwrite existing directory`,
		Args: cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			providerName := strings.ToLower(args[0])

			// Validate provider name
			if err := validateProviderName(providerName); err != nil {
				return err
			}

			// Determine output directory
			targetDir := outputDir
			if targetDir == "" {
				// Default to pkg/provider/<name>
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				targetDir = filepath.Join(cwd, "pkg", "provider", providerName)
			}

			// Check if directory exists
			if _, err := os.Stat(targetDir); err == nil {
				if !force {
					return fmt.Errorf("directory %s already exists, use --force to overwrite", targetDir)
				}
				// Remove existing directory if force is set
				if err := os.RemoveAll(targetDir); err != nil {
					return fmt.Errorf("failed to remove existing directory: %w", err)
				}
			}

			// Generate the provider scaffold
			if err := generateProvider(providerName, targetDir); err != nil {
				return fmt.Errorf("failed to generate provider: %w", err)
			}

			fmt.Printf("Successfully generated provider scaffold at %s\n", targetDir)
			fmt.Println("\nNext steps:")
			fmt.Printf("  1. Import the provider in main.go: _ \"github.com/Cray-HPE/cani/pkg/provider/%s\"\n", providerName)
			fmt.Println("  2. Implement the TODO methods in each file")
			fmt.Println("  3. Run 'make bin' to build with your new provider")

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: pkg/provider/<name>)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing directory")

	return cmd
}

// validateProviderName checks that the provider name is valid
func validateProviderName(name string) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Must be a valid Go package name (lowercase letters, numbers, underscores)
	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("provider name must start with a lowercase letter and contain only lowercase letters, numbers, and underscores")
	}

	// Reserved names
	reserved := []string{"provider", "internal", "cmd", "pkg", "main", "test"}
	for _, r := range reserved {
		if name == r {
			return fmt.Errorf("provider name '%s' is reserved", name)
		}
	}

	return nil
}
