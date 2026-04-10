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

func TestDefaultSlugForClass_HillBlade(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeBlade,
		Xname:    ParseXname("x9000c1s0"),
		Hardware: import_.SlsHardware{Xname: "x9000c1s0", Class: "Hill"},
	}
	slug := resolveSlug(cl)
	// Hill blades should use Mountain (CrayEX) defaults, not River.
	if slug != "hpe-crayex-ex235a-compute-blade" {
		t.Errorf("resolveSlug(Hill blade) = %q, want %q", slug, "hpe-crayex-ex235a-compute-blade")
	}
}

func TestDefaultSlugForClass_HillNodeCard(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeNodeCard,
		Xname:    ParseXname("x9000c1s0b0"),
		Hardware: import_.SlsHardware{Xname: "x9000c1s0b0", Class: "Hill"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-crayex-ex235a-compute-blade-bard-peak-node-card" {
		t.Errorf("resolveSlug(Hill nodecard) = %q, want %q", slug, "hpe-crayex-ex235a-compute-blade-bard-peak-node-card")
	}
}

func TestDefaultSlugForClass_HillNode(t *testing.T) {
	cl := CsmClassification{
		CaniType: devicetypes.TypeNode,
		Xname:    ParseXname("x9000c1s0b0n0"),
		Hardware: import_.SlsHardware{Xname: "x9000c1s0b0n0", Class: "Hill"},
	}
	slug := resolveSlug(cl)
	if slug != "hpe-crayex-ex235a-compute-node" {
		t.Errorf("resolveSlug(Hill node) = %q, want %q", slug, "hpe-crayex-ex235a-compute-node")
	}
}

func TestClassesForSlug_DefaultSlug(t *testing.T) {
	classes := ClassesForSlug("hpe-crayex-ex235a-compute-blade")
	if !classes[ClassMountain] {
		t.Error("expected Mountain for EX235A")
	}
	if !classes[ClassHill] {
		t.Error("expected Hill for EX235A (inherits from Mountain)")
	}
	if classes[ClassRiver] {
		t.Error("unexpected River for EX235A")
	}
}

func TestClassesForSlug_NonDefaultSlug(t *testing.T) {
	// EX235N is not in defaultSlugs but shares the "hpe-crayex" family.
	classes := ClassesForSlug("hpe-crayex-ex235n-compute-blade")
	if !classes[ClassMountain] {
		t.Error("expected Mountain for EX235N (family match)")
	}
	if !classes[ClassHill] {
		t.Error("expected Hill for EX235N (inherits from Mountain)")
	}
	if classes[ClassRiver] {
		t.Error("unexpected River for EX235N")
	}
}

func TestClassesForSlug_RiverSlug(t *testing.T) {
	classes := ClassesForSlug("hpe-dl380-gen-11")
	if !classes[ClassRiver] {
		t.Error("expected River for DL380")
	}
	if classes[ClassMountain] {
		t.Error("unexpected Mountain for DL380")
	}
}

func TestSlugFamily(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"hpe-crayex-ex235a-compute-blade", "hpe-crayex"},
		{"hpe-crayex-ex235n-compute-blade", "hpe-crayex"},
		{"hpe-dl380-gen-11", "hpe-dl380"},
		{"hpe-aruba-8325-32c", "hpe-aruba"},
		{"cray-xd225v", "cray-xd225v"},
		{"hpe-ex2000", "hpe-ex2000"},
	}
	for _, tt := range tests {
		got := slugFamily(tt.slug)
		if got != tt.want {
			t.Errorf("slugFamily(%q) = %q, want %q", tt.slug, got, tt.want)
		}
	}
}
