/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package show

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Base formats that are always available
var baseFormats = []string{"table", "json", "tree"}

// NewCommand creates the parent "show" command.
func NewCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "show",
		Short: "Show items from the inventory",
		Long:  `Show items from the inventory.`,
		Args:  cli.NoArgs,
		RunE:  show,
	}

	// Define valid sort keys
	validSortKeys := []string{"name", "type", "id", "status", "vendor", "model"}
	cmd.PersistentFlags().StringP("sort", "s", "name", fmt.Sprintf("Sort by this field (%s)", strings.Join(validSortKeys, ", ")))

	// Get all valid formats (base + provider-registered)
	validFormatKeys := getAllValidFormats()

	cmd.PersistentFlags().StringP("format", "o", "table", fmt.Sprintf("Output format (%s)", strings.Join(validFormatKeys, ", ")))
	cmd.PersistentFlags().Bool("no-color", false, "Disable colorized output")
	cmd.PersistentFlags().StringP("file", "f", "", "Load inventory from a YAML file instead of the datastore")

	// Tree detail modifiers (inherited by all subcommands)
	cmd.PersistentFlags().StringSlice("with", []string{"empty-us"},
		"Include extra detail in tree output (modules, interfaces, cables, empty-us)")

	// Add validation
	cmd.PreRunE = func(cmd *cli.Command, args []string) error {

		sortKey, _ := cmd.Flags().GetString("sort")
		if !contains(validSortKeys, sortKey) {
			return fmt.Errorf("invalid sort key '%s'. Valid options: %s",
				sortKey, strings.Join(validSortKeys, ", "))
		}

		formatKey, _ := cmd.Flags().GetString("format")
		if !contains(validFormatKeys, formatKey) {
			return fmt.Errorf("invalid format '%s'. Valid options: %s",
				formatKey, strings.Join(validFormatKeys, ", "))
		}

		return validateWithFlag(cmd)
	}

	// Add noun-based subcommands
	cmd.AddCommand(newLocationCommand())
	cmd.AddCommand(newRackShowCommand())
	cmd.AddCommand(newDeviceShowCommand())
	cmd.AddCommand(newModuleCommand())
	cmd.AddCommand(newCableCommand())
	cmd.AddCommand(newFruCommand())
	cmd.AddCommand(newInterfaceCommand())
	cmd.AddCommand(newMetadataCommand())
	cmd.AddCommand(newVLANShowCommand())
	cmd.AddCommand(newPrefixShowCommand())
	cmd.AddCommand(newIPShowCommand())

	return cmd
}

// loadInventory loads the inventory from a file or the configured datastore.
func loadInventory(cmd *cli.Command, args []string) (*devicetypes.Inventory, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		return loadInventoryFromFile(filePath)
	}

	if err := store.Setup(cmd); err != nil {
		return nil, fmt.Errorf("failed to set device store: %w", err)
	}
	return datastores.Datastore.Load()
}

func show(cmd *cli.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintAllTables(inv)
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := visual.BuildFullTree(inv, tf)
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		output, err := json.MarshalIndent(inv, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal inventory: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// BaseFormats returns the format names that every show subcommand accepts.
func BaseFormats() []string {
	return append([]string(nil), baseFormats...)
}

func getAllValidFormats() []string {
	return BaseFormats()
}

// inventoryYAML is an intermediate struct for YAML parsing with string keys
type inventoryYAML struct {
	Devices map[string]*devicetypes.CaniDeviceType `yaml:"devices"`
	Cables  map[string]*devicetypes.CaniCableType  `yaml:"cables,omitempty"`
}

// loadInventoryFromFile loads an inventory from a YAML file
func loadInventoryFromFile(filePath string) (*devicetypes.Inventory, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse into intermediate struct with string keys
	rawInv := &inventoryYAML{}
	if err := yaml.Unmarshal(data, rawInv); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to proper Inventory with UUID keys
	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Cables:  make(map[uuid.UUID]*devicetypes.CaniCableType),
	}

	// Convert device string keys to UUIDs
	for idStr, device := range rawInv.Devices {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid device UUID %q: %w", idStr, err)
		}
		if device != nil {
			device.ID = id
		}
		inv.Devices[id] = device
	}

	// Convert cable string keys to UUIDs
	for idStr, cable := range rawInv.Cables {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid cable UUID %q: %w", idStr, err)
		}
		if cable != nil {
			cable.ID = id
		}
		inv.Cables[id] = cable
	}

	return inv, nil
}
