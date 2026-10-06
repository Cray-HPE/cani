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

package cmdtest_test

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/remove"
	"github.com/Cray-HPE/cani/cmd/update"
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/testutil/cmdtest"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// ipamInventory seeds a Global 10.0.0.0/16 holding 10.0.0.5 and a Global VRF
// blue, the smallest shape every IPAM verb can act on.
func ipamInventory() *devicetypes.Inventory {
	inventory := devicetypes.NewInventory()
	prefix, address, vrf := uuid.New(), uuid.New(), uuid.New()
	inventory.Prefixes[prefix] = &devicetypes.CaniPrefix{ID: prefix, Prefix: "10.0.0.0/16", PrefixLen: 16, Namespace: "Global"}
	inventory.IPAddresses[address] = &devicetypes.CaniIPAddress{ID: address, Host: "10.0.0.5", Address: "10.0.0.5/16", Parent: prefix}
	inventory.VRFs[vrf] = &devicetypes.CaniVRF{ID: vrf, Name: "blue", Namespace: "Global"}
	return inventory
}

// ipamVerbs builds the parent command each IPAM verb hangs off.
var ipamVerbs = map[string]func() *cli.Command{
	"add":    add.NewCommand,
	"remove": remove.NewCommand,
	"update": update.NewCommand,
}

// ipamFingerprint runs one IPAM verb through a fresh harness and returns a
// UUID-free rendering of the resulting inventory.
func ipamFingerprint(t *testing.T, parent, name string, flags map[string][]string, args ...string) string {
	t.Helper()

	root := ipamVerbs[parent]()
	harness := cmdtest.New(t, root, subcommand(t, root, name), ipamInventory())

	if err := harness.Run(t, flags, args...); err != nil {
		t.Fatalf("%s %s: %v", parent, name, err)
	}
	return cmdtest.Fingerprint(harness.Inventory())
}

// ipamFingerprints renders every IPAM CRUD verb once, under whatever providers
// are currently registered.
func ipamFingerprints(t *testing.T) map[string]string {
	t.Helper()
	return map[string]string{
		"add prefix":    ipamFingerprint(t, "add", "prefix", map[string][]string{"vrf": {"blue"}}, "10.0.1.0/24"),
		"update prefix": ipamFingerprint(t, "update", "prefix", map[string][]string{"vrf": {"blue"}}, "10.0.0.0/16"),
		"remove prefix": ipamFingerprint(t, "remove", "prefix", map[string][]string{"cascade": {"true"}}, "10.0.0.0/16"),
		"remove ip":     ipamFingerprint(t, "remove", "ip", nil, "10.0.0.5"),
		"remove vrf":    ipamFingerprint(t, "remove", "vrf", nil, "blue"),
	}
}

// assertIPAMBaseline guards the IPAM fingerprints against being vacuous: each
// verb must have left the mark its scenario predicts.
func assertIPAMBaseline(t *testing.T, baseline map[string]string) {
	t.Helper()
	for _, guard := range []struct{ verb, want string }{
		{"add prefix", `prefix 10.0.1.0/24`},
		{"add prefix", `"vrfs":["<id>"]`},
		{"update prefix", `"vrfs":["<id>"]`},
		{"remove prefix", `vrf blue`},
		{"remove ip", `prefix 10.0.0.0/16`},
		{"remove vrf", `ip 10.0.0.5/16`},
	} {
		if !strings.Contains(baseline[guard.verb], guard.want) {
			t.Fatalf("%s baseline lacks %s:\n%s", guard.verb, guard.want, baseline[guard.verb])
		}
	}
	for _, gone := range []struct{ verb, text string }{
		{"remove prefix", `prefix 10.0.0.0/16`},
		{"remove prefix", `ip 10.0.0.5/16`},
		{"remove ip", `ip 10.0.0.5/16`},
		{"remove vrf", `vrf blue`},
	} {
		if strings.Contains(baseline[gone.verb], gone.text) {
			t.Fatalf("%s baseline still contains %s:\n%s", gone.verb, gone.text, baseline[gone.verb])
		}
	}
}
