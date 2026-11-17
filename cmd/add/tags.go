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
	"slices"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/customfields"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// labels holds the role, tags and custom fields an add stamps on each item.
// The role and tags are read once from --role and --tag; a noun without a
// --role flag stamps none. The custom fields are added by withCustomFields
// once the inventory that defines them is loaded.
type labels struct {
	role   string
	tags   []string
	fields customfields.Edit
}

// labelsFromFlags reads --role and --tag; an unset flag leaves the item's
// field as the hardware type defines it.
func labelsFromFlags(cmd *cli.Command) labels {
	role, _ := cmd.Flags().GetString(flagRole)
	tags, _ := cmd.Flags().GetStringArray("tag")
	return labels{role: role, tags: tags}
}

// withCustomFields adds the --custom-field values, checked against the
// inventory's definitions for objects of contentType.
func (l labels) withCustomFields(cmd *cli.Command, inv *devicetypes.Inventory, contentType string) (labels, error) {
	fields, err := customfields.FromFlags(cmd, inv, contentType)
	l.fields = fields
	return l, err
}

// apply sets the role that was given, appends the tags and sets the custom
// fields. The tag list and each custom-field value are copied, so items
// built from one hardware type never share them.
func (l labels) apply(meta *devicetypes.ObjectMeta) {
	if l.role != "" {
		meta.Role = l.role
	}
	if len(l.tags) > 0 {
		meta.Tags = slices.Concat(meta.Tags, l.tags)
	}
	l.fields.Stamp(meta)
}
