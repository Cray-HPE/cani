/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

var (
	sampleRackID        = uuid.MustParse("00000000-0000-0000-0001-000000000001")
	sampleNodeAID       = uuid.MustParse("00000000-0000-0000-0002-000000000001")
	sampleNodeBID       = uuid.MustParse("00000000-0000-0000-0002-000000000002")
	sampleSwitchID      = uuid.MustParse("00000000-0000-0000-0003-000000000001")
	sampleNodeAIfaceID  = uuid.MustParse("00000000-0000-0000-0004-000000000001")
	sampleSwitchIfaceID = uuid.MustParse("00000000-0000-0000-0005-000000000001")
	sampleCableID       = uuid.MustParse("00000000-0000-0000-0006-000000000001")
)

func sampleRackInventory() *devicetypes.Inventory {
	inventory := devicetypes.NewInventory()

	inventory.Racks[sampleRackID] = &devicetypes.CaniRackType{
		ID:      sampleRackID,
		Name:    "Rack-006U",
		UHeight: 6,
		Devices: []uuid.UUID{sampleNodeAID, sampleNodeBID, sampleSwitchID},
	}
	inventory.Devices[sampleRackID] = &devicetypes.CaniDeviceType{
		ID:       sampleRackID,
		Name:     "Rack-006U",
		Type:     devicetypes.Rack,
		Children: []uuid.UUID{sampleNodeAID, sampleNodeBID, sampleSwitchID},
	}
	inventory.Devices[sampleNodeAID] = &devicetypes.CaniDeviceType{
		ID:           sampleNodeAID,
		Name:         "Node-A",
		Type:         devicetypes.TypeNode,
		Model:        "HPE DL360",
		UHeight:      2,
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive), Role: "compute"},
		Parent:       sampleRackID,
		Rack:         sampleRackID,
		RackPosition: 1,
		Interfaces: []devicetypes.InterfaceSpec{
			{ID: sampleNodeAIfaceID, Name: "eth0", Type: devicetypes.InterfacesElemTypeA1000BaseT},
		},
	}
	inventory.Devices[sampleNodeBID] = &devicetypes.CaniDeviceType{
		ID:           sampleNodeBID,
		Name:         "Node-B",
		Type:         devicetypes.TypeNode,
		Model:        "HPE DL360",
		UHeight:      1,
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive), Role: "compute"},
		Parent:       sampleRackID,
		Rack:         sampleRackID,
		RackPosition: 4,
	}
	inventory.Devices[sampleSwitchID] = &devicetypes.CaniDeviceType{
		ID:           sampleSwitchID,
		Name:         "Leaf-1",
		Type:         devicetypes.TypeSwitch,
		Model:        "Aruba 8325",
		UHeight:      1,
		ObjectMeta:   devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive), Role: "leaf"},
		Parent:       sampleRackID,
		Rack:         sampleRackID,
		RackPosition: 6,
		Interfaces: []devicetypes.InterfaceSpec{
			{ID: sampleSwitchIfaceID, Name: "1/1/1", Type: devicetypes.InterfacesElemTypeA1000BaseT},
		},
	}
	inventory.Interfaces[sampleNodeAIfaceID] = &devicetypes.CaniInterface{
		ID:            sampleNodeAIfaceID,
		Name:          "eth0",
		InterfaceType: devicetypes.InterfacesElemTypeA1000BaseT,
		DeviceID:      sampleNodeAID,
	}
	inventory.Interfaces[sampleSwitchIfaceID] = &devicetypes.CaniInterface{
		ID:            sampleSwitchIfaceID,
		Name:          "1/1/1",
		InterfaceType: devicetypes.InterfacesElemTypeA1000BaseT,
		DeviceID:      sampleSwitchID,
	}
	inventory.Cables[sampleCableID] = &devicetypes.CaniCableType{
		ID:                 sampleCableID,
		Slug:               "cat6",
		Label:              "node-a-to-leaf",
		CableType:          "cat6",
		ObjectMeta:         devicetypes.ObjectMeta{Status: string(devicetypes.StatusConnected)},
		TerminationA:       sampleNodeAIfaceID,
		TerminationB:       sampleSwitchIfaceID,
		TerminationADevice: sampleNodeAID,
		TerminationBDevice: sampleSwitchID,
		TerminationAPort:   "eth0",
		TerminationBPort:   "1/1/1",
	}

	return inventory
}

func sampleRackView(t *testing.T) (*devicetypes.Inventory, *RackView) {
	t.Helper()

	inventory := sampleRackInventory()
	rackView, err := BuildRackVisualization(inventory, sampleRackID)
	if err != nil {
		t.Fatalf("BuildRackVisualization returned error: %v", err)
	}

	return inventory, rackView
}

func assertExactOutput(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("output mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func expectedRackASCIIOutput(showCables bool) string {
	const (
		innerWidth   = 78
		uColumnWidth = 5
		contentWidth = 72
	)

	lines := []string{
		"┌" + strings.Repeat("─", innerWidth) + "┐",
		"│" + strings.Repeat(" ", 34) + "Rack-006U" + strings.Repeat(" ", 35) + "│",
		"├" + strings.Repeat("─", uColumnWidth) + "┬" + strings.Repeat("─", contentWidth) + "┤",
		rackGoldenSlotLine(6, "█ Leaf-1", contentWidth),
		rackGoldenSlotLine(5, "░ [EMPTY]", contentWidth),
		rackGoldenSlotLine(4, "█ Node-B", contentWidth),
		rackGoldenSlotLine(3, "░ [EMPTY]", contentWidth),
		rackGoldenSlotLine(2, "▓ (continued)", contentWidth),
		rackGoldenSlotLine(1, "█ Node-A (2U)", contentWidth),
		"└" + strings.Repeat("─", uColumnWidth) + "┴" + strings.Repeat("─", contentWidth) + "┘",
		"  Summary: 3 devices, 4/6 U occupied, 2 U empty",
	}
	if showCables {
		lines = append(lines,
			"",
			"  Cable Connections (1 cables):",
			"    • node-a-to-leaf [Node-A:eth0] ←→ [Leaf-1:1/1/1]",
		)
	}

	return strings.Join(lines, "\n") + "\n"
}

func rackGoldenSlotLine(position int, content string, contentWidth int) string {
	return fmt.Sprintf("│ U%-3d│%s│", position, rightPadVisible(content, contentWidth))
}

func rightPadVisible(value string, width int) string {
	visibleWidth := utf8.RuneCountInString(value)
	if visibleWidth >= width {
		return value
	}
	return value + strings.Repeat(" ", width-visibleWidth)
}

func assertRackSlot(t *testing.T, rackView *RackView, position int, deviceName string, isStart, isContinued bool) {
	t.Helper()

	slot := rackView.GetSlot(position)
	if slot == nil {
		t.Fatalf("slot U%d is empty, want %s", position, deviceName)
	}
	if slot.Device.Name != deviceName {
		t.Fatalf("slot U%d device = %q, want %q", position, slot.Device.Name, deviceName)
	}
	if slot.IsStart != isStart {
		t.Fatalf("slot U%d IsStart = %t, want %t", position, slot.IsStart, isStart)
	}
	if slot.IsContinued != isContinued {
		t.Fatalf("slot U%d IsContinued = %t, want %t", position, slot.IsContinued, isContinued)
	}
}

func assertRackSlotEmpty(t *testing.T, rackView *RackView, position int) {
	t.Helper()

	if rackView.GetSlot(position) != nil {
		t.Fatalf("slot U%d is occupied, want empty", position)
	}
}

// TestFindAllRacks verifies rack discovery returns the rack device from a mixed
// inventory.
//
// Why it matters: rack rendering starts by discovering rack-typed devices before
// building the user-visible diagram.
// Inputs: an inventory with one rack and three child devices. Outputs: exactly
// one discovered rack named Rack-006U.
// Data choice: the fixture includes non-rack children so the test proves type
// filtering rather than simple inventory length.
func TestFindAllRacks(t *testing.T) {
	inventory := sampleRackInventory()

	racks := FindAllRacks(inventory)

	if len(racks) != 1 {
		t.Fatalf("len(racks) = %d, want 1", len(racks))
	}
	if racks[0].Name != "Rack-006U" {
		t.Fatalf("rack name = %q, want %q", racks[0].Name, "Rack-006U")
	}
}

// TestBuildRackVisualization verifies rack height, slot occupancy, and device
// counts are derived from inventory relationships.
//
// Why it matters: the rack diagram can only be trusted if the intermediate view
// preserves rack identity and physical U occupancy correctly.
// Inputs: a 6U rack inventory with three positioned devices. Outputs: rack name,
// height, unique device count, occupied U count, and empty U count.
// Data choice: one 2U node plus two 1U devices proves that physical occupancy is
// distinct from unique device count.
func TestBuildRackVisualization(t *testing.T) {
	_, rackView := sampleRackView(t)

	if rackView.Rack.Name != "Rack-006U" {
		t.Fatalf("rack name = %q, want %q", rackView.Rack.Name, "Rack-006U")
	}
	if rackView.Height != 6 {
		t.Fatalf("rack height = %d, want 6", rackView.Height)
	}
	if rackView.DeviceCount() != 3 {
		t.Fatalf("device count = %d, want 3", rackView.DeviceCount())
	}
	if rackView.OccupiedCount() != 4 {
		t.Fatalf("occupied count = %d, want 4", rackView.OccupiedCount())
	}
	if rackView.EmptyCount() != 2 {
		t.Fatalf("empty count = %d, want 2", rackView.EmptyCount())
	}
}

// TestRackPositions verifies each U slot points to the expected device and slot
// role.
//
// Why it matters: top-to-bottom rack diagrams rely on exact U placement, start
// markers, and continued markers for multi-U hardware.
// Inputs: a rack view with a 2U node at U1, a 1U node at U4, and a switch at U6.
// Outputs: slot names and IsStart/IsContinued flags for occupied and empty U rows.
// Data choice: gaps at U3 and U5 make empty-slot assertions meaningful instead
// of only checking occupied rows.
func TestRackPositions(t *testing.T) {
	_, rackView := sampleRackView(t)

	testCases := []struct {
		name        string
		position    int
		deviceName  string
		isStart     bool
		isContinued bool
	}{
		{name: "node start", position: 1, deviceName: "Node-A", isStart: true},
		{name: "node continuation", position: 2, deviceName: "Node-A", isContinued: true},
		{name: "node b", position: 4, deviceName: "Node-B", isStart: true},
		{name: "switch", position: 6, deviceName: "Leaf-1", isStart: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertRackSlot(t, rackView, testCase.position, testCase.deviceName, testCase.isStart, testCase.isContinued)
		})
	}

	for _, emptyPosition := range []int{3, 5} {
		assertRackSlotEmpty(t, rackView, emptyPosition)
	}
}

// TestRenderRackASCII verifies classic rack rendering matches the exact diagram
// users see when cables are hidden.
//
// Why it matters: this package is responsible for human-friendly output, so
// box-drawing glyphs, U ordering, spacing, and summary text are user-facing API.
// Inputs: a rendered 6U rack with color disabled. Outputs: the complete rack
// diagram string without ANSI escapes or cable rows.
// Data choice: the fixture includes occupied, continued, and empty rows so a
// formatting regression changes the golden-style output.
func TestRenderRackASCII(t *testing.T) {
	_, rackView := sampleRackView(t)

	var output bytes.Buffer
	options := RenderOptions{NoColor: true}

	if err := RenderRackASCII(&output, rackView, options); err != nil {
		t.Fatalf("RenderRackASCII returned error: %v", err)
	}

	assertExactOutput(t, output.String(), expectedRackASCIIOutput(false))
}

// TestRenderRackASCIIRowsHaveStableDisplayWidth verifies every boxed rack row
// renders as an 80-column line when Unicode marker glyphs are present.
//
// Why it matters: rack diagrams use box-drawing and occupancy glyphs, so byte
// length padding can make the borders visibly drift in a terminal.
// Inputs: a no-color classic rack diagram with occupied, continued, and empty
// marker glyphs. Outputs: each boxed row has an 80-rune visible width.
// Data choice: the sample fixture includes multiple non-ASCII markers that would
// expose byte-count padding regressions.
func TestRenderRackASCIIRowsHaveStableDisplayWidth(t *testing.T) {
	_, rackView := sampleRackView(t)

	var output bytes.Buffer
	options := RenderOptions{NoColor: true}

	if err := RenderRackASCII(&output, rackView, options); err != nil {
		t.Fatalf("RenderRackASCII returned error: %v", err)
	}

	for _, line := range strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n") {
		if !strings.HasPrefix(line, "│") && !strings.HasPrefix(line, "├") && !strings.HasPrefix(line, "┌") && !strings.HasPrefix(line, "└") {
			continue
		}
		if got := utf8.RuneCountInString(line); got != terminalWidth {
			t.Fatalf("visible width for %q = %d, want %d", line, got, terminalWidth)
		}
	}
}

// TestRenderRackWithCables verifies classic rack rendering includes the exact
// cable section users see when cable display is enabled.
//
// Why it matters: cable output uses special arrow and bullet glyphs whose layout
// should be protected from accidental formatting changes.
// Inputs: a rendered 6U rack, its inventory, and ShowCables enabled. Outputs: the
// complete rack diagram plus cable count and resolved endpoint labels.
// Data choice: one cable between a node and switch keeps ordering deterministic
// while proving interface UUIDs resolve to device and port names.
func TestRenderRackWithCables(t *testing.T) {
	inventory, rackView := sampleRackView(t)

	var output bytes.Buffer
	options := RenderOptions{
		NoColor:    true,
		ShowCables: true,
		Inventory:  inventory,
	}

	if err := RenderRackASCII(&output, rackView, options); err != nil {
		t.Fatalf("RenderRackASCII returned error: %v", err)
	}

	assertExactOutput(t, output.String(), expectedRackASCIIOutput(true))
}

// TestRenderAllRacks verifies rendering through the package-level rack entry
// point emits the same user-visible classic diagram for discovered racks.
//
// Why it matters: callers usually render all racks from an inventory rather than
// calling RenderRackASCII directly.
// Inputs: an inventory with one rack and NoColor enabled. Outputs: the complete
// no-cable classic rack diagram.
// Data choice: a single rack avoids map-order ambiguity while still covering the
// RenderAllRacksTo discovery and build path.
func TestRenderAllRacks(t *testing.T) {
	inventory := sampleRackInventory()

	var output bytes.Buffer
	options := RenderOptions{NoColor: true}

	if err := RenderAllRacksTo(&output, inventory, options); err != nil {
		t.Fatalf("RenderAllRacksTo returned error: %v", err)
	}

	assertExactOutput(t, output.String(), expectedRackASCIIOutput(false))
}

// TestRenderAllRacksWithFilter verifies matching and non-matching rack filters
// produce the expected visible output.
//
// Why it matters: rack filtering is a user-facing selector, and misses should be
// reported clearly instead of producing an empty diagram.
// Inputs: one inventory with a matching filter and then a non-matching filter.
// Outputs: the matching rack diagram and the explicit no-match message.
// Data choice: mixed-case filter text proves filtering is case-insensitive while
// NonExistent exercises the miss branch.
func TestRenderAllRacksWithFilter(t *testing.T) {
	inventory := sampleRackInventory()

	var output bytes.Buffer
	options := RenderOptions{
		NoColor:    true,
		RackFilter: "rack-006u",
	}

	if err := RenderAllRacksTo(&output, inventory, options); err != nil {
		t.Fatalf("RenderAllRacksTo matching filter returned error: %v", err)
	}
	assertExactOutput(t, output.String(), expectedRackASCIIOutput(false))

	output.Reset()
	options.RackFilter = "NonExistent"
	if err := RenderAllRacksTo(&output, inventory, options); err != nil {
		t.Fatalf("RenderAllRacksTo non-matching filter returned error: %v", err)
	}
	assertExactOutput(t, output.String(), "No racks matching 'NonExistent' found in inventory.\n")
}

// TestEmptyInventory verifies an inventory without racks renders the rackless
// summary path instead of failing.
//
// Why it matters: users can inspect partial inventories before rack data exists,
// and the visual layer should still produce a useful message.
// Inputs: an initialized inventory with no devices or cables. Outputs: zero rack
// discovery results and a rackless summary containing the no-racks title.
// Data choice: devicetypes.NewInventory supplies initialized maps so the test
// covers the empty-valid case rather than nil-map behavior.
func TestEmptyInventory(t *testing.T) {
	inventory := devicetypes.NewInventory()

	racks := FindAllRacks(inventory)
	if len(racks) != 0 {
		t.Fatalf("len(racks) = %d, want 0", len(racks))
	}

	var output bytes.Buffer
	options := RenderOptions{NoColor: true}

	if err := RenderAllRacksTo(&output, inventory, options); err != nil {
		t.Fatalf("RenderAllRacksTo returned error: %v", err)
	}
	if !strings.Contains(output.String(), "Inventory Summary (No Racks Defined)") {
		t.Fatalf("rackless summary missing no-racks title:\n%s", output.String())
	}
}

// TestDeviceWithoutPosition verifies child devices without a rack position are
// preserved in the unpositioned-device list.
//
// Why it matters: a device missing U placement should not disappear from the
// visual model, because users need to notice incomplete inventory data.
// Inputs: one rack with one child device whose RackPosition is zero. Outputs: one
// unpositioned device named Unpositioned-Server.
// Data choice: the child has a valid parent relationship but no U position to
// isolate the unpositioned branch.
func TestDeviceWithoutPosition(t *testing.T) {
	rackID := uuid.MustParse("00000000-0000-0000-0007-000000000001")
	serverID := uuid.MustParse("00000000-0000-0000-0008-000000000001")
	inventory := devicetypes.NewInventory()
	inventory.Racks[rackID] = &devicetypes.CaniRackType{
		ID:      rackID,
		Name:    "Test-Rack",
		UHeight: 6,
		Devices: []uuid.UUID{serverID},
	}
	inventory.Devices[serverID] = &devicetypes.CaniDeviceType{
		ID:     serverID,
		Name:   "Unpositioned-Server",
		Type:   devicetypes.TypeNode,
		Parent: rackID,
	}

	rackView, err := BuildRackVisualization(inventory, rackID)
	if err != nil {
		t.Fatalf("BuildRackVisualization returned error: %v", err)
	}

	if len(rackView.UnpositionedDevices) != 1 {
		t.Fatalf("len(UnpositionedDevices) = %d, want 1", len(rackView.UnpositionedDevices))
	}
	if rackView.UnpositionedDevices[0].Name != "Unpositioned-Server" {
		t.Fatalf("unpositioned device = %q, want %q", rackView.UnpositionedDevices[0].Name, "Unpositioned-Server")
	}
}
