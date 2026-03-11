package import_

// NetworkSettings holds network configuration from HPCM.
type NetworkSettings struct {
	Name           string `json:"name,omitempty"`
	DefaultGateway string `json:"defaultGateway,omitempty"`
	Nics           []Nic  `json:"nics,omitempty"`
	IpAddress      string `json:"ipAddress,omitempty"`
	MacAddress     string `json:"macAddress,omitempty"`
	SubnetMask     string `json:"subnetMask,omitempty"`
	MgmtServerIp   string `json:"mgmtServerIp,omitempty"`
}
