package redfish

import import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"

// Redfish implements the provider.Provider interface
type Redfish struct {
	Options  *ImportOptions
	RawRoots []import_.ServiceRoot
}

// New creates a new Redfish provider instance
func New() *Redfish {
	return &Redfish{}
}

// ClearRoots resets the raw root storage for a fresh import.
func (p *Redfish) ClearRoots() {
	p.RawRoots = nil
}

// SetRoots stores raw ServiceRoots from the import phase.
func (p *Redfish) SetRoots(roots []import_.ServiceRoot) {
	p.RawRoots = roots
}

// GetRoots returns the raw ServiceRoots for the transform phase.
func (p *Redfish) GetRoots() []import_.ServiceRoot {
	return p.RawRoots
}

func (p *Redfish) Slug() string {
	return "redfish"
}
