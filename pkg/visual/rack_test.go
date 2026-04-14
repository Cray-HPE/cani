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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// getTestFixturePath returns the path to test fixtures relative to the test file
func getTestFixturePath(t *testing.T, filename string) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	return filepath.Join(cwd, "..", "..", "testdata", "fixtures", "cani", filename)
}

// loadInventoryFromFile loads a test inventory from YAML
func loadInventoryFromFile(t *testing.T, filePath string) *devicetypes.Inventory {
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test fixture %s: %v", filePath, err)
	}

	// Parse into intermediate struct with string keys
	type inventoryYAML struct {
		Devices map[string]*devicetypes.CaniDeviceType `yaml:"devices"`
		Cables  map[string]*devicetypes.CaniCableType  `yaml:"cables,omitempty"`
	}

	rawInv := &inventoryYAML{}
	if err := yaml.Unmarshal(data, rawInv); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Cables:  make(map[uuid.UUID]*devicetypes.CaniCableType),
	}

	for idStr, device := range rawInv.Devices {
		id, err := uuid.Parse(idStr)
		if err != nil {
			t.Fatalf("Invalid device UUID %q: %v", idStr, err)
		}
		if device != nil {
			device.ID = id
		}
		inv.Devices[id] = device
	}

	for idStr, cable := range rawInv.Cables {
		id, err := uuid.Parse(idStr)
		if err != nil {
			t.Fatalf("Invalid cable UUID %q: %v", idStr, err)
		}
		if cable != nil {
			cable.ID = id
		}
		inv.Cables[id] = cable
	}

	return inv
}

func TestLoadSampleInventory(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)

	// Verify correct number of devices loaded
	if len(inv.Devices) != 11 {
		t.Errorf("Expected 11 devices (1 rack + 9 servers + 1 switch), got %d", len(inv.Devices))
	}

	// Verify correct number of cables loaded
	if len(inv.Cables) != 9 {
		t.Errorf("Expected 9 cables, got %d", len(inv.Cables))
	}
}

func TestFindAllRacks(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)
	racks := FindAllRacks(inv)

	if len(racks) != 1 {
		t.Errorf("Expected 1 rack, got %d", len(racks))
	}

	if len(racks) > 0 && racks[0].Name != "Rack-001" {
		t.Errorf("Expected rack name 'Rack-001', got '%s'", racks[0].Name)
	}
}

func TestBuildRackVisualization(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)
	racks := FindAllRacks(inv)

	if len(racks) == 0 {
		t.Fatal("No racks found in inventory")
	}

	rackView, err := BuildRackVisualization(inv, racks[0].ID)
	if err != nil {
		t.Fatalf("Failed to build rack visualization: %v", err)
	}

	// Verify rack properties
	if rackView.Rack.Name != "Rack-001" {
		t.Errorf("Expected rack name 'Rack-001', got '%s'", rackView.Rack.Name)
	}

	if rackView.Height != 48 {
		t.Errorf("Expected rack height 48U, got %d", rackView.Height)
	}

	// Verify device count (9 servers + 1 switch = 10 devices)
	deviceCount := 0
	for _, slot := range rackView.Slots {
		if slot.Device != nil {
			deviceCount++
		}
	}
	// Each 2U server occupies 2 slots, so we should count unique devices
	uniqueDevices := make(map[uuid.UUID]bool)
	for _, slot := range rackView.Slots {
		if slot.Device != nil {
			uniqueDevices[slot.Device.ID] = true
		}
	}
	if len(uniqueDevices) != 10 {
		t.Errorf("Expected 10 unique devices in rack, got %d", len(uniqueDevices))
	}
}

func TestRackPositions(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)
	racks := FindAllRacks(inv)

	if len(racks) == 0 {
		t.Fatal("No racks found in inventory")
	}

	rackView, err := BuildRackVisualization(inv, racks[0].ID)
	if err != nil {
		t.Fatalf("Failed to build rack visualization: %v", err)
	}

	// Check expected positions
	expectedPositions := map[string]int{
		"Server-001": 1,
		"Server-002": 3,
		"Server-003": 5,
		"Server-004": 7,
		"Server-005": 9,
		"Server-006": 11,
		"Server-007": 13,
		"Server-008": 15,
		"Server-009": 17,
		"Switch-001": 47,
	}

	for name, expectedPos := range expectedPositions {
		slot := rackView.Slots[expectedPos]
		if slot.Device == nil {
			t.Errorf("Expected device '%s' at U%d, but slot is empty", name, expectedPos)
			continue
		}
		if slot.Device.Name != name {
			t.Errorf("Expected device '%s' at U%d, got '%s'", name, expectedPos, slot.Device.Name)
		}
	}
}

func TestRenderRackASCII(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)
	racks := FindAllRacks(inv)

	if len(racks) == 0 {
		t.Fatal("No racks found in inventory")
	}

	rackView, err := BuildRackVisualization(inv, racks[0].ID)
	if err != nil {
		t.Fatalf("Failed to build rack visualization: %v", err)
	}

	var buf bytes.Buffer
	opts := RenderOptions{NoColor: true}

	if err := RenderRackASCII(&buf, rackView, opts); err != nil {
		t.Fatalf("Failed to render rack ASCII: %v", err)
	}

	output := buf.String()

	// Verify output contains expected elements
	expectedStrings := []string{
		"Rack-001",
		"Server-001",
		"Server-009",
		"Switch-001",
		"U48",
		"U1",
		"[EMPTY]",
		"Summary:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestRenderRackWithCables(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)
	racks := FindAllRacks(inv)

	if len(racks) == 0 {
		t.Fatal("No racks found in inventory")
	}

	rackView, err := BuildRackVisualization(inv, racks[0].ID)
	if err != nil {
		t.Fatalf("Failed to build rack visualization: %v", err)
	}

	var buf bytes.Buffer
	opts := RenderOptions{
		NoColor:    true,
		ShowCables: true,
		Inventory:  inv,
	}

	if err := RenderRackASCII(&buf, rackView, opts); err != nil {
		t.Fatalf("Failed to render rack ASCII: %v", err)
	}

	output := buf.String()

	// Verify cable section is present
	if !strings.Contains(output, "Cable Connections") {
		t.Error("Expected output to contain 'Cable Connections'")
	}

	// Verify cable count
	if !strings.Contains(output, "9 cables") {
		t.Error("Expected output to contain '9 cables'")
	}

	// Verify some cable connections are shown
	expectedCableStrings := []string{
		"Server-001",
		"Switch-001",
		"Gig-E 1",
		"←→",
	}

	for _, expected := range expectedCableStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected cable output to contain '%s'", expected)
		}
	}
}

func TestRenderAllRacks(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)

	var buf bytes.Buffer
	opts := RenderOptions{NoColor: true}

	if err := RenderAllRacksTo(&buf, inv, opts); err != nil {
		t.Fatalf("Failed to render all racks: %v", err)
	}

	output := buf.String()

	// Verify output is not empty
	if len(output) == 0 {
		t.Error("Expected non-empty output")
	}

	// Verify rack is present
	if !strings.Contains(output, "Rack-001") {
		t.Error("Expected output to contain 'Rack-001'")
	}
}

func TestRenderAllRacksWithFilter(t *testing.T) {
	testFile := getTestFixturePath(t, "sample-rack-inventory.yaml")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test fixture not found: %s", testFile)
	}

	inv := loadInventoryFromFile(t, testFile)

	// Test with matching filter
	var buf bytes.Buffer
	opts := RenderOptions{
		NoColor:    true,
		RackFilter: "Rack-001",
	}

	if err := RenderAllRacksTo(&buf, inv, opts); err != nil {
		t.Fatalf("Failed to render filtered racks: %v", err)
	}

	if !strings.Contains(buf.String(), "Rack-001") {
		t.Error("Expected matching rack to be rendered")
	}

	// Test with non-matching filter
	buf.Reset()
	opts.RackFilter = "NonExistent"

	if err := RenderAllRacksTo(&buf, inv, opts); err != nil {
		t.Fatalf("Failed to render with non-matching filter: %v", err)
	}

	if !strings.Contains(buf.String(), "No racks matching") {
		t.Error("Expected 'No racks matching' message for non-matching filter")
	}
}

func TestEmptyInventory(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
		Cables:  make(map[uuid.UUID]*devicetypes.CaniCableType),
	}

	racks := FindAllRacks(inv)
	if len(racks) != 0 {
		t.Errorf("Expected 0 racks in empty inventory, got %d", len(racks))
	}

	var buf bytes.Buffer
	opts := RenderOptions{NoColor: true}

	if err := RenderAllRacksTo(&buf, inv, opts); err != nil {
		t.Fatalf("Failed to render empty inventory: %v", err)
	}

	if !strings.Contains(buf.String(), "No Racks Defined") {
		t.Error("Expected 'No Racks Defined' message for empty inventory")
	}
}

func TestDeviceWithoutPosition(t *testing.T) {
	rackID := uuid.MustParse("00000000-0000-0000-0001-000000000001")
	serverID := uuid.MustParse("00000000-0000-0000-0002-000000000001")

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			rackID: {
				ID:       rackID,
				Name:     "Test-Rack",
				Type:     devicetypes.Type("rack"),
				Children: []uuid.UUID{serverID},
			},
			serverID: {
				ID:     serverID,
				Name:   "Unpositioned-Server",
				Type:   devicetypes.Type("server"),
				Parent: rackID,
				// No RackPosition set
			},
		},
	}

	rackView, err := BuildRackVisualization(inv, rackID)
	if err != nil {
		t.Fatalf("Failed to build rack visualization: %v", err)
	}

	// Device should be in unpositioned list
	if len(rackView.UnpositionedDevices) != 1 {
		t.Errorf("Expected 1 unpositioned device, got %d", len(rackView.UnpositionedDevices))
	}

	if len(rackView.UnpositionedDevices) > 0 && rackView.UnpositionedDevices[0].Name != "Unpositioned-Server" {
		t.Errorf("Expected unpositioned device 'Unpositioned-Server', got '%s'", rackView.UnpositionedDevices[0].Name)
	}
}
