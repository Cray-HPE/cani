package export

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestExport(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		testName    string
		inv         devicetypes.Inventory
		contains    string
		expectedErr bool
	}{
		{
			testName: "TestExport passing test",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {Name: "server-01", Type: devicetypes.Type("server")},
				},
			},
			contains:    "Summary: 0 locations, 0 racks, 1 devices, 0 modules, 0 cables",
			expectedErr: false,
		},
		{
			testName:    "TestExport failing test",
			inv:         devicetypes.Inventory{},
			contains:    "Summary: 0 locations, 0 racks, 0 devices, 0 modules, 0 cables",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := Export(tt.inv)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()

			if tt.expectedErr && err == nil {
				t.Errorf("Export() expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("Export() unexpected error: %v", err)
				return
			}
			if !strings.Contains(got, tt.contains) {
				t.Errorf("Export() = %q expecting: \n%q\n", tt.contains, got)
			}
		})
	}
}
func TestPrintLocation(t *testing.T) {
	tests := []struct {
		testName string
		location *devicetypes.CaniLocationType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "TestPrintLocation passing test",
			location: &devicetypes.CaniLocationType{
				Name:         "location-01",
				LocationType: "site",
			},
			inv:      &devicetypes.Inventory{},
			expected: "📍 location-01 (site)",
		},
		{
			testName: "TestPrintLocation failing test",
			location: nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printLocation(tt.location, tt.inv, 0)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := strings.TrimSpace(buf.String())

			if got != tt.expected {
				t.Errorf("printLocation() output = %q, want %q", got, tt.expected)
			}
		})
	}

}
func TestPrintRack(t *testing.T) {
	tests := []struct {
		testName string
		rack     *devicetypes.CaniRackType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "TestPrintRack passing test",
			rack: &devicetypes.CaniRackType{
				Name:    "rack-01",
				UHeight: 42,
			},
			inv:      &devicetypes.Inventory{},
			expected: "🗄️  rack-01 [42U]\n  ┌─────────────────────────────────────────────────────────┐\n  └─────────────────────────────────────────────────────────┘",
		},
		{
			testName: "TestPrintRack failing test",
			rack:     nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printRack(tt.rack, tt.inv, 0)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := strings.TrimSpace(buf.String())

			if got != tt.expected {
				t.Errorf("printRack() output = %q, want %q", got, tt.expected)
			}
		})
	}
}
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
			testName: "Passing test",
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
			testName: "Failing test",
			rack:     &devicetypes.CaniRackType{},
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printRackCables(tt.rack, tt.inv, 0)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := strings.TrimSpace(buf.String())

			if got != tt.expected {
				t.Errorf("printRackCables() output = %q, want %q", got, tt.expected)
			}
		})
	}
}
func TestPrintDevice(t *testing.T) {
	tests := []struct {
		testName string
		device   *devicetypes.CaniDeviceType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "Passing test",
			device: &devicetypes.CaniDeviceType{
				Name:  "server-01",
				Type:  devicetypes.Type("server"),
				Model: "ProLiant DL360",
			},
			inv:      &devicetypes.Inventory{},
			expected: "🖥️  server-01 (server) - ProLiant DL360",
		},
		{
			testName: "Failing test",
			device:   nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printDevice(tt.device, tt.inv, 0)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := strings.TrimSpace(buf.String())

			if got != tt.expected {
				t.Errorf("printDevice() output = %q, expected %q", got, tt.expected)
			}
		})
	}
}
func TestPrintModule(t *testing.T) {
	tests := []struct {
		testName string
		module   *devicetypes.CaniModuleType
		inv      *devicetypes.Inventory
		expected string
	}{
		{
			testName: "Passing test",
			module: &devicetypes.CaniModuleType{
				Name:          "gpu-a100",
				ModuleBayName: "bay-0",
				Slug:          "nvidia-a100",
			},
			inv:      &devicetypes.Inventory{},
			expected: "📦 gpu-a100 [bay-0] - nvidia-a100",
		},
		{
			testName: "Failing test",
			module:   nil,
			inv:      &devicetypes.Inventory{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printModule(tt.module, tt.inv, 0)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := strings.TrimSpace(buf.String())

			if got != tt.expected {
				t.Errorf("printModule() output = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestGetDeviceName(t *testing.T) {

	deviceID := uuid.New()

	tests := []struct {
		testName    string
		inv         *devicetypes.Inventory
		deviceID    uuid.UUID
		expectedErr bool
		expected    string
	}{
		{
			testName: "Passing test",
			inv: &devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						Name: "test-device",
					},
				},
			},
			deviceID:    deviceID,
			expectedErr: false,
			expected:    "test-device",
		},
		{
			testName: "Failing test",
			inv: &devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						Name: "test-device",
					},
				},
			},
			deviceID:    deviceID,
			expectedErr: true,
			expected:    "something-else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := getDeviceName(tt.inv, tt.deviceID)

			if tt.expectedErr {

				if got != "test-device" {
					t.Errorf("Expected %q but got %q", "test-device", got)
				}

				return
			}
			if got != tt.expected {
				t.Errorf("getDeviceName() = %q, expected %q", got, tt.expected)
			}
		})
	}
}
