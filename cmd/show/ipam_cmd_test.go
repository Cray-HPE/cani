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
package show

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// twoNamespaceIPAM seeds 10.0.0.0/24 holding 10.0.0.5 and a VRF blue in both
// Global and tenant-a; each record's text is identical across the namespaces.
func twoNamespaceIPAM() *devicetypes.Inventory {
	inventory := devicetypes.NewInventory()
	for _, namespace := range []string{"Global", "tenant-a"} {
		prefix, address, vrf := uuid.New(), uuid.New(), uuid.New()
		inventory.Prefixes[prefix] = &devicetypes.CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24", Namespace: namespace, Description: namespace}
		inventory.IPAddresses[address] = &devicetypes.CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/24", Parent: prefix, Description: namespace}
		inventory.VRFs[vrf] = &devicetypes.CaniVRF{ID: vrf, Name: "blue", Namespace: namespace, Description: namespace}
	}
	return inventory
}

// stdoutOf runs a show command through the harness and returns what it printed.
func stdoutOf(t *testing.T, target *cli.Command, flags map[string][]string) string {
	t.Helper()
	harness := cmdtest.New(t, NewCommand(), target, twoNamespaceIPAM())
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	previous := os.Stdout
	os.Stdout = writer
	runErr := harness.Run(t, flags)
	os.Stdout = previous
	_ = writer.Close()
	var output bytes.Buffer
	_, _ = io.Copy(&output, reader)
	if runErr != nil {
		t.Fatalf("%s: %v", target.Name(), runErr)
	}
	if harness.Store.Saves != 0 {
		t.Fatalf("%s saved %d time(s); show must never write", target.Name(), harness.Store.Saves)
	}
	return output.String()
}

// TestShowIPAMFiltersByNamespace verifies --namespace limits show prefix, ip
// and vrf to that namespace, and that omitting it lists every namespace.
//
// Why it matters: add, update and remove are namespace-aware; a listing that
// could not be scoped the same way would leave the operator unable to check
// what those commands did in a tenant.
// Inputs: identical prefix, address and VRF text in Global and tenant-a;
// each command run with --namespace tenant-a and without.
// Outputs: JSON naming only the tenant-a record when filtered, and both when
// not; no save in either case.
// Data choice: each record's description is its namespace, so the output
// shows which copy was listed without decoding UUIDs.
func TestShowIPAMFiltersByNamespace(t *testing.T) {
	for _, target := range []func() *cli.Command{newPrefixShowCommand, newIPShowCommand, newVRFShowCommand} {
		cmd := target()
		t.Run(cmd.Name(), func(t *testing.T) {
			// Act.
			filtered := stdoutOf(t, target(), map[string][]string{"format": {"json"}, flagNamespace: {"tenant-a"}})
			all := stdoutOf(t, target(), map[string][]string{"format": {"json"}})

			// Assert.
			if strings.Count(filtered, `"description": "tenant-a"`) != 1 || strings.Contains(filtered, `"description": "Global"`) {
				t.Errorf("--namespace tenant-a listed:\n%s", filtered)
			}
			if strings.Count(all, `"description": "`) != 2 {
				t.Errorf("unfiltered listing does not hold both namespaces:\n%s", all)
			}
		})
	}
}
