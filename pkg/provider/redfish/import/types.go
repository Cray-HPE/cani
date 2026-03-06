package import_

import (
	"encoding/json"
	"fmt"
)

// ServiceRoot represents a Redfish ServiceRoot response from a BMC.
// Only fields useful for inventory classification and metadata are included.
type ServiceRoot struct {
	OdataType      string  `json:"@odata.type"`
	ID             string  `json:"Id"`
	Name           string  `json:"Name"`
	Product        string  `json:"Product"`
	Vendor         string  `json:"Vendor"`
	UUID           string  `json:"UUID"`
	RedfishVersion string  `json:"RedfishVersion"`
	Oem            OemData `json:"Oem"`
}

// OemData holds vendor-specific OEM extensions.
type OemData struct {
	Hpe *HpeOem `json:"Hpe,omitempty"`
}

// HpeOem holds HPE iLO-specific OEM data.
type HpeOem struct {
	Manager  []HpeManager `json:"Manager,omitempty"`
	Moniker  HpeMoniker   `json:"Moniker,omitempty"`
	Sessions HpeSessions  `json:"Sessions,omitempty"`
}

// HpeManager holds iLO manager details.
type HpeManager struct {
	DefaultLanguage        string    `json:"DefaultLanguage,omitempty"`
	FQDN                   string    `json:"FQDN,omitempty"`
	HostName               string    `json:"HostName,omitempty"`
	ManagerFirmwareVersion string    `json:"ManagerFirmwareVersion,omitempty"`
	ManagerType            string    `json:"ManagerType,omitempty"`
	Status                 HpeStatus `json:"Status,omitempty"`
}

// HpeStatus holds a simple health status.
type HpeStatus struct {
	Health string `json:"Health,omitempty"`
}

// HpeMoniker holds HPE product naming metadata.
type HpeMoniker struct {
	PRODTAG string `json:"PRODTAG,omitempty"`
	SYSFAM  string `json:"SYSFAM,omitempty"`
	VENDABR string `json:"VENDABR,omitempty"`
	VENDNAM string `json:"VENDNAM,omitempty"`
	PRODNAM string `json:"PRODNAM,omitempty"`
	PRODGEN string `json:"PRODGEN,omitempty"`
	PRODABR string `json:"PRODABR,omitempty"`
	BMC     string `json:"BMC,omitempty"`
}

// HpeSessions holds session info including the server name.
type HpeSessions struct {
	ServerName string `json:"ServerName,omitempty"`
}

// ParseServiceRoots parses JSON data containing either a single ServiceRoot
// object or a JSON array of ServiceRoot objects. Returns the parsed roots.
func ParseServiceRoots(data []byte) ([]ServiceRoot, error) {
	// Try array first.
	var roots []ServiceRoot
	if err := json.Unmarshal(data, &roots); err == nil && len(roots) > 0 {
		return roots, nil
	}

	// Fall back to single object.
	var root ServiceRoot
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parsing Redfish ServiceRoot JSON: %w", err)
	}

	// Validate we got something useful.
	if root.ID == "" && root.Product == "" && root.UUID == "" {
		return nil, fmt.Errorf("JSON does not appear to be a Redfish ServiceRoot")
	}

	return []ServiceRoot{root}, nil
}

// ManagerType returns the BMC manager type (e.g. "iLO 6"), or empty string.
func (r *ServiceRoot) ManagerType() string {
	if r.Oem.Hpe != nil && len(r.Oem.Hpe.Manager) > 0 {
		return r.Oem.Hpe.Manager[0].ManagerType
	}
	return ""
}

// ManagerFirmwareVersion returns the BMC firmware version, or empty string.
func (r *ServiceRoot) ManagerFirmwareVersion() string {
	if r.Oem.Hpe != nil && len(r.Oem.Hpe.Manager) > 0 {
		return r.Oem.Hpe.Manager[0].ManagerFirmwareVersion
	}
	return ""
}

// ManagerFQDN returns the BMC FQDN, or empty string.
func (r *ServiceRoot) ManagerFQDN() string {
	if r.Oem.Hpe != nil && len(r.Oem.Hpe.Manager) > 0 {
		return r.Oem.Hpe.Manager[0].FQDN
	}
	return ""
}

// ManagerHostName returns the BMC hostname, or empty string.
func (r *ServiceRoot) ManagerHostName() string {
	if r.Oem.Hpe != nil && len(r.Oem.Hpe.Manager) > 0 {
		return r.Oem.Hpe.Manager[0].HostName
	}
	return ""
}

// ManagerHealth returns the BMC health status, or empty string.
func (r *ServiceRoot) ManagerHealth() string {
	if r.Oem.Hpe != nil && len(r.Oem.Hpe.Manager) > 0 {
		return r.Oem.Hpe.Manager[0].Status.Health
	}
	return ""
}

// ProductTag returns the HPE product tag (e.g. "HPE iLO 6"), or empty string.
func (r *ServiceRoot) ProductTag() string {
	if r.Oem.Hpe != nil {
		return r.Oem.Hpe.Moniker.PRODTAG
	}
	return ""
}

// SystemFamily returns the HPE system family (e.g. "ProLiant"), or empty string.
func (r *ServiceRoot) SystemFamily() string {
	if r.Oem.Hpe != nil {
		return r.Oem.Hpe.Moniker.SYSFAM
	}
	return ""
}
