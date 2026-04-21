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
package visual

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// Column widths for metadata tables.
const (
	ColMetaName = 25
	ColColor    = 10
	ColDesc     = 30
	ColCTypes   = 30
	ColWeight   = 8
)

// PrintMetadataTable renders inventory metadata as grouped tables.
func PrintMetadataTable(inv *devicetypes.Inventory) {
	if inv.Metadata == nil || (len(inv.Metadata.Roles) == 0 &&
		len(inv.Metadata.Statuses) == 0 &&
		len(inv.Metadata.Tags) == 0) {
		fmt.Println("(no metadata defined)")
		return
	}

	if len(inv.Metadata.Roles) > 0 {
		fmt.Printf("\nRoles (%d):\n", len(inv.Metadata.Roles))
		PrintRolesTable(inv.Metadata.Roles)
	}
	if len(inv.Metadata.Statuses) > 0 {
		fmt.Printf("\nStatuses (%d):\n", len(inv.Metadata.Statuses))
		PrintEntriesTable(inv.Metadata.Statuses)
	}
	if len(inv.Metadata.Tags) > 0 {
		fmt.Printf("\nTags (%d):\n", len(inv.Metadata.Tags))
		PrintEntriesTable(inv.Metadata.Tags)
	}
	fmt.Println()
}

// PrintRolesTable renders role entries with a WEIGHT column.
func PrintRolesTable(entries []devicetypes.MetadataEntry) {
	header := Col("NAME", ColMetaName) + "  " +
		Col("COLOR", ColColor) + "  " +
		Col("DESCRIPTION", ColDesc) + "  " +
		Col("CONTENT TYPES", ColCTypes) + "  " +
		Col("WEIGHT", ColWeight)
	sep := Col(strings.Repeat("-", ColMetaName), ColMetaName) + "  " +
		Col(strings.Repeat("-", ColColor), ColColor) + "  " +
		Col(strings.Repeat("-", ColDesc), ColDesc) + "  " +
		Col(strings.Repeat("-", ColCTypes), ColCTypes) + "  " +
		Col(strings.Repeat("-", ColWeight), ColWeight)

	fmt.Println(header)
	fmt.Println(sep)
	for _, e := range entries {
		fmt.Println(
			Col(e.Name, ColMetaName) + "  " +
				Col(e.Color, ColColor) + "  " +
				Col(e.Description, ColDesc) + "  " +
				Col(strings.Join(e.ContentTypes, ","), ColCTypes) + "  " +
				Col(strconv.Itoa(e.Weight), ColWeight),
		)
	}
}

// PrintEntriesTable renders metadata entries without a WEIGHT column.
func PrintEntriesTable(entries []devicetypes.MetadataEntry) {
	header := Col("NAME", ColMetaName) + "  " +
		Col("COLOR", ColColor) + "  " +
		Col("DESCRIPTION", ColDesc) + "  " +
		Col("CONTENT TYPES", ColCTypes)
	sep := Col(strings.Repeat("-", ColMetaName), ColMetaName) + "  " +
		Col(strings.Repeat("-", ColColor), ColColor) + "  " +
		Col(strings.Repeat("-", ColDesc), ColDesc) + "  " +
		Col(strings.Repeat("-", ColCTypes), ColCTypes)

	fmt.Println(header)
	fmt.Println(sep)
	for _, e := range entries {
		fmt.Println(
			Col(e.Name, ColMetaName) + "  " +
				Col(e.Color, ColColor) + "  " +
				Col(e.Description, ColDesc) + "  " +
				Col(strings.Join(e.ContentTypes, ","), ColCTypes),
		)
	}
}
