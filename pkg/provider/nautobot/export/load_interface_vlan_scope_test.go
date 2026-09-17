package export

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestEnrichInterfacesKeepsVLANLocationScope verifies equal VIDs in different
// locations resolve to the correct Nautobot UUID in each interface PATCH.
//
// Why it matters: a global VID map silently attaches one switch to another
// location's VLAN when both sites use the same native or tagged VLAN number.
// Inputs: two located switches and a locationless switch, using VLAN 2000.
// Outputs: each located switch uses its own VLAN; the locationless one is
// skipped and reports both VLAN references as unresolved.
// Data choice: duplicate VIDs with distinct UUIDs expose scope loss regardless
// of Go map iteration order.
func TestEnrichInterfacesKeepsVLANLocationScope(t *testing.T) {
	inventory := devicetypes.NewInventory()
	createdDevices := make(map[string]uuid.UUID)
	createdVLANs := make(map[uuid.UUID]uuid.UUID)
	expected := make(map[string]uuid.UUID)
	patches := make(map[string]interfaceVLANPayload)

	exporter, cleanup := newExporterWithServer(t, recordInterfaceVLANPatches(t, patches))
	defer cleanup()

	for _, name := range []string{"switch-east", "switch-west"} {
		deviceID, locationID, vlanID := uuid.New(), uuid.New(), uuid.New()
		nautobotDeviceID, nautobotVLANID, interfaceID := uuid.New(), uuid.New(), uuid.New()
		inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{
			ID: deviceID, Name: name, Location: locationID,
			Interfaces: []devicetypes.InterfaceSpec{{Name: "port1", UntaggedVLAN: 2000, TaggedVLANs: []int{2000}}},
		}
		inventory.VLANs[vlanID] = &devicetypes.CaniVLAN{ID: vlanID, VID: 2000, Location: locationID}
		createdDevices[name] = nautobotDeviceID
		createdVLANs[vlanID] = nautobotVLANID
		exporter.Cache.CacheInterface(nautobotDeviceID, "port1", &CachedItem{ID: interfaceID, Name: "port1"})
		expected["/dcim/interfaces/"+interfaceID.String()+"/"] = nautobotVLANID
	}
	unlocatedID, nautobotUnlocatedID := uuid.New(), uuid.New()
	inventory.Devices[unlocatedID] = &devicetypes.CaniDeviceType{
		ID: unlocatedID, Name: "switch-unlocated",
		Interfaces: []devicetypes.InterfaceSpec{{Name: "port1", UntaggedVLAN: 2000, TaggedVLANs: []int{2000}}},
	}
	createdDevices["switch-unlocated"] = nautobotUnlocatedID
	exporter.Cache.CacheInterface(nautobotUnlocatedID, "port1", &CachedItem{ID: uuid.New(), Name: "port1"})

	result := &LoadResult{}
	if err := exporter.enrichInterfaces(context.Background(), inventory, createdDevices, createdVLANs, result); err != nil {
		t.Fatalf("enrichInterfaces() error = %v", err)
	}
	if len(result.Errors) != 0 || result.IfacesUnresolvedRefs != 2 {
		t.Fatalf("enrichment errors=%v unresolved=%d", result.Errors, result.IfacesUnresolvedRefs)
	}
	assertInterfaceVLANPatches(t, patches, expected)
}

func recordInterfaceVLANPatches(t *testing.T, patches map[string]interfaceVLANPayload) http.HandlerFunc {
	t.Helper()
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodGet:
			_, _ = writer.Write([]byte(`{"count":0,"results":[]}`))
		case http.MethodPatch:
			var payload interfaceVLANPayload
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Errorf("decode interface PATCH: %v", err)
			}
			patches[request.URL.Path] = payload
			_, _ = writer.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected %s %s", request.Method, request.URL.Path)
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func assertInterfaceVLANPatches(t *testing.T, patches map[string]interfaceVLANPayload, expected map[string]uuid.UUID) {
	t.Helper()
	if len(patches) != len(expected) {
		t.Fatalf("received %d interface PATCHes, want %d", len(patches), len(expected))
	}
	for requestPath, vlanID := range expected {
		payload, received := patches[requestPath]
		if !received {
			t.Fatalf("missing PATCH %s; received %v", requestPath, patches)
		}
		if payload.UntaggedVLAN.ID != vlanID || !reflect.DeepEqual(payload.TaggedVLANs, []interfaceVLANReference{{ID: vlanID}}) {
			t.Errorf("PATCH %s VLANs = %+v, want native and tagged VLAN %s", requestPath, payload, vlanID)
		}
	}
}

type interfaceVLANPayload struct {
	UntaggedVLAN interfaceVLANReference   `json:"untagged_vlan"`
	TaggedVLANs  []interfaceVLANReference `json:"tagged_vlans"`
}

type interfaceVLANReference struct {
	ID uuid.UUID `json:"id"`
}

// TestBuildVIDMapScopeFallbacks verifies local scope and ambiguity are preserved
// when converting inventory VLAN identities to exported UUIDs.
//
// Why it matters: fallback must not silently select a different VLAN when the
// intended local VLAN is ambiguous or could not be created in Nautobot.
// Inputs: local, remote, unscoped, duplicate, and unexported VLAN candidates.
// Outputs: the unique local UUID, an unscoped fallback, or no mapping.
// Data choice: every candidate uses VID 2000 so only scope and identity differ.
func TestBuildVIDMapScopeFallbacks(t *testing.T) {
	locationID, otherLocationID := uuid.New(), uuid.New()
	local := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 2000, Location: locationID}
	remote := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 2000, Location: otherLocationID}
	unscoped := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 2000}
	duplicateLocal := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 2000, Location: locationID}
	duplicateUnscoped := &devicetypes.CaniVLAN{ID: uuid.New(), VID: 2000}
	created := map[uuid.UUID]uuid.UUID{
		local.ID: uuid.New(), remote.ID: uuid.New(), unscoped.ID: uuid.New(),
		duplicateLocal.ID: uuid.New(), duplicateUnscoped.ID: uuid.New(),
	}
	cases := []struct {
		name       string
		locationID uuid.UUID
		vlans      []*devicetypes.CaniVLAN
		created    map[uuid.UUID]uuid.UUID
		wantID     uuid.UUID
	}{
		{"own location", locationID, []*devicetypes.CaniVLAN{local, remote}, created, created[local.ID]},
		{"unlocated device", uuid.Nil, []*devicetypes.CaniVLAN{local, remote}, created, uuid.Nil},
		{"other location", uuid.New(), []*devicetypes.CaniVLAN{local, remote}, created, uuid.Nil},
		{"unscoped fallback", locationID, []*devicetypes.CaniVLAN{remote, unscoped}, created, created[unscoped.ID]},
		{"local wins fallback", locationID, []*devicetypes.CaniVLAN{unscoped, local, remote}, created, created[local.ID]},
		{"unscoped device", uuid.Nil, []*devicetypes.CaniVLAN{unscoped, local}, created, created[unscoped.ID]},
		{"failed local blocks fallback", locationID, []*devicetypes.CaniVLAN{local, unscoped}, map[uuid.UUID]uuid.UUID{unscoped.ID: created[unscoped.ID]}, uuid.Nil},
		{"nil remote UUID", locationID, []*devicetypes.CaniVLAN{local}, map[uuid.UUID]uuid.UUID{local.ID: uuid.Nil}, uuid.Nil},
		{"ambiguous local blocks fallback", locationID, []*devicetypes.CaniVLAN{local, duplicateLocal, unscoped}, created, uuid.Nil},
		{"partially exported ambiguity", locationID, []*devicetypes.CaniVLAN{local, duplicateLocal, unscoped}, map[uuid.UUID]uuid.UUID{local.ID: created[local.ID], unscoped.ID: created[unscoped.ID]}, uuid.Nil},
		{"ambiguous unscoped", uuid.Nil, []*devicetypes.CaniVLAN{unscoped, duplicateUnscoped}, created, uuid.Nil},
		{"local overrides unscoped ambiguity", locationID, []*devicetypes.CaniVLAN{local, unscoped, duplicateUnscoped}, created, created[local.ID]},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			inventory := devicetypes.NewInventory()
			inventory.VLANs[uuid.New()] = nil
			for _, vlan := range testCase.vlans {
				inventory.VLANs[vlan.ID] = vlan
			}
			want := make(map[int]uuid.UUID)
			if testCase.wantID != uuid.Nil {
				want[2000] = testCase.wantID
			}

			got := buildVIDMap(inventory, testCase.created, testCase.locationID)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("buildVIDMap() = %v, want %v", got, want)
			}
		})
	}
}
