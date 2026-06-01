package devicetypes

import "github.com/google/uuid"

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

// GetID returns the unique identifier.
func (ip *CaniIPAddress) GetID() uuid.UUID {
	if ip == nil {
		return uuid.Nil
	}
	return ip.ID
}
