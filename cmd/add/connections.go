/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package add

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/devicetypes/connections"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// newConnectionsCommand creates the "add connections" subcommand.
func newConnectionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connections <file.yaml|file.csv>",
		Short: "Add cable connections from a YAML connection map or CSV file.",
		Long: `Add cable connections from a declarative YAML file or a CSV file.
The CSV format is auto-detected from the header row:

Human CSV (one row per cable):
  a_device,a_port,b_device,b_port[,type,label,color,length,length_unit,status]
  Use a_device="_defaults" row for cable defaults (type, color, status, length_unit).

Nautobot interfaces CSV (exported from Nautobot, paired by UUID):
  name,device__name,id,cable_peer[,cable__pk,type,status__name,label]

YAML (.yaml, .yml): A connection map with brace-expansion patterns.

Use the 'generate' subcommand to create a connection map from a topology
pattern (star, leaf-spine, ring) without writing YAML by hand.

Examples:
  cani alpha add connections topology.yaml
  cani alpha add connections topology.yaml --dry-run
  cani alpha add connections interfaces.csv --dry-run
  cani alpha add connections cables.csv --dry-run
  cani alpha add connections generate star --hub switch-01 --spokes "node-{01..48}"`,
		Args: cobra.MaximumNArgs(1),
		RunE: addConnections,
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be created without modifying the inventory")

	cmd.AddCommand(newGenerateCommand())

	return cmd
}

func addConnections(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading connection map: %w", err)
	}

	cm, err := parseConnectionFile(args[0], data)
	if err != nil {
		return err
	}

	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	resolved, resolveErrs := connections.ResolveConnectionMap(cm, inventory)
	for _, e := range resolveErrs {
		log.Printf("WARNING: %v", e)
	}

	if len(resolved) == 0 && len(resolveErrs) > 0 {
		return fmt.Errorf("no connections resolved; %d errors", len(resolveErrs))
	}

	if dryRun {
		return printDryRun(resolved, inventory)
	}

	created, applyErrs := connections.ApplyConnections(resolved, inventory)
	for _, e := range applyErrs {
		log.Printf("WARNING: %v", e)
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}

	log.Printf("%d cable(s) created from %s", created, args[0])
	if len(resolveErrs) > 0 {
		log.Printf("%d connection(s) skipped due to errors", len(resolveErrs))
	}
	return nil
}

// printDryRun displays what connections would be created.
func printDryRun(conns []connections.ResolvedConnection, inv *devicetypes.Inventory) error {
	fmt.Printf("Dry run: %d cable(s) would be created:\n\n", len(conns))
	for i, c := range conns {
		aName := connectionDeviceName(c.ADevice, inv)
		bName := connectionDeviceName(c.BDevice, inv)
		cableType := c.Cable.Type
		if cableType == "" {
			cableType = "(default)"
		}
		fmt.Printf("  %3d. %s:%s --[%s]-- %s:%s\n",
			i+1, aName, c.APort, cableType, bName, c.BPort)
		if c.AMac != "" {
			fmt.Printf("       %s:%s mac=%s\n", aName, c.APort, c.AMac)
		}
		if c.BMac != "" {
			fmt.Printf("       %s:%s mac=%s\n", bName, c.BPort, c.BMac)
		}
	}
	return nil
}

// connectionDeviceName returns the device or module name for a UUID, or the UUID string if not found.
func connectionDeviceName(id uuid.UUID, inv *devicetypes.Inventory) string {
	if id == uuid.Nil {
		return "(none)"
	}
	if dev, ok := inv.Devices[id]; ok && dev.Name != "" {
		return dev.Name
	}
	if mod, ok := inv.Modules[id]; ok && mod.Name != "" {
		return mod.Name
	}
	return id.String()
}

// parseConnectionFile detects the file format from the extension and
// parses it into a ConnectionMap. Supports .yaml/.yml (declarative
// connection map) and .csv (Nautobot interfaces export).
func parseConnectionFile(path string, data []byte) (*connections.ConnectionMap, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".csv":
		cm, err := connections.ParseCSV(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("parsing CSV connections: %w", err)
		}
		return cm, nil

	case ".yaml", ".yml":
		var cm connections.ConnectionMap
		if err := yaml.Unmarshal(data, &cm); err != nil {
			return nil, fmt.Errorf("parsing connection map: %w", err)
		}
		if cm.Version == "" {
			return nil, fmt.Errorf("connection map missing required 'version' field")
		}
		return &cm, nil

	default:
		return nil, fmt.Errorf("unsupported file format %q (use .yaml, .yml, or .csv)", ext)
	}
}
