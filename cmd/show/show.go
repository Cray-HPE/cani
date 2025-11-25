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

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Base formats that are always available
var baseFormats = []string{"json"}

// NewCommand creates the parent "show" command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show items from the inventory",
		Long:  `Show items from the inventory.`,
		// Args:    cobra.ArbitraryArgs,
		RunE: show,
	}

	// Define valid sort keys
	validSortKeys := []string{"name", "type", "id", "status", "vendor", "model"}
	cmd.PersistentFlags().StringP("sort", "s", "name", fmt.Sprintf("Sort by this field (%s)", strings.Join(validSortKeys, ", ")))

	// Get all valid formats (base + provider-registered)
	validFormatKeys := getAllValidFormats()

	cmd.PersistentFlags().StringP("format", "o", "json", fmt.Sprintf("Output format (%s)", strings.Join(validFormatKeys, ", ")))

	// Visual mode flags
	cmd.PersistentFlags().BoolP("visual", "v", false, "Display ASCII visualization of rack layout")
	cmd.PersistentFlags().String("rack", "", "Filter to specific rack by name (used with --visual)")
	cmd.PersistentFlags().Bool("no-color", false, "Disable colorized output (used with --visual)")
	cmd.PersistentFlags().StringP("file", "f", "", "Load inventory from YAML file (used with --visual)")
	cmd.PersistentFlags().Bool("show-cables", false, "Show cable connections in visual output")

	// Compact rack view flags
	cmd.PersistentFlags().Bool("rack-view", false, "Display compact ASCII rack view with device symbols")
	cmd.PersistentFlags().Int("columns", 0, "Number of rack columns before wrapping (0=auto, used with --rack-view)")
	cmd.PersistentFlags().String("cable-type", "", "Filter cables by type (e.g., 'dac', 'cat6', used with --rack-view)")
	cmd.PersistentFlags().CountP("verbose", "V", "Verbose output: -V shows legend, -VV shows all cables (used with --rack-view)")
	cmd.PersistentFlags().Bool("show-routing", false, "Display cable routing with branching visualization (1 rack per line)")

	// Add validation
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		sortKey, _ := cmd.Flags().GetString("sort")
		if !contains(validSortKeys, sortKey) {
			return fmt.Errorf("invalid sort key '%s'. Valid options: %s",
				sortKey, strings.Join(validSortKeys, ", "))
		}

		return nil
	}

	// Add noun-based subcommands
	cmd.AddCommand(newLocationCommand())
	cmd.AddCommand(newRackShowCommand())
	cmd.AddCommand(newDeviceShowCommand())
	cmd.AddCommand(newModuleCommand())
	cmd.AddCommand(newCableCommand())
	cmd.AddCommand(ListCablesCmd)
	cmd.AddCommand(ListInterfacesCmd)

	return cmd
}

// loadInventory loads the inventory from a file or the configured datastore.
func loadInventory(cmd *cobra.Command, args []string) (*devicetypes.Inventory, error) {
	filePath, _ := cmd.Flags().GetString("file")
	visualMode, _ := cmd.Flags().GetBool("visual")

	if filePath != "" && visualMode {
		return loadInventoryFromFile(filePath)
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return nil, fmt.Errorf("failed to set device store: %w", err)
	}
	return datastores.Datastore.Load()
}

func show(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	// Check for cable routing view mode
	showRouting, _ := cmd.Flags().GetBool("show-routing")
	if showRouting {
		rackFilter, _ := cmd.Flags().GetString("rack")
		noColor, _ := cmd.Flags().GetBool("no-color")
		verbose, _ := cmd.Flags().GetCount("verbose")
		cableType, _ := cmd.Flags().GetString("cable-type")

		opts := visual.CompactRenderOptions{
			NoColor:    noColor,
			RackFilter: rackFilter,
			Verbose:    verbose,
			CableType:  cableType,
		}

		return visual.RenderCompactRacksWithCables(inv, opts)
	}

	// Check for compact rack view mode
	rackViewMode, _ := cmd.Flags().GetBool("rack-view")
	if rackViewMode {
		rackFilter, _ := cmd.Flags().GetString("rack")
		noColor, _ := cmd.Flags().GetBool("no-color")
		columns, _ := cmd.Flags().GetInt("columns")
		verbose, _ := cmd.Flags().GetCount("verbose")
		cableType, _ := cmd.Flags().GetString("cable-type")

		opts := visual.CompactRenderOptions{
			NoColor:    noColor,
			RackFilter: rackFilter,
			Columns:    columns,
			Verbose:    verbose,
			CableType:  cableType,
		}

		return visual.RenderCompactRacks(inv, opts)
	}

	// Check for visual mode
	visualMode, _ := cmd.Flags().GetBool("visual")
	if visualMode {
		rackFilter, _ := cmd.Flags().GetString("rack")
		noColor, _ := cmd.Flags().GetBool("no-color")
		showCables, _ := cmd.Flags().GetBool("show-cables")

		opts := visual.RenderOptions{
			NoColor:    noColor,
			RackFilter: rackFilter,
			ShowCables: showCables,
			Inventory:  inv,
		}

		return visual.RenderAllRacks(inv, opts)
	}

	// Output full inventory in JSON format
	output, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}
	fmt.Println(string(output))

	return nil
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

func getAllValidFormats() []string {
	allFormats := make([]string, len(baseFormats))
	copy(allFormats, baseFormats)
	return allFormats
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
