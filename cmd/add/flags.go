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

import "github.com/Cray-HPE/cani/internal/cli"

// Shared CLI flag names used across multiple "add" subcommands. Centralizing
// them avoids duplicated string literals and keeps flag wiring consistent.
const (
	flagDescription = "description"
	flagLocation    = "location"
	flagInterface   = "interface"
	flagDNSName     = "dns-name"
	flagNamespace   = "namespace"
	flagBayFilter   = "bay-filter"
	flagDryRun      = "dry-run"
)

// Asset-tracking flags shared by every noun that records them.
const (
	flagSerial   = "serial"
	flagAssetTag = "asset-tag"
)

// identity holds the asset-tracking fields an add stamps on a device, rack
// or module, read once from the persistent --serial and --asset-tag flags.
type identity struct {
	serial   string
	assetTag string
}

// identityFromFlags reads the identity flags; an unset flag leaves the
// item's field as the hardware type defines it.
func identityFromFlags(cmd *cli.Command) identity {
	serial, _ := cmd.Flags().GetString(flagSerial)
	assetTag, _ := cmd.Flags().GetString(flagAssetTag)
	return identity{serial: serial, assetTag: assetTag}
}

// apply stamps the fields that were given onto the item's serial and asset
// tag. The same values go on every item of a multi-item add, as --serial
// has always done.
func (id identity) apply(serial, assetTag *string) {
	if id.serial != "" {
		*serial = id.serial
	}
	if id.assetTag != "" {
		*assetTag = id.assetTag
	}
}
