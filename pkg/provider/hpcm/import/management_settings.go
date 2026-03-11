package import_

// ManagementSettings holds management card configuration from HPCM.
type ManagementSettings struct {
	CardType       string `json:"cardType,omitempty"`
	CardIpAddress  string `json:"cardIpAddress,omitempty"`
	CardMacAddress string `json:"cardMacAddress,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	Channel        int32  `json:"channel,omitempty"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
}
