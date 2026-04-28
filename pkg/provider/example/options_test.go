package example

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestGetDefaultOptions(t *testing.T) {
	p := New()
	opts := p.GetDefaultOptions()
	if opts == nil {
		t.Fatal("GetDefaultOptions() returned nil")
	}
}

func TestGetOptionsStruct(t *testing.T) {
	p := New()
	s := p.GetOptionsStruct()
	if _, ok := s.(*Options); !ok {
		t.Errorf("GetOptionsStruct() returned %T, want *Options", s)
	}
}

func TestGetImportOptionsStruct(t *testing.T) {
	p := New()
	s := p.GetImportOptionsStruct()
	if _, ok := s.(*ImportOptions); !ok {
		t.Errorf("GetImportOptionsStruct() returned %T, want *ImportOptions", s)
	}
}

func TestGetImportDefaults(t *testing.T) {
	p := New()
	opts := p.GetImportDefaults()
	if opts == nil {
		t.Fatal("GetImportDefaults() returned nil")
	}
}

func TestBindImportFlags(t *testing.T) {
	p := New()
	cmd := &cobra.Command{Use: "test"}
	if err := p.BindImportFlags(cmd); err != nil {
		t.Errorf("BindImportFlags() error = %v", err)
	}
}

func TestGetExportOptionsStruct(t *testing.T) {
	p := New()
	s := p.GetExportOptionsStruct()
	if _, ok := s.(*ExportOptions); !ok {
		t.Errorf("GetExportOptionsStruct() returned %T, want *ExportOptions", s)
	}
}

func TestGetExportDefaults(t *testing.T) {
	p := New()
	opts := p.GetExportDefaults()
	if opts == nil {
		t.Fatal("GetExportDefaults() returned nil")
	}
}

func TestBindExportFlags(t *testing.T) {
	p := New()
	cmd := &cobra.Command{Use: "test"}
	if err := p.BindExportFlags(cmd); err != nil {
		t.Errorf("BindExportFlags() error = %v", err)
	}
}
