package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
	"github.com/google/uuid"
)

func singleResultDevice(t *testing.T, result *devicetypes.TransformResult) (uuid.UUID, *devicetypes.CaniDeviceType) {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Devices) != 1 {
		t.Fatalf("Devices = %d, want 1", len(result.Devices))
	}
	for id, dev := range result.Devices {
		if dev == nil {
			t.Fatalf("Devices[%s] is nil", id)
		}
		return id, dev
	}
	t.Fatal("result has no device despite len(result.Devices) == 1")
	return uuid.Nil, nil
}

func assertUnsupportedCoreTypesEmpty(t *testing.T, result *devicetypes.TransformResult) {
	t.Helper()
	checks := map[string]int{
		"Locations":   len(result.Locations),
		"Racks":       len(result.Racks),
		"Modules":     len(result.Modules),
		"Cables":      len(result.Cables),
		"Frus":        len(result.Frus),
		"Prefixes":    len(result.Prefixes),
		"IPAddresses": len(result.IPAddresses),
		"VLANs":       len(result.VLANs),
	}
	for name, got := range checks {
		if got != 0 {
			t.Errorf("%s = %d, want 0 for ServiceRoot-only transform", name, got)
		}
	}
	if result.Metadata != nil {
		t.Errorf("Metadata = %+v, want nil for ServiceRoot-only transform", result.Metadata)
	}
}

// TestTransformRoots_MapsNautobotDeviceAnchorFields verifies a Redfish root that
// matches the local device-type library fills the CANI fields used as Nautobot
// Device and DeviceType anchors.
//
// Why it matters: Nautobot export resolves Device.device_type from the CANI slug
// and creates DeviceType records from local library values, so Redfish transform
// must preserve the raw device identity while carrying the matched slug and
// template fields forward.
// Inputs: one ServiceRoot whose Product is a known HPE library slug and whose
// Vendor is empty. Outputs: the single transformed device with Name/ID/provider
// metadata from Redfish, Type as the Redfish device classification, and Slug,
// Manufacturer, Model, PartNumber, and UHeight from the library.
// Data choice: the DL380 Gen11 8SFF fixture has stable manufacturer, model, part
// number, and U-height fields, and using its slug drives the exact-match branch.
func TestTransformRoots_MapsNautobotDeviceAnchorFields(t *testing.T) {
	lib, ok := devicetypes.GetBySlug("hpe-proliant-dl380-gen11-8sff")
	if !ok {
		t.Fatal("expected embedded library fixture hpe-proliant-dl380-gen11-8sff")
	}
	root := testRoot()
	root.Product = lib.Slug
	root.Vendor = ""

	result, err := transformRoots([]import_.ServiceRoot{root}, nil)
	if err != nil {
		t.Fatalf("transformRoots() error: %v", err)
	}
	deviceID, dev := singleResultDevice(t, result)

	if dev.ID != deviceID || dev.ID == uuid.Nil {
		t.Errorf("device identity = key %s / ID %s, want matching non-nil UUIDs", deviceID, dev.ID)
	}
	if dev.Name != lib.Slug {
		t.Errorf("Name = %q, want Redfish Product %q", dev.Name, lib.Slug)
	}
	if dev.Slug != lib.Slug {
		t.Errorf("Slug = %q, want %q", dev.Slug, lib.Slug)
	}
	if dev.Manufacturer != lib.Manufacturer {
		t.Errorf("Manufacturer = %q, want library manufacturer %q", dev.Manufacturer, lib.Manufacturer)
	}
	if dev.Model != lib.Model {
		t.Errorf("Model = %q, want %q", dev.Model, lib.Model)
	}
	if dev.PartNumber != lib.PartNumber {
		t.Errorf("PartNumber = %q, want %q", dev.PartNumber, lib.PartNumber)
	}
	if dev.UHeight != lib.UHeight {
		t.Errorf("UHeight = %d, want %d", dev.UHeight, lib.UHeight)
	}
	if dev.Type != devicetypes.TypeNode {
		t.Errorf(hwTypeErrFmt, dev.Type, devicetypes.TypeNode)
	}
	redfishMeta, ok := dev.GetProviderSubMap(providerKeyRedfish)
	if !ok {
		t.Fatal("ProviderMetadata missing redfish sub-map")
	}
	if redfishMeta[metaKeyRedfishUUID] != root.UUID {
		t.Errorf("redfish_uuid = %v, want %q", redfishMeta[metaKeyRedfishUUID], root.UUID)
	}
	if redfishMeta["redfish_version"] != root.RedfishVersion {
		t.Errorf("redfish_version = %v, want %q", redfishMeta["redfish_version"], root.RedfishVersion)
	}
	assertUnsupportedCoreTypesEmpty(t, result)
}
