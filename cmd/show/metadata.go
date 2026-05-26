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
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// newMetadataCommand creates the "show metadata" subcommand.
func newMetadataCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "metadata",
		Short: "Show metadata definitions (roles, statuses, tags) in the inventory.",
		Long:  "Show metadata definitions (roles, statuses, tags) in the inventory.",
		RunE:  showMetadata,
	}
}

func showMetadata(cmd *cobra.Command, args []string) error {
	inv, err := loadInventory(cmd, args)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "table":
		printMetadataTable(inv)
		return nil
	default:
		return marshalAndPrint(inv.Metadata)
	}
}

// printMetadataTable renders inventory metadata as grouped tables.
func printMetadataTable(inv *devicetypes.Inventory) {
	if inv.Metadata == nil || (len(inv.Metadata.Roles) == 0 &&
		len(inv.Metadata.Statuses) == 0 &&
		len(inv.Metadata.Tags) == 0) {
		fmt.Println("(no metadata defined)")
		return
	}

	if len(inv.Metadata.Roles) > 0 {
		fmt.Printf("\nRoles (%d):\n", len(inv.Metadata.Roles))
		printRolesTable(inv.Metadata.Roles)
	}
	if len(inv.Metadata.Statuses) > 0 {
		fmt.Printf("\nStatuses (%d):\n", len(inv.Metadata.Statuses))
		printEntriesTable(inv.Metadata.Statuses)
	}
	if len(inv.Metadata.Tags) > 0 {
		fmt.Printf("\nTags (%d):\n", len(inv.Metadata.Tags))
		printEntriesTable(inv.Metadata.Tags)
	}
	fmt.Println()
}

// Column widths for metadata tables.
const (
	colMetaName = 25
	colColor    = 10
	colDesc     = 30
	colCTypes   = 30
	colWeight   = 8
)

// printRolesTable renders role entries with a WEIGHT column.
func printRolesTable(entries []devicetypes.MetadataEntry) {
	header := col("NAME", colMetaName) + "  " +
		col("COLOR", colColor) + "  " +
		col("DESCRIPTION", colDesc) + "  " +
		col("CONTENT TYPES", colCTypes) + "  " +
		col("WEIGHT", colWeight)
	sep := col(strings.Repeat("-", colMetaName), colMetaName) + "  " +
		col(strings.Repeat("-", colColor), colColor) + "  " +
		col(strings.Repeat("-", colDesc), colDesc) + "  " +
		col(strings.Repeat("-", colCTypes), colCTypes) + "  " +
		col(strings.Repeat("-", colWeight), colWeight)

	fmt.Println(header)
	fmt.Println(sep)
	for _, e := range entries {
		fmt.Println(
			col(e.Name, colMetaName) + "  " +
				col(e.Color, colColor) + "  " +
				col(e.Description, colDesc) + "  " +
				col(strings.Join(e.ContentTypes, ","), colCTypes) + "  " +
				col(strconv.Itoa(e.Weight), colWeight),
		)
	}
}

// printEntriesTable renders metadata entries without a WEIGHT column.
func printEntriesTable(entries []devicetypes.MetadataEntry) {
	header := col("NAME", colMetaName) + "  " +
		col("COLOR", colColor) + "  " +
		col("DESCRIPTION", colDesc) + "  " +
		col("CONTENT TYPES", colCTypes)
	sep := col(strings.Repeat("-", colMetaName), colMetaName) + "  " +
		col(strings.Repeat("-", colColor), colColor) + "  " +
		col(strings.Repeat("-", colDesc), colDesc) + "  " +
		col(strings.Repeat("-", colCTypes), colCTypes)

	fmt.Println(header)
	fmt.Println(sep)
	for _, e := range entries {
		fmt.Println(
			col(e.Name, colMetaName) + "  " +
				col(e.Color, colColor) + "  " +
				col(e.Description, colDesc) + "  " +
				col(strings.Join(e.ContentTypes, ","), colCTypes),
		)
	}
}
