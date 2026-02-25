package ochami

import import_ "github.com/Cray-HPE/cani/pkg/provider/ochami/import"

// Ochami implements the provider.Provider interface
type Ochami struct {
	Options *ImportOptions
	// RawRecords holds all raw JSON records from the import phase
	RawRecords []import_.JSONDeviceRecord
}

// New creates a new Ochami provider instance
func New() *Ochami {
	return &Ochami{}
}

// ClearRecords resets the raw record storage for a fresh import
func (p *Ochami) ClearRecords() {
	p.RawRecords = nil
}

// SetRecords stores raw records from the import phase
func (p *Ochami) SetRecords(records []import_.JSONDeviceRecord) {
	p.RawRecords = records
}

// GetRecords returns the raw records for the transform phase
func (p *Ochami) GetRecords() []import_.JSONDeviceRecord {
	return p.RawRecords
}

func (p *Ochami) Slug() string {
	return "ochami"
}
