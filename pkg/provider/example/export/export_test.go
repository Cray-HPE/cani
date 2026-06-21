package export

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestExport verifies Export renders the top-level inventory banner, available
// device-only content, and the summary line for populated and empty inventories.
//
// Why it matters: Export is the example provider's operator-facing output path,
// so it must always produce a readable summary and render standalone devices when
// no location or rack hierarchy exists.
// Inputs: an inventory with one standalone device and an empty inventory.
// Outputs: nil errors and stdout containing the expected summary lines.
// Data choice: the device-only case forces Export's final content branch, while
// the empty case proves the summary still renders without inventory objects.
func TestExport(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		testName string
		inv      devicetypes.Inventory
		contains string
	}{
		{
			testName: "standalone device inventory",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {Name: "server-01", Type: devicetypes.Type("server")},
				},
			},
			contains: "Summary: 0 locations, 0 racks, 1 devices, 0 modules, 0 cables",
		},
		{
			testName: "empty inventory",
			inv:      devicetypes.Inventory{},
			contains: "Summary: 0 locations, 0 racks, 0 devices, 0 modules, 0 cables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := captureStdout(t, func() {
				if err := Export(tt.inv); err != nil {
					t.Errorf("Export() unexpected error: %v", err)
				}
			})

			if !strings.Contains(got, tt.contains) {
				t.Errorf("Export() = %q expecting: \n%q\n", tt.contains, got)
			}
		})
	}
}

// TestPrintLocation verifies printLocation renders a location line and returns
// silently for nil locations.
//
// Why it matters: location rendering is recursive, so the base line and nil guard
// must be stable before children and racks are walked.
// Inputs: a simple site location and a nil location. Outputs: the rendered line
// or no stdout.
// Data choice: a single site location isolates the direct format, while nil
// exercises the defensive early return.
func TestPrintLocation(t *testing.T) {
	tests := []struct {
		testName string
		location *devicetypes.CaniLocationType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "location line",
			location: &devicetypes.CaniLocationType{
				Name:         "location-01",
				LocationType: "site",
			},
			inv:      &devicetypes.Inventory{},
			expected: "📍 location-01 (site)",
		},
		{
			testName: "nil location prints nothing",
			location: nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := strings.TrimSpace(captureStdout(t, func() { printLocation(tt.location, tt.inv, 0) }))

			if got != tt.expected {
				t.Errorf("printLocation() output = %q, want %q", got, tt.expected)
			}
		})
	}

}

// TestPrintRack verifies printRack renders the rack frame and returns silently
// for a nil rack.
//
// Why it matters: rack output is the frame around device placement, and nil rack
// references can appear when parent relationships are incomplete.
// Inputs: a 42U rack with no devices and a nil rack. Outputs: the rack frame or
// no stdout.
// Data choice: an empty 42U rack isolates frame rendering without cable/device
// noise, while nil exercises the defensive branch.
func TestPrintRack(t *testing.T) {
	tests := []struct {
		testName string
		rack     *devicetypes.CaniRackType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "empty rack frame",
			rack: &devicetypes.CaniRackType{
				Name:    "rack-01",
				UHeight: 42,
			},
			inv:      &devicetypes.Inventory{},
			expected: "🗄️  rack-01 [42U]\n  ┌─────────────────────────────────────────────────────────┐\n  └─────────────────────────────────────────────────────────┘",
		},
		{
			testName: "nil rack prints nothing",
			rack:     nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := strings.TrimSpace(captureStdout(t, func() { printRack(tt.rack, tt.inv, 0) }))

			if got != tt.expected {
				t.Errorf("printRack() output = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPrintRackCables verifies printRackCables renders a rack cable with a known
// A endpoint and unknown B endpoint, and prints nothing when no cables exist.
//
// Why it matters: partially resolved cables are still useful in the export, but
// an inventory with no cables should not emit an empty Cables section.
// Inputs: a rack with one device/interface and one cable terminating at that
// interface, plus an empty rack/inventory. Outputs: a cable line or no stdout.
// Data choice: one missing B termination proves the unknown fallback text, while
// an empty inventory isolates the no-cables guard.
func TestPrintRackCables(t *testing.T) {
	deviceID := uuid.New()
	ifaceID := uuid.New()
	cableID := uuid.New()

	tests := []struct {
		testName string
		rack     *devicetypes.CaniRackType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "one known endpoint",
			rack: &devicetypes.CaniRackType{
				Devices: []uuid.UUID{deviceID},
			},
			inv: &devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID:   deviceID,
						Name: "server-01",
						Interfaces: []devicetypes.InterfaceSpec{
							{ID: ifaceID, Name: "eth0"},
						},
					},
				},
				Cables: map[uuid.UUID]*devicetypes.CaniCableType{
					cableID: {Slug: "cat6", TerminationA: ifaceID},
				},
				Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
					ifaceID: {ID: ifaceID, Name: "eth0", DeviceID: deviceID},
				},
			},
			expected: "Cables:\n  ⚡ [cat6] server-01:eth0 ══ unknown:?",
		},
		{
			testName: "no cables prints nothing",
			rack:     &devicetypes.CaniRackType{},
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := strings.TrimSpace(captureStdout(t, func() { printRackCables(tt.rack, tt.inv, 0) }))

			if got != tt.expected {
				t.Errorf("printRackCables() output = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPrintDevice verifies printDevice renders a device line and returns
// silently for a nil device.
//
// Why it matters: devices are the common leaf export item, and nil child
// references must not break recursive rendering.
// Inputs: one modeled server device and a nil device. Outputs: the rendered line
// or no stdout.
// Data choice: the server has Name, Type, and Model populated to prove the full
// line format, while nil isolates the guard branch.
func TestPrintDevice(t *testing.T) {
	tests := []struct {
		testName string
		device   *devicetypes.CaniDeviceType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "device line",
			device: &devicetypes.CaniDeviceType{
				Name:  "server-01",
				Type:  devicetypes.Type("server"),
				Model: "ProLiant DL360",
			},
			inv:      &devicetypes.Inventory{},
			expected: "🖥️  server-01 (server) - ProLiant DL360",
		},
		{
			testName: "nil device prints nothing",
			device:   nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := strings.TrimSpace(captureStdout(t, func() { printDevice(tt.device, tt.inv, 0) }))

			if got != tt.expected {
				t.Errorf("printDevice() output = %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestPrintModule verifies printModule renders a module line and returns
// silently for nil modules.
//
// Why it matters: module output identifies hardware installed inside devices,
// and nil module references should not interrupt device rendering.
// Inputs: one GPU module with bay and slug fields plus a nil module. Outputs:
// the rendered line or no stdout.
// Data choice: Name, ModuleBayName, and Slug are the exact fields printModule
// formats, and nil exercises the guard branch.
func TestPrintModule(t *testing.T) {
	tests := []struct {
		testName string
		module   *devicetypes.CaniModuleType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "module line",
			module: &devicetypes.CaniModuleType{
				Name:          "gpu-a100",
				ModuleBayName: "bay-0",
				Slug:          "nvidia-a100",
			},
			inv:      &devicetypes.Inventory{},
			expected: "📦 gpu-a100 [bay-0] - nvidia-a100",
		},
		{
			testName: "nil module prints nothing",
			module:   nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := strings.TrimSpace(captureStdout(t, func() { printModule(tt.module, tt.inv, 0) }))

			if got != tt.expected {
				t.Errorf("printModule() output = %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestGetDeviceName verifies getDeviceName returns the inventory device name
// when the ID is present.
//
// Why it matters: cable and hierarchy renderers need stable human-readable names
// for referenced device UUIDs.
// Inputs: an inventory with one device keyed by UUID. Outputs: that device's
// Name field.
// Data choice: a single known device isolates the found branch; the miss branch
// is covered separately by TestGetDeviceName_NotFound.
func TestGetDeviceName(t *testing.T) {
	deviceID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			deviceID: {Name: "test-device"},
		},
	}

	got := getDeviceName(inv, deviceID)

	if got != "test-device" {
		t.Errorf("getDeviceName() = %q, expected %q", got, "test-device")
	}
}

// captureStdout redirects os.Stdout while fn runs and returns everything it
// printed. It centralizes the pipe-swap boilerplate so each new test stays flat.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return buf.String()
}

// TestExport_LocationsPath verifies Export walks the location hierarchy when the
// inventory has locations.
//
// Why it matters: locations are the top of the inventory tree, so the
// location-first branch is the primary rendering path for any real site and must
// emit the location line plus an accurate summary.
// Inputs: an inventory with a single location and nothing else. Outputs: stdout
// containing the location line and the "1 locations" summary.
// Data choice: exactly one location with zero racks/devices isolates the
// locations branch from the racks-only and devices-only fallbacks.
func TestExport_LocationsPath(t *testing.T) {
	inv := devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			uuid.New(): {Name: "hall-A", LocationType: "site"},
		},
	}

	out := captureStdout(t, func() {
		if err := Export(inv); err != nil {
			t.Errorf("Export() error = %v, want nil", err)
		}
	})

	if !strings.Contains(out, "hall-A (site)") {
		t.Errorf("Export() output missing location line; got:\n%s", out)
	}
	if !strings.Contains(out, "1 locations, 0 racks, 0 devices, 0 modules, 0 cables") {
		t.Errorf("Export() output missing expected summary; got:\n%s", out)
	}
}

// TestExport_RacksOnlyPath verifies Export prints racks directly when there are
// no locations but racks exist.
//
// Why it matters: inventories imported before location data is attached still
// need a readable layout, so the racks-only fallback must render each rack rather
// than print nothing.
// Inputs: an inventory with one rack and no locations. Outputs: stdout containing
// the rack line and the "1 racks" summary.
// Data choice: a single 42U rack with no locations forces the else-if racks
// branch and proves the rack renderer is reached.
func TestExport_RacksOnlyPath(t *testing.T) {
	inv := devicetypes.Inventory{
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			uuid.New(): {Name: "rack-9", UHeight: 42},
		},
	}

	out := captureStdout(t, func() {
		if err := Export(inv); err != nil {
			t.Errorf("Export() error = %v, want nil", err)
		}
	})

	if !strings.Contains(out, "rack-9 [42U]") {
		t.Errorf("Export() output missing rack line; got:\n%s", out)
	}
	if !strings.Contains(out, "0 locations, 1 racks, 0 devices, 0 modules, 0 cables") {
		t.Errorf("Export() output missing expected summary; got:\n%s", out)
	}
}

// TestPrintLocation_ChildrenAndRacks verifies a location renders its child
// locations and its racks beneath it.
//
// Why it matters: the location tree is recursive, so a parent must descend into
// both nested locations and the racks it directly contains, or whole branches of
// the inventory would be invisible.
// Inputs: a parent location referencing one child location and one rack, with
// both present in the inventory. Outputs: stdout containing the parent, the
// child, and the rack lines, parent first.
// Data choice: one child plus one rack exercises both child loops at once while
// keeping the expected order unambiguous.
func TestPrintLocation_ChildrenAndRacks(t *testing.T) {
	childLocID := uuid.New()
	rackID := uuid.New()
	inv := &devicetypes.Inventory{
		Locations: map[uuid.UUID]*devicetypes.CaniLocationType{
			childLocID: {Name: "row-1", LocationType: "row"},
		},
		Racks: map[uuid.UUID]*devicetypes.CaniRackType{
			rackID: {Name: "rack-7", UHeight: 42},
		},
	}
	parent := &devicetypes.CaniLocationType{
		Name:         "datacenter",
		LocationType: "site",
		Children:     []uuid.UUID{childLocID},
		Racks:        []uuid.UUID{rackID},
	}

	out := captureStdout(t, func() { printLocation(parent, inv, 0) })

	for _, want := range []string{"datacenter (site)", "row-1 (row)", "rack-7 [42U]"} {
		if !strings.Contains(out, want) {
			t.Errorf("printLocation() output missing %q; got:\n%s", want, out)
		}
	}
	if strings.Index(out, "datacenter") > strings.Index(out, "row-1") {
		t.Errorf("expected parent before child location; got:\n%s", out)
	}
}

// TestPrintRack_DevicesSorted verifies devices in a rack are printed top-to-bottom
// by U-position, using the occupied-slot map when present and the device's
// RackPosition as a fallback.
//
// Why it matters: a rack diagram is only correct if higher U-numbers print first
// and both the slot-map and fallback position sources agree, otherwise operators
// read the physical layout upside down or with gaps.
// Inputs: a rack with a 2U device pinned at U20 via OccupiedSlots and a 1U device
// at RackPosition 10, listed bottom-first to force a sort. Outputs: stdout with
// the U20 device before the U10 device, the correct U ranges, and a "device"
// hardware-type fallback for the untyped node.
// Data choice: one slot-mapped multi-U device and one fallback single-U device
// cover both the start-U branches, both U-range formats, and the empty-Type
// fallback in a single render.
func TestPrintRack_DevicesSorted(t *testing.T) {
	topID := uuid.New()
	botID := uuid.New()
	rack := &devicetypes.CaniRackType{
		Name:    "rack-A",
		UHeight: 42,
		Devices: []uuid.UUID{botID, topID}, // bottom-first to prove sorting
		OccupiedSlots: map[int]map[string]uuid.UUID{
			20: {"front": topID},
		},
	}
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			topID: {ID: topID, Name: "top-node", Type: devicetypes.Type("server"), UHeight: 2},
			botID: {ID: botID, Name: "bot-node", Type: devicetypes.Type(""), RackPosition: 10, UHeight: 1},
		},
	}

	out := captureStdout(t, func() { printRack(rack, inv, 0) })

	for _, want := range []string{"U20-U21", "U10", "top-node", "bot-node", "server", "device"} {
		if !strings.Contains(out, want) {
			t.Errorf("printRack() output missing %q; got:\n%s", want, out)
		}
	}
	if strings.Index(out, "top-node") > strings.Index(out, "bot-node") {
		t.Errorf("expected U20 device before U10 device; got:\n%s", out)
	}
}

// TestPrintDevice_ChildrenAndModules verifies a device renders its child devices
// and its modules, and falls back to a generic type label when Type is empty.
//
// Why it matters: chassis-style hardware nests blades and modules, so the device
// renderer must recurse into children and list modules or composed systems would
// appear empty.
// Inputs: a parent device with an empty Type referencing one child device, plus a
// module whose ParentDevice is the parent. Outputs: stdout containing the parent
// (typed "device"), the child, and the module, parent first.
// Data choice: an untyped parent with exactly one child and one module exercises
// the children loop, the module loop, and the hardware-type fallback together.
func TestPrintDevice_ChildrenAndModules(t *testing.T) {
	parentID := uuid.New()
	childID := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			parentID: {ID: parentID, Name: "chassis", Type: devicetypes.Type(""), Model: "DL", Children: []uuid.UUID{childID}},
			childID:  {ID: childID, Name: "blade-1", Type: devicetypes.Type("blade"), Model: "X"},
		},
		Modules: map[uuid.UUID]*devicetypes.CaniModuleType{
			uuid.New(): {Name: "nic-0", ModuleBayName: "bay-0", Slug: "mlx", ParentDevice: parentID},
		},
	}

	out := captureStdout(t, func() { printDevice(inv.Devices[parentID], inv, 0) })

	for _, want := range []string{"chassis (device) - DL", "blade-1 (blade)", "nic-0 [bay-0] - mlx"} {
		if !strings.Contains(out, want) {
			t.Errorf("printDevice() output missing %q; got:\n%s", want, out)
		}
	}
	if strings.Index(out, "chassis") > strings.Index(out, "blade-1") {
		t.Errorf("expected parent before child device; got:\n%s", out)
	}
}

// TestPrintRackCables_EdgeCases verifies the two remaining cable branches: both
// endpoints resolving to named devices, and cables that exist but touch no device
// in this rack.
//
// Why it matters: a cable list must name both ends when known and stay silent
// when no cable belongs to the rack, so operators are not shown phantom or
// half-resolved links.
// Inputs: (1) a rack holding both cable endpoints with named interfaces; (2) a
// rack holding an unrelated device while the only cable joins two other devices.
// Outputs: (1) a fully named "A:port ══ B:port" line; (2) no output at all.
// Data choice: two fully resolved endpoints prove the deviceB/ifaceB branches,
// and a cable wholly outside the rack proves the empty-rackCables early return.
func TestPrintRackCables_EdgeCases(t *testing.T) {
	t.Run("both endpoints resolve", func(t *testing.T) {
		devAID, devBID := uuid.New(), uuid.New()
		ifAID, ifBID := uuid.New(), uuid.New()
		rack := &devicetypes.CaniRackType{Devices: []uuid.UUID{devAID, devBID}}
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				devAID: {ID: devAID, Name: "leaf-1", Interfaces: []devicetypes.InterfaceSpec{{ID: ifAID, Name: "eth0"}}},
				devBID: {ID: devBID, Name: "spine-1", Interfaces: []devicetypes.InterfaceSpec{{ID: ifBID, Name: "eth1"}}},
			},
			Cables: map[uuid.UUID]*devicetypes.CaniCableType{
				uuid.New(): {Slug: "qsfp", TerminationA: ifAID, TerminationB: ifBID},
			},
			Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
				ifAID: {ID: ifAID, Name: "eth0", DeviceID: devAID},
				ifBID: {ID: ifBID, Name: "eth1", DeviceID: devBID},
			},
		}

		out := captureStdout(t, func() { printRackCables(rack, inv, 0) })

		if !strings.Contains(out, "[qsfp] leaf-1:eth0 ══ spine-1:eth1") {
			t.Errorf("printRackCables() missing fully resolved cable; got:\n%s", out)
		}
	})

	t.Run("cables exist but none in this rack", func(t *testing.T) {
		inRackID := uuid.New()
		otherAID, otherBID := uuid.New(), uuid.New()
		ifAID, ifBID := uuid.New(), uuid.New()
		rack := &devicetypes.CaniRackType{Devices: []uuid.UUID{inRackID}}
		inv := &devicetypes.Inventory{
			Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
				inRackID: {ID: inRackID, Name: "in-rack"},
				otherAID: {ID: otherAID, Name: "other-a", Interfaces: []devicetypes.InterfaceSpec{{ID: ifAID, Name: "eth0"}}},
				otherBID: {ID: otherBID, Name: "other-b", Interfaces: []devicetypes.InterfaceSpec{{ID: ifBID, Name: "eth1"}}},
			},
			Cables: map[uuid.UUID]*devicetypes.CaniCableType{
				uuid.New(): {Slug: "qsfp", TerminationA: ifAID, TerminationB: ifBID},
			},
			Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
				ifAID: {ID: ifAID, Name: "eth0", DeviceID: otherAID},
				ifBID: {ID: ifBID, Name: "eth1", DeviceID: otherBID},
			},
		}

		out := captureStdout(t, func() { printRackCables(rack, inv, 0) })

		if strings.TrimSpace(out) != "" {
			t.Errorf("printRackCables() should print nothing when no cable is in rack; got:\n%s", out)
		}
	})
}

// TestGetDeviceName_NotFound verifies getDeviceName falls back to the truncated
// UUID when the device is absent from the inventory.
//
// Why it matters: callers render a label for every referenced ID, so a missing
// device must still yield a stable, human-scannable token instead of an empty
// string or panic.
// Inputs: an empty device map and a random UUID. Outputs: the first eight
// characters of that UUID's string form.
// Data choice: an empty map guarantees the lookup misses, isolating the fallback
// branch from the already-tested found path.
func TestGetDeviceName_NotFound(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{}}

	got := getDeviceName(inv, id)

	if want := id.String()[:8]; got != want {
		t.Errorf("getDeviceName() = %q, want %q", got, want)
	}
}
