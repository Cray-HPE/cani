package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestExportFullMetadata(t *testing.T) {
	inv := devicetypes.NewInventory()

	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:           nodeID,
		Name:         "node-fallback",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"ochami": map[string]any{
				"xname":        "x3000c0s1b0n0",
				"ip":           "10.1.0.1",
				"boot_mac":     "aa:bb:cc:dd:ee:01",
				"nid":          42,
				"hostname":     "nid000042",
				"host_aliases": []string{"compute-1", "worker-1"},
			},
		}},
	}

	bmcID := uuid.New()
	inv.Devices[bmcID] = &devicetypes.CaniDeviceType{
		ID:           bmcID,
		Name:         "bmc-fallback",
		Type: devicetypes.Type("bmc"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"ochami": map[string]any{
				"xname": "x3000c0s1b0",
				"ip":    "10.1.0.100",
				"mac":   "aa:bb:cc:dd:ee:ff",
			},
		}},
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()

	for _, want := range []string{
		"xname: x3000c0s1b0n0",
		"ip: 10.1.0.1",
		"boot_mac: aa:bb:cc:dd:ee:01",
		"nid: 42",
		"hostname: nid000042",
		"- compute-1",
		"- worker-1",
		"xname: x3000c0s1b0",
		"ip: 10.1.0.100",
		"mac: aa:bb:cc:dd:ee:ff",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestExportPartialMetadata(t *testing.T) {
	inv := devicetypes.NewInventory()

	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:           nodeID,
		Name:         "node-1",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"ochami": map[string]any{
				"xname": "x3000c0s2b0n0",
			},
		}},
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "xname: x3000c0s2b0n0") {
		t.Errorf("output missing xname\ngot:\n%s", out)
	}
	if strings.Contains(out, "nid:") {
		t.Errorf("output should omit nid when nil\ngot:\n%s", out)
	}
}

func TestExportFallbackToName(t *testing.T) {
	inv := devicetypes.NewInventory()

	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:           nodeID,
		Name:         "my-node-name",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"ochami": map[string]any{
				"ip": "10.0.0.1",
			},
		}},
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "xname: my-node-name") {
		t.Errorf("expected fallback to device Name\ngot:\n%s", out)
	}
}

func TestExportEmptyInventory(t *testing.T) {
	inv := devicetypes.NewInventory()

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "bmcs: []") {
		t.Errorf("expected empty bmcs list\ngot:\n%s", out)
	}
	if !strings.Contains(out, "nodes: []") {
		t.Errorf("expected empty nodes list\ngot:\n%s", out)
	}
}

func TestExportSorting(t *testing.T) {
	inv := devicetypes.NewInventory()

	for _, xname := range []string{"x3000c0s3b0n0", "x3000c0s1b0n0", "x3000c0s2b0n0"} {
		id := uuid.New()
		inv.Devices[id] = &devicetypes.CaniDeviceType{
			ID:           id,
			Type: devicetypes.Type("node"),
			ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
				"ochami": map[string]any{"xname": xname},
			}},
		}
	}

	for _, xname := range []string{"x3000c0s3b0", "x3000c0s1b0", "x3000c0s2b0"} {
		id := uuid.New()
		inv.Devices[id] = &devicetypes.CaniDeviceType{
			ID:           id,
			Type: devicetypes.Type("bmc"),
			ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
				"ochami": map[string]any{"xname": xname},
			}},
		}
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()

	idx1 := strings.Index(out, "x3000c0s1b0n0")
	idx2 := strings.Index(out, "x3000c0s2b0n0")
	idx3 := strings.Index(out, "x3000c0s3b0n0")
	if idx1 >= idx2 || idx2 >= idx3 {
		t.Errorf("nodes not sorted by xname\ngot:\n%s", out)
	}
}

func TestExportCsmFallback(t *testing.T) {
	inv := devicetypes.NewInventory()

	// Node with only CSM metadata — no ochami sub-map.
	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:           nodeID,
		Name:         "nid001000",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"csm": map[string]any{
				"xname":   "x9000c1s0b0n0",
				"nid":     1000,
				"role":    "Compute",
				"aliases": []string{"nid001000"},
			},
		}},
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()

	for _, want := range []string{
		"xname: x9000c1s0b0n0",
		"nid: 1000",
		"hostname: nid001000",
		"- nid001000",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestExportOchamiOverridesCsm(t *testing.T) {
	inv := devicetypes.NewInventory()

	// Node with both CSM and ochami metadata — ochami should win.
	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:           nodeID,
		Name:         "fallback",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"csm": map[string]any{
				"xname":   "x9000c1s0b0n0",
				"nid":     1000,
				"aliases": []string{"nid001000"},
			},
			"ochami": map[string]any{
				"xname":        "x9000c1s0b0n0",
				"nid":          9999,
				"hostname":     "custom-host",
				"host_aliases": []string{"custom-alias"},
			},
		}},
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()

	// Ochami values should win
	if !strings.Contains(out, "nid: 9999") {
		t.Errorf("expected ochami nid to override csm\ngot:\n%s", out)
	}
	if !strings.Contains(out, "hostname: custom-host") {
		t.Errorf("expected ochami hostname to override csm\ngot:\n%s", out)
	}
	if !strings.Contains(out, "- custom-alias") {
		t.Errorf("expected ochami host_aliases to override csm\ngot:\n%s", out)
	}
	// CSM aliases should NOT appear
	if strings.Contains(out, "nid001000") {
		t.Errorf("csm aliases should be overridden by ochami\ngot:\n%s", out)
	}
}

func TestExportSkipsNonNodeNonBMC(t *testing.T) {
	inv := devicetypes.NewInventory()

	chassisID := uuid.New()
	inv.Devices[chassisID] = &devicetypes.CaniDeviceType{
		ID:           chassisID,
		Name:         "chassis-1",
		Type: devicetypes.Type("chassis"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"ochami": map[string]any{"xname": "x3000c0"},
		}},
	}

	var buf bytes.Buffer
	if err := Export(*inv, &buf); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "x3000c0") {
		t.Errorf("chassis should not appear in output\ngot:\n%s", out)
	}
}
