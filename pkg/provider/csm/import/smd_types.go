package import_

// SmdComponentList is the top-level SMD/HSM state components response.
type SmdComponentList struct {
	Components []SmdComponent `json:"Components"`
}

// SmdComponent represents a single hardware component from HSM State Manager.
type SmdComponent struct {
	ID      string `json:"ID"`
	Type    string `json:"Type"`
	State   string `json:"State"`
	Flag    string `json:"Flag"`
	Enabled bool   `json:"Enabled"`
	Role    string `json:"Role,omitempty"`
	SubRole string `json:"SubRole,omitempty"`
	NID     int    `json:"NID,omitempty"`
	NetType string `json:"NetType,omitempty"`
	Arch    string `json:"Arch,omitempty"`
	Class   string `json:"Class,omitempty"`
	Locked  bool   `json:"Locked,omitempty"`
}
