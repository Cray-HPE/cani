/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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

// TestLoad_SendsListedProviderMetadataOnADeviceCreate verifies a whole
// export sends a device's custom fields, the provider-metadata key the
// export options list, and no other provider metadata.
//
// Why it matters: Load builds the device mapper itself; if the listed keys
// stopped at the exporter's options, the option would be accepted and have
// no effect on the objects an export writes most of.
// Inputs: an exporter listing csm.xname; one device with three custom
// fields whose csm bucket holds xname and nid; a fake Nautobot that records
// the device create.
// Outputs: the create's custom_fields are the three custom fields and
// xname, without nid.
// Data choice: nid sits in the same bucket as the listed key, so a
// projection that sent the whole bucket once any of it was listed shows.
func TestLoad_SendsListedProviderMetadataOnADeviceCreate(t *testing.T) {
	// Arrange.
	var sent map[string]any
	defined := []string{"decommissioned", "spare_slots", "tier", "xname"}
	e, cleanup := newExporterWithServer(t, deviceCreateEchoingDefined(defined, &sent))
	defer cleanup()
	seedDeviceRefs(t, e)
	e.Options.DefaultLocation, e.Options.DefaultStatus, e.Options.DefaultRole = "DC1", "Active", "Compute"
	e.Options.ProviderMetadataFields = []ProviderMetadataField{{Provider: "csm", Key: "xname", CustomField: "xname"}}
	device := customFieldDevice("compute-001")
	device.ProviderMetadata["csm"].(map[string]any)["nid"] = 1001
	inv := &devicetypes.Inventory{Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{uuid.New(): device}}

	// Act.
	err := e.Load(inv)

	// Assert.
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]any{"decommissioned": false, "spare_slots": float64(0), "tier": "gold", "xname": "x3000c0s1b0"}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("device create custom_fields = %v, want %v", sent, want)
	}
}

// TestCreateVLAN_SendsListedProviderMetadata verifies an object the
// exporter writes without the device mapper takes the listed keys from the
// same export options.
//
// Why it matters: locations, racks, modules, cables and the IPAM objects
// read the projection from the exporter rather than from a mapper; were
// that source empty, the option would reach devices and nothing else.
// Inputs: an exporter listing csm.sls_network as the custom field network;
// a VLAN with no custom fields whose csm bucket holds sls_network; a fake
// Nautobot that records the VLAN create.
// Outputs: the create's custom_fields are network alone.
// Data choice: the VLAN has no custom fields of its own, so the payload
// exists only if the listed key was read, and the rename shows the key is
// sent under the name the option gives it.
func TestCreateVLAN_SendsListedProviderMetadata(t *testing.T) {
	// Arrange.
	var captured []byte
	e, cleanup := newExporterWithServer(t, vlanCreateServer(http.StatusCreated, uuid.New(), &captured))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Options.ProviderMetadataFields = []ProviderMetadataField{{Provider: "csm", Key: "sls_network", CustomField: "network"}}
	vlan := &devicetypes.CaniVLAN{VID: 100, Name: "vlan100"}
	vlan.ProviderMetadata = map[string]any{"csm": map[string]any{"sls_network": "NMN", "sls_type": "ethernet"}}

	// Act.
	_, err := e.createVLAN(context.Background(), vlan, uuid.Nil, &LoadResult{})

	// Assert.
	if err != nil {
		t.Fatalf("createVLAN: %v", err)
	}
	var payload struct {
		CustomFields map[string]any `json:"custom_fields"`
	}
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("decode VLAN payload: %v\nbody: %s", err, captured)
	}
	if want := map[string]any{"network": "NMN"}; !reflect.DeepEqual(payload.CustomFields, want) {
		t.Errorf("VLAN create custom_fields = %v, want %v", payload.CustomFields, want)
	}
}
