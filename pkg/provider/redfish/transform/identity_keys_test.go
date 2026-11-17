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
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
)

// TestTransformDeclaresTheBMCNamesAsIdentity verifies the result Transform
// returns declares the BMC's FQDN and host name as the keys that identify a
// device.
//
// Why it matters: the merge compares only the keys a result declares; two
// servers of one product name reached through different BMCs must stay two
// records, which the declared keys are what keeps them.
// Inputs: one ServiceRoot through the provider getter.
// Outputs: IdentityKeys equal to ["bmc_fqdn", "bmc_hostname"], in that
// order, since the first key both devices hold decides.
// Data choice: the test root, so the declaration is shown on a result that
// also holds a device.
func TestTransformDeclaresTheBMCNamesAsIdentity(t *testing.T) {
	// Arrange.
	withProviderGetter(t, []import_.ServiceRoot{testRoot()})

	// Act.
	result, err := Transform(devicetypes.Inventory{})

	// Assert.
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	if want := []string{"bmc_fqdn", "bmc_hostname"}; !reflect.DeepEqual(result.IdentityKeys, want) {
		t.Errorf("IdentityKeys = %v, want %v", result.IdentityKeys, want)
	}
}
