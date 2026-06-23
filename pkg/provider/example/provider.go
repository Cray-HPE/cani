package example

import import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"

// Example implements the provider.Provider interface
type Example struct {
	Options *ImportOptions
	// RawRecords holds all raw CSV records from the import phase
	RawRecords []import_.CsvRecord
	// DcimData holds parsed DCIM CSV data (multi-section format)
	DcimData *import_.DcimCSV
}

// New creates a new Example provider instance
func New() *Example {
	return &Example{}
}

// ClearRecords resets the raw record storage for a fresh import
func (p *Example) ClearRecords() {
	p.RawRecords = nil
}

// SetRecords stores raw records from the import phase
func (p *Example) SetRecords(records []import_.CsvRecord) {
	p.RawRecords = records
}

// GetRecords returns the raw records for the transform phase
func (p *Example) GetRecords() []import_.CsvRecord {
	return p.RawRecords
}

// SetDcimRecords stores parsed DCIM CSV data
func (p *Example) SetDcimRecords(data *import_.DcimCSV) {
	p.DcimData = data
}

// GetDcimRecords returns the DCIM CSV data for transform
func (p *Example) GetDcimRecords() *import_.DcimCSV {
	return p.DcimData
}

// ClearDcimRecords resets DCIM CSV data
func (p *Example) ClearDcimRecords() {
	p.DcimData = nil
}

// IsDcimImport returns true if a DCIM CSV was parsed
func (p *Example) IsDcimImport() bool {
	return p.DcimData != nil
}

func (p *Example) Slug() string {
	return "example"
}
