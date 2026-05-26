package example

import (
	"testing"

	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.RawRecords != nil {
		t.Errorf("New() RawRecords should be nil, got %v", p.RawRecords)
	}
}

func TestSlug(t *testing.T) {
	p := New()
	if got := p.Slug(); got != "example" {
		t.Errorf("Slug() = %q, want %q", got, "example")
	}
}

func TestRecordManagement(t *testing.T) {
	p := New()

	records := []import_.CsvRecord{
		{PartNumber: "P1", Description: "Device 1", Quantity: 1},
		{PartNumber: "P2", Description: "Device 2", Quantity: 2},
	}

	p.SetRecords(records)

	got := p.GetRecords()
	if len(got) != 2 {
		t.Errorf("GetRecords() len = %d, want 2", len(got))
	}
	if got[0].PartNumber != "P1" {
		t.Errorf("GetRecords()[0].PartNumber = %q, want %q", got[0].PartNumber, "P1")
	}

	p.ClearRecords()
	if got := p.GetRecords(); got != nil {
		t.Errorf("GetRecords() after ClearRecords() = %v, want nil", got)
	}
}
