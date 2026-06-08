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
package add

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// newMetadataCommand creates the "add metadata" subcommand group.
func newMetadataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Create metadata definitions (roles, statuses, tags) in the local inventory.",
		Long: `Create metadata definitions in the local inventory.

Metadata definitions are stored under providerMetadata in the datastore
and are used during export to create the corresponding objects in Nautobot.

Use a subcommand to create a specific metadata type:
  cani alpha add metadata role <name>
  cani alpha add metadata status <name>
  cani alpha add metadata tag <name>`,
	}

	cmd.PersistentFlags().StringSlice("content-types", nil, "Content types (e.g. dcim.device,dcim.rack)")
	cmd.PersistentFlags().String("color", "", "Color (hex, e.g. aa1409)")
	cmd.PersistentFlags().String("description", "", "Description")

	cmd.AddCommand(newMetadataRoleCommand())
	cmd.AddCommand(newMetadataStatusCommand())
	cmd.AddCommand(newMetadataTagCommand())

	return cmd
}

// rejectMetadataFlag enforces mutual exclusivity: the "add metadata"
// subcommand cannot be combined with the --metadata flag.
func rejectMetadataFlag(cmd *cobra.Command) error {
	if f := cmd.Flags().Lookup("metadata"); f != nil && f.Changed {
		return fmt.Errorf("--metadata flag cannot be used with the 'add metadata' subcommand; they are mutually exclusive")
	}
	return nil
}

func newMetadataRoleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "role <name>",
		Short: "Create a role definition in the inventory.",
		Args:  cobra.ExactArgs(1),
		RunE:  addMetadataRole,
	}
}

func newMetadataStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status <name>",
		Short: "Create a status definition in the inventory.",
		Args:  cobra.ExactArgs(1),
		RunE:  addMetadataStatus,
	}
}

func newMetadataTagCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <name>",
		Short: "Create a tag definition in the inventory.",
		Args:  cobra.ExactArgs(1),
		RunE:  addMetadataTag,
	}
}

func addMetadataRole(cmd *cobra.Command, args []string) error {
	if err := rejectMetadataFlag(cmd); err != nil {
		return err
	}
	entry := buildEntry(cmd, args[0])
	entry.Weight = 1000
	return saveMetadata(cmd, "roles", entry)
}

func addMetadataStatus(cmd *cobra.Command, args []string) error {
	if err := rejectMetadataFlag(cmd); err != nil {
		return err
	}
	normalized, err := devicetypes.ValidateUserStatus(args[0])
	if err != nil {
		return err
	}
	entry := buildEntry(cmd, string(normalized))
	return saveMetadata(cmd, "statuses", entry)
}

func addMetadataTag(cmd *cobra.Command, args []string) error {
	if err := rejectMetadataFlag(cmd); err != nil {
		return err
	}
	return saveMetadata(cmd, "tags", buildEntry(cmd, args[0]))
}

func buildEntry(cmd *cobra.Command, name string) devicetypes.MetadataEntry {
	ct, _ := cmd.Flags().GetStringSlice("content-types")
	color, _ := cmd.Flags().GetString("color")
	desc, _ := cmd.Flags().GetString("description")
	return devicetypes.MetadataEntry{
		Name:         name,
		Color:        color,
		Description:  desc,
		ContentTypes: ct,
	}
}

func saveMetadata(cmd *cobra.Command, kind string, entry devicetypes.MetadataEntry) error {
	if err := datastores.SetDeviceStore(cmd, nil); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}
	if err := inventory.AddMetadata(kind, entry); err != nil {
		return err
	}
	if err := datastores.Datastore.Save(inventory); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	log.Printf("Added %s %q to inventory", singularKind(kind), entry.Name)
	return nil
}

// singularKind returns the singular form of a metadata kind name.
func singularKind(kind string) string {
	switch kind {
	case "statuses":
		return "status"
	default:
		return strings.TrimSuffix(kind, "s")
	}
}

// ParseMetadataFlags parses --metadata key=value pairs into a map.
// The parsed values are handed to each provider's MetadataApplier, which
// decides where to store them.
func ParseMetadataFlags(pairs []string) (map[string]string, error) {
	result := make(map[string]string, len(pairs))
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --metadata value %q; expected key=value", p)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}
