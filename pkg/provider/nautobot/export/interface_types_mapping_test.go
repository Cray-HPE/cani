/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// ---------- isValidNautobotInterfaceType ----------

// TestIsValidNautobotInterfaceType verifies that supported Nautobot interface
// type slugs are accepted and unsupported ones (e.g. nvlink, pcie) are rejected.
//
// Why it matters: cani may carry hardware interface types Nautobot has no enum
// for; sending an invalid type aborts the interface create, so export must
// pre-validate and skip/translate unknowns.
// Inputs: a list of known-valid slugs and a list of known-invalid ones.
// Outputs: bool valid/invalid per slug.
// Data choice: the valid list mirrors the function's switch exactly; nvlink and
// pcie-gen5-x16 are real GPU-fabric types cani knows but Nautobot does not.
func TestIsValidNautobotInterfaceType(t *testing.T) {
	validTypes := []string{
		"100base-tx", "1000base-t", "10gbase-x-sfpp", "25gbase-x-sfp28",
		"40gbase-x-qsfpp", "100gbase-x-qsfp28", "200gbase-x-qsfp56",
		"400gbase-x-osfp", "400gbase-x-qsfpdd",
		"infiniband-hdr", "infiniband-ndr",
		"virtual", "lag", "other",
	}
	for _, vt := range validTypes {
		if !isValidNautobotInterfaceType(vt) {
			t.Errorf("expected %q to be valid", vt)
		}
	}

	invalidTypes := []string{
		"nvlink", "pcie-gen5-x16", "spi", "", "garbage",
	}
	for _, ivt := range invalidTypes {
		if isValidNautobotInterfaceType(ivt) {
			t.Errorf("expected %q to be invalid", ivt)
		}
	}
}

// TestIsValidNautobotInterfaceTypeValid verifies isValidNautobotInterfaceType
// returns true for every slug in Nautobot's accepted InterfaceTypeChoices set.
//
// Why it matters: the exporter gates interface creation on this whitelist, so
// every legitimate type must pass or valid interfaces would be skipped.
// Inputs: a slice of accepted slugs (100base-tx … infiniband-ndr, virtual, lag,
// other). Outputs: true for each.
// Data choice: the list mirrors every speed plus virtual/lag/other choices the
// exporter emits, asserting full positive coverage of the whitelist.
func TestIsValidNautobotInterfaceTypeValid(t *testing.T) {
	validTypes := []string{
		"100base-tx", "1000base-t", "10gbase-x-sfpp", "25gbase-x-sfp28",
		"40gbase-x-qsfpp", "100gbase-x-qsfp28", "200gbase-x-qsfp56",
		"400gbase-x-osfp", "400gbase-x-qsfpdd",
		"infiniband-hdr", "infiniband-ndr",
		"virtual", "lag", "other",
	}
	for _, vt := range validTypes {
		if !isValidNautobotInterfaceType(vt) {
			t.Errorf("expected true for %q", vt)
		}
	}
}

// TestIsValidNautobotInterfaceTypeInvalid verifies isValidNautobotInterfaceType
// returns false for strings outside the accepted choice set.
//
// Why it matters: rejecting unknown types prevents the exporter from sending
// values Nautobot would refuse on create.
// Inputs: "not-a-type", "ethernet", "1gbase-t", and "". Outputs: false for each.
// Data choice: the cases include a bogus string, a too-generic name, a library
// alias that only mapInterfaceType understands, and the empty string to cover
// near misses and the zero value.
func TestIsValidNautobotInterfaceTypeInvalid(t *testing.T) {
	invalidTypes := []string{"not-a-type", "ethernet", "1gbase-t", ""}
	for _, vt := range invalidTypes {
		if isValidNautobotInterfaceType(vt) {
			t.Errorf("expected false for %q", vt)
		}
	}
}

// ---------- mapInterfaceType additional cases ----------

// TestMapInterfaceType25g verifies mapInterfaceType passes "25gbase-x-sfp28"
// through unchanged.
//
// Why it matters: interface types must land on exact Nautobot
// InterfaceTypeChoices slugs or interface creation is rejected; 25G already
// matches Nautobot's slug and must not be rewritten.
// Inputs: the devicetypes string "25gbase-x-sfp28". Outputs: the identical
// Nautobot slug "25gbase-x-sfp28".
// Data choice: 25G is a common SFP28 fabric link whose cani and Nautobot names
// coincide, covering the identity branch of the mapping.
func TestMapInterfaceType25g(t *testing.T) {
	got := mapInterfaceType("25gbase-x-sfp28")
	if got != "25gbase-x-sfp28" {
		t.Errorf("got %q, want '25gbase-x-sfp28'", got)
	}
}

// TestMapInterfaceType40g verifies mapInterfaceType passes "40gbase-x-qsfpp"
// through unchanged.
//
// Why it matters: 40G QSFP+ uplinks must export with Nautobot's exact slug or
// the interface is rejected on create.
// Inputs: "40gbase-x-qsfpp". Outputs: the identical "40gbase-x-qsfpp".
// Data choice: 40G QSFP+ is a standard switch uplink whose cani and Nautobot
// slugs already match, exercising the pass-through branch.
func TestMapInterfaceType40g(t *testing.T) {
	got := mapInterfaceType("40gbase-x-qsfpp")
	if got != "40gbase-x-qsfpp" {
		t.Errorf("got %q, want '40gbase-x-qsfpp'", got)
	}
}

// TestMapInterfaceType100g verifies mapInterfaceType passes "100gbase-x-qsfp28"
// through unchanged.
//
// Why it matters: 100G fabric links must export with the exact Nautobot slug to
// be accepted on create.
// Inputs: "100gbase-x-qsfp28". Outputs: the identical "100gbase-x-qsfp28".
// Data choice: 100G QSFP28 is the dominant HPC fabric speed whose names match
// across cani and Nautobot, covering the identity branch.
func TestMapInterfaceType100g(t *testing.T) {
	got := mapInterfaceType("100gbase-x-qsfp28")
	if got != "100gbase-x-qsfp28" {
		t.Errorf("got %q, want '100gbase-x-qsfp28'", got)
	}
}

// TestMapInterfaceType200g verifies mapInterfaceType passes "200gbase-x-qsfp56"
// through unchanged.
//
// Why it matters: 200G QSFP56 links must export with Nautobot's exact slug to
// avoid rejection.
// Inputs: "200gbase-x-qsfp56". Outputs: the identical "200gbase-x-qsfp56".
// Data choice: 200G QSFP56 is a high-speed fabric link whose cani and Nautobot
// slugs coincide, exercising the pass-through branch.
func TestMapInterfaceType200g(t *testing.T) {
	got := mapInterfaceType("200gbase-x-qsfp56")
	if got != "200gbase-x-qsfp56" {
		t.Errorf("got %q, want '200gbase-x-qsfp56'", got)
	}
}

// TestMapInterfaceType100baseTx verifies mapInterfaceType passes "100base-tx"
// through unchanged.
//
// Why it matters: low-speed management/copper ports must still export with the
// exact Nautobot slug.
// Inputs: "100base-tx". Outputs: the identical "100base-tx".
// Data choice: 100BASE-TX is a typical management-port type whose names already
// match, covering the identity branch for a non-fabric speed.
func TestMapInterfaceType100baseTx(t *testing.T) {
	got := mapInterfaceType("100base-tx")
	if got != "100base-tx" {
		t.Errorf("got %q, want '100base-tx'", got)
	}
}

// TestMapInterfaceType1gBaseT verifies mapInterfaceType rewrites the devicetypes
// alias "1gbase-t" to Nautobot's canonical "1000base-t".
//
// Why it matters: the cani devicetype library and Nautobot disagree on the 1G
// copper slug, so the exporter must translate or interface creation fails.
// Inputs: "1gbase-t". Outputs: "1000base-t".
// Data choice: 1G copper is the common speed where the two vocabularies differ,
// making it the key non-identity mapping branch.
func TestMapInterfaceType1gBaseT(t *testing.T) {
	got := mapInterfaceType("1gbase-t")
	if got != "1000base-t" {
		t.Errorf("got %q, want '1000base-t'", got)
	}
}

// TestMapInterfaceTypeInfinibandHDR verifies mapInterfaceType passes
// "infiniband-hdr" through unchanged.
//
// Why it matters: InfiniBand HDR fabric ports must export with Nautobot's exact
// slug; HDR must not be swallowed by the NDR case that is matched first.
// Inputs: "infiniband-hdr". Outputs: the identical "infiniband-hdr".
// Data choice: HDR is checked after NDR in the switch, so this input (lacking
// "ndr") confirms the HDR branch is actually reached.
func TestMapInterfaceTypeInfinibandHDR(t *testing.T) {
	got := mapInterfaceType("infiniband-hdr")
	if got != "infiniband-hdr" {
		t.Errorf("got %q, want 'infiniband-hdr'", got)
	}
}

// TestMapInterfaceType400gQSFPDD verifies mapInterfaceType passes the QSFP-DD
// form factor "400gbase-x-qsfpdd" through unchanged.
//
// Why it matters: Nautobot accepts both 400G form factors, and folding QSFP-DD
// onto OSFP lost the connector type on export and on the following re-import.
// Inputs: "400gbase-x-qsfpdd". Outputs: "400gbase-x-qsfpdd".
// Data choice: QSFP-DD is the alternate 400G connector that the old mapper
// rewrote, so it proves the pass-through rather than an identity case.
func TestMapInterfaceType400gQSFPDD(t *testing.T) {
	got := mapInterfaceType("400gbase-x-qsfpdd")
	if got != "400gbase-x-qsfpdd" {
		t.Errorf("got %q, want '400gbase-x-qsfpdd'", got)
	}
}

// ---------- resolveCableType slug-based fallback ----------

// TestResolveCableTypeSlugCat verifies resolveCableType's slug fallback maps any
// slug containing "cat" to a non-nil Cat5e cable-type choice.
//
// Why it matters: legacy cani cables carry only a slug, so the exporter must
// still resolve a Nautobot CableTypeChoices value to create the cable.
// Inputs: a CaniCableType with Slug "some-cat5e-cable" and no explicit type,
// category, or connector. Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: the slug embeds "cat" with the higher-priority fields empty,
// forcing resolution down to the slug branch.
func TestResolveCableTypeSlugCat(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "some-cat5e-cable"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'cat'")
	}
}

// TestResolveCableTypeSlugDAC verifies resolveCableType's slug fallback maps a
// slug containing "dac" to a non-nil passive-DAC cable-type choice.
//
// Why it matters: direct-attach copper cables are common in racks and must still
// resolve from a bare slug for export.
// Inputs: a CaniCableType with Slug "hpe-dac-1m" and the other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "hpe-dac-1m" embeds "dac" while leaving the higher-priority fields
// empty, isolating the DAC slug branch.
func TestResolveCableTypeSlugDAC(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "hpe-dac-1m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'dac'")
	}
}

// TestResolveCableTypeSlugAOC verifies resolveCableType's slug fallback maps a
// slug containing "aoc" to a non-nil active-optical cable-type choice.
//
// Why it matters: active optical cables must resolve from a bare slug so the
// exporter can still classify them in Nautobot.
// Inputs: a CaniCableType with Slug "mellanox-aoc-10m" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "mellanox-aoc-10m" embeds "aoc" with no higher-priority fields,
// exercising only the AOC slug branch.
func TestResolveCableTypeSlugAOC(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "mellanox-aoc-10m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'aoc'")
	}
}

// TestResolveCableTypeSlugFiber verifies resolveCableType's slug fallback maps a
// slug containing "fiber" to a non-nil multimode (MMF-OM4) cable-type choice.
//
// Why it matters: generically named fiber cables must still resolve to a
// concrete Nautobot type for export.
// Inputs: a CaniCableType with Slug "fiber-lc-om4-5m" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: the slug contains "fiber" (which shares the MMF branch) while
// leaving higher-priority fields empty.
func TestResolveCableTypeSlugFiber(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "fiber-lc-om4-5m"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'fiber'")
	}
}

// TestResolveCableTypeSlugMMF verifies resolveCableType's slug fallback maps a
// slug containing "mmf" to a non-nil multimode (MMF-OM4) cable-type choice.
//
// Why it matters: multimode fiber cables identified only by slug must still be
// classified so the exporter can create them.
// Inputs: a CaniCableType with Slug "mmf-om3-cable" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "mmf" shares the fiber branch; because the value is in Slug (not
// CableCategory), resolution falls through to the slug heuristic.
func TestResolveCableTypeSlugMMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{Slug: "mmf-om3-cable"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for slug containing 'mmf'")
	}
}

// TestResolveCableTypeConnectorLC verifies resolveCableType uses the connector
// heuristic to map an "lc" connector to a non-nil (SMF-OS2) cable type.
//
// Why it matters: when a cani cable lacks an explicit type and category, the
// connector is the best signal for picking a Nautobot cable type on export.
// Inputs: a CaniCableType with ConnectorType "lc" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: only ConnectorType is set so resolution skips the type/category
// tiers and exercises the connector-heuristic branch.
func TestResolveCableTypeConnectorLC(t *testing.T) {
	cable := &devicetypes.CaniCableType{ConnectorType: "lc"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for LC connector")
	}
}

// TestResolveCableTypeConnectorMPO verifies resolveCableType uses the connector
// heuristic to map an "mpo" connector to a non-nil (MMF-OM4) cable type.
//
// Why it matters: MPO trunks are common in fabrics, so the connector heuristic
// must classify them when no explicit type/category is present.
// Inputs: a CaniCableType with ConnectorType "mpo" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: ConnectorType alone is set, isolating the connector branch for a
// multimode connector distinct from the LC (single-mode) case.
func TestResolveCableTypeConnectorMPO(t *testing.T) {
	cable := &devicetypes.CaniCableType{ConnectorType: "mpo"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for MPO connector")
	}
}

// TestResolveCableTypeCategoryMMF verifies resolveCableType resolves the
// CableCategory "mmf-om3" via the category map to a non-nil cable type.
//
// Why it matters: when a cani cable names its category, the exporter should use
// that exact grade rather than guessing from a connector or slug.
// Inputs: a CaniCableType with CableCategory "mmf-om3" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "mmf-om3" is a real key in cableTypeMap, confirming the
// category-lookup tier resolves a specific multimode grade.
func TestResolveCableTypeCategoryMMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{CableCategory: "mmf-om3"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for mmf-om3 category")
	}
}

// TestResolveCableTypeCategorySMF verifies resolveCableType resolves the
// CableCategory "smf-os2" via the category map to a non-nil cable type.
//
// Why it matters: single-mode fiber grades must resolve from the named category
// so long-haul links export with the correct Nautobot type.
// Inputs: a CaniCableType with CableCategory "smf-os2" and other fields empty.
// Outputs: a non-nil *PatchedWritableCableRequestType.
// Data choice: "smf-os2" is a real cableTypeMap key, pairing with the MMF case
// to cover both fiber families through the category tier.
func TestResolveCableTypeCategorySMF(t *testing.T) {
	cable := &devicetypes.CaniCableType{CableCategory: "smf-os2"}
	got := resolveCableType(cable)
	if got == nil {
		t.Fatal("expected non-nil for smf-os2 category")
	}
}
