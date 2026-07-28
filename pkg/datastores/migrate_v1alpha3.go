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

package datastores

import (
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// migrateV1Alpha3 upgrades a v1alpha3 inventory to v1alpha4 in place. The new
// contract adds VRFs, interface VLAN/LAG fields, and device relationship fields.
// Existing records need no conversion, but the VRF collection must be writable.
func migrateV1Alpha3(inv *devicetypes.Inventory) {
	if inv.VRFs == nil {
		inv.VRFs = make(map[uuid.UUID]*devicetypes.CaniVRF)
	}
	inv.SchemaVersion = devicetypes.SchemaVersionV1Alpha4
}
