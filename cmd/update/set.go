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
package update

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/customfields"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// applyTagFlag replaces the item's tags with the --tag values when the flag
// was given, and reports whether it was.
func applyTagFlag(cmd *cli.Command, meta *devicetypes.ObjectMeta) bool {
	if !cmd.Flags().Changed("tag") {
		return false
	}
	meta.Tags, _ = cmd.Flags().GetStringArray("tag")
	return true
}

// applyLabelFlags applies the flags `update` hands to every subcommand that
// edit what an item is labelled with: --tag replaces its tags, and
// --custom-field and --clear-custom-field edit its custom fields as an
// object of contentType. It reports whether any of them was given.
func applyLabelFlags(cmd *cli.Command, inv *devicetypes.Inventory, meta *devicetypes.ObjectMeta, contentType string) (bool, error) {
	if err := customfields.Apply(cmd, inv, meta, contentType); err != nil {
		return false, err
	}
	tagged := applyTagFlag(cmd, meta)
	return tagged || cmd.Flags().Changed(customfields.FlagSet) || cmd.Flags().Changed(customfields.FlagClear), nil
}

// parseSetFlags splits --set key=value pairs into a map.
func parseSetFlags(pairs []string) (map[string]string, error) {
	result := make(map[string]string, len(pairs))
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --set value %q; expected key=value", p)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}
