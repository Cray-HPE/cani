package transform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

func fixtureDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "testdata", "fixtures")
}

func loadFixtures(t *testing.T) (*import_.SlsDumpstate, *import_.SmdComponentList) {
	t.Helper()
	slsData, err := os.ReadFile(filepath.Join(fixtureDir(), "csm/simulator/sls.json"))
	if err != nil {
		t.Fatalf("read sls.json: %v", err)
	}
	sls, err := import_.ParseSlsDumpstate(slsData)
	if err != nil {
		t.Fatalf("ParseSlsDumpstate: %v", err)
	}
	smdData, err := os.ReadFile(filepath.Join(fixtureDir(), "csm/simulator/smd.json"))
	if err != nil {
		t.Fatalf("read smd.json: %v", err)
	}
	smd, err := import_.ParseSmdComponents(smdData)
	if err != nil {
		t.Fatalf("ParseSmdComponents: %v", err)
	}
	return sls, smd
}

func TestTransformSls_River(t *testing.T) {
	sls, smd := loadFixtures(t)
	smdMap := buildSmdMap(smd)
	result, err := transformSls(sls, smdMap, nil)
	if err != nil {
		t.Fatalf("transformSls: %v", err)
	}
	if len(result.Devices) == 0 {
		t.Fatal("expected devices")
	}
	if len(result.Racks) == 0 {
		t.Fatal("expected racks")
	}

	// Verify we have at least one cabinet, chassis, blade, and node
	var cabinets, chassis, blades, nodes int
	for _, dev := range result.Devices {
		switch dev.Type {
		case devicetypes.TypeCabinet:
			cabinets++
		case devicetypes.TypeChassis:
			chassis++
		case devicetypes.TypeBlade:
			blades++
		case devicetypes.TypeNode:
			nodes++
		}
	}
	if cabinets == 0 {
		t.Error("expected at least one cabinet device")
	}
	if chassis == 0 {
		t.Error("expected at least one chassis device")
	}
	if blades == 0 {
		t.Error("expected at least one blade device")
	}
	if nodes == 0 {
		t.Error("expected at least one node device")
	}
	t.Logf("River: %d cabinets, %d chassis, %d blades, %d nodes, %d racks",
		cabinets, chassis, blades, nodes, len(result.Racks))
}

func TestTransformSls_Mountain(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(fixtureDir(), "sls/small_mountain.json"))
	if err != nil {
		t.Fatalf("read small_mountain.json: %v", err)
	}
	sls, err := import_.ParseSlsDumpstate(data)
	if err != nil {
		t.Fatalf("ParseSlsDumpstate: %v", err)
	}
	result, err := transformSls(sls, nil, nil)
	if err != nil {
		t.Fatalf("transformSls: %v", err)
	}
	if len(result.Devices) == 0 {
		t.Fatal("expected devices")
	}
	t.Logf("Mountain: %d devices, %d racks", len(result.Devices), len(result.Racks))
}

func TestTransformSls_Empty(t *testing.T) {
	sls := &import_.SlsDumpstate{Hardware: map[string]import_.SlsHardware{}}
	result, err := transformSls(sls, nil, nil)
	if err != nil {
		t.Fatalf("transformSls: %v", err)
	}
	if len(result.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(result.Devices))
	}
}

func TestTransformSls_NodeMetadata(t *testing.T) {
	sls, smd := loadFixtures(t)
	smdMap := buildSmdMap(smd)
	result, err := transformSls(sls, smdMap, nil)
	if err != nil {
		t.Fatalf("transformSls: %v", err)
	}

	// Find a node and check it has CSM metadata
	for _, dev := range result.Devices {
		if dev.Type != devicetypes.TypeNode {
			continue
		}
		csm, ok := dev.ProviderMetadata["csm"].(map[string]any)
		if !ok {
			t.Error("node missing csm provider metadata")
			continue
		}
		if _, ok := csm["xname"]; !ok {
			t.Error("node csm metadata missing xname")
		}
		return // just check one
	}
	t.Error("no nodes found in transform result")
}

func TestTransformSls_ParentChain(t *testing.T) {
	sls, smd := loadFixtures(t)
	smdMap := buildSmdMap(smd)
	result, err := transformSls(sls, smdMap, nil)
	if err != nil {
		t.Fatalf("transformSls: %v", err)
	}

	// Build device lookup
	devByID := result.Devices

	// Every non-cabinet device must have a non-nil Parent
	for id, dev := range devByID {
		if dev.Type == devicetypes.TypeCabinet {
			// Cabinet's parent should be a rack UUID
			if dev.Parent == uuid.Nil {
				t.Errorf("cabinet %q (%s) has nil Parent (should be rack UUID)", dev.Name, id)
			}
			if _, ok := result.Racks[dev.Parent]; !ok {
				t.Errorf("cabinet %q parent %s not found in racks", dev.Name, dev.Parent)
			}
			continue
		}
		if dev.Parent == uuid.Nil {
			t.Errorf("device %q (%s, type=%s) has nil Parent", dev.Name, id, dev.Type)
		}
	}

	// Walk every device up to a rack — verify reachability
	for id, dev := range devByID {
		found := false
		visited := make(map[uuid.UUID]bool)
		cur := dev.Parent
		for cur != uuid.Nil && !visited[cur] {
			visited[cur] = true
			if _, ok := result.Racks[cur]; ok {
				found = true
				break
			}
			parent, ok := devByID[cur]
			if !ok {
				break
			}
			cur = parent.Parent
		}
		if !found {
			t.Errorf("device %q (%s, type=%s) cannot reach a rack via parent chain",
				dev.Name, id, dev.Type)
		}
	}
}

func TestClassifyHardware(t *testing.T) {
	hw := import_.SlsHardware{
		Xname:      "x3000",
		TypeString: "Cabinet",
		Class:      "River",
	}
	cl := classifyHardware(hw, nil)
	if cl.CaniType != devicetypes.TypeCabinet {
		t.Errorf("CaniType = %v, want TypeCabinet", cl.CaniType)
	}
	if cl.Skip {
		t.Error("expected Skip=false for Cabinet")
	}
	if cl.Warning != "" {
		t.Errorf("unexpected warning: %s", cl.Warning)
	}
}

func TestClassifyHardware_Skip(t *testing.T) {
	hw := import_.SlsHardware{
		Xname:      "x3000c0b0",
		TypeString: "ChassisBMC",
		Class:      "River",
	}
	cl := classifyHardware(hw, nil)
	if !cl.Skip {
		t.Error("expected Skip=true for ChassisBMC")
	}
}

func TestClassifyHardware_Warning(t *testing.T) {
	// A MgmtSwitch in a Mountain cabinet should produce a warning.
	hw := import_.SlsHardware{
		Xname:      "x1000c0w1",
		TypeString: "MgmtSwitch",
		Class:      "Mountain",
	}
	cl := classifyHardware(hw, nil)
	if cl.Warning == "" {
		t.Error("expected warning for MgmtSwitch in Mountain cabinet")
	}
	if cl.CaniType != devicetypes.TypeMgmtSwitch {
		t.Errorf("CaniType = %v, want TypeMgmtSwitch", cl.CaniType)
	}
	if cl.Skip {
		t.Error("mismatch should still classify, not skip")
	}
}

func TestClassifyHardware_NoWarningMatchingClass(t *testing.T) {
	// A MgmtSwitch in a River cabinet should produce no warning.
	hw := import_.SlsHardware{
		Xname:      "x3000c0w1",
		TypeString: "MgmtSwitch",
		Class:      "River",
	}
	cl := classifyHardware(hw, nil)
	if cl.Warning != "" {
		t.Errorf("unexpected warning: %s", cl.Warning)
	}
}

func TestClassifyHardware_ClassFromXname(t *testing.T) {
	// When Class is empty, infer from cabinet number range.
	hw := import_.SlsHardware{
		Xname:      "x1000c0r0",
		TypeString: "RouterModule",
		Class:      "", // no explicit class
	}
	cl := classifyHardware(hw, nil)
	// x1000 → Mountain, RouterModule is valid in Mountain → no warning.
	if cl.Warning != "" {
		t.Errorf("unexpected warning: %s", cl.Warning)
	}
	if cl.CaniType != devicetypes.TypeHsnSwitch {
		t.Errorf("CaniType = %v, want TypeHsnSwitch", cl.CaniType)
	}
}

func TestClassifyHardware_CabinetPDU(t *testing.T) {
	hw := import_.SlsHardware{
		Xname:      "x3000m0",
		TypeString: "CabinetPDUController",
		Class:      "River",
	}
	cl := classifyHardware(hw, nil)
	if cl.CaniType != devicetypes.TypeCabinetPDU {
		t.Errorf("CaniType = %v, want TypeCabinetPDU", cl.CaniType)
	}
	if cl.Skip {
		t.Error("expected Skip=false for CabinetPDUController")
	}
	if cl.Warning != "" {
		t.Errorf("unexpected warning for PDU in River: %s", cl.Warning)
	}
}

func TestClassifyHardware_MgmtCDUSwitch(t *testing.T) {
	hw := import_.SlsHardware{
		Xname:      "x1000d0",
		TypeString: "MgmtCDUSwitch",
		Class:      "Mountain",
	}
	cl := classifyHardware(hw, nil)
	if cl.CaniType != devicetypes.TypeCDU {
		t.Errorf("CaniType = %v, want TypeCDU", cl.CaniType)
	}
	if cl.Skip {
		t.Error("expected Skip=false for MgmtCDUSwitch")
	}
	if cl.Warning != "" {
		t.Errorf("unexpected warning for CDU in Mountain: %s", cl.Warning)
	}
}

// ---------- Idempotency ----------

// buildExistingFromResult populates an Inventory from a TransformResult,
// simulating what the merge pipeline does so we can pass it as "existing"
// on a second transform run.
func buildExistingFromResult(result *devicetypes.TransformResult) *devicetypes.Inventory {
	inv := devicetypes.NewInventory()
	for id, dev := range result.Devices {
		inv.Devices[id] = dev
	}
	for id, rack := range result.Racks {
		inv.Racks[id] = rack
	}
	for id, loc := range result.Locations {
		inv.Locations[id] = loc
	}
	for id, mod := range result.Modules {
		inv.Modules[id] = mod
	}
	for id, fru := range result.Frus {
		inv.Frus[id] = fru
	}
	inv.RebuildProviderKeyIndex()
	return inv
}

func TestTransformSls_Idempotent_River(t *testing.T) {
	sls, smd := loadFixtures(t)
	smdMap := buildSmdMap(smd)

	// First pass — no existing.
	first, err := transformSls(sls, smdMap, nil)
	if err != nil {
		t.Fatalf("first pass: %v", err)
	}

	existing := buildExistingFromResult(first)

	// Second pass — existing populated from first run.
	second, err := transformSls(sls, smdMap, existing)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}

	// Device count must match.
	if len(first.Devices) != len(second.Devices) {
		t.Fatalf("device count mismatch: first=%d second=%d",
			len(first.Devices), len(second.Devices))
	}

	// Every device UUID from the first pass must appear in the second.
	for id := range first.Devices {
		if _, ok := second.Devices[id]; !ok {
			t.Errorf("device %s from first pass missing in second", id)
		}
	}

	// Rack UUIDs must be identical.
	if len(first.Racks) != len(second.Racks) {
		t.Fatalf("rack count mismatch: first=%d second=%d",
			len(first.Racks), len(second.Racks))
	}
	for id := range first.Racks {
		if _, ok := second.Racks[id]; !ok {
			t.Errorf("rack %s from first pass missing in second", id)
		}
	}
}

func TestTransformSls_Idempotent_Mountain(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(fixtureDir(), "sls/small_mountain.json"))
	if err != nil {
		t.Fatalf("read small_mountain.json: %v", err)
	}
	sls, err := import_.ParseSlsDumpstate(data)
	if err != nil {
		t.Fatalf("ParseSlsDumpstate: %v", err)
	}

	first, err := transformSls(sls, nil, nil)
	if err != nil {
		t.Fatalf("first pass: %v", err)
	}

	existing := buildExistingFromResult(first)

	second, err := transformSls(sls, nil, existing)
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}

	for id := range first.Devices {
		if _, ok := second.Devices[id]; !ok {
			t.Errorf("device %s from first pass missing in second", id)
		}
	}
	for id := range first.Racks {
		if _, ok := second.Racks[id]; !ok {
			t.Errorf("rack %s from first pass missing in second", id)
		}
	}
}
