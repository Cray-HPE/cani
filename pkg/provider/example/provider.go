package example

import import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"

// Example implements the provider.Provider interface
type Example struct {
	Options *ImportOptions
	// RawRecords holds all raw CSV records from the import phase
	RawRecords []import_.CsvRecord
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

func (p *Example) Slug() string {
	return "example"
}
