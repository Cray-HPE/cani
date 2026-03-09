package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

func TestResolveSlug_RiverCabinet(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeCabinet,
		Xname:    ParseXname("x3000"),
		Hardware: import_.SlsHardware{Xname: "x3000", Class: "River"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-eia-cabinet" {
		t.Errorf("resolveSlug(River cabinet) = %q, want %q", slug, "hpe-eia-cabinet")
	}
}

func TestResolveSlug_RiverChassis(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeChassis,
		Xname:    ParseXname("x3000c0"),
		Hardware: import_.SlsHardware{Xname: "x3000c0", Class: "River"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-eia-chassis" {
		t.Errorf("resolveSlug(River chassis) = %q, want %q", slug, "hpe-eia-chassis")
	}
}

func TestResolveSlug_MountainCabinet(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeCabinet,
		Xname:    ParseXname("x1000"),
		Hardware: import_.SlsHardware{Xname: "x1000", Class: "Mountain"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-ex2000" {
		t.Errorf("resolveSlug(Mountain cabinet) = %q, want %q", slug, "hpe-ex2000")
	}
}

func TestResolveSlug_MountainChassis(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeChassis,
		Xname:    ParseXname("x1000c0"),
		Hardware: import_.SlsHardware{Xname: "x1000c0", Class: "Mountain"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-crayex-chassis" {
		t.Errorf("resolveSlug(Mountain chassis) = %q, want %q", slug, "hpe-crayex-chassis")
	}
}

func TestResolveSlug_ClassFromXname(t *testing.T) {
	// Empty Class field but x3000 → River.
	cl := CsmClassification{
		CaniType: devicetypes.TypeCabinet,
		Xname:    ParseXname("x3000"),
		Hardware: import_.SlsHardware{Xname: "x3000", Class: ""},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-eia-cabinet" {
		t.Errorf("resolveSlug(empty class, x3000) = %q, want %q", slug, "hpe-eia-cabinet")
	}
}

func TestResolveSlug_NoMatchReturnEmpty(t *testing.T) {
	// A type with no entry in defaultSlugs should return "".
	cl := CsmClassification{
		CaniType: devicetypes.TypeCDU,
		Xname:    ParseXname("x3000c0"),
		Hardware: import_.SlsHardware{Xname: "x3000c0", Class: "River"},
	}
	slug := resolveSlug(cl)
	if slug != "" {
		t.Errorf("resolveSlug(River CDU) = %q, want empty", slug)
	}
}

func TestResolveSlug_RiverNode(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeNode,
		Xname:    ParseXname("x3000c0s1b0n0"),
		Hardware: import_.SlsHardware{Xname: "x3000c0s1b0n0", Class: "River"},
	}
	slug := resolveSlug(cl)
	if slug != "cray-xd225v" {
		t.Errorf("resolveSlug(River node) = %q, want %q", slug, "cray-xd225v")
	}
}

func TestResolveSlug_RiverMgmtSwitch(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeMgmtSwitch,
		Xname:    ParseXname("x3000c0w1"),
		Hardware: import_.SlsHardware{Xname: "x3000c0w1", Class: "River"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-aruba-6300m-48g" {
		t.Errorf("resolveSlug(River MgmtSwitch) = %q, want %q", slug, "hpe-aruba-6300m-48g")
	}
}

func TestResolveSlug_RiverSpineSwitch(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeSwitch,
		Xname:    ParseXname("x3000c0h1s1"),
		Hardware: import_.SlsHardware{Xname: "x3000c0h1s1", Class: "River"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-aruba-8325-32c" {
		t.Errorf("resolveSlug(River spine) = %q, want %q", slug, "hpe-aruba-8325-32c")
	}
}

func TestDefaultSlugForClass_Hill(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeCabinet,
		Xname:    ParseXname("x5000"),
		Hardware: import_.SlsHardware{Xname: "x5000", Class: "Hill"},
	}
	slug := resolveSlug(cl)
	// Hill cabinets use the same EX2000 enclosure as Mountain.
	if slug != "hpe-ex2000" {
		t.Errorf("resolveSlug(Hill cabinet) = %q, want %q", slug, "hpe-ex2000")
	}
}
