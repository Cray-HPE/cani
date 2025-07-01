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
	"fmt"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// Base formats that are always available
var baseFormats = []string{"json"}

// NewCommand creates the parent "add" command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show",
		Short:   "Show items from the inventory",
		Long:    `Show items from the inventory.`,
		PreRunE: provider.GetActiveProvider,
		// Args:    cobra.ArbitraryArgs,
		RunE: show,
	}

	// Define valid sort keys
	validSortKeys := []string{"name", "type", "id", "status", "vendor", "model"}
	cmd.PersistentFlags().StringP("sort", "s", "name", fmt.Sprintf("Sort by this field (%s)", strings.Join(validSortKeys, ", ")))

	// Get all valid formats (base + provider-registered)
	validFormatKeys := getAllValidFormats()

	cmd.PersistentFlags().StringP("format", "o", "json", fmt.Sprintf("Output format (%s)", strings.Join(validFormatKeys, ", ")))

	// Add validation
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		sortKey, _ := cmd.Flags().GetString("sort")
		if !contains(validSortKeys, sortKey) {
			return fmt.Errorf("invalid sort key '%s'. Valid options: %s",
				sortKey, strings.Join(validSortKeys, ", "))
		}

		// formatKey, _ := cmd.Flags().GetString("format")
		// if !contains(ValidFormatKeys, formatKey) {
		// 	return fmt.Errorf("invalid format key '%s'. Valid options: %s",
		// 		formatKey, strings.Join(ValidFormatKeys, ", "))
		// }
		return provider.GetActiveProvider(cmd, args)
	}

	return cmd
}

func show(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load device store: %w", err)
	}

	devices := []*devicetypes.CaniDeviceType{}
	if cmd.Flags().Changed("sort") {
		for _, device := range inv.Devices {
			devices = append(devices, device)
		}

	}
	sort.Slice(devices, func(i, j int) bool {
		sortKey, _ := cmd.Flags().GetString("sort")
		switch sortKey {
		case "name":
			return devices[i].Name < devices[j].Name
		case "type":
			return devices[i].Type < devices[j].Type
		case "id":
			return devices[i].ID.String() < devices[j].ID.String()
		case "status":
			return devices[i].Status < devices[j].Status
		case "vendor":
			return devices[i].Vendor < devices[j].Vendor
		case "model":
			return devices[i].Model < devices[j].Model
		default:
			return devices[i].Name < devices[j].Name // Default sort by name
		}
	})

	// The provider can print their own list output, which may include other formats
	if err := provider.ActiveProvider.Show(cmd, args, devices); err != nil {
		return err
	}

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
	// Start with base formats
	allFormats := make([]string, len(baseFormats))
	// copy(allFormats, baseFormats)

	// // Add provider-registered formats
	// providerFormats := registry.GetValidFormats("show")
	// allFormats = append(allFormats, providerFormats...)

	return allFormats
}

func validateFlags(cmd *cobra.Command, args []string) error {
	// Validate sort key (existing code)
	validSortKeys := []string{"name", "type", "id", "status", "vendor", "model"}
	sortKey, _ := cmd.Flags().GetString("sort")

	if !contains(validSortKeys, sortKey) {
		return fmt.Errorf("invalid sort key '%s'. Valid options: %s",
			sortKey, strings.Join(validSortKeys, ", "))
	}

	// Validate format
	format, _ := cmd.Flags().GetString("format")
	allFormats := getAllValidFormats()

	if !contains(allFormats, format) {
		return fmt.Errorf("invalid format '%s'. Valid options: %s",
			format, strings.Join(allFormats, ", "))
	}

	return provider.GetActiveProvider(cmd, args)
}
