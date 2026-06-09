/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package csm

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/csm/commands"
)

// hmnVlanKey is the CSM provider metadata key for a cabinet's HMN VLAN.
const hmnVlanKey = "hmnVlan"

// OnRackAdded implements provider.RackPostAddHook.
// It detects CSM cabinet provider defaults on a rack type and, when
// present, computes the next cabinet number and VLAN, sets the
// appropriate provider metadata on the rack, and logs the results.
func (p *Csm) OnRackAdded(r *devicetypes.CaniRackType, inventory *devicetypes.Inventory) error {
	defaults := commands.DecodeRackCSMDefaults(r.ProviderDefaults)
	if defaults == nil || defaults.Ordinal == 0 {
		return nil
	}

	log.Println("Querying inventory to suggest Cabinet")

	cabinetNum := nextCabinetNumber(inventory, defaults.Ordinal)
	xname := fmt.Sprintf("x%d", cabinetNum)

	vlan := nextVlan(inventory, defaults.StartingHmnVlan, defaults.EndingHmnVlan)

	log.Printf("Suggested cabinet number: %d", cabinetNum)
	log.Printf("Suggested VLAN ID: %d", vlan)

	r.SetProviderMeta("csm", "xname", xname)
	r.SetProviderMeta("csm", "class", defaults.Class)
	if vlan != 0 {
		r.SetProviderMeta("csm", hmnVlanKey, vlan)
	}
	r.Status = string(devicetypes.StatusStaged)
	r.Name = xname

	log.Println("Cabinet was successfully staged to be added to the system")
	log.Printf("Cabinet Number: %d", cabinetNum)

	return nil
}

// nextCabinetNumber finds the next unused cabinet number starting from
// the base ordinal by scanning existing racks for CSM xnames.
func nextCabinetNumber(inventory *devicetypes.Inventory, baseOrdinal int) int {
	used := make(map[int]bool)
	for _, rack := range inventory.Racks {
		if sub, ok := rack.GetProviderSubMap("csm"); ok {
			recordCabinetXnameOrdinal(sub, used)
		}
	}
	// Also check devices for backwards compatibility with imports
	// that create Device entries for cabinets.
	for _, dev := range inventory.Devices {
		if sub, ok := dev.GetProviderSubMap("csm"); ok {
			recordCabinetXnameOrdinal(sub, used)
		}
	}

	for n := baseOrdinal; ; n++ {
		if !used[n] {
			return n
		}
	}
}

// recordCabinetXnameOrdinal extracts the numeric ordinal from a CSM
// sub-map's xname (e.g. "x1000") and marks it used.
func recordCabinetXnameOrdinal(sub map[string]any, used map[int]bool) {
	x, _ := sub["xname"].(string)
	if len(x) <= 1 || x[0] != 'x' {
		return
	}
	var num int
	if _, err := fmt.Sscanf(x, "x%d", &num); err == nil {
		used[num] = true
	}
}

// nextVlan finds the next unused HMN VLAN in the given range by
// scanning existing racks.
func nextVlan(inventory *devicetypes.Inventory, startVlan, endVlan int) int {
	if startVlan == 0 {
		return 0
	}
	used := make(map[int]bool)
	for _, rack := range inventory.Racks {
		sub, ok := rack.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		if v := intFromAny(sub[hmnVlanKey]); v != 0 {
			used[v] = true
		}
	}
	for _, dev := range inventory.Devices {
		sub, ok := dev.GetProviderSubMap("csm")
		if !ok {
			continue
		}
		if v := intFromAny(sub[hmnVlanKey]); v != 0 {
			used[v] = true
		}
	}

	for v := startVlan; v <= endVlan; v++ {
		if !used[v] {
			return v
		}
	}
	return 0
}

// intFromAny converts a numeric interface{} to int.
func intFromAny(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}
