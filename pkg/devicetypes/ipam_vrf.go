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
	"strings"

	(
	"errors"
	"fmt"

	"github.com/google/uuid"
)
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

// GetID returns the unique identifier.
func (v *CaniVRF) GetID() uuid.UUID {
	if v == nil {
		return uuid.Nil
	}
	return v.ID
}

// FindVRFByNameOrID looks up a VRF by UUID string or exact name (case-insensitive).
func (inv *Inventory) FindVRFByNameOrID(arg string) (*CaniVRF, error) {
	if id, err := uuid.Parse(arg); err == nil {
		if vrf, ok := inv.VRFs[id]; ok {
			return vrf, nil
		}
		return nil, fmt.Errorf("VRF with UUID %q not found", arg)
	}
	for _, vrf := range inv.VRFs {
		if strings.EqualFold(vrf.Name, arg) {
			return vrf, nil
		}
	}
	return nil, fmt.Errorf("VRF %q not found", arg)
}

// Validate checks the VRF for internal consistency.
func (v *CaniVRF) Validate() error {
	if v == nil {
		return errors.New("cannot validate nil CaniVRF")
	}
	if v.Name == "" {
		return fmt.Errorf("VRF %s: name is required", v.ID)
	}
	return nil
}
