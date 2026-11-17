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
	"reflect"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/ochami/import"
)

// fakeOChaMiSource hands Transform the records a test built.
type fakeOChaMiSource struct{ records []import_.JSONDeviceRecord }

func (f fakeOChaMiSource) GetRecords() []import_.JSONDeviceRecord { return f.records }

// TestTransformDeclaresTheXnameAsIdentity verifies the result Transform
// returns declares the xname as the one key that identifies a device, the
// same key CSM declares, so a node both providers import is one record.
//
// Why it matters: the merge compares only the keys a result declares, and
// reads a declared key from whichever bucket holds it; declaring xname is
// what lets an OChaMi import of a node CSM already recorded be told apart
// from, or joined to, that record by xname.
// Inputs: one node record with an xname, through the provider getter.
// Outputs: IdentityKeys equal to ["xname"].
// Data choice: a single node keeps the fixture small; the declaration does
// not depend on what the records hold.
func TestTransformDeclaresTheXnameAsIdentity(t *testing.T) {
	// Arrange.
	orig := providerGetter
	t.Cleanup(func() { providerGetter = orig })
	SetProviderGetter(func() interface {
		GetRecords() []import_.JSONDeviceRecord
	} {
		return fakeOChaMiSource{records: []import_.JSONDeviceRecord{{
			DeviceType: "Node", SerialNumber: "NODE-SN-001", Manufacturer: "HPE",
			Properties: import_.Properties{Xname: "x3000c0s1b0n0"},
		}}}
	})

	// Act.
	result, err := Transform(*devicetypes.NewInventory())

	// Assert.
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	if want := []string{"xname"}; !reflect.DeepEqual(result.IdentityKeys, want) {
		t.Errorf("IdentityKeys = %v, want %v", result.IdentityKeys, want)
	}
}
