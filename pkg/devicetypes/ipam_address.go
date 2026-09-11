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

	"github.com/google/uuid"
)

// IPAddressType classifies how an IP address is used.
type IPAddressType string

const (
	IPAddressTypeHost  IPAddressType = "host"
	IPAddressTypeDHCP  IPAddressType = "dhcp"
	IPAddressTypeSLAAC IPAddressType = "slaac"
)

// IPAddressRole indicates a special function of an IP address.
type IPAddressRole string

const (
	IPAddressRoleLoopback  IPAddressRole = "loopback"
	IPAddressRoleSecondary IPAddressRole = "secondary"
	IPAddressRoleAnycast   IPAddressRole = "anycast"
	IPAddressRoleVIP       IPAddressRole = "vip"
	IPAddressRoleVRRP      IPAddressRole = "vrrp"
	IPAddressRoleHSRP      IPAddressRole = "hsrp"
	IPAddressRoleGLBP      IPAddressRole = "glbp"
)

// CaniIPAddress represents a single host address with its subnet mask.
// IP addresses are organized under their parent prefix and can be
// assigned to one or more interfaces.
type CaniIPAddress struct {
	// Identity
	ID         uuid.UUID `json:"id" yaml:"id"`
	Host       string    `json:"host" yaml:"host"`              // IP without mask: "10.0.0.1"
	MaskLength int       `json:"maskLength" yaml:"mask_length"` // Prefix length: 24
	Address    string    `json:"address" yaml:"address"`        // Combined CIDR: "10.0.0.1/24"
	IPVersion  int       `json:"ipVersion" yaml:"ip_version"`   // 4 or 6

	// Classification
	Type        IPAddressType `json:"type,omitempty" yaml:"type,omitempty"`        // host, dhcp, slaac
	IPRole      IPAddressRole `json:"ipRole,omitempty" yaml:"ip_role,omitempty"`   // loopback, vip, etc.
	DNSName     string        `json:"dnsName,omitempty" yaml:"dns_name,omitempty"` // Optional forward DNS name
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`

	// Relationships
	Parent     uuid.UUID   `json:"parent,omitempty" yaml:"parent,omitempty"`         // Parent prefix (auto-computed)
	Interfaces []uuid.UUID `json:"interfaces,omitempty" yaml:"interfaces,omitempty"` // Assigned interface IDs
	NATInside  uuid.UUID   `json:"natInside,omitempty" yaml:"nat_inside,omitempty"`  // NAT inside IP (optional)

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`
}

// Validate checks that the address is valid without mutating the receiver.
func (ip *CaniIPAddress) Validate() error {
	if ip == nil {
		return fmt.Errorf("cannot validate nil CaniIPAddress")
	}
	copy := *ip
	return ParseIPAddress(&copy)
}

// GetID returns the unique identifier.
func (ip *CaniIPAddress) GetID() uuid.UUID {
	if ip == nil {
		return uuid.Nil
	}
	return ip.ID
}

// GetSlug returns the address as its portable natural key.
func (ip *CaniIPAddress) GetSlug() string {
	if ip == nil {
		return ""
	}
	return ip.Address
}

// GetStatus returns the current status.
func (ip *CaniIPAddress) GetStatus() string {
	if ip == nil {
		return ""
	}
	return ip.Status
}
