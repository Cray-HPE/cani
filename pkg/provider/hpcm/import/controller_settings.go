package import_

// ControllerSettings holds controller configuration from HPCM.
type ControllerSettings struct {
	Type       string `json:"type,omitempty"`
	IpAddress  string `json:"ipAddress,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Channel    int32  `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}
