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
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// fakeCSMSource hands Transform the SLS and SMD data a test loaded.
type fakeCSMSource struct {
	sls *import_.SlsDumpstate
	smd *import_.SmdComponentList
}

func (f fakeCSMSource) GetSls() *import_.SlsDumpstate     { return f.sls }
func (f fakeCSMSource) GetSmd() *import_.SmdComponentList { return f.smd }

// TestTransformDeclaresTheXnameAsIdentity verifies the result Transform
// returns declares the xname as the one key that identifies a device.
//
// Why it matters: the merge compares only the keys a result declares; a
// result that forgot to declare the xname would let two nodes of one name
// with different xnames fold into one record on import.
// Inputs: the river SLS and SMD fixtures through the provider getter.
// Outputs: IdentityKeys equal to ["xname"], and no other key.
// Data choice: the full fixture, so the declaration is shown on the path
// the import command takes rather than on an empty result.
func TestTransformDeclaresTheXnameAsIdentity(t *testing.T) {
	// Arrange.
	sls, smd := loadFixtures(t)
	orig := providerGetter
	t.Cleanup(func() { providerGetter = orig })
	SetProviderGetter(func() interface {
		GetSls() *import_.SlsDumpstate
		GetSmd() *import_.SmdComponentList
	} {
		return fakeCSMSource{sls: sls, smd: smd}
	})

	// Act.
	result, err := Transform(devicetypes.Inventory{})

	// Assert.
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	if want := []string{"xname"}; !reflect.DeepEqual(result.IdentityKeys, want) {
		t.Errorf("IdentityKeys = %v, want %v", result.IdentityKeys, want)
	}
}
