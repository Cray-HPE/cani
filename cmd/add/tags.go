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
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

// collectProviderMetadata parses --metadata key=value flags into a map.
// Returns nil if the flag was not set or parsing fails.
func collectProviderMetadata(cmd *cobra.Command) map[string]string {
	pairs, _ := cmd.Flags().GetStringArray("metadata")
	if len(pairs) == 0 {
		return nil
	}
	m, err := ParseMetadataFlags(pairs)
	if err != nil {
		return nil
	}
	return m
}

// applyTagsToDevice appends tags to a device.
func applyTagsToDevice(device *devicetypes.CaniDeviceType, tags []string) {
	if len(tags) > 0 {
		device.Tags = append(device.Tags, tags...)
	}
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

// applyTagsToRack appends tags to a rack.
func applyTagsToRack(rack *devicetypes.CaniRackType, tags []string) {
	if len(tags) > 0 {
		rack.Tags = append(rack.Tags, tags...)
	}
}

// applyProviderMetadataToRack merges metadata into the rack's provider metadata.
func applyProviderMetadataToRack(rack *devicetypes.CaniRackType, meta map[string]string) {
	applyProviderMetadataMap(&rack.ProviderMetadata, meta)
}
