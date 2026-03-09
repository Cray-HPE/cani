package import_

// SlsDumpstate is the top-level SLS state returned by /dumpstate.
type SlsDumpstate struct {
	Hardware map[string]SlsHardware `json:"Hardware"`
	Networks map[string]SlsNetwork  `json:"Networks"`
}

// SlsHardware represents a single hardware entry in SLS.
type SlsHardware struct {
	Parent          string         `json:"Parent"`
	Xname           string         `json:"Xname"`
	Children        []string       `json:"Children,omitempty"`
	Type            string         `json:"Type"`
	Class           string         `json:"Class"`
	TypeString      string         `json:"TypeString"`
	LastUpdated     int64          `json:"LastUpdated,omitempty"`
	LastUpdatedTime string         `json:"LastUpdatedTime,omitempty"`
	ExtraProperties map[string]any `json:"ExtraProperties,omitempty"`
}

// SlsNetwork represents an SLS network definition.
type SlsNetwork struct {
	Name            string                     `json:"Name"`
	FullName        string                     `json:"FullName"`
	IPRanges        []string                   `json:"IPRanges"`
	Type            string                     `json:"Type"`
	LastUpdated     int64                      `json:"LastUpdated,omitempty"`
	LastUpdatedTime string                     `json:"LastUpdatedTime,omitempty"`
	ExtraProperties *SlsNetworkExtraProperties `json:"ExtraProperties,omitempty"`
}

// SlsNetworkExtraProperties contains network-level metadata and subnets.
type SlsNetworkExtraProperties struct {
	CIDR               string      `json:"CIDR,omitempty"`
	MTU                int         `json:"MTU,omitempty"`
	Comment            string      `json:"Comment,omitempty"`
	Subnets            []SlsSubnet `json:"Subnets,omitempty"`
	SystemDefaultRoute string      `json:"SystemDefaultRoute,omitempty"`
	VlanRange          []int       `json:"VlanRange,omitempty"`
}

// SlsSubnet represents a subnet within an SLS network.
type SlsSubnet struct {
	Name            string             `json:"Name"`
	FullName        string             `json:"FullName"`
	CIDR            string             `json:"CIDR"`
	Gateway         string             `json:"Gateway,omitempty"`
	VlanID          int                `json:"VlanID,omitempty"`
	DHCPStart       string             `json:"DHCPStart,omitempty"`
	DHCPEnd         string             `json:"DHCPEnd,omitempty"`
	MetalLBPoolName string             `json:"MetalLBPoolName,omitempty"`
	IPReservations  []SlsIPReservation `json:"IPReservations,omitempty"`
}

// SlsIPReservation represents an IP reservation within a subnet.
type SlsIPReservation struct {
	Name      string   `json:"Name"`
	IPAddress string   `json:"IPAddress"`
	Comment   string   `json:"Comment,omitempty"`
	Aliases   []string `json:"Aliases,omitempty"`
}

// Typed extra-properties structs for known hardware types.
// These are decoded from the generic ExtraProperties map via DecodeExtraProperties.

// SlsNodeExtraProperties holds extra properties for Node hardware.
type SlsNodeExtraProperties struct {
	NID     int      `json:"NID,omitempty"`
	Role    string   `json:"Role,omitempty"`
	SubRole string   `json:"SubRole,omitempty"`
	Aliases []string `json:"Aliases,omitempty"`
}

// SlsCabinetExtraProperties holds extra properties for Cabinet hardware.
type SlsCabinetExtraProperties struct {
	Networks map[string]map[string]SlsCabinetNetwork `json:"Networks,omitempty"`
}

// SlsCabinetNetwork holds CIDR, gateway, and VLAN for a cabinet network.
type SlsCabinetNetwork struct {
	CIDR    string `json:"CIDR"`
	Gateway string `json:"Gateway"`
	VLan    int    `json:"VLan"`
}

// SlsMgmtSwitchExtraProperties holds extra properties for MgmtSwitch hardware.
type SlsMgmtSwitchExtraProperties struct {
	Aliases          []string `json:"Aliases,omitempty"`
	Brand            string   `json:"Brand,omitempty"`
	IP4Addr          string   `json:"IP4addr,omitempty"`
	Model            string   `json:"Model,omitempty"`
	SNMPAuthPassword string   `json:"SNMPAuthPassword,omitempty"`
	SNMPAuthProtocol string   `json:"SNMPAuthProtocol,omitempty"`
	SNMPPrivPassword string   `json:"SNMPPrivPassword,omitempty"`
	SNMPPrivProtocol string   `json:"SNMPPrivProtocol,omitempty"`
	SNMPUsername     string   `json:"SNMPUsername,omitempty"`
}

// SlsMgmtHLSwitchExtraProperties holds extra properties for MgmtHLSwitch (spine) hardware.
type SlsMgmtHLSwitchExtraProperties struct {
	Aliases []string `json:"Aliases,omitempty"`
	Brand   string   `json:"Brand,omitempty"`
	IP4Addr string   `json:"IP4addr,omitempty"`
	Model   string   `json:"Model,omitempty"`
}

// SlsMgmtSwitchConnectorExtraProperties holds extra properties for MgmtSwitchConnector.
type SlsMgmtSwitchConnectorExtraProperties struct {
	NodeNics   []string `json:"NodeNics,omitempty"`
	VendorName string   `json:"VendorName,omitempty"`
}

// SlsRouterBMCExtraProperties holds extra properties for RouterBMC hardware.
type SlsRouterBMCExtraProperties struct {
	Password string `json:"Password,omitempty"`
	Username string `json:"Username,omitempty"`
}
