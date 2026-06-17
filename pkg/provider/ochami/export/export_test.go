package export

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type failingWriter struct{}

func (f failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

// TestExportMatchesOpenCHAMIFileFormat verifies Export emits the OpenCHAMI
// nodes.yaml root shape with bmcs and nodes arrays of xname/mac/ip entries.
//
// Why it matters: ex-bootstrap consumes this minimal FileFormat/Entry contract;
// extra or missing fields would make the export drift from the target inventory
// format.
// Inputs: an inventory with one node and one BMC carrying ochami xname, mac, and
// ip metadata. Outputs: decoded YAML with one BMC entry, one node entry, and no
// keys outside xname/mac/ip.
// Data choice: distinct BMC and node values prove each bucket is populated from
// the right device type.
func TestExportMatchesOpenCHAMIFileFormat(t *testing.T) {
	inv := devicetypes.NewInventory()
	nodeID := uuid.New()
	bmcID := uuid.New()
	inv.Devices[nodeID] = deviceWithMeta(nodeID, "node", "node-fallback", map[string]any{
		"xname": "x3000c0s1b0n0",
		"mac":   "aa:bb:cc:dd:ee:01",
		"ip":    "10.1.0.1",
	})
	inv.Devices[bmcID] = deviceWithMeta(bmcID, "bmc", "bmc-fallback", map[string]any{
		"xname": "x3000c0s1b0",
		"mac":   "aa:bb:cc:dd:ee:ff",
		"ip":    "10.1.0.100",
	})

	output := exportToString(t, *inv)
	payload := decodePayload(t, output)

	assertEntries(t, "bmcs", payload.BMCs, []openChamiEntry{{
		Xname: "x3000c0s1b0",
		MAC:   "aa:bb:cc:dd:ee:ff",
		IP:    "10.1.0.100",
	}})
	assertEntries(t, "nodes", payload.Nodes, []openChamiEntry{{
		Xname: "x3000c0s1b0n0",
		MAC:   "aa:bb:cc:dd:ee:01",
		IP:    "10.1.0.1",
	}})
	assertOnlyOpenCHAMIEntryKeys(t, output)
}

// TestExportUsesCsmFallbackAndBootMACAsMAC verifies CSM metadata can populate
// the minimal OpenCHAMI entry fields when ochami metadata is absent.
//
// Why it matters: inventories transformed from CSM often store node NIC identity
// as boot_mac, but OpenCHAMI expects the output field to be mac.
// Inputs: a node with CSM xname, ip, boot_mac, nid, and aliases metadata.
// Outputs: one node entry whose mac equals boot_mac and whose YAML omits nid,
// hostname, and host_aliases.
// Data choice: including legacy extra CSM keys proves they are ignored rather
// than serialized into the OpenCHAMI file.
func TestExportUsesCsmFallbackAndBootMACAsMAC(t *testing.T) {
	inv := devicetypes.NewInventory()
	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:   nodeID,
		Name: "node-fallback",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"csm": map[string]any{
				"xname":    "x9000c1s0b0n0",
				"ip":       "10.9.0.1",
				"boot_mac": "aa:bb:cc:dd:ee:09",
				"nid":      1000,
				"aliases":  []string{"nid001000"},
			},
		}},
	}

	output := exportToString(t, *inv)
	payload := decodePayload(t, output)

	assertEntries(t, "nodes", payload.Nodes, []openChamiEntry{{
		Xname: "x9000c1s0b0n0",
		MAC:   "aa:bb:cc:dd:ee:09",
		IP:    "10.9.0.1",
	}})
	assertOnlyOpenCHAMIEntryKeys(t, output)
}

// TestExportOchamiOverridesCsm verifies ochami metadata wins over CSM fallback
// values for the minimal entry fields.
//
// Why it matters: provider-specific ochami metadata represents the export target
// contract and must not be overwritten by legacy CSM values when both exist.
// Inputs: one node with conflicting csm and ochami xname, mac/boot_mac, and ip
// values. Outputs: one node entry containing the ochami values only.
// Data choice: every exported field differs between providers, so any precedence
// regression is visible.
func TestExportOchamiOverridesCsm(t *testing.T) {
	inv := devicetypes.NewInventory()
	nodeID := uuid.New()
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{
		ID:   nodeID,
		Name: "fallback",
		Type: devicetypes.Type("node"),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"csm": map[string]any{
				"xname":    "x9000c1s0b0n0",
				"ip":       "10.9.0.1",
				"boot_mac": "aa:bb:cc:dd:ee:09",
			},
			"ochami": map[string]any{
				"xname": "x9000c1s0b0n0-override",
				"ip":    "10.9.0.2",
				"mac":   "aa:bb:cc:dd:ee:10",
			},
		}},
	}

	payload := decodePayload(t, exportToString(t, *inv))

	assertEntries(t, "nodes", payload.Nodes, []openChamiEntry{{
		Xname: "x9000c1s0b0n0-override",
		MAC:   "aa:bb:cc:dd:ee:10",
		IP:    "10.9.0.2",
	}})
}

// TestExportFallbackToNameAndEmptyInventory verifies missing xname values fall
// back to the CANI device name and empty inventories emit empty bmcs/nodes lists.
//
// Why it matters: an export should remain valid when metadata is partial or when
// the inventory has no exportable devices.
// Inputs: a node with only ip/mac metadata and a separate empty inventory.
// Outputs: the node entry uses Name as xname, and the empty inventory decodes to
// zero BMC and node entries.
// Data choice: omitting xname isolates the fallback path, while a zero-valued
// inventory proves the root YAML shape is still present.
func TestExportFallbackToNameAndEmptyInventory(t *testing.T) {
	inv := devicetypes.NewInventory()
	nodeID := uuid.New()
	inv.Devices[nodeID] = deviceWithMeta(nodeID, "node", "my-node-name", map[string]any{
		"ip":  "10.0.0.1",
		"mac": "aa:bb:cc:dd:ee:11",
	})

	payload := decodePayload(t, exportToString(t, *inv))
	assertEntries(t, "nodes", payload.Nodes, []openChamiEntry{{
		Xname: "my-node-name",
		MAC:   "aa:bb:cc:dd:ee:11",
		IP:    "10.0.0.1",
	}})

	emptyPayload := decodePayload(t, exportToString(t, *devicetypes.NewInventory()))
	if len(emptyPayload.BMCs) != 0 || len(emptyPayload.Nodes) != 0 {
		t.Fatalf("empty payload = %+v, want no bmcs or nodes", emptyPayload)
	}
}

// TestExportSortingAndFiltering verifies BMC and node entries are sorted by
// xname, nil devices are ignored, and unsupported device types are skipped.
//
// Why it matters: deterministic ordering keeps generated inventory diffs stable,
// and only BMC/node records belong in the OpenCHAMI file while nil map entries
// should not panic.
// Inputs: unsorted node and BMC devices plus one nil entry and one chassis
// device. Outputs: sorted node and BMC entries with the nil/chassis entries
// absent.
// Data choice: three values per bucket prove ordering beyond a single comparison;
// the nil and chassis entries exercise both skip branches.
func TestExportSortingAndFiltering(t *testing.T) {
	inv := devicetypes.NewInventory()
	inv.Devices[uuid.Nil] = nil
	for _, xname := range []string{"x3000c0s3b0n0", "x3000c0s1b0n0", "x3000c0s2b0n0"} {
		id := uuid.New()
		inv.Devices[id] = deviceWithMeta(id, "node", "", map[string]any{"xname": xname})
	}
	for _, xname := range []string{"x3000c0s3b0", "x3000c0s1b0", "x3000c0s2b0"} {
		id := uuid.New()
		inv.Devices[id] = deviceWithMeta(id, "bmc", "", map[string]any{"xname": xname})
	}
	chassisID := uuid.New()
	inv.Devices[chassisID] = deviceWithMeta(chassisID, "chassis", "chassis-1", map[string]any{"xname": "x3000c0"})

	payload := decodePayload(t, exportToString(t, *inv))

	wantNodes := []openChamiEntry{{Xname: "x3000c0s1b0n0"}, {Xname: "x3000c0s2b0n0"}, {Xname: "x3000c0s3b0n0"}}
	wantBMCs := []openChamiEntry{{Xname: "x3000c0s1b0"}, {Xname: "x3000c0s2b0"}, {Xname: "x3000c0s3b0"}}
	assertEntries(t, "nodes", payload.Nodes, wantNodes)
	assertEntries(t, "bmcs", payload.BMCs, wantBMCs)
}

// TestExportReturnsWriteError verifies writer failures are returned to the
// caller.
//
// Why it matters: command callers need export write failures to fail the command
// instead of silently dropping generated YAML.
// Inputs: an empty inventory and a writer that always returns an error. Outputs:
// the write error from Export.
// Data choice: an empty inventory still marshals valid YAML, isolating the writer
// failure from payload construction.
func TestExportReturnsWriteError(t *testing.T) {
	if err := Export(*devicetypes.NewInventory(), failingWriter{}); err == nil {
		t.Fatal("Export() error = nil, want write error")
	}
}

func deviceWithMeta(id uuid.UUID, deviceType, name string, meta map[string]any) *devicetypes.CaniDeviceType {
	return &devicetypes.CaniDeviceType{
		ID:   id,
		Name: name,
		Type: devicetypes.Type(deviceType),
		ObjectMeta: devicetypes.ObjectMeta{ProviderMetadata: map[string]any{
			"ochami": meta,
		}},
	}
}

func exportToString(t *testing.T, inv devicetypes.Inventory) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Export(inv, &buf); err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	return buf.String()
}

func decodePayload(t *testing.T, output string) openChamiPayload {
	t.Helper()
	var payload openChamiPayload
	if err := yaml.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v for output:\n%s", err, output)
	}
	return payload
}

func assertEntries(t *testing.T, label string, got, want []openChamiEntry) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", label, got, want)
	}
}

func assertOnlyOpenCHAMIEntryKeys(t *testing.T, output string) {
	t.Helper()
	var raw map[string][]map[string]any
	if err := yaml.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("yaml.Unmarshal(raw) error = %v for output:\n%s", err, output)
	}
	wantKeys := map[string]bool{"xname": true, "mac": true, "ip": true}
	for bucket, entries := range raw {
		for index, entry := range entries {
			for key := range entry {
				if !wantKeys[key] {
					t.Fatalf("%s[%d] has unexpected key %q in output:\n%s", bucket, index, key, output)
				}
			}
		}
	}
}
