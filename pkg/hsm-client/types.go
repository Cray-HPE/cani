package hms_client

// From: https://github.com/Cray-HPE/hms-smd/blob/master/cmd/smd/smd-api.go
type HMSValues struct {
	Arch    []string `json:"Arch,omitempty"`
	Class   []string `json:"Class,omitempty"`
	Flag    []string `json:"Flag,omitempty"`
	NetType []string `json:"NetType,omitempty"`
	Role    []string `json:"Role,omitempty"`
	SubRole []string `json:"SubRole,omitempty"`
	State   []string `json:"State,omitempty"`
	Type    []string `json:"Type,omitempty"`
}
