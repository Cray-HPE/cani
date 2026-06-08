/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/spf13/cobra"
)

// newLocationCommand creates the "show location" subcommand.
func newLocationCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "location [name|uuid]",
		Short: "List locations in the inventory.",
		Long:  "List locations, or show a single location by name or UUID.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  showLocations,
	}
}

func showLocations(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleLocation(cmd, inv, args[0])
	}

	locations := make([]*devicetypes.CaniLocationType, 0, len(inv.Locations))
	for _, loc := range inv.Locations {
		locations = append(locations, loc)
	}
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].Name < locations[j].Name
	})

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintLocationTable(locations, inv)
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := visual.BuildLocationTree(inv, tf)
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(locations)
	}
}

func showSingleLocation(cmd *cobra.Command, inv *devicetypes.Inventory, arg string) error {
	loc, err := findLocationByNameOrUUID(arg, inv)
	if err != nil {
		return err
	}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		visual.PrintLocationTable([]*devicetypes.CaniLocationType{loc}, inv)
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := []visual.TreeNode{visual.LocationToTreeNode(loc, inv, tf)}
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(loc)
	}
}

// newRackShowCommand creates the "show rack" subcommand.
func newRackShowCommand() *cobra.Command {
	// Merge base formats with rack-specific visual formats.
	rackFormats := append(BaseFormats(), visual.ValidRackFormats()...)

	cmd := &cobra.Command{
		Use:   "rack [name|uuid]",
		Short: "List racks in the inventory.",
		Long: fmt.Sprintf(`List racks, or show a single rack by name or UUID.
When a rack is specified, the default output is a visual detail view.

Output formats (-o):
  table     Tabular list of racks
  json      JSON output
  tree      Hierarchical tree view
  classic   Full-height ASCII rack with device symbols
  minimap   Compact 2-character-wide rack columns
  detail    Single-rack minimap with right-side annotations
  routing   Cable routing with branching visualization

Examples:
  cani show rack                        # table of all racks
  cani show rack MyRack                 # detail view of one rack
  cani show rack -o minimap             # compact minimap of all racks
  cani show rack -o routing -VV         # cable routing with all cables
  cani show rack -o json                # JSON output

Valid formats: %s`, strings.Join(rackFormats, ", ")),
		Args: cobra.MaximumNArgs(1),
		RunE: showRacks,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			formatKey, _ := cmd.Flags().GetString("format")
			if !contains(rackFormats, formatKey) {
				return fmt.Errorf("invalid format '%s'. Valid options: %s",
					formatKey, strings.Join(rackFormats, ", "))
			}
			return validateWithFlag(cmd)
		},
	}

	cmd.Flags().Int("columns", 0, "Number of rack columns before wrapping (0=auto, used with minimap)")
	cmd.Flags().CountP("verbose", "V", "Verbose output: -V shows legend, -VV shows all cables")
	cmd.Flags().BoolP("labels", "l", false, "Show A/B termination labels on routing view")
	cmd.Flags().BoolP("interactive", "I", false, "Interactive toggle mode for routing view")

	return cmd
}

func showRacks(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		return showSingleRack(cmd, inv, args[0])
	}

	format, _ := cmd.Flags().GetString("format")
	noColor, _ := cmd.Flags().GetBool("no-color")
	verbose, _ := cmd.Flags().GetCount("verbose")
	columns, _ := cmd.Flags().GetInt("columns")
	showLabels, _ := cmd.Flags().GetBool("labels")
	interactive, _ := cmd.Flags().GetBool("interactive")

	// Handle rack-specific visual formats.
	if rf := visual.RackFormat(format); rf == visual.RackFormatClassic ||
		rf == visual.RackFormatMinimap ||
		rf == visual.RackFormatDetail ||
		rf == visual.RackFormatRouting {
		return visual.RenderRack(inv, rf, visual.CompactRenderOptions{
			NoColor:     noColor,
			Columns:     columns,
			Verbose:     verbose,
			ShowLabels:  showLabels,
			Interactive: interactive,
			Inventory:   inv,
		})
	}

	// Fall through to base format dispatch.
	racks := make([]*devicetypes.CaniRackType, 0, len(inv.Racks))
	for _, rack := range inv.Racks {
		racks = append(racks, rack)
	}
	sort.Slice(racks, func(i, j int) bool {
		return racks[i].Name < racks[j].Name
	})

	switch format {
	case "table":
		visual.PrintRackTable(racks, inv)
		return nil
	case "tree":
		tf := treeFilterFromCmd(cmd)
		nodes := visual.BuildRackTree(racks, inv, tf)
		visual.RenderTreeOutput(nodes, tf.NoColor)
		return nil
	default:
		return marshalAndPrint(racks)
	}
}
