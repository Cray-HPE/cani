package transform

// CsmMetadata holds CSM-specific metadata stored in ProviderMetadata["csm"].
type CsmMetadata struct {
	Xname   string   `json:"xname,omitempty"`
	Class   string   `json:"class,omitempty"`
	Role    string   `json:"role,omitempty"`
	SubRole string   `json:"subRole,omitempty"`
	NID     int      `json:"nid,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
	State   string   `json:"state,omitempty"`
	HMNVlan int      `json:"hmnVlan,omitempty"`
}

// toProviderMetadata converts CsmMetadata to the generic map form
// stored on CaniDeviceType.ProviderMetadata.
func toProviderMetadata(m CsmMetadata) map[string]any {
	md := map[string]any{}
	if m.Xname != "" {
		md["xname"] = m.Xname
	}
	if m.Class != "" {
		md["class"] = m.Class
	}
	if m.Role != "" {
		md["role"] = m.Role
	}
	if m.SubRole != "" {
		md["subRole"] = m.SubRole
	}
	if m.NID != 0 {
		md["nid"] = m.NID
	}
	if len(m.Aliases) > 0 {
		md["aliases"] = m.Aliases
	}
	if m.State != "" {
		md["state"] = m.State
	}
	if m.HMNVlan != 0 {
		md["hmnVlan"] = m.HMNVlan
	}
	return md
}
