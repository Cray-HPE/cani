package datastores

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestJSONStorePreservesAuthoredInterfaceFields verifies interface instance
// data survives conversion to embedded specs and repeated save/load cycles.
//
// Why it matters: loading must not discard management or custom-field data.
// Inputs: an added interface with all authored metadata and switchport fields.
// Outputs: the complete interface is unchanged after each round trip.
// Data choice: explicit nonzero fields expose omissions in either conversion.
func TestJSONStorePreservesAuthoredInterfaceFields(t *testing.T) {
	inventory := devicetypes.NewInventory()
	deviceID := uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{ID: deviceID, Name: "switch"}
	iface := &devicetypes.CaniInterface{
		ID: uuid.New(), Name: "mgmt0", DeviceID: deviceID, InterfaceType: "1000base-t",
		MgmtOnly: true, Label: "Management", MacAddress: "02:00:00:00:00:01",
		Lag: "bond0", Mode: "tagged", UntaggedVLAN: 10, TaggedVLANs: []int{20, 30},
		VRF: "operations", Description: "Out-of-band management", ContentType: "dcim.interface",
		ObjectMeta: devicetypes.ObjectMeta{
			Status: string(devicetypes.StatusActive), Role: "management", Tags: []string{"critical"},
			Tenant: "operations", CustomFields: map[string]any{"owner": "network-team"},
			ExternalIDs:      map[string]uuid.UUID{"source": uuid.New()},
			ProviderMetadata: map[string]any{"source": map[string]any{"port": "mgmt0"}},
		},
	}
	if err := inventory.AddInterface(iface); err != nil {
		t.Fatalf("AddInterface() error = %v", err)
	}
	store := &JSONStore{Path: filepath.Join(t.TempDir(), "inventory.json")}

	for range 2 {
		if err := store.Save(inventory); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		loaded, err := store.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got := loaded.Interfaces[iface.ID]; !reflect.DeepEqual(got, iface) {
			t.Fatalf("loaded interface = %#v, want %#v", got, iface)
		}
		inventory = loaded
	}
}

// TestJSONStoreRebuildsInterfaceIPAssignments verifies interface IP lists are
// derived from address assignments rather than stale serialized reverse links.
//
// Why it matters: assignment identity must survive rebuilding interface objects.
// Inputs: one assigned IP and a stale reverse list on its interface.
// Outputs: repeated rebuilds produce only the actual address UUID, once.
// Data choice: a different cached UUID exposes trusting or dropping reverse data.
func TestJSONStoreRebuildsInterfaceIPAssignments(t *testing.T) {
	inventory := devicetypes.NewInventory()
	deviceID, interfaceID, addressID := uuid.New(), uuid.New(), uuid.New()
	inventory.Devices[deviceID] = &devicetypes.CaniDeviceType{ID: deviceID, Name: "host"}
	iface := &devicetypes.CaniInterface{
		ID: interfaceID, DeviceID: deviceID, Name: "eth0", IPAddresses: []uuid.UUID{uuid.New()},
	}
	if err := inventory.AddInterface(iface); err != nil {
		t.Fatalf("AddInterface() error = %v", err)
	}
	inventory.IPAddresses[addressID] = &devicetypes.CaniIPAddress{
		ID: addressID, Address: "10.0.0.1/24", Interfaces: []uuid.UUID{interfaceID},
	}
	store := &JSONStore{Path: filepath.Join(t.TempDir(), "inventory.json")}
	if err := store.Save(inventory); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := loaded.RebuildDerivedState().Err(); err != nil {
		t.Fatalf("rebuild error = %v", err)
	}

	if got := loaded.Interfaces[interfaceID].IPAddresses; !reflect.DeepEqual(got, []uuid.UUID{addressID}) {
		t.Errorf("interface IPs = %v, want [%s]", got, addressID)
	}
}
