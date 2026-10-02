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
package devicetypes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// CaniVRF represents a virtual routing and forwarding instance.
type CaniVRF struct {
	// Identity
	ID          uuid.UUID `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	RD          string    `json:"rd,omitempty" yaml:"rd,omitempty"` // Route distinguisher (RFC 4364)
	Namespace   string    `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`

	// Device assignments (Nautobot vrf-device-assignments)
	Devices []uuid.UUID `json:"devices,omitempty" yaml:"devices,omitempty"`

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`
}

// Validate checks that the VRF has a portable natural key.
func (v *CaniVRF) Validate() error {
	if v == nil {
		return fmt.Errorf("cannot validate nil CaniVRF")
	}
	if strings.TrimSpace(v.Name) == "" {
		return fmt.Errorf("VRF name must not be empty")
	}
	return nil
}

// GetID returns the unique identifier.
func (v *CaniVRF) GetID() uuid.UUID {
	if v == nil {
		return uuid.Nil
	}
	return v.ID
}

// GetSlug returns the VRF name as its portable natural key.
func (v *CaniVRF) GetSlug() string {
	if v == nil {
		return ""
	}
	return v.Name
}

// GetStatus returns the current status.
func (v *CaniVRF) GetStatus() string {
	if v == nil {
		return ""
	}
	return v.Status
}

// AddVRFDevice assigns the device to the VRF, as AddVRF requires of a new
// VRF's devices, and reports whether the assignment was new.
func (inv *Inventory) AddVRFDevice(vrf *CaniVRF, deviceID uuid.UUID) (bool, error) {
	if inv.Devices[deviceID] == nil {
		return false, fmt.Errorf("vrf %s: device %s not found", vrf.Name, deviceID)
	}
	if slices.Contains(vrf.Devices, deviceID) {
		return false, nil
	}
	vrf.Devices = append(vrf.Devices, deviceID)
	return true, nil
}

// RemoveDevice drops the device's assignment and reports whether it was
// assigned. The device need not exist, so an assignment left behind by a
// device that is gone can be dropped too.
func (v *CaniVRF) RemoveDevice(deviceID uuid.UUID) bool {
	if !slices.Contains(v.Devices, deviceID) {
		return false
	}
	v.Devices = removeUUID(v.Devices, deviceID)
	return true
}
