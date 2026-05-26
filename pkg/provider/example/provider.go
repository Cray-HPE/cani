package example

import import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"

// Example implements the provider.Provider interface
type Example struct {
	Options *ImportOptions
	// RawRecords holds all raw CSV records from the import phase
	RawRecords []import_.CsvRecord
	// SystemData holds parsed system CSV data (multi-section format)
	SystemData *import_.SystemCSV
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

// SetSystemRecords stores parsed system CSV data
func (p *Example) SetSystemRecords(data *import_.SystemCSV) {
	p.SystemData = data
}

// GetSystemRecords returns the system CSV data for transform
func (p *Example) GetSystemRecords() *import_.SystemCSV {
	return p.SystemData
}

// ClearSystemRecords resets system CSV data
func (p *Example) ClearSystemRecords() {
	p.SystemData = nil
}

// IsSystemImport returns true if a system CSV was parsed
func (p *Example) IsSystemImport() bool {
	return p.SystemData != nil
}

func (p *Example) Slug() string {
	return "example"
}
