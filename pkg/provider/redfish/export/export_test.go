package export

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestExportWritesNodeAsServiceRoot verifies a CANI node device is emitted as a
// Redfish-compatible ServiceRoot JSON object with BMC manager metadata.
//
// Why it matters: Redfish export is the provider's wire-format boundary, so the
// exported JSON must preserve the device identity and management endpoint shape
// consumed by Redfish import/transform workflows.
// Inputs: an inventory with one TypeNode device containing ID, name, model, and
// vendor. Outputs: stdout JSON containing one decoded serviceRoot with exact
// top-level and OEM manager fields.
// Data choice: explicit UUID/name/model/vendor values make every mapped field
// independently assertable.
func TestExportWritesNodeAsServiceRoot(t *testing.T) {
	deviceID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	inventory := devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
		deviceID: {
			ID:     deviceID,
			Name:   "node-01",
			Model:  "ProLiant DL325 Gen11",
			Vendor: "HPE",
			Type:   devicetypes.TypeNode,
		},
	}}

	output, err := captureStdout(t, func() error { return Export(inventory, false) })
	if err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	roots := decodeServiceRoots(t, output)
	if len(roots) != 1 {
		t.Fatalf("decoded roots = %d, want 1", len(roots))
	}
	root := roots[0]
	assertField(t, "OdataID", root.OdataID, "/redfish/v1")
	assertField(t, "OdataType", root.OdataType, "#ServiceRoot.v1_13_0.ServiceRoot")
	assertField(t, "ID", root.ID, "RootService")
	assertField(t, "UUID", root.UUID, deviceID.String())
	assertField(t, "Product", root.Product, "ProLiant DL325 Gen11")
	assertField(t, "Vendor", root.Vendor, "HPE")
	assertField(t, "RedfishVersion", root.RedfishVersion, "1.13.0")
	if root.Oem.Hpe == nil {
		t.Fatal("Oem.Hpe = nil, want manager metadata")
	}
	if len(root.Oem.Hpe.Manager) != 1 {
		t.Fatalf("Manager entries = %d, want 1", len(root.Oem.Hpe.Manager))
	}
	manager := root.Oem.Hpe.Manager[0]
	assertField(t, "Manager.FQDN", manager.FQDN, "node-01-bmc.local")
	assertField(t, "Manager.HostName", manager.HostName, "node-01-bmc")
	assertField(t, "Manager.ManagerType", manager.ManagerType, "iLO 6")
}

// TestExportSkipsNilAndNonNodeDevices verifies export ignores nil device slots
// and devices whose type is not TypeNode.
//
// Why it matters: Redfish export models BMC-discovered server roots only; switch,
// module, or corrupt nil inventory entries must not leak into the JSON payload.
// Inputs: an inventory with one node, one switch, and one nil device pointer.
// Outputs: stdout JSON containing only the node's ServiceRoot.
// Data choice: the skipped switch has a distinct UUID/model so a mistaken export
// would be visible in the decoded result.
func TestExportSkipsNilAndNonNodeDevices(t *testing.T) {
	nodeID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	switchID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002")
	inventory := devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
		nodeID: {
			ID:           nodeID,
			Name:         "node-02",
			Model:        "Node Model",
			Manufacturer: "HPE",
			Type:         devicetypes.TypeNode,
		},
		switchID: {
			ID:     switchID,
			Name:   "switch-01",
			Model:  "Switch Model",
			Vendor: "Aruba",
			Type:   devicetypes.TypeSwitch,
		},
		uuid.Nil: nil,
	}}

	output, err := captureStdout(t, func() error { return Export(inventory, false) })
	if err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	roots := decodeServiceRoots(t, output)

	if len(roots) != 1 {
		t.Fatalf("decoded roots = %d, want 1", len(roots))
	}
	assertField(t, "UUID", roots[0].UUID, nodeID.String())
	assertField(t, "Product", roots[0].Product, "Node Model")
	assertField(t, "Vendor", roots[0].Vendor, "HPE")
}

// TestExportEmptyInventoryWritesNull verifies an inventory without devices emits
// valid JSON null from the nil roots slice.
//
// Why it matters: callers should be able to pipe or parse Redfish export output
// even when there are no node devices to export.
// Inputs: an empty Inventory value. Outputs: stdout containing null and a nil
// decoded root slice.
// Data choice: a zero-value inventory exercises the no-devices path without map
// initialization noise.
func TestExportEmptyInventoryWritesNull(t *testing.T) {
	output, err := captureStdout(t, func() error { return Export(devicetypes.Inventory{}, false) })
	if err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	if strings.TrimSpace(output) != "null" {
		t.Errorf("Export() output = %q, want null", output)
	}
	if roots := decodeServiceRoots(t, output); roots != nil {
		t.Fatalf("decoded roots = %+v, want nil", roots)
	}
}

// TestExportReturnsEncodeError verifies stdout write failures are wrapped with
// Redfish export context.
//
// Why it matters: command callers rely on Export returning output failures so a
// broken pipe or closed descriptor does not look like a successful export.
// Inputs: os.Stdout temporarily replaced with a closed file and an inventory with
// one node. Outputs: a non-nil error containing the Redfish encoding context.
// Data choice: a closed temp file deterministically makes json.Encoder fail
// without external services or timing.
func TestExportReturnsEncodeError(t *testing.T) {
	closedStdout, err := os.CreateTemp(t.TempDir(), "closed-stdout")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	if err := closedStdout.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	originalStdout := os.Stdout
	os.Stdout = closedStdout
	t.Cleanup(func() { os.Stdout = originalStdout })
	inventory := devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
		uuid.New(): {Name: "node-03", Type: devicetypes.TypeNode},
	}}

	err = Export(inventory, false)

	if err == nil {
		t.Fatal("Export() error = nil, want encoding error")
	}
	if !strings.Contains(err.Error(), "encoding Redfish ServiceRoots") {
		t.Errorf("Export() error = %q, want Redfish encoding context", err.Error())
	}
}

// TestExportDryRunWritesNoPayload verifies dry-run mode emits no Redfish JSON to
// stdout, matching the dry-run semantics other providers' exporters honor.
//
// Why it matters: a uniform --dry-run contract lets operators preview an export
// without producing the wire payload, regardless of which provider they target.
// Inputs: an inventory with one node device and dryRun=true. Outputs: empty
// stdout and a nil error.
// Data choice: a single node device would otherwise produce exactly one
// ServiceRoot, so empty stdout proves the payload was suppressed, not merely absent.
func TestExportDryRunWritesNoPayload(t *testing.T) {
	inventory := devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
		uuid.New(): {Name: "node-dry", Type: devicetypes.TypeNode},
	}}

	output, err := captureStdout(t, func() error { return Export(inventory, true) })
	if err != nil {
		t.Fatalf("Export() dry-run error = %v, want nil", err)
	}
	if strings.TrimSpace(output) != "" {
		t.Errorf("Export() dry-run stdout = %q, want empty", output)
	}
}

func captureStdout(t *testing.T, run func() error) (string, error) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	originalStdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = originalStdout
		_ = reader.Close()
	})

	runErr := run()
	if err := writer.Close(); err != nil {
		t.Fatalf("stdout pipe close error = %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("stdout pipe read error = %v", err)
	}
	return string(output), runErr
}

func decodeServiceRoots(t *testing.T, output string) []serviceRoot {
	t.Helper()
	var roots []serviceRoot
	if err := json.Unmarshal([]byte(output), &roots); err != nil {
		t.Fatalf("json.Unmarshal() error = %v for output %q", err, output)
	}
	return roots
}

func assertField(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}
