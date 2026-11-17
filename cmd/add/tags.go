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
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// labels holds the role and tags an add stamps on each item, read once from
// the --role and --tag flags. A noun without a --role flag stamps tags only.
type labels struct {
	role string
	tags []string
}

// labelsFromFlags reads --role and --tag; an unset flag leaves the item's
// field as the hardware type defines it.
func labelsFromFlags(cmd *cli.Command) labels {
	role, _ := cmd.Flags().GetString(flagRole)
	tags, _ := cmd.Flags().GetStringArray("tag")
	return labels{role: role, tags: tags}
}

// apply sets the role that was given and appends the tags. The tag list is
// copied, so items built from one hardware type never share it.
func (l labels) apply(meta *devicetypes.ObjectMeta) {
	if l.role != "" {
		meta.Role = l.role
	}
	if len(l.tags) > 0 {
		meta.Tags = slices.Concat(meta.Tags, l.tags)
	}
}

// collectProviderMetadata returns the --metadata key=value pairs, already
// validated by the flag parser, or nil when the flag was not set.
func collectProviderMetadata(cmd *cli.Command) map[string]string {
	meta, _ := cmd.Flags().GetStringToString("metadata")
	return meta
}

// applyProviderMetadataMap lets each registered provider merge metadata
// into its own section of the given provider-metadata map.
func applyProviderMetadataMap(pm *map[string]any, meta map[string]string) {
	if len(meta) == 0 {
		return
	}
	for _, p := range provider.GetProviders() {
		if ma, ok := p.(provider.MetadataApplier); ok {
			ma.ApplyMetadata(pm, meta)
		}
	}
}

// applyProviderMetadataToDevice merges metadata into the device's provider metadata.
func applyProviderMetadataToDevice(device *devicetypes.CaniDeviceType, meta map[string]string) {
	applyProviderMetadataMap(&device.ProviderMetadata, meta)
}

// applyProviderMetadataToRack merges metadata into the rack's provider metadata.
func applyProviderMetadataToRack(rack *devicetypes.CaniRackType, meta map[string]string) {
	applyProviderMetadataMap(&rack.ProviderMetadata, meta)
}
