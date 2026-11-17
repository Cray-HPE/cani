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
package devicetypes

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestRebuildDerivedStateReportsProviderMetadataOutsideABucket verifies that
// a provider-metadata value stored beside the provider buckets is reported
// for the operator, by object and key, on every kind of object, without
// failing the inventory.
//
// Why it matters: provider metadata belongs to the provider whose bucket
// holds it, and nothing reads a value outside one. HPCM's rack number sat
// outside a bucket for that reason unnoticed; a value nothing reads should
// be said to be so, not kept in silence.
// Inputs: an inventory in which a rack, a VLAN and an interface spec of a
// device each hold one value outside a bucket beside a proper bucket, and a
// module holds buckets only.
// Outputs: one unresolved line for each of the three, naming the object and
// the key and not the value, and no error.
// Data choice: one object from DCIM, one from IPAM and one interface spec,
// so the report is seen to cover collections of different shapes; the
// values are distinctive so a leak into a line would be found.
func TestRebuildDerivedStateReportsProviderMetadataOutsideABucket(t *testing.T) {
	// Arrange.
	inv := NewInventory()
	rackID, vlanID, deviceID, moduleID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "rack-9000", ObjectMeta: ObjectMeta{
		ProviderMetadata: map[string]any{"hpcm": map[string]any{"rack_number": 9000}, "rack_number": "value-9000"},
	}}
	inv.VLANs[vlanID] = &CaniVLAN{ID: vlanID, VID: 100, Name: "storage", ObjectMeta: ObjectMeta{
		ProviderMetadata: map[string]any{"note": "value-note"},
	}}
	inv.Devices[deviceID] = &CaniDeviceType{ID: deviceID, Name: "leaf-1",
		ObjectMeta: ObjectMeta{ProviderMetadata: map[string]any{"csm": map[string]any{"xname": "x3000c0w1"}}},
		Interfaces: []InterfaceSpec{{Name: "1/1/1", ProviderMetadata: map[string]any{"speed": "value-speed"}}},
	}
	inv.Modules[moduleID] = &CaniModuleType{ID: moduleID, Name: "nic-1", ObjectMeta: ObjectMeta{
		ProviderMetadata: map[string]any{"hpcm": map[string]any{"geoloc": "x9000c1s0b0n0"}, "redfish": map[string]any{}},
	}}

	// Act.
	result := inv.RebuildDerivedState()

	// Assert.
	var reported []string
	for _, line := range result.Unresolved {
		if strings.Contains(line, "provider metadata") {
			reported = append(reported, line)
		}
	}
	slices.Sort(reported)
	const remedy = ` is outside every provider's bucket, so nothing reads it; move it into a bucket or into customFields`
	want := []string{
		`VLAN "storage" (` + vlanID.String() + `): provider metadata "note"` + remedy,
		`interface "1/1/1" of device "leaf-1": provider metadata "speed"` + remedy,
		`rack "rack-9000" (` + rackID.String() + `): provider metadata "rack_number"` + remedy,
	}
	if !slices.Equal(reported, want) {
		t.Errorf("reported = %q\nwant %q", reported, want)
	}
	if strings.Contains(strings.Join(reported, "\n"), "value-") {
		t.Errorf("a report line prints a value: %q", reported)
	}
	if err := result.Err(); err != nil {
		t.Errorf("Err() = %v, want the inventory still usable", err)
	}
}

// withRegisteredBuckets makes only the given provider slugs known for the
// rest of the test, restoring whatever was registered before.
func withRegisteredBuckets(t *testing.T, slugs ...string) {
	t.Helper()
	bucketsMu.Lock()
	saved := knownBuckets
	knownBuckets = make(map[string]bool, len(slugs))
	for _, slug := range slugs {
		knownBuckets[slug] = true
	}
	bucketsMu.Unlock()
	t.Cleanup(func() {
		bucketsMu.Lock()
		knownBuckets = saved
		bucketsMu.Unlock()
	})
}

// TestValidateProviderMetadataReportsBucketsNoProviderOwns verifies that a
// bucket named for a provider that is not built in is reported once, with
// the number of objects holding it and one example, while a registered
// provider's bucket is not.
//
// Why it matters: a bucket is a provider's own record and nothing else reads
// it. One named for no provider, left by a build that had the provider or
// by a typo, is as unread as a value outside every bucket, and until now
// only the latter was reported. One line per bucket name keeps a whole
// import's worth of one stray bucket from burying command output.
// Inputs: csm registered as the one provider; a device and a rack holding
// an "acme" bucket, the device also holding a csm bucket.
// Outputs: one unresolved line naming "acme", counting 2 and giving the
// device as the example, and none for csm.
// Data choice: the device's label sorts before the rack's, so the example
// is seen to be chosen by label rather than by map order.
func TestValidateProviderMetadataReportsBucketsNoProviderOwns(t *testing.T) {
	// Arrange.
	withRegisteredBuckets(t, "csm")
	inv := NewInventory()
	deviceID, rackID := uuid.New(), uuid.New()
	inv.Devices[deviceID] = &CaniDeviceType{ID: deviceID, Name: "n1", ObjectMeta: ObjectMeta{
		ProviderMetadata: map[string]any{"acme": map[string]any{"tag": 1}, "csm": map[string]any{"xname": "x3000c0s1b0n0"}},
	}}
	inv.Racks[rackID] = &CaniRackType{ID: rackID, Name: "x3000", ObjectMeta: ObjectMeta{
		ProviderMetadata: map[string]any{"acme": map[string]any{"tag": 2}},
	}}

	// Act.
	result := inv.validateProviderMetadata()

	// Assert.
	if len(result.Unresolved) != 1 {
		t.Fatalf("Unresolved = %q, want one line", result.Unresolved)
	}
	for _, want := range []string{`bucket "acme" is held by 2 object(s)`, `e.g. device "n1" (` + deviceID.String() + `)`, `no provider "acme" is built in`} {
		if !strings.Contains(result.Unresolved[0], want) {
			t.Errorf("Unresolved[0] = %q, want it to contain %q", result.Unresolved[0], want)
		}
	}
}

// TestValidateProviderMetadataSkipsBucketsWhenNoProviderIsRegistered
// verifies that with no provider registered, buckets are not judged at all.
//
// Why it matters: the model is used on its own, in its tests and in tools
// that load an inventory without the providers; with nothing registered
// every bucket would look unowned, and every inventory would be reported.
// Inputs: no registered provider; a device holding an "acme" bucket.
// Outputs: nothing reported.
// Data choice: the same bucket the test above reports, so registration
// alone makes the difference.
func TestValidateProviderMetadataSkipsBucketsWhenNoProviderIsRegistered(t *testing.T) {
	// Arrange.
	withRegisteredBuckets(t)
	inv := NewInventory()
	deviceID := uuid.New()
	inv.Devices[deviceID] = &CaniDeviceType{ID: deviceID, Name: "n1", ObjectMeta: ObjectMeta{
		ProviderMetadata: map[string]any{"acme": map[string]any{"tag": 1}},
	}}

	// Act.
	result := inv.validateProviderMetadata()

	// Assert.
	if len(result.Unresolved) != 0 {
		t.Errorf("Unresolved = %q, want none", result.Unresolved)
	}
}
