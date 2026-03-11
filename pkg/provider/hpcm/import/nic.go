package import_

import "time"

// Nic represents a network interface card from HPCM.
type Nic struct {
	Name             string            `json:"name,omitempty"`
	ID               int64             `json:"id,omitempty"`
	UUID             string            `json:"uuid,omitempty"`
	Etag             string            `json:"etag,omitempty"`
	CreationTime     time.Time         `json:"creationTime,omitempty"`
	ModificationTime time.Time         `json:"modificationTime,omitempty"`
	DeletionTime     time.Time         `json:"deletionTime,omitempty"`
	Links            map[string]string `json:"links,omitempty"`
	Network          string            `json:"network,omitempty"`
	IpAddress        string            `json:"ipAddress,omitempty"`
	Ipv6Address      string            `json:"ipv6Address,omitempty"`
	MacAddress       string            `json:"macAddress,omitempty"`
	BondingMaster    string            `json:"bondingMaster,omitempty"`
	BondingMode      string            `json:"bondingMode,omitempty"`
	InterfaceName    string            `json:"interfaceName,omitempty"`
	Managed          bool              `json:"managed,omitempty"`
	Type             string            `json:"type,omitempty"`
	Node             string            `json:"node,omitempty"`
	NodeName         string            `json:"nodeName,omitempty"`
	Controller       string            `json:"controller,omitempty"`
	ControllerName   string            `json:"controllerName,omitempty"`
	NetworkName      string            `json:"networkName,omitempty"`
	Attributes       map[string]any    `json:"attributes,omitempty"`
}
