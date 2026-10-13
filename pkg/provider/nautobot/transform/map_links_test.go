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
package transform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// setNBTags fills a generated `Tags *[]struct{Id ...}` field from tag UUIDs
// through JSON, so the test never names the anonymous element type.
func setNBTags(t *testing.T, field any, ids ...uuid.UUID) {
	t.Helper()
	refs := make([]map[string]string, len(ids))
	for i, id := range ids {
		refs[i] = map[string]string{"id": id.String()}
	}
	raw, _ := json.Marshal(refs)
	if err := json.Unmarshal(raw, field); err != nil {
		t.Fatalf("setting tags: %v", err)
	}
}

func nbTag(id uuid.UUID, name string) nautobotapi.Tag {
	return nautobotapi.Tag{Id: oaPtr(id), Name: name}
}

func external(nbID uuid.UUID) devicetypes.ObjectMeta {
	return devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: nbID}}
}

// TestLinkTagsNamesTagsOnRecordsAndInterfaceSpecs verifies that LinkTags
// resolves tag references to sorted names on a device, a prefix and a
// device's interface spec, and drops a reference the catalog does not know.
//
// Why it matters: Nautobot returns tags as bare references; without the
// catalog join a tag set in Nautobot would never reach the datastore, and
// the round-trip oracle compares tags per family.
// Inputs: a two-tag catalog; a device tagged with both plus an unknown id, a
// prefix tagged with one, an interface of the device tagged with the other.
// Outputs: device ["alpha","beta"], prefix ["beta"], interface spec
// ["alpha"]; an untagged record keeps nil.
// Data choice: the device's references are given in reverse order to show
// the names come out sorted.
func TestLinkTagsNamesTagsOnRecordsAndInterfaceSpecs(t *testing.T) {
	alpha, beta, unknown := uuid.New(), uuid.New(), uuid.New()
	devNB, ifaceNB, prefixNB, quietNB := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	raw := &RawData{Tags: []nautobotapi.Tag{nbTag(alpha, "alpha"), nbTag(beta, "beta")}}
	raw.Devices = []nautobotapi.Device{{Id: oaPtr(devNB)}, {Id: oaPtr(quietNB)}}
	setNBTags(t, &raw.Devices[0].Tags, beta, unknown, alpha)
	raw.Interfaces = []nautobotapi.Interface{{Id: oaPtr(ifaceNB)}}
	setNBRef(&raw.Interfaces[0].Device, devNB)
	setNBTags(t, &raw.Interfaces[0].Tags, alpha)
	raw.Prefixes = []nautobotapi.Prefix{{Id: oaPtr(prefixNB)}}
	setNBTags(t, &raw.Prefixes[0].Tags, beta)

	devID, quietID, prefixID := uuid.New(), uuid.New(), uuid.New()
	result := &devicetypes.TransformResult{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			devID:   {ID: devID, ObjectMeta: external(devNB), Interfaces: []devicetypes.InterfaceSpec{{ID: ifaceNB, Name: "eth0"}}},
			quietID: {ID: quietID, ObjectMeta: external(quietNB)},
		},
		Prefixes: map[uuid.UUID]*devicetypes.CaniPrefix{prefixID: {ID: prefixID, ObjectMeta: external(prefixNB)}},
	}

	LinkTags(result, raw)

	if got := result.Devices[devID].Tags; !reflect.DeepEqual(got, []string{"alpha", "beta"}) {
		t.Fatalf("device tags = %v", got)
	}
	if got := result.Devices[devID].Interfaces[0].Tags; !reflect.DeepEqual(got, []string{"alpha"}) {
		t.Fatalf("interface tags = %v", got)
	}
	if got := result.Prefixes[prefixID].Tags; !reflect.DeepEqual(got, []string{"beta"}) {
		t.Fatalf("prefix tags = %v", got)
	}
	if got := result.Devices[quietID].Tags; got != nil {
		t.Fatalf("untagged device must keep nil tags, got %v", got)
	}
}

// TestLinkPrimaryIPsResolvesThroughImportedAddresses verifies that a device's
// primary IPv4 and IPv6 are set to the cani UUIDs of the imported addresses
// carrying the referenced Nautobot UUIDs, and left unset when the address
// was not imported.
//
// Why it matters: addresses get fresh cani UUIDs on import, so the device
// reference must be translated or it would dangle and fail the rebuild.
// Inputs: one device whose primary_ip4 names an imported address and whose
// primary_ip6 names an address the import left out.
// Outputs: PrimaryIPv4 equals the imported address's cani UUID; PrimaryIPv6
// stays Nil.
// Data choice: both families on one device exercise the hit and the miss in
// a single pass.
func TestLinkPrimaryIPsResolvesThroughImportedAddresses(t *testing.T) {
	devNB, ip4NB, ip6NB := uuid.New(), uuid.New(), uuid.New()
	raw := []nautobotapi.Device{{Id: oaPtr(devNB)}}
	setNBRef(&raw[0].PrimaryIp4, ip4NB)
	setNBRef(&raw[0].PrimaryIp6, ip6NB)

	devID, ip4ID := uuid.New(), uuid.New()
	result := &devicetypes.TransformResult{
		Devices:     map[uuid.UUID]*devicetypes.CaniDeviceType{devID: {ID: devID, ObjectMeta: external(devNB)}},
		IPAddresses: map[uuid.UUID]*devicetypes.CaniIPAddress{ip4ID: {ID: ip4ID, Address: "10.0.0.5/24", ObjectMeta: external(ip4NB)}},
	}

	LinkPrimaryIPs(result, raw)

	device := result.Devices[devID]
	if device.PrimaryIPv4 != ip4ID || device.PrimaryIPv6 != uuid.Nil {
		t.Fatalf("primary IPs = v4 %s, v6 %s; want v4 %s, v6 nil", device.PrimaryIPv4, device.PrimaryIPv6, ip4ID)
	}
}

// TestTagNamesToleratesNonSliceInput verifies that tagNames returns nil for a
// nil pointer and for a value that is not a slice, instead of panicking on an
// unexpected generated shape.
//
// Why it matters: the helper is applied to every family through reflection,
// so a shape it does not recognise must degrade to "no tags".
// Inputs: a typed nil pointer and a plain string.
// Outputs: nil for both.
// Data choice: the two inputs cover the nil-pointer and wrong-kind branches.
func TestTagNamesToleratesNonSliceInput(t *testing.T) {
	var none *[]struct{ Id *string }
	for _, in := range []any{none, "not a slice"} {
		if got := tagNames(in, map[uuid.UUID]string{}); got != nil {
			t.Fatalf("tagNames(%s) = %v, want nil", fmt.Sprintf("%T", in), got)
		}
	}
}
