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

// ---------- getDeviceInterfaceSpecs ----------

// TestGetDeviceInterfaceSpecsWithInterfaces verifies that getDeviceInterfaceSpecs
// returns one spec per instantiated device interface, preserving name and type
// and deriving speed (e.g. infiniband-ndr -> 400000000 Kbps).
//
// Why it matters: when the device-type library already supplies interfaces,
// export must use them verbatim so Nautobot interface records match the real
// hardware rather than a fabricated fallback set.
// Inputs: a node with three interfaces (eth0, mgmt0, ib0). Outputs: three specs
// in order with derived speeds.
// Data choice: mixing 1000base-t and infiniband-ndr exercises both type
// pass-through and the speed-derivation lookup.
func TestGetDeviceInterfaceSpecsWithInterfaces(t *testing.T) {
	mgmt := true
	dev := &devicetypes.CaniDeviceType{
		Type: "node",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "eth0", Type: "1000base-t"},
			{Name: "mgmt0", Type: "1000base-t", MgmtOnly: &mgmt},
			{Name: "ib0", Type: "infiniband-ndr"},
		},
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(specs))
	}
	if specs[0].Name != "eth0" || specs[0].Type != "1000base-t" {
		t.Errorf("spec[0] = %+v", specs[0])
	}
	if specs[2].Name != "ib0" || specs[2].Type != "infiniband-ndr" || specs[2].Speed != 400000000 {
		t.Errorf("spec[2] = %+v", specs[2])
	}
}

// TestGetDeviceInterfaceSpecsPreservesDescription verifies that Description
// is carried from the device-type InterfaceSpec to the export interfaceSpec.
func TestGetDeviceInterfaceSpecsPreservesDescription(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: "switch",
		Interfaces: []devicetypes.InterfaceSpec{
			{Name: "lag256", Type: "lag", Description: "ISL link"},
			{Name: "1/1/1", Type: "100gbase-x-qsfp28"},
		},
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(specs))
	}
	if specs[0].Description != "ISL link" {
		t.Errorf("spec[0].Description = %q, want %q", specs[0].Description, "ISL link")
	}
	if specs[1].Description != "" {
		t.Errorf("spec[1].Description = %q, want empty", specs[1].Description)
	}
}

// TestGetDeviceInterfaceSpecsFallbackBlade verifies that getDeviceInterfaceSpecs
// synthesizes a fallback interface set (iLO + 4 Ethernet) for a blade/node with
// no instantiated interfaces, leading with the management iLO port.
//
// Why it matters: devices lacking a populated interface list still need
// interfaces in Nautobot; the type-based fallback guarantees a sensible default
// so exported servers are not left interface-less.
// Inputs: a Blade with Model "ProLiant DL360" and no Interfaces.
// Outputs: at least 5 specs, the first named "iLO".
// Data choice: a non-InfiniBand model keeps the count at the base 5 (no
// ib0/ib1 added).
func TestGetDeviceInterfaceSpecsFallbackBlade(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type:  devicetypes.Type(devicetypes.Blade),
		Model: "ProLiant DL360",
	}

	specs := getDeviceInterfaceSpecs(dev)
	// Should have iLO + 4 eth = 5 interfaces
	if len(specs) < 5 {
		t.Fatalf("expected at least 5 specs for blade, got %d", len(specs))
	}
	if specs[0].Name != "iLO" {
		t.Errorf("first spec should be iLO, got %q", specs[0].Name)
	}
}

// TestGetDeviceInterfaceSpecsFallbackHSNSwitch verifies that
// getDeviceInterfaceSpecs builds 65 specs for an HSNSwitch with no interfaces:
// one mgmt0 plus 64 InfiniBand-NDR osfp ports.
//
// Why it matters: high-speed-network switches export with a fixed dense port
// layout; the exact count and osfp type must match so fabric topology is
// represented correctly in Nautobot.
// Inputs: a device with Type HSNSwitch and no Interfaces.
// Outputs: 65 specs; specs[0]=="mgmt0", specs[1].Type=="infiniband-ndr".
// Data choice: asserting the count plus the first two specs pins down both the
// management port and the osfp block.
func TestGetDeviceInterfaceSpecsFallbackHSNSwitch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.HSNSwitch),
	}

	specs := getDeviceInterfaceSpecs(dev)
	// mgmt0 + 64 osfp = 65
	if len(specs) != 65 {
		t.Fatalf("expected 65 specs for HSN switch, got %d", len(specs))
	}
	if specs[0].Name != "mgmt0" {
		t.Errorf("first spec should be mgmt0, got %q", specs[0].Name)
	}
	if specs[1].Type != "infiniband-ndr" {
		t.Errorf("osfp type = %q, want infiniband-ndr", specs[1].Type)
	}
}

// TestGetDeviceInterfaceSpecsFallbackMgmtSwitch verifies that
// getDeviceInterfaceSpecs builds 53 specs for a MgmtSwitch with no interfaces:
// mgmt0 + 48 copper ports + 4 SFP uplinks.
//
// Why it matters: management switches have a standard Aruba-2930F-style port
// profile; the precise count ensures all switch ports are exported to Nautobot.
// Inputs: a device with Type MgmtSwitch and no Interfaces. Outputs: 53 specs.
// Data choice: the count (1+48+4) is the simplest assertion that all three
// port groups were emitted.
func TestGetDeviceInterfaceSpecsFallbackMgmtSwitch(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.MgmtSwitch),
	}

	specs := getDeviceInterfaceSpecs(dev)
	// mgmt0 + 48 ports + 4 sfp = 53
	if len(specs) != 53 {
		t.Fatalf("expected 53 specs for mgmt switch, got %d", len(specs))
	}
}

// TestGetDeviceInterfaceSpecsFallbackPDU verifies that getDeviceInterfaceSpecs
// emits exactly one mgmt0 interface for a CabinetPDU with no interfaces.
//
// Why it matters: PDUs are network-managed only; exporting a single management
// interface (and nothing else) keeps their Nautobot representation accurate.
// Inputs: a device with Type CabinetPDU and no Interfaces.
// Outputs: a single spec named "mgmt0".
// Data choice: the PDU is the minimal fallback case, asserting both length 1
// and the mgmt0 name.
func TestGetDeviceInterfaceSpecsFallbackPDU(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type: devicetypes.Type(devicetypes.CabinetPDU),
	}

	specs := getDeviceInterfaceSpecs(dev)
	if len(specs) != 1 || specs[0].Name != "mgmt0" {
		t.Fatalf("expected 1 mgmt spec for PDU, got %d", len(specs))
	}
}

// TestGetDeviceInterfaceSpecsInfinibandNode verifies that getDeviceInterfaceSpecs
// adds two InfiniBand ports (ib0, ib1) to the node fallback when the model name
// signals InfiniBand, yielding iLO + 4 eth + 2 ib = 7 specs.
//
// Why it matters: nodes with IB adapters must export their fabric interfaces;
// the model-name heuristic (containsInfiniband) is what triggers the extra
// ports so HSN connectivity is captured.
// Inputs: a Node with Model "Quantum InfiniBand NDR Switch" and no Interfaces.
// Outputs: 7 specs; the last two named "ib0" and "ib1".
// Data choice: a model containing "infiniband"/"ndr" reliably fires the IB
// branch on top of the base node set.
func TestGetDeviceInterfaceSpecsInfinibandNode(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{
		Type:  devicetypes.Type(devicetypes.Node),
		Model: "Quantum InfiniBand NDR Switch",
	}

	specs := getDeviceInterfaceSpecs(dev)
	// Should have iLO + 4 eth + 2 ib = 7
	if len(specs) != 7 {
		t.Fatalf("expected 7 specs for IB node, got %d", len(specs))
	}
	// Last two should be IB
	if specs[5].Name != "ib0" || specs[6].Name != "ib1" {
		t.Errorf("IB interfaces: %q, %q", specs[5].Name, specs[6].Name)
	}
}

// ---------- getSpeedForType additional ----------

// TestGetSpeedForType200g verifies getSpeedForType returns 200000000 Kbps for
// "200gbase-x-qsfp56".
//
// Why it matters: Nautobot interfaces carry an explicit speed, so the exporter
// must derive the right Kbps from the type or it records wrong link speeds.
// Inputs: "200gbase-x-qsfp56". Outputs: 200000000 (200 Gbps in Kbps).
// Data choice: 200G shares its speed case with InfiniBand HDR, so this asserts
// the Ethernet half of that combined branch.
func TestGetSpeedForType200g(t *testing.T) {
	if got := getSpeedForType("200gbase-x-qsfp56"); got != 200000000 {
		t.Errorf("got %d, want 200000000", got)
	}
}

// TestGetSpeedForTypeInfinibandHDR verifies getSpeedForType returns 200000000
// Kbps for "infiniband-hdr".
//
// Why it matters: InfiniBand HDR ports must report 200 Gbps in Nautobot just
// like 200G Ethernet, so fabric link speeds stay accurate.
// Inputs: "infiniband-hdr". Outputs: 200000000 (200 Gbps in Kbps).
// Data choice: HDR is the InfiniBand member of the shared 200G speed case,
// confirming both members resolve identically.
func TestGetSpeedForTypeInfinibandHDR(t *testing.T) {
	if got := getSpeedForType("infiniband-hdr"); got != 200000000 {
		t.Errorf("got %d, want 200000000", got)
	}
}

// TestGetSpeedForTypeInfinibandNDR verifies getSpeedForType returns 400000000
// Kbps for "infiniband-ndr".
//
// Why it matters: InfiniBand NDR ports must report 400 Gbps so the exporter
// records the correct fabric speed in Nautobot.
// Inputs: "infiniband-ndr". Outputs: 400000000 (400 Gbps in Kbps).
// Data choice: NDR shares the 400G speed case with 400G Ethernet, so this
// asserts the InfiniBand member of that branch.
func TestGetSpeedForTypeInfinibandNDR(t *testing.T) {
	if got := getSpeedForType("infiniband-ndr"); got != 400000000 {
		t.Errorf("got %d, want 400000000", got)
	}
}

// TestGetSpeedForType400gOSFP verifies getSpeedForType returns 400000000 Kbps for
// "400gbase-x-osfp".
//
// Why it matters: 400G OSFP links are the fastest exported ports and must carry
// the correct 400 Gbps speed in Nautobot.
// Inputs: "400gbase-x-osfp". Outputs: 400000000 (400 Gbps in Kbps).
// Data choice: OSFP is the Ethernet member of the shared 400G case (with QSFP-DD
// and InfiniBand NDR), asserting that branch from the Ethernet side.
func TestGetSpeedForType400gOSFP(t *testing.T) {
	if got := getSpeedForType("400gbase-x-osfp"); got != 400000000 {
		t.Errorf("got %d, want 400000000", got)
	}
}

// TestGetSpeedForType25g verifies getSpeedForType returns 25000000 Kbps for
// "25gbase-x-sfp28".
//
// Why it matters: 25G server links must report the correct speed so Nautobot
// reflects real edge connectivity.
// Inputs: "25gbase-x-sfp28". Outputs: 25000000 (25 Gbps in Kbps).
// Data choice: 25G has its own standalone speed case, chosen to cover a
// mid-range edge speed distinct from the shared high-speed branches.
func TestGetSpeedForType25g(t *testing.T) {
	if got := getSpeedForType("25gbase-x-sfp28"); got != 25000000 {
		t.Errorf("got %d, want 25000000", got)
	}
}

// TestGetSpeedForTypeUnknown verifies getSpeedForType falls back to 1000000 Kbps
// (1 Gbps) for an unrecognized interface type.
//
// Why it matters: unknown types must still yield a sane default speed so the
// exporter can create the interface rather than failing or sending zero.
// Inputs: "some-unknown-type". Outputs: 1000000 (default 1 Gbps in Kbps).
// Data choice: a deliberately bogus type string forces the switch's default
// branch.
func TestGetSpeedForTypeUnknown(t *testing.T) {
	if got := getSpeedForType("some-unknown-type"); got != 1000000 {
		t.Errorf("got %d, want 1000000 (default)", got)
	}
}

// ---------- containsInfiniband additional ----------

// TestContainsInfinibandModel verifies containsInfiniband returns true when a
// device model name contains "infiniband".
//
// Why it matters: the exporter uses this signal to decide whether to synthesize
// InfiniBand interfaces, so IB hardware must be recognized by model name.
// Inputs: a CaniDeviceType with Model "Mellanox InfiniBand Switch". Outputs:
// true.
// Data choice: a realistic Mellanox IB switch name exercises the literal
// "infiniband" substring match (case-insensitive).
func TestContainsInfinibandModel(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "Mellanox InfiniBand Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'infiniband'")
	}
}

// TestContainsInfinibandNDR verifies containsInfiniband returns true when a model
// name contains the generation token "ndr".
//
// Why it matters: IB hardware is often named by generation rather than the word
// "infiniband", so the exporter must also key off "ndr".
// Inputs: a CaniDeviceType with Model "NDR-400G Switch". Outputs: true.
// Data choice: an NDR-generation switch name lacking "infiniband" proves the
// "ndr" substring branch independently.
func TestContainsInfinibandNDR(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "NDR-400G Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'ndr'")
	}
}

// TestContainsInfinibandHDR verifies containsInfiniband returns true when a model
// name contains the generation token "hdr".
//
// Why it matters: HDR-generation IB switches must be detected so the exporter
// adds their fabric interfaces.
// Inputs: a CaniDeviceType with Model "QM8700 HDR Switch". Outputs: true.
// Data choice: the QM8700 is a real HDR switch; its name carries "hdr" but not
// "infiniband", isolating the "hdr" branch.
func TestContainsInfinibandHDR(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "QM8700 HDR Switch"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'hdr'")
	}
}

// TestContainsInfinibandMCX verifies containsInfiniband returns true when a model
// name contains the adapter token "mcx".
//
// Why it matters: Mellanox ConnectX adapters use "MCX" part numbers, so the
// exporter must recognize them as InfiniBand-capable.
// Inputs: a CaniDeviceType with Model "MCX653105A-HDAT". Outputs: true.
// Data choice: a genuine ConnectX-6 part number exercises the "mcx" branch via a
// raw SKU rather than a descriptive name.
func TestContainsInfinibandMCX(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "MCX653105A-HDAT"}
	if !containsInfiniband(dev) {
		t.Error("expected true for model containing 'mcx'")
	}
}

// TestContainsInfinibandFalse verifies containsInfiniband returns false for a
// standard Ethernet server model.
//
// Why it matters: false positives would make the exporter add phantom InfiniBand
// interfaces to non-IB hardware.
// Inputs: a CaniDeviceType with Model "HPE ProLiant DL380". Outputs: false.
// Data choice: a ubiquitous non-IB server whose name contains none of the
// infiniband/ndr/hdr/mcx tokens, covering the negative path.
func TestContainsInfinibandFalse(t *testing.T) {
	dev := &devicetypes.CaniDeviceType{Model: "HPE ProLiant DL380"}
	if containsInfiniband(dev) {
		t.Error("expected false for standard server model")
	}
}

// ---------- getSpeedForType 40g and 100g ----------

// TestGetSpeedForType40g verifies getSpeedForType maps the 40G QSFP+ interface
// type string to 40000000 kbps.
//
// Why it matters: Nautobot stores interface speed in kbps, so cani must convert
// each cani interface type to the correct numeric speed when exporting NICs.
// Inputs: the type slug "40gbase-x-qsfpp". Outputs: the integer speed in kbps.
// Data choice: the canonical 40G QSFP+ slug pins the high-speed mapping that
// fabric uplinks rely on.
func TestGetSpeedForType40g(t *testing.T) {
	if got := getSpeedForType("40gbase-x-qsfpp"); got != 40000000 {
		t.Errorf("got %d, want 40000000", got)
	}
}

// TestGetSpeedForType100g verifies getSpeedForType maps the 100G QSFP28
// interface type string to 100000000 kbps.
//
// Why it matters: exported interface speeds must match Nautobot's kbps field so
// 100G fabric links report their real bandwidth.
// Inputs: the type slug "100gbase-x-qsfp28". Outputs: the integer kbps speed.
// Data choice: the 100G QSFP28 slug is the most common high-speed switch uplink,
// making it the key mapping to guard.
func TestGetSpeedForType100g(t *testing.T) {
	if got := getSpeedForType("100gbase-x-qsfp28"); got != 100000000 {
		t.Errorf("got %d, want 100000000", got)
	}
}

// TestGetSpeedForType10g verifies getSpeedForType maps the 10G SFP+ interface
// type string to 10000000 kbps.
//
// Why it matters: server and ToR links are frequently 10G, so this conversion
// keeps exported Nautobot interface speeds accurate.
// Inputs: the type slug "10gbase-x-sfpp". Outputs: the integer kbps speed.
// Data choice: the 10G SFP+ slug represents the common mid-speed case between
// the 1G default and the 40G/100G fabric tiers.
func TestGetSpeedForType10g(t *testing.T) {
	if got := getSpeedForType("10gbase-x-sfpp"); got != 10000000 {
		t.Errorf("got %d, want 10000000", got)
	}
}

// TestGetSpeedForType100baseTx verifies getSpeedForType maps the 100BASE-TX
// copper interface type string to 100000 kbps.
//
// Why it matters: management/BMC ports are often 100M copper, and exporting the
// wrong speed would misrepresent them in Nautobot.
// Inputs: the type slug "100base-tx". Outputs: the integer kbps speed.
// Data choice: 100BASE-TX is the slowest mapped tier, guarding the low end away
// from the 1Gbps default fallback.
func TestGetSpeedForType100baseTx(t *testing.T) {
	if got := getSpeedForType("100base-tx"); got != 100000 {
		t.Errorf("got %d, want 100000", got)
	}
}
