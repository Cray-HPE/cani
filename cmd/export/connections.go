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
package export

import (
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes/connections"
	"gopkg.in/yaml.v3"
)

// newConnectionsCommand creates the "export connections" subcommand.
func newConnectionsCommand() *cli.Command {
	cmd := &cli.Command{
		Use:   "connections",
		Short: "Export current cable connections as YAML or CSV.",
		Long: `Export all cable connections from the inventory as a YAML or CSV file.
The output can be edited and re-applied with 'cani alpha add connections'.

Example:
  cani alpha export connections > topology.yaml
  cani alpha export connections --format csv > topology.csv`,
		Args: cli.NoArgs,
		RunE: exportConnections,
	}
	cmd.Flags().String("format", "yaml", "Output format: yaml or csv")
	return cmd
}

func exportConnections(cmd *cli.Command, args []string) error {
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	cm := connections.InventoryToConnectionMap(inventory)

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "yaml":
		enc := yaml.NewEncoder(os.Stdout)
		enc.SetIndent(2)
		return enc.Encode(cm)
	case "csv":
		return connections.WriteCSV(os.Stdout, cm)
	default:
		return fmt.Errorf("unsupported format %q: use yaml or csv", format)
	}
}
