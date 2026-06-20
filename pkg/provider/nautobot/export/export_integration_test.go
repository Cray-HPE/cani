package export

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

const (
	testNautobotURL   = "http://localhost:8081/api"
	testNautobotToken = "0123456789abcdef0123456789abcdef01234567"
)

// loadTypesOnce ensures the community device type library is loaded exactly
// once per test run so that GetBySlug / CreateDeviceTypeFromLocal can resolve
// the real slugs used in the fixture.
var loadTypesOnce sync.Once

func ensureTypesLoaded(t *testing.T) {
	t.Helper()
	loadTypesOnce.Do(func() {
		if err := devicetypes.LoadAll(nil, []string{config.DefaultTypesRepo}, true, false); err != nil {
			t.Fatalf("load device type library: %v", err)
		}
	})
}

// skipUnlessNautobot skips the test when Nautobot is unreachable.
func skipUnlessNautobot(t *testing.T) {
	t.Helper()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, testNautobotURL+"/status/", nil)
	req.Header.Set("Authorization", "Token "+testNautobotToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("Nautobot not reachable: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("Nautobot returned %d", resp.StatusCode)
	}
}

// fixtureDir returns the path to testdata/fixtures/cani.
func fixtureDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "testdata", "fixtures", "cani")
}

// loadSpineLeafInventory reads and unmarshals the spine-leaf fixture.
func loadSpineLeafInventory(t *testing.T) *devicetypes.Inventory {
	t.Helper()
	path := filepath.Join(fixtureDir(), "spine_leaf_inventory.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var inv devicetypes.Inventory
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	// Reverse indices and derived FKs (device.Rack, rack.Devices) are not
	// serialized; rebuild them from the forward FKs as production Load does so
	// the export pipeline sees device->rack placement.
	inv.RebuildDerivedState()
	return &inv
}

// newTestProvider creates a Nautobot provider wired to the local instance,
// bypassing cli/config config loading.
func newTestProvider(t *testing.T) *Exporter {
	t.Helper()
	client, err := NewNautobotClient(testNautobotURL, testNautobotToken)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	ctx := context.Background()
	cache := NewLookupCache(client)
	cache.SetContext(ctx)
	cache.SetCreateDeviceTypes(true)
	cache.SetCreateLocationTypes(true)
	cache.SetCreateLocations(true)
	cache.SetCreateStatuses(true)
	cache.SetCreateRoles(true)

	return &Exporter{
		Client: client,
		Cache:  cache,
		Options: &ExporterOpts{
			DefaultLocation:     "HPC-DataCenter",
			DefaultRole:         "Generic",
			DefaultStatus:       "Active",
			CreateDeviceTypes:   true,
			CreateLocationTypes: true,
			CreateModuleTypes:   true,
			CreateLocations:     true,
			CreateStatuses:      true,
			CreateRoles:         true,
			Merge:               false,
			DryRun:              false,
		},
	}
}

// TestExportSpineLeafDryRun verifies a dry-run export of the spine-leaf fixture
// resolves device types and maps locations, racks, and devices, asserting 12
// devices and 2 racks created.
//
// Why it matters: this end-to-end run drives the real phase pipeline and the
// community device-type library against a (skippable) live Nautobot, catching
// wiring regressions that fake-server unit tests miss; dry-run validates the
// read/mapping path without mutating Nautobot.
// Inputs: spine_leaf_inventory.json with DryRun=true. Outputs: populated
// LoadResult counts; phases 3-6 need real device IDs so they are only logged.
// Data choice: the spine-leaf fixture is a small but complete fabric whose exact
// 12-device, 2-rack shape makes the count assertions meaningful.
func TestExportSpineLeafDryRun(t *testing.T) {
	skipUnlessNautobot(t)
	ensureTypesLoaded(t)
	inv := loadSpineLeafInventory(t)

	e := newTestProvider(t)
	e.Options.DryRun = true

	// Replicate the Load() pipeline, bypassing setupExportFromConfig.
	ctx := context.Background()
	e.Cache.SetContext(ctx)

	result := &LoadResult{
		Created:          make([]string, 0),
		Updated:          make([]string, 0),
		Skipped:          make([]string, 0),
		Errors:           make([]string, 0),
		Conflicts:        make([]ConflictInfo, 0),
		LocationsCreated: make([]string, 0),
		LocationsSkipped: make([]string, 0),
		RacksCreated:     make([]string, 0),
	}

	mapper := NewDeviceMapper(e.Cache, &MapperOpts{
		DefaultLocation: e.Options.DefaultLocation,
		DefaultRole:     e.Options.DefaultRole,
		DefaultStatus:   e.Options.DefaultStatus,
	})
	mapper.SetInventory(inv)

	// ---- Phase 0: Locations (errors are failures) ----
	var phase0Errs int
	createdLocationIDs, err := e.loadLocations(ctx, inv, result)
	if err != nil {
		t.Errorf("phase 0 (locations): %v", err)
		phase0Errs++
	}
	_ = createdLocationIDs

	// ---- Phase 1: Racks (errors are failures) ----
	var phase1Errs int
	createdRackIDs := make(map[uuid.UUID]uuid.UUID)
	for rackID, rack := range inv.Racks {
		if rack == nil || rack.Name == "" {
			continue
		}
		existing, err := e.Cache.GetRackByName(rack.Name)
		if err != nil {
			t.Errorf("phase 1 rack %s: lookup error: %v", rack.Name, err)
			phase1Errs++
			continue
		}
		if existing != nil {
			createdRackIDs[rackID] = existing.ID
			continue
		}
		nautobotRackID, err := e.createRackFromCaniRack(ctx, rack, inv, mapper, result)
		if err != nil {
			t.Errorf("phase 1 rack %s: create error: %v", rack.Name, err)
			phase1Errs++
			continue
		}
		createdRackIDs[rackID] = nautobotRackID
	}

	// ---- Phase 2: Devices (errors are failures) ----
	var phase2Errs int
	createdDeviceIDs := make(map[string]uuid.UUID)
	for _, device := range inv.Devices {
		if device == nil || device.Name == "" {
			continue
		}
		category := devicetypes.ClassifyForNautobot(string(device.Type))
		if category != devicetypes.CategoryDevice {
			continue
		}
		existing, err := e.Cache.GetDeviceByName(device.Name)
		if err != nil {
			t.Errorf("phase 2 device %s: lookup error: %v", device.Name, err)
			phase2Errs++
			continue
		}
		if existing != nil {
			createdDeviceIDs[device.Name] = existing.ID
			continue
		}
		nautobotID, err := e.createDeviceWithID(ctx, device, mapper, result)
		if err != nil {
			t.Errorf("phase 2 device %s: create error: %v", device.Name, err)
			phase2Errs++
			continue
		}
		if nautobotID != uuid.Nil {
			createdDeviceIDs[device.Name] = nautobotID
		}
	}

	// ---- Phases 3-6: require real device IDs ----
	// In dry-run mode the devices are not actually created in Nautobot, so
	// interface / module / FRU / cable lookups will fail. We log these as
	// informational and only count them for the summary.
	var downstreamErrs int

	// Phase 3: Interfaces (bulk)
	if err := e.loadInterfaces(ctx, inv, createdDeviceIDs, result); err != nil {
		t.Logf("phase 3 iface: %v", err)
		downstreamErrs++
	}

	// Phase 4: Modules
	if err := e.loadModules(ctx, inv, createdDeviceIDs, result); err != nil {
		t.Logf("phase 4 modules: %v", err)
		downstreamErrs++
	}

	// Phase 5: FRUs
	if err := e.loadFrus(ctx, inv, createdDeviceIDs, result); err != nil {
		t.Logf("phase 5 frus: %v", err)
		downstreamErrs++
	}

	// Phase 6: Cables
	for cableID, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if err := e.createCaniCableType(ctx, cableID, cable, inv, createdDeviceIDs, result); err != nil {
			t.Logf("phase 6 cable %s: %v", cable.Label, err)
			downstreamErrs++
		}
	}

	// Summary
	e.printLoadSummary(result)

	t.Logf("Locations created/skipped: %v / %v", result.LocationsCreated, result.LocationsSkipped)
	t.Logf("Racks created: %v", result.RacksCreated)
	t.Logf("Devices created: %v", result.Created)
	t.Logf("Interfaces created: %d", result.IfacesCreated)
	t.Logf("Modules created: %d", result.ModulesCreated)
	t.Logf("FRUs created: %d", result.FrusCreated)
	t.Logf("Cables created: %d", result.CablesCreated)
	t.Logf("Phase 0-2 errors: %d, downstream (3-6) errors: %d (expected in dry-run)", phase0Errs+phase1Errs+phase2Errs, downstreamErrs)

	// Validate counts
	if got := len(result.Created); got != 12 {
		t.Errorf("expected 12 devices created, got %d", got)
	}
	if got := len(result.RacksCreated); got != 2 {
		t.Errorf("expected 2 racks created, got %d", got)
	}
}

// TestExportSpineLeafLive verifies the full seven-phase export of the spine-leaf
// fixture succeeds against a live Nautobot, failing on any phase error.
//
// Why it matters: only a live run exercises the real creates (locations, racks,
// devices, interfaces, modules, FRUs, cables) and their dependency ordering,
// confirming the exporter actually persists a coherent topology end to end.
// Inputs: spine_leaf_inventory.json with DryRun=false; gated by CANI_LIVE_TEST=1
// and a reachable Nautobot. Outputs: a LoadResult whose accumulated Errors must
// be empty.
// Data choice: reusing the same spine-leaf fixture as the dry-run keeps the
// read-path and write-path tests comparable on identical data.
func TestExportSpineLeafLive(t *testing.T) {
	if os.Getenv("CANI_LIVE_TEST") == "" {
		t.Skip("set CANI_LIVE_TEST=1 to run live export tests")
	}
	skipUnlessNautobot(t)
	ensureTypesLoaded(t)
	inv := loadSpineLeafInventory(t)

	e := newTestProvider(t)
	e.Options.DryRun = false

	ctx := context.Background()
	e.Cache.SetContext(ctx)

	result := &LoadResult{
		Created:          make([]string, 0),
		Updated:          make([]string, 0),
		Skipped:          make([]string, 0),
		Errors:           make([]string, 0),
		Conflicts:        make([]ConflictInfo, 0),
		LocationsCreated: make([]string, 0),
		LocationsSkipped: make([]string, 0),
		RacksCreated:     make([]string, 0),
	}

	mapper := NewDeviceMapper(e.Cache, &MapperOpts{
		DefaultLocation: e.Options.DefaultLocation,
		DefaultRole:     e.Options.DefaultRole,
		DefaultStatus:   e.Options.DefaultStatus,
	})
	mapper.SetInventory(inv)

	// Phase 0: Locations
	if _, err := e.loadLocations(ctx, inv, result); err != nil {
		t.Logf("location phase: %v", err)
	}

	// Phase 1: Racks
	createdRackIDs := make(map[uuid.UUID]uuid.UUID)
	for rackID, rack := range inv.Racks {
		if rack == nil || rack.Name == "" {
			continue
		}
		existing, _ := e.Cache.GetRackByName(rack.Name)
		if existing != nil {
			createdRackIDs[rackID] = existing.ID
			continue
		}
		nautobotRackID, err := e.createRackFromCaniRack(ctx, rack, inv, mapper, result)
		if err != nil {
			result.Errors = append(result.Errors, rack.Name+": "+err.Error())
			continue
		}
		createdRackIDs[rackID] = nautobotRackID
	}

	// Phase 2: Devices
	createdDeviceIDs := make(map[string]uuid.UUID)
	for _, device := range inv.Devices {
		if device == nil || device.Name == "" {
			continue
		}
		category := devicetypes.ClassifyForNautobot(string(device.Type))
		if category != devicetypes.CategoryDevice {
			continue
		}
		existing, _ := e.Cache.GetDeviceByName(device.Name)
		if existing != nil {
			createdDeviceIDs[device.Name] = existing.ID
			continue
		}
		nautobotID, err := e.createDeviceWithID(ctx, device, mapper, result)
		if err != nil {
			result.Errors = append(result.Errors, device.Name+": "+err.Error())
			continue
		}
		if nautobotID != uuid.Nil {
			createdDeviceIDs[device.Name] = nautobotID
		}
	}

	// Phase 3: Interfaces (bulk)
	if err := e.loadInterfaces(ctx, inv, createdDeviceIDs, result); err != nil {
		result.Errors = append(result.Errors, "interfaces: "+err.Error())
	}

	// Phase 4: Modules
	if err := e.loadModules(ctx, inv, createdDeviceIDs, result); err != nil {
		result.Errors = append(result.Errors, "modules: "+err.Error())
	}

	// Phase 5: FRUs
	if err := e.loadFrus(ctx, inv, createdDeviceIDs, result); err != nil {
		result.Errors = append(result.Errors, "frus: "+err.Error())
	}

	// Phase 6: Cables
	for cableID, cable := range inv.Cables {
		if cable == nil {
			continue
		}
		if err := e.createCaniCableType(ctx, cableID, cable, inv, createdDeviceIDs, result); err != nil {
			result.Errors = append(result.Errors, cable.Label+": "+err.Error())
		}
	}

	e.printLoadSummary(result)

	for _, e := range result.Errors {
		t.Errorf("export error: %s", e)
	}
}
