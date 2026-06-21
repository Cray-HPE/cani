package transform

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
	"github.com/Cray-HPE/cani/pkg/visual"
	"github.com/google/uuid"
)

// --- Inference and parsing tests ---

// TestInferHardwareType verifies a free-text hardware description is classified
// into a rack, switch, node, or cable category (or empty when unknown).
//
// Why it matters: classification routes each CSV row to the correct builder, so
// the importer must reliably map product descriptions to a hardware category.
// Inputs: a table of HPE product descriptions. Outputs: the inferred category
// string.
// Data choice: at least one description per category plus an unrecognized device
// and a memory kit exercise every branch including the empty-string fallthrough.
func TestInferHardwareType(t *testing.T) {
	tests := []struct {
		name, description, want string
	}{
		{"48U rack", "HPE 48U 800mmx1200mm G2 Enterprise Shock Rack", "rack"},
		{"cabinet", "HPE Cabinet G2", "rack"},
		{"aruba switch", "HPE Aruba Networking 8360-48Y6C v2", "switch"},
		{"generic switch", "HPE 48-Port Ethernet Switch", "switch"},
		{"proliant server", "HPE ProLiant DL380 Gen11", "node"},
		{"blade server", "HPE BladeSystem c7000 Blade", "node"},
		{"cat6 cable", "HPE Cat6 RJ45 M/M 2m", "cable"},
		{"cat5e cable", "HPE CAT5e RJ45 2.3m Cable", "cable"},
		{"DAC cable", "HPE 400G QSFP-DD DAC 3m", "cable"},
		{"direct attach cable", "HPE 100Gb Direct Attach Copper Cable", "cable"},
		{"AOC cable", "HPE Aruba 100G QSFP28 15m AOC", "cable"},
		{"active optical cable", "HPE Active Optical Cable 30m", "cable"},
		{"OM4 fiber", "HPE Premier Flex LC LC OM4 15m", "cable"},
		{"power jumper", "HPE C19 C20 250V 16A 2m Jumper", "cable"},
		{"RJ45 cable", "HPE RJ45 to RJ45 Cat5e Black", "cable"},
		{"QSFP cable", "HPE QSFP28 to QSFP28 Cable", "cable"},
		{"generic cable", "HPE Data Cable 3m", "cable"},
		{"unknown device", "XD670", ""},
		{"memory module", "HPE 64GB DDR5 Memory Kit", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferHardwareType(tt.description); got != tt.want {
				t.Errorf("inferHardwareType(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

// TestInferCableTypeSlug verifies a cable description is mapped to its canonical
// cable-type slug.
//
// Why it matters: cables need a precise library type for the BOM, so the
// importer must distinguish copper, fiber, DAC, AOC, and power cables from free
// text.
// Inputs: cable descriptions. Outputs: the resolved cable-type slug constant.
// Data choice: one description per slug family (cat5/5e/6/6a, DAC, AOC, OM3/4,
// SMF, power) plus a QSFP-without-type that must fall through to "other".
func TestInferCableTypeSlug(t *testing.T) {
	tests := []struct {
		name, description, want string
	}{
		{"cat5 cable", "HPE CAT5 RJ45 1.2m Cable", cableTypeCat5e},
		{"cat5e cable", "HPE CAT5e RJ45 2.3m Cable", cableTypeCat5e},
		{"cat6 cable", "HPE Cat6 RJ45 M/M 2m", cableTypeCat6},
		{"cat6a cable", "HPE Cat6a Shielded Cable 3m", cableTypeCat6a},
		{"DAC cable", "HPE 400G QSFP-DD DAC 3m", cableTypeDacPassive},
		{"direct attach cable", "HPE 100Gb Direct Attach Copper Cable", cableTypeDacPassive},
		{"AOC cable", "HPE Aruba 100G QSFP28 15m AOC", cableTypeAoc},
		{"active optical cable", "HPE Active Optical Cable 30m", cableTypeAoc},
		{"OM3 fiber", "HPE LC LC OM3 2F 30m", cableTypeMmfOm4},
		{"OM4 fiber", "HPE Premier Flex LC LC OM4 15m", cableTypeMmfOm4},
		{"MMF fiber", "HPE InfiniBand NDR MPO MPO MM 10m", cableTypeMmfOm4},
		{"SMF fiber", "HPE InfiniBand NDR MPO MPO SM 10m", cableTypeSmf},
		{"single mode fiber", "HPE Single Mode Fiber Cable 20m", cableTypeSmf},
		{"power jumper", "HPE C19 C20 250V 16A 2m Jumper", cableTypePower},
		{"power cord", "HPE Power Cord 2m", cableTypePower},
		{"generic cable", "HPE Data Cable 3m", cableTypeOther},
		{"QSFP without type", "HPE QSFP28 Cable", cableTypeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferCableTypeSlug(tt.description); got != tt.want {
				t.Errorf("inferCableTypeSlug(%q) = %q, want %q", tt.description, got, tt.want)
			}
		})
	}
}

// TestResolveCableTypeSlug verifies the resolver still yields a non-empty slug
// when the part number is unknown by falling back to description inference.
//
// Why it matters: every cable must carry a type even when its part number is not
// in the library, so the description fallback must never produce an empty slug.
// Inputs: an unknown or blank part number paired with a description. Outputs: a
// non-empty slug (only non-emptiness is asserted, not the specific value).
// Data choice: a fake part number, a blank part number, and a "Mystery Widget"
// description drive the description-match and unknown-everything fallthroughs.
func TestResolveCableTypeSlug(t *testing.T) {
	tests := []struct {
		name, partNumber, description string
	}{
		{"fallback to description for cat6", "FAKE-PN-123", "Cat6 RJ45 patch cable 2m"},
		{"fallback to description for DAC", "", "400G QSFP-DD DAC 3m"},
		{"unknown description returns other", "", "Mystery Widget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveCableTypeSlug(tt.partNumber, tt.description); got == "" {
				t.Error("expected non-empty cable type slug")
			}
		})
	}
}

// TestInferCableType verifies an interface's electrical type is mapped to the
// cable slug used to connect it.
//
// Why it matters: auto-generated cables must match the interface media, so the
// importer picks copper (cat6/cat6a) for BASE-T ports and DAC for SFP/QSFP ports.
// Inputs: interface element types. Outputs: the cable-type slug.
// Data choice: 1000BASE-T→cat6, 10GBASE-T→cat6a, the SFP/QSFP family→DAC, and an
// unknown type→other cover every branch including the default.
func TestInferCableType(t *testing.T) {
	tests := []struct {
		name      string
		ifaceType devicetypes.InterfacesElemType
		want      string
	}{
		{"1000base-t", devicetypes.InterfacesElemTypeA1000BaseT, cableTypeCat6},
		{"10gbase-t", devicetypes.InterfacesElemTypeA10GbaseT, cableTypeCat6a},
		{"10gbase-x-sfpp", devicetypes.InterfacesElemTypeA10GbaseXSfpp, cableTypeDacPassive},
		{"25gbase-x-sfp28", devicetypes.InterfacesElemTypeA25GbaseXSfp28, cableTypeDacPassive},
		{"40gbase-x-qsfpp", devicetypes.InterfacesElemTypeA40GbaseXQsfpp, cableTypeDacPassive},
		{"100gbase-x-qsfp28", devicetypes.InterfacesElemTypeA100GbaseXQsfp28, cableTypeDacPassive},
		{"400gbase-x-qsfpdd", devicetypes.InterfacesElemTypeA400GbaseXQsfpdd, cableTypeDacPassive},
		{"unknown type", devicetypes.InterfacesElemType("unknown"), cableTypeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferCableType(tt.ifaceType); got != tt.want {
				t.Errorf("inferCableType(%q) = %q, want %q", tt.ifaceType, got, tt.want)
			}
		})
	}
}

// TestParseLengthFromDescription verifies a numeric length and unit embedded in
// a product description are extracted.
//
// Why it matters: cable length is often only present in the description text, so
// the importer must pull it out for documentation and BOM output.
// Inputs: descriptions with embedded lengths. Outputs: numeric length and unit.
// Data choice: meters, feet, decimal meters, centimeters, and a no-length
// description cover each unit plus the (0, "") miss.
func TestParseLengthFromDescription(t *testing.T) {
	tests := []struct {
		name, description string
		wantLength        float64
		wantUnit          string
	}{
		{"meters", "HPE Cat6 RJ45 M/M 2m", 2, "m"},
		{"feet", "HPE Cable 10ft", 10, "ft"},
		{"decimal meters", "HPE 1.5m DAC Cable", 1.5, "m"},
		{"centimeters", "HPE 50cm Patch Cable", 50, "cm"},
		{"no length", "HPE Generic Cable", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, unit := parseLengthFromDescription(tt.description)
			if length != tt.wantLength {
				t.Errorf("length = %v, want %v", length, tt.wantLength)
			}
			if unit != tt.wantUnit {
				t.Errorf("unit = %q, want %q", unit, tt.wantUnit)
			}
		})
	}
}

// TestParseCableLength verifies a standalone length string is split into a
// number and unit, defaulting the unit to meters.
//
// Why it matters: explicit connection rows carry a free-form length field, so
// the importer must split "10ft" reliably and default a bare number to meters.
// Inputs: length strings. Outputs: numeric length and unit.
// Data choice: a bare "5"→m default, empty and "abc" miss cases, and a
// whitespace-padded " 3m " together prove defaulting, the misses, and trimming.
func TestParseCableLength(t *testing.T) {
	tests := []struct {
		name, input string
		wantLength  float64
		wantUnit    string
	}{
		{"meters", "3m", 3, "m"},
		{"feet", "10ft", 10, "ft"},
		{"decimal", "1.5m", 1.5, "m"},
		{"no unit defaults to m", "5", 5, "m"},
		{"empty string", "", 0, ""},
		{"non-numeric", "abc", 0, ""},
		{"with spaces", " 3m ", 3, "m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length, unit := parseCableLength(tt.input)
			if length != tt.wantLength {
				t.Errorf("length = %v, want %v", length, tt.wantLength)
			}
			if unit != tt.wantUnit {
				t.Errorf("unit = %q, want %q", unit, tt.wantUnit)
			}
		})
	}
}

// TestGenerateCableLabel verifies a cable label appends a zero-padded ordinal
// only when more than one cable shares a description.
//
// Why it matters: bulk cables created from one row need unique, stable labels,
// so a quantity greater than one must produce -001/-002 suffixes while a single
// cable stays bare.
// Inputs: a description, index, and total count. Outputs: the label string.
// Data choice: total=1 (no suffix), first/second of three (-001/-002), and the
// tenth of twenty (-010) prove the bare case and three-digit zero padding.
func TestGenerateCableLabel(t *testing.T) {
	tests := []struct {
		name, description string
		index, total      int
		want              string
	}{
		{"single cable", "HPE Cat6 Cable", 0, 1, "HPE Cat6 Cable"},
		{"first of multiple", "HPE Cat6 Cable", 0, 3, "HPE Cat6 Cable-001"},
		{"second of multiple", "HPE Cat6 Cable", 1, 3, "HPE Cat6 Cable-002"},
		{"tenth cable", "HPE DAC Cable", 9, 20, "HPE DAC Cable-010"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateCableLabel(tt.description, tt.index, tt.total); got != tt.want {
				t.Errorf("generateCableLabel(%q, %d, %d) = %q, want %q",
					tt.description, tt.index, tt.total, got, tt.want)
			}
		})
	}
}

// --- Lookup helper tests ---

// TestFindDeviceByName verifies a device is looked up by name, returning the
// device on a hit and nil on a miss.
//
// Why it matters: cable endpoints and parenting reference devices by name, so
// the importer must resolve a name to the right device or report its absence.
// Inputs: an inventory with one named device, queried for that name and a
// missing name. Outputs: the matching device pointer or nil.
// Data choice: a single registered "server-01" isolates both the hit (ID match)
// and the miss path with no ambiguity.
func TestFindDeviceByName(t *testing.T) {
	deviceID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {ID: deviceID, Name: "server-01"},
		},
	}

	t.Run("found", func(t *testing.T) {
		if got := findDeviceByName(inv, "server-01"); got == nil || got.ID != deviceID {
			t.Errorf("findDeviceByName() = %v, want device with ID %s", got, deviceID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if got := findDeviceByName(inv, "nonexistent"); got != nil {
			t.Errorf("findDeviceByName() = %v, want nil", got)
		}
	})
}

// TestFindAvailableInterface verifies the finder returns a device's first
// interface with no connected cable, or nil when none is free.
//
// Why it matters: auto-cabling must claim only unused ports, so the finder skips
// occupied interfaces and reports when a device is fully patched.
// Inputs: devices with a mix of connected/free interfaces, all connected, and
// none at all. Outputs: the free interface pointer or nil.
// Data choice: eth0 connected + eth1 free proves occupied ports are skipped; the
// all-connected and no-interface cases prove both nil paths.
func TestFindAvailableInterface(t *testing.T) {
	cableID := uuid.New()
	inv := &devicetypes.Inventory{}

	t.Run("returns unconnected interface", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Interfaces: []devicetypes.InterfaceSpec{
				{Name: "eth0", ConnectedCable: &cableID},
				{Name: "eth1"},
			},
		}
		if got := findAvailableInterface(inv, device); got == nil || got.Name != "eth1" {
			t.Errorf("findAvailableInterface() = %v, want eth1", got)
		}
	})

	t.Run("all connected returns nil", func(t *testing.T) {
		device := &devicetypes.CaniDeviceType{
			Interfaces: []devicetypes.InterfaceSpec{{Name: "eth0", ConnectedCable: &cableID}},
		}
		if got := findAvailableInterface(inv, device); got != nil {
			t.Errorf("findAvailableInterface() = %v, want nil", got)
		}
	})

	t.Run("no interfaces returns nil", func(t *testing.T) {
		if got := findAvailableInterface(inv, &devicetypes.CaniDeviceType{}); got != nil {
			t.Errorf("findAvailableInterface() = %v, want nil", got)
		}
	})
}

// --- Connection logic tests ---

// TestLinkInterfacesToCable verifies a cable's A/B terminations are linked to
// their interfaces, erroring when a termination interface is missing.
//
// Why it matters: a cable must attach to real interfaces on both ends, so
// linking fails loudly rather than leaving a dangling termination.
// Inputs: an inventory with both interfaces registered (success) and one with
// neither (error). Outputs: nil or an error.
// Data choice: registering A and B in both the device and Interfaces maps proves
// the happy path; empty maps with random termination IDs force the error.
func TestLinkInterfacesToCable(t *testing.T) {
	ifaceAID, ifaceBID := uuid.New(), uuid.New()
	deviceAID, deviceBID := uuid.New(), uuid.New()
	cableID := uuid.New()

	t.Run("links both interfaces", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceAID: {ID: deviceAID, Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceAID, Name: "eth0"}}},
				deviceBID: {ID: deviceBID, Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceBID, Name: "eth0"}}},
			},
			Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
				ifaceAID: {ID: ifaceAID, DeviceID: deviceAID},
				ifaceBID: {ID: ifaceBID, DeviceID: deviceBID},
			},
		}
		cable := &devicetypes.CaniCableType{ID: cableID, TerminationA: ifaceAID, TerminationB: ifaceBID}
		if err := linkInterfacesToCable(inv, cable); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing interface returns error", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices:    map[uuid.UUID]*devicetypes.CaniDeviceType{},
			Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{},
		}
		cable := &devicetypes.CaniCableType{ID: cableID, TerminationA: uuid.New(), TerminationB: uuid.New()}
		if err := linkInterfacesToCable(inv, cable); err == nil {
			t.Error("expected error but got none")
		}
	})
}

// TestGroupDevicesByConfigGroup verifies devices are grouped by their
// example-provider ConfigGroup, skipping devices without metadata or that are nil.
//
// Why it matters: auto-cabling pairs devices by config group, so the grouping
// must read provider metadata and ignore devices that cannot be grouped.
// Inputs: a device with ConfigGroup "0200", one without metadata, and a nil
// device. Outputs: a map of config group to device slice.
// Data choice: the three cases isolate the populated group, the missing-metadata
// skip, and the nil-device skip.
func TestGroupDevicesByConfigGroup(t *testing.T) {
	deviceID := uuid.New()

	t.Run("groups by config group metadata", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				deviceID: {
					ID: deviceID,
					ObjectMeta: devicetypes.ObjectMeta{
						ProviderMetadata: map[string]any{
							"example": map[string]any{"ConfigGroup": "0200"},
						},
					},
				},
			},
		}
		if result := groupDevicesByConfigGroup(inv); len(result["0200"]) != 1 {
			t.Errorf("expected 1 device in group 0200, got %d", len(result["0200"]))
		}
	})

	t.Run("skips devices without metadata", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: {ID: deviceID}},
		}
		if result := groupDevicesByConfigGroup(inv); len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("skips nil devices", func(t *testing.T) {
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{deviceID: nil},
		}
		if result := groupDevicesByConfigGroup(inv); len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})
}

// TestFindRelatedDeviceGroups verifies the device groups a cable group connects
// to, excluding rack groups, the cable's own group, and malformed short keys.
//
// Why it matters: hub-spoke cabling links a cable group (e.g. 0900) to device
// groups (0200/0300), so racks and the cable's own group must be filtered out.
// Inputs: a cable group and a map of device groups. Outputs: the count of
// related groups.
// Data choice: 0200/0300 included, 0100 (rack) excluded, the own 0900 excluded,
// and a one-character key cover each filter branch.
func TestFindRelatedDeviceGroups(t *testing.T) {
	deviceID := uuid.New()
	tests := []struct {
		name           string
		cableGroup     string
		devicesByGroup map[string][]*devicetypes.CaniDeviceType
		wantLen        int
	}{
		{"finds non-rack non-cable groups", "0900", map[string][]*devicetypes.CaniDeviceType{
			"0200": {{ID: deviceID}}, "0300": {{ID: deviceID}},
		}, 2},
		{"excludes rack group 01XX", "0900", map[string][]*devicetypes.CaniDeviceType{
			"0100": {{ID: deviceID}}, "0200": {{ID: deviceID}},
		}, 1},
		{"excludes own cable group", "0900", map[string][]*devicetypes.CaniDeviceType{
			"0900": {{ID: deviceID}},
		}, 0},
		{"short cable group returns empty", "1", map[string][]*devicetypes.CaniDeviceType{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findRelatedDeviceGroups(tt.cableGroup, tt.devicesByGroup); len(got) != tt.wantLen {
				t.Errorf("findRelatedDeviceGroups() returned %d groups, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestAutoConnectCables verifies cable auto-connection runs across related
// device groups and tolerates groups with no relations without panicking.
//
// Why it matters: auto-cabling wires switch hubs to node spokes by config group,
// so the importer must attempt connections and stay safe when no devices match.
// Inputs: a switch (0200) + node (0300) with a 0900 cable, and an empty
// inventory with a 0900 cable. Outputs: mutated cable terminations and device
// endpoint fields; the empty case asserts the cable remains unconnected.
// Data choice: the switch/node config groups model the real 0900→0200/0300
// cabling relation, while the empty inventory isolates the no-related-groups path.
func TestAutoConnectCables(t *testing.T) {
	t.Run("connects cables to related device groups", func(t *testing.T) {
		switchID, nodeID := uuid.New(), uuid.New()
		switchIfaceID, nodeIfaceID := uuid.New(), uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				switchID: {
					ID: switchID, Name: "switch-01", Type: "switch",
					Interfaces: []devicetypes.InterfaceSpec{{ID: switchIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
					ObjectMeta: devicetypes.ObjectMeta{
						ProviderMetadata: map[string]any{"example": map[string]any{"ConfigGroup": "0200"}},
					},
				},
				nodeID: {
					ID: nodeID, Name: "server-01", Type: "node",
					Interfaces: []devicetypes.InterfaceSpec{{ID: nodeIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
					ObjectMeta: devicetypes.ObjectMeta{
						ProviderMetadata: map[string]any{"example": map[string]any{"ConfigGroup": "0300"}},
					},
				},
			},
		}
		cable := devicetypes.NewCable("cat6", "test-cable")
		autoConnectCables(inv, map[string][]*devicetypes.CaniCableType{"0900": {cable}})
		if cable.TerminationA != nodeIfaceID {
			t.Errorf("TerminationA = %s, want spoke interface %s", cable.TerminationA, nodeIfaceID)
		}
		if cable.TerminationB != switchIfaceID {
			t.Errorf("TerminationB = %s, want hub interface %s", cable.TerminationB, switchIfaceID)
		}
		if cable.TerminationADevice != nodeID {
			t.Errorf("TerminationADevice = %s, want spoke device %s", cable.TerminationADevice, nodeID)
		}
		if cable.TerminationBDevice != switchID {
			t.Errorf("TerminationBDevice = %s, want hub device %s", cable.TerminationBDevice, switchID)
		}
		if cable.TerminationAPort != "eth0" || cable.TerminationBPort != "eth0" {
			t.Errorf("ports = %q/%q, want eth0/eth0", cable.TerminationAPort, cable.TerminationBPort)
		}
	})

	t.Run("no related groups does not panic", func(t *testing.T) {
		inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}
		cable := devicetypes.NewCable("cat6", "test-cable")
		autoConnectCables(inv, map[string][]*devicetypes.CaniCableType{"0900": {cable}})
		if cable.TerminationA != uuid.Nil || cable.TerminationB != uuid.Nil {
			t.Errorf("terminations = %s/%s, want both Nil", cable.TerminationA, cable.TerminationB)
		}
	})
}

// TestConnectCablesHubSpoke verifies cables are connected between hub and spoke
// devices, left unconnected when there are no hubs or no remaining spokes.
//
// Why it matters: hub-spoke is the core auto-cabling primitive, so it must
// terminate cables on both ends when possible and stop cleanly when endpoints
// run out.
// Inputs: hub/spoke sets with controlled interface counts and cable pools.
// Outputs: mutated cable terminations.
// Data choice: one hub + one spoke proves a full connection; no hubs proves the
// unconnected path; two cables with one spoke proves the pool stops at the spoke
// count.
func TestConnectCablesHubSpoke(t *testing.T) {
	t.Run("connects hub to spoke", func(t *testing.T) {
		switchID, nodeID := uuid.New(), uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				switchID: {ID: switchID, Name: "switch-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0"}}},
				nodeID:   {ID: nodeID, Name: "server-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "eth0"}}},
			},
		}
		cable := devicetypes.NewCable("cat6", "test-cable")
		connectCablesHubSpoke(inv, []*devicetypes.CaniCableType{cable},
			[]*devicetypes.CaniDeviceType{inv.Devices[switchID]},
			[]*devicetypes.CaniDeviceType{inv.Devices[nodeID]})
		if cable.TerminationA == uuid.Nil || cable.TerminationB == uuid.Nil {
			t.Error("expected cable to be connected to both hub and spoke")
		}
	})

	t.Run("no hubs leaves cable unconnected", func(t *testing.T) {
		inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}
		cable := devicetypes.NewCable("cat6", "test")
		connectCablesHubSpoke(inv, []*devicetypes.CaniCableType{cable}, nil, nil)
		if cable.TerminationA != uuid.Nil {
			t.Error("expected cable to remain unconnected")
		}
	})

	t.Run("more cables than spokes", func(t *testing.T) {
		hubID, spokeID := uuid.New(), uuid.New()
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				hubID:   {ID: hubID, Name: "sw-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "e0"}, {ID: uuid.New(), Name: "e1"}}},
				spokeID: {ID: spokeID, Name: "srv-01", Interfaces: []devicetypes.InterfaceSpec{{ID: uuid.New(), Name: "e0"}}},
			},
		}
		cables := []*devicetypes.CaniCableType{
			devicetypes.NewCable("cat6", "c1"),
			devicetypes.NewCable("cat6", "c2"),
		}
		connectCablesHubSpoke(inv, cables,
			[]*devicetypes.CaniDeviceType{inv.Devices[hubID]},
			[]*devicetypes.CaniDeviceType{inv.Devices[spokeID]})
		if cables[0].TerminationA == uuid.Nil {
			t.Error("first cable should be connected")
		}
		if cables[1].TerminationA != uuid.Nil {
			t.Error("second cable should remain unconnected (no more spokes)")
		}
	})
}

// --- Cable creation tests ---

// TestCreateCableFromExplicitRecord verifies a cable is built from an explicit
// source/dest record, erroring when any endpoint device or port is missing.
//
// Why it matters: explicit connection rows must resolve all four endpoints to
// real interfaces, so a typo in any field aborts cable creation rather than
// producing a half-wired cable.
// Inputs: a fully wired switch/server pair plus records that each break one of
// the four endpoint fields. Outputs: the cable (terminations set to the resolved
// interface IDs) or an error.
// Data choice: the success case asserts both terminations resolve to the right
// interface IDs; the four error rows each null out exactly one endpoint to
// isolate every lookup failure.
func TestCreateCableFromExplicitRecord(t *testing.T) {
	srcDeviceID, dstDeviceID := uuid.New(), uuid.New()
	srcIfaceID, dstIfaceID := uuid.New(), uuid.New()

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDeviceID: {
				ID: srcDeviceID, Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
			},
			dstDeviceID: {
				ID: dstDeviceID, Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		rec := import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0"}
		cable, err := createCableFromExplicitRecord(inv, rec)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cable == nil {
			t.Fatal("expected cable but got nil")
		}
		if cable.TerminationA != srcIfaceID {
			t.Errorf("TerminationA = %v, want %v", cable.TerminationA, srcIfaceID)
		}
		if cable.TerminationB != dstIfaceID {
			t.Errorf("TerminationB = %v, want %v", cable.TerminationB, dstIfaceID)
		}
	})

	errorCases := []struct {
		name string
		rec  import_.CsvRecord
	}{
		{"source device not found", import_.CsvRecord{SourceDevice: "nonexistent", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0"}},
		{"source port not found", import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "nonexistent", DestDevice: "server-01", DestPort: "eth0"}},
		{"dest device not found", import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "nonexistent", DestPort: "eth0"}},
		{"dest port not found", import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "nonexistent"}},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := createCableFromExplicitRecord(inv, tt.rec); err == nil {
				t.Error("expected error but got none")
			}
		})
	}
}

// TestTransformCables verifies cable records are transformed into cables,
// yielding none for empty input and one cable per unit for product rows.
//
// Why it matters: the cable pass turns BOM rows into inventory cables, so an
// empty batch is a no-op and a quantity-3 product row must expand to three
// cables.
// Inputs: nil records and a quantity-3 C7536A product row. Outputs: the created
// cable slice.
// Data choice: C7536A with quantity 3 is a real cable part number that proves
// quantity expansion; nil input proves the empty no-op.
func TestTransformCables(t *testing.T) {
	t.Run("empty records", func(t *testing.T) {
		inv := &devicetypes.Inventory{Cables: make(map[uuid.UUID]*devicetypes.CaniCableType)}
		tally := &visual.StepTally{}
		recordNum := 0
		cables, err := transformCables(inv, nil, false, visual.ETLOptions{}, tally, &recordNum, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cables) != 0 {
			t.Errorf("expected 0 cables, got %d", len(cables))
		}
	})

	t.Run("product records", func(t *testing.T) {
		inv := &devicetypes.Inventory{Cables: make(map[uuid.UUID]*devicetypes.CaniCableType)}
		tally := &visual.StepTally{}
		recordNum := 0
		records := []import_.CsvRecord{{PartNumber: "C7536A", Description: "HPE Cat6 RJ45 M/M 2m", Quantity: 3}}
		cables, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cables) != 3 {
			t.Errorf("expected 3 cables, got %d", len(cables))
		}
	})
}

// --- Step info builder tests ---

// TestBuildCableStepInfo verifies the step-through display info for an explicit
// cable, marking the cable-type field derived when the record omits it.
//
// Why it matters: step mode shows operators each cable's field mappings, so the
// display must report type/quantity/created items and flag inferred fields.
// Inputs: a cable with two terminations and records with and without a
// CableType. Outputs: a step-info struct (HwType, Quantity, Mappings,
// CreatedItems, IsDerived).
// Data choice: an explicit "cat6" gives three concrete mappings and one created
// item; the empty CableType case proves the third mapping is flagged derived.
func TestBuildCableStepInfo(t *testing.T) {
	t.Run("explicit cable", func(t *testing.T) {
		cable := devicetypes.NewCable("cat6", "switch:e0 ↔ server:e0")
		cable.SetTerminations(uuid.New(), uuid.New())
		rec := import_.CsvRecord{
			SourceDevice: "switch-01", SourcePort: "eth0",
			DestDevice: "server-01", DestPort: "eth0",
			CableType: "cat6",
		}
		info := buildCableStepInfo(rec, cable)
		if info.HwType != "cable" {
			t.Errorf("HwType = %q, want %q", info.HwType, "cable")
		}
		if info.Quantity != 1 {
			t.Errorf("Quantity = %d, want 1", info.Quantity)
		}
		if len(info.Mappings) != 3 {
			t.Errorf("len(Mappings) = %d, want 3", len(info.Mappings))
		}
		if len(info.CreatedItems) != 1 {
			t.Errorf("len(CreatedItems) = %d, want 1", len(info.CreatedItems))
		}
	})

	t.Run("derived cable type", func(t *testing.T) {
		cable := devicetypes.NewCable("cat6", "label")
		cable.SetTerminations(uuid.New(), uuid.New())
		rec := import_.CsvRecord{
			SourceDevice: "sw", SourcePort: "e0",
			DestDevice: "srv", DestPort: "e0",
			CableType: "",
		}
		info := buildCableStepInfo(rec, cable)
		if len(info.Mappings) >= 3 && !info.Mappings[2].IsDerived {
			t.Error("expected CableType mapping to be marked as derived when empty")
		}
	})
}

// TestBuildCableProductStepInfo verifies the step-through display info for a
// bulk cable product row.
//
// Why it matters: step mode must summarize multi-cable product rows, so the
// display reports the quantity, type, and one created item per cable.
// Inputs: two cat6 cables from a quantity-2 C7536A row. Outputs: a step-info
// struct.
// Data choice: two cables with quantity 2 prove the created-items list tracks
// each cable and the mapping and quantity counts match.
func TestBuildCableProductStepInfo(t *testing.T) {
	cables := []*devicetypes.CaniCableType{
		devicetypes.NewCable("cat6", "Cable-001"),
		devicetypes.NewCable("cat6", "Cable-002"),
	}
	rec := import_.CsvRecord{PartNumber: "C7536A", Description: "HPE Cat6 RJ45 Cable 2m", Quantity: 2}
	info := buildCableProductStepInfo(rec, cables)
	if info.Quantity != 2 {
		t.Errorf("Quantity = %d, want 2", info.Quantity)
	}
	if info.HwType != "cable" {
		t.Errorf("HwType = %q, want %q", info.HwType, "cable")
	}
	if len(info.CreatedItems) != 2 {
		t.Errorf("len(CreatedItems) = %d, want 2", len(info.CreatedItems))
	}
	if len(info.Mappings) != 2 {
		t.Errorf("len(Mappings) = %d, want 2", len(info.Mappings))
	}
}

// --- Explicit cable connection tests ---

// TestCreateCableFromExplicitRecord_WithLength verifies an explicit record's
// cable-length string is parsed onto the created cable.
//
// Why it matters: cable length feeds documentation and bill-of-materials
// output, so a "5m" field on a connection row must surface as a numeric length
// and unit rather than being dropped.
// Inputs: an inventory with a wired switch/server pair and a record carrying
// CableLength "5m". Outputs: a cable whose Length is 5 and LengthUnit is "m".
// Data choice: "5m" is the canonical "<number><unit>" form, the smallest input
// that proves both the numeric and unit halves are parsed.
func TestCreateCableFromExplicitRecord_WithLength(t *testing.T) {
	srcDevID, dstDevID := uuid.New(), uuid.New()
	srcIfaceID, dstIfaceID := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDevID: {ID: srcDevID, Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
			dstDevID: {ID: dstDevID, Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
		},
	}

	rec := import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0", CableLength: "5m"}
	cable, err := createCableFromExplicitRecord(inv, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cable.Length == nil || *cable.Length != 5 {
		t.Errorf("cable Length = %v, want 5", cable.Length)
	}
	if cable.LengthUnit != "m" {
		t.Errorf("cable LengthUnit = %q, want %q", cable.LengthUnit, "m")
	}
}

// TestTransformCables_ExplicitRecords verifies an explicit source/dest record
// becomes a cable whose interfaces are linked on both ends.
//
// Why it matters: explicit connection rows are how operators wire named ports
// together, so the transform must both create the cable and mark each endpoint
// interface as occupied, preventing a later pass from reusing a taken port.
// Inputs: an inventory with two devices, their interface specs, and the
// matching inv.Interfaces entries, plus one explicit record. Outputs: one
// cable with both terminations set and the source interface's ConnectedCable
// pointing at it.
// Data choice: registering the interfaces in both the device specs and the
// inv.Interfaces map is the minimum needed for GetInterfaceByID to resolve, so
// the linking step is exercised rather than short-circuited.
func TestTransformCables_ExplicitRecords(t *testing.T) {
	srcDevID, dstDevID := uuid.New(), uuid.New()
	srcIfaceID, dstIfaceID := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDevID: {ID: srcDevID, Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
			dstDevID: {ID: dstDevID, Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
		},
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			srcIfaceID: {ID: srcIfaceID, DeviceID: srcDevID},
			dstIfaceID: {ID: dstIfaceID, DeviceID: dstDevID},
		},
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{},
	}

	records := []import_.CsvRecord{
		{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0", CableType: "cat6"},
	}
	tally := &visual.StepTally{}
	recordNum := 0
	cables, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cables) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(cables))
	}

	var cable *devicetypes.CaniCableType
	for _, c := range cables {
		cable = c
	}
	if cable.TerminationA != srcIfaceID {
		t.Errorf("TerminationA = %v, want %v", cable.TerminationA, srcIfaceID)
	}
	if cable.TerminationB != dstIfaceID {
		t.Errorf("TerminationB = %v, want %v", cable.TerminationB, dstIfaceID)
	}
	if got := inv.Devices[srcDevID].Interfaces[0].ConnectedCable; got == nil || *got != cable.ID {
		t.Error("expected source interface ConnectedCable to point to the created cable")
	}
}

// TestTransformCables_ExplicitRecordError verifies an explicit record naming a
// missing device aborts the cable pass with an error.
//
// Why it matters: a cable to a non-existent device cannot be wired, so the
// transform must fail loudly rather than emit a dangling cable that would break
// downstream export.
// Inputs: an empty-device inventory and one explicit record referencing devices
// that do not exist. Outputs: a non-nil error from transformCables.
// Data choice: both endpoints are absent so the very first lookup in the
// explicit path fails, proving the error is propagated out of the loop.
func TestTransformCables_ExplicitRecordError(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
		Cables:  map[uuid.UUID]*devicetypes.CaniCableType{},
	}
	records := []import_.CsvRecord{
		{SourceDevice: "ghost-switch", SourcePort: "eth0", DestDevice: "ghost-server", DestPort: "eth0"},
	}
	tally := &visual.StepTally{}
	recordNum := 0
	if _, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records)); err == nil {
		t.Fatal("expected error for explicit cable record with missing devices")
	}
}

// --- Cable type resolution cascade ---

// TestResolveCableTypeSlug verifies the three-level resolution cascade:
// part-number lookup, slug lookup, then description-pattern inference.
//
// Why it matters: cable rows arrive with inconsistent identifiers, so the
// resolver must try the authoritative library keys before falling back to
// fuzzy description matching, otherwise cables would be mistyped in the BOM.
// Inputs: a real cable part number, a real cable slug, and a free-text
// description. Outputs: the canonical slug for each.
// Data choice: C7536A and its slug hpe-cat5-rj45-4-3m-cable are real library
// entries, so each cascade branch resolves against actual data rather than a
// fixture that could drift from production.
func TestResolveCableTypeSlug_Cascade(t *testing.T) {
	tests := []struct {
		name, partNumber, description, want string
	}{
		{"by part number", "C7536A", "ignored description", "hpe-cat5-rj45-4-3m-cable"},
		{"by slug from description", "", "hpe-cat5-rj45-4-3m-cable", "hpe-cat5-rj45-4-3m-cable"},
		{"by pattern inference", "", "Generic DAC direct attach copper", cableTypeDacPassive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveCableTypeSlug(tt.partNumber, tt.description); got != tt.want {
				t.Errorf("resolveCableTypeSlug(%q, %q) = %q, want %q", tt.partNumber, tt.description, got, tt.want)
			}
		})
	}
}

// TestParseLengthFromDescription_Overflow verifies a numeric token too large
// for float64 is rejected rather than silently parsed.
//
// Why it matters: a malformed or absurd length in a product description must
// not crash or produce an infinite length on a cable; the parser should treat
// it as "no length" so the cable simply carries no measurement.
// Inputs: a 400-digit number followed by "m". Outputs: zero length and empty
// unit.
// Data choice: 400 nines exceeds the float64 range, forcing strconv.ParseFloat
// to return a range error — the only way to exercise the parse-failure branch
// since the regex already guarantees the token is otherwise numeric.
func TestParseLengthFromDescription_Overflow(t *testing.T) {
	desc := strings.Repeat("9", 400) + "m"
	length, unit := parseLengthFromDescription(desc)
	if length != 0 || unit != "" {
		t.Errorf("parseLengthFromDescription overflow = (%v, %q), want (0, \"\")", length, unit)
	}
}

// TestParseCableLength_InvalidNumber verifies a token that matches the numeric
// pattern but is not a valid float is rejected as no length.
//
// Why it matters: the length regex admits multiple dots, so a malformed value
// like "1.2.3m" reaches strconv.ParseFloat; the parser must treat the parse
// failure as "no length" rather than propagate an error or a bogus number.
// Inputs: the string "1.2.3m". Outputs: zero length and empty unit.
// Data choice: "1.2.3m" passes the regex (digits and dots followed by a unit)
// yet fails ParseFloat, the only shape that reaches the parse-error branch.
func TestParseCableLength_InvalidNumber(t *testing.T) {
	length, unit := parseCableLength("1.2.3m")
	if length != 0 || unit != "" {
		t.Errorf("parseCableLength(%q) = (%v, %q), want (0, \"\")", "1.2.3m", length, unit)
	}
}

// TestLinkInterfacesToCable_MissingB verifies linking fails when the B-side
// termination interface is absent from the inventory index.
//
// Why it matters: a cable whose endpoint interface cannot be resolved would
// leave a dangling reference, so the linker must surface an error instead of
// marking a phantom interface as occupied.
// Inputs: an inventory where only the A interface is registered and a cable
// whose B termination points at an unregistered ID. Outputs: a non-nil error.
// Data choice: registering A but not B isolates the B-side failure path,
// proving the second lookup is checked independently of the first.
func TestLinkInterfacesToCable_MissingB(t *testing.T) {
	devID := uuid.New()
	aID, bID := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			devID: {ID: devID, Name: "d", Interfaces: []devicetypes.InterfaceSpec{{ID: aID, Name: "a"}}},
		},
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			aID: {ID: aID, DeviceID: devID},
		},
	}
	cable := devicetypes.NewCable("cat6", "x")
	cable.SetTerminations(aID, bID)
	if err := linkInterfacesToCable(inv, cable); err == nil {
		t.Fatal("expected error when termination B interface is missing from inventory")
	}
}

// TestTransformCables_ProductRecords verifies cable product rows create one
// cable per quantity, parse the length, and trigger auto-connect bookkeeping.
//
// Why it matters: product rows (no explicit endpoints) represent bulk cable
// purchases, so the transform must expand quantity into individual cables,
// initialize the inventory cable map lazily, and route them through the
// config-group auto-connect step.
// Inputs: a bare inventory (nil Cables map) and one product row with quantity
// 2, a "2m" length, and a config group. Outputs: two cables each 2 meters long.
// Data choice: quantity 2 proves the per-unit loop runs more than once, and a
// config group exercises the cablesByGroup path that single-quantity tests skip.
func TestTransformCables_ProductRecords(t *testing.T) {
	inv := &devicetypes.Inventory{}
	records := []import_.CsvRecord{
		{PartNumber: "C7536A", Description: "HPE Cat6 Cable 2m", Quantity: 2, ConfigGroup: "0900"},
	}
	tally := &visual.StepTally{}
	recordNum := 0
	cables, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cables) != 2 {
		t.Fatalf("expected 2 cables, got %d", len(cables))
	}
	for _, c := range cables {
		if c.Length == nil || *c.Length != 2 {
			t.Errorf("cable Length = %v, want 2", c.Length)
		}
	}
}

// TestTransformCables_LinkError verifies the explicit-record pass aborts when an
// endpoint interface is not registered in the inventory index.
//
// Why it matters: createCableFromExplicitRecord can succeed on the device's own
// interface slice while the inventory-wide index is incomplete, so the pass
// must still fail closed rather than emit an unlinked cable.
// Inputs: two devices with interface specs but no inv.Interfaces map entries,
// and one explicit record. Outputs: a non-nil error from transformCables.
// Data choice: omitting only the inv.Interfaces map isolates the link step as
// the failure point, since device lookup and cable creation both succeed first.
func TestTransformCables_LinkError(t *testing.T) {
	srcDevID, dstDevID := uuid.New(), uuid.New()
	srcIfaceID, dstIfaceID := uuid.New(), uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDevID: {ID: srcDevID, Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
			dstDevID: {ID: dstDevID, Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
		},
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{},
	}
	records := []import_.CsvRecord{
		{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0", CableType: "cat6"},
	}
	tally := &visual.StepTally{}
	recordNum := 0
	if _, err := transformCables(inv, records, false, visual.ETLOptions{}, tally, &recordNum, len(records)); err == nil {
		t.Fatal("expected link error when interfaces are not registered in inventory")
	}
}

// wiredCableInventory builds an inventory with two devices fully indexed so an
// explicit cable record links successfully.
func wiredCableInventory(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	srcDevID, dstDevID := uuid.New(), uuid.New()
	srcIfaceID, dstIfaceID := uuid.New(), uuid.New()
	return &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			srcDevID: {ID: srcDevID, Name: "switch-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: srcIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
			dstDevID: {ID: dstDevID, Name: "server-01",
				Interfaces: []devicetypes.InterfaceSpec{{ID: dstIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT}}},
		},
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			srcIfaceID: {ID: srcIfaceID, DeviceID: srcDevID},
			dstIfaceID: {ID: dstIfaceID, DeviceID: dstDevID},
		},
		Cables: map[uuid.UUID]*devicetypes.CaniCableType{},
	}
}

// TestTransformCables_ExplicitStepMode verifies step-through prompting around an
// explicit cable, both when the user continues and when the prompt is interrupted.
//
// Why it matters: step mode lets an operator review each cable interactively, so
// the transform must advance the tally and emit a prompt on success yet abort
// the whole pass if the prompt stream closes unexpectedly.
// Inputs: a wired inventory, one explicit record, and a redirected stdin that
// either supplies a newline or reaches EOF. Outputs: a tally increment and no
// error on success; a non-nil error on interruption.
// Data choice: an EOF stdin is the simplest deterministic way to make the
// interactive prompt return an error without a real terminal.
func TestTransformCables_ExplicitStepMode(t *testing.T) {
	rec := import_.CsvRecord{SourceDevice: "switch-01", SourcePort: "eth0", DestDevice: "server-01", DestPort: "eth0", CableType: "cat6"}
	opts := visual.ETLOptions{NoColor: true, Writer: io.Discard}

	t.Run("continues on enter", func(t *testing.T) {
		withStdin(t, "\n")
		inv := wiredCableInventory(t)
		tally := &visual.StepTally{}
		recordNum := 0
		if _, err := transformCables(inv, []import_.CsvRecord{rec}, true, opts, tally, &recordNum, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tally.Cables != 1 {
			t.Errorf("tally.Cables = %d, want 1", tally.Cables)
		}
	})

	t.Run("aborts when prompt interrupted", func(t *testing.T) {
		withStdin(t, "")
		inv := wiredCableInventory(t)
		tally := &visual.StepTally{}
		recordNum := 0
		if _, err := transformCables(inv, []import_.CsvRecord{rec}, true, opts, tally, &recordNum, 1); err == nil {
			t.Fatal("expected error when explicit-cable step prompt is interrupted")
		}
	})
}

// TestTransformCables_ProductStepMode verifies step-through prompting around a
// cable product row on both the continue and interrupt paths.
//
// Why it matters: product rows can expand into many cables, and step mode must
// summarize the batch with a single prompt while still failing the pass if the
// prompt stream closes.
// Inputs: a bare inventory, one product row (quantity 2), and a redirected
// stdin supplying a newline or EOF. Outputs: a tally of 2 and no error on
// success; a non-nil error on interruption.
// Data choice: quantity 2 ensures the batch tally branch (len > 0) is taken,
// and the EOF case deterministically drives the prompt error.
func TestTransformCables_ProductStepMode(t *testing.T) {
	rec := import_.CsvRecord{PartNumber: "C7536A", Description: "HPE Cat6 Cable 2m", Quantity: 2, ConfigGroup: "0900"}
	opts := visual.ETLOptions{NoColor: true, Writer: io.Discard}

	t.Run("continues on enter", func(t *testing.T) {
		withStdin(t, "\n")
		inv := &devicetypes.Inventory{}
		tally := &visual.StepTally{}
		recordNum := 0
		if _, err := transformCables(inv, []import_.CsvRecord{rec}, true, opts, tally, &recordNum, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tally.Cables != 2 {
			t.Errorf("tally.Cables = %d, want 2", tally.Cables)
		}
	})

	t.Run("aborts when prompt interrupted", func(t *testing.T) {
		withStdin(t, "")
		inv := &devicetypes.Inventory{}
		tally := &visual.StepTally{}
		recordNum := 0
		if _, err := transformCables(inv, []import_.CsvRecord{rec}, true, opts, tally, &recordNum, 1); err == nil {
			t.Fatal("expected error when product-cable step prompt is interrupted")
		}
	})
}

// TestConnectCablesHubSpoke verifies the hub-spoke auto-connect handles running
// out of cables and skipping endpoints that have no free interface.
//
// Why it matters: auto-connect pairs spokes to hubs using a finite cable pool,
// so it must stop cleanly when cables run out and skip any device whose
// interfaces are all occupied rather than panic or over-subscribe a port.
// Inputs: small hub/spoke sets with controlled interface counts and a
// one-element cable pool. Outputs: a connected first cable, or an untouched
// cable when an endpoint has no available interface.
// Data choice: two spokes with one cable forces the pool-exhausted break, while
// zero-interface devices isolate the spoke-skip and hub-skip branches.
func TestConnectCablesHubSpoke_EdgeCases(t *testing.T) {
	mkDev := func(name string, ifaces int) *devicetypes.CaniDeviceType {
		d := &devicetypes.CaniDeviceType{ID: uuid.New(), Name: name}
		for i := 0; i < ifaces; i++ {
			d.Interfaces = append(d.Interfaces, devicetypes.InterfaceSpec{ID: uuid.New(), Name: fmt.Sprintf("e%d", i)})
		}
		return d
	}

	t.Run("breaks when cables exhausted", func(t *testing.T) {
		inv := &devicetypes.Inventory{}
		cables := []*devicetypes.CaniCableType{devicetypes.NewCable("cat6", "c1")}
		connectCablesHubSpoke(inv, cables,
			[]*devicetypes.CaniDeviceType{mkDev("hub", 2)},
			[]*devicetypes.CaniDeviceType{mkDev("s1", 1), mkDev("s2", 1)})
		if cables[0].TerminationA == uuid.Nil {
			t.Error("expected first cable to be connected")
		}
	})

	t.Run("skips spoke with no free interface", func(t *testing.T) {
		inv := &devicetypes.Inventory{}
		cables := []*devicetypes.CaniCableType{devicetypes.NewCable("cat6", "c1")}
		connectCablesHubSpoke(inv, cables,
			[]*devicetypes.CaniDeviceType{mkDev("hub", 1)},
			[]*devicetypes.CaniDeviceType{mkDev("s1", 0)})
		if cables[0].TerminationA != uuid.Nil {
			t.Error("expected no connection when spoke has no free interface")
		}
	})

	t.Run("skips when hub has no free interface", func(t *testing.T) {
		inv := &devicetypes.Inventory{}
		cables := []*devicetypes.CaniCableType{devicetypes.NewCable("cat6", "c1")}
		connectCablesHubSpoke(inv, cables,
			[]*devicetypes.CaniDeviceType{mkDev("hub", 0)},
			[]*devicetypes.CaniDeviceType{mkDev("s1", 1)})
		if cables[0].TerminationA != uuid.Nil {
			t.Error("expected no connection when hub has no free interface")
		}
	})
}
