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
package transform

import (
	import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"
)

// hpcmBucket is the provider-metadata bucket that holds everything HPCM
// records about an object and the inventory has no field for. None of it is
// stored as a custom field: a custom field is an attribute an operator
// defines, and another provider's export would send it as one.
const hpcmBucket = "hpcm"

// nodeMetadata gathers what HPCM records about a node: its UUID, its
// aliases and its location.
func nodeMetadata(node import_.Node) map[string]any {
	meta := make(map[string]any)
	if node.UUID != "" {
		meta[metaKeyHpcmUUID] = node.UUID
	}
	if len(node.Aliases) > 0 {
		meta[metaKeyAliases] = node.Aliases
	}
	if node.Location != nil {
		meta["location"] = locationToMap(node.Location)
	}
	return meta
}

// inBucket returns the provider metadata of an object whose HPCM bucket
// holds meta.
func inBucket(meta map[string]any) map[string]any {
	return map[string]any{hpcmBucket: meta}
}
