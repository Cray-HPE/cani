package csm

import (
	"github.com/Cray-HPE/cani/pkg/provider/csm/client"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// Csm implements the provider.Provider interface
type Csm struct {
	Options *ImportOptions
	Client  *client.Client
	RawSls  *import_.SlsDumpstate
	RawSmd  *import_.SmdComponentList
}

// New creates a new Csm provider instance
func New() *Csm {
	return &Csm{}
}

// ClearRawData resets all raw import data for a fresh import.
func (p *Csm) ClearRawData() {
	p.RawSls = nil
	p.RawSmd = nil
}

// SetSls stores the parsed SLS dumpstate from the import phase.
func (p *Csm) SetSls(sls *import_.SlsDumpstate) {
	p.RawSls = sls
}

// GetSls returns the SLS dumpstate for the transform phase.
func (p *Csm) GetSls() *import_.SlsDumpstate {
	return p.RawSls
}

// SetSmd stores the parsed SMD component list from the import phase.
func (p *Csm) SetSmd(smd *import_.SmdComponentList) {
	p.RawSmd = smd
}

// GetSmd returns the SMD component list for the transform phase.
func (p *Csm) GetSmd() *import_.SmdComponentList {
	return p.RawSmd
}

func (p *Csm) Slug() string {
	return "csm"
}

// SetClient stores an authenticated API client on the provider.
func (p *Csm) SetClient(c *client.Client) {
	p.Client = c
}

// GetClient returns the authenticated API client.
func (p *Csm) GetClient() *client.Client {
	return p.Client
}
