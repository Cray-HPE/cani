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
package add

import (
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// attemptAutoStage tries to stage devices matching the resolved slug via any
// registered provider. It returns true (handled) when one or more devices were
// staged and the inventory was saved.
func attemptAutoStage(inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType, qty int) (bool, error) {
	// Snapshot which devices are already staged before staging.
	alreadyStaged := snapshotStagedDevices(inventory)

	staged := stageDevices(inventory, device, qty)
	if staged == 0 {
		return false, nil
	}

	if err := datastores.Datastore.Save(inventory); err != nil {
		return false, fmt.Errorf("failed to save inventory: %w", err)
	}
	logStagedDevices(inventory, alreadyStaged)
	log.Printf("%d device(s) added", staged)
	return true, nil
}

// snapshotStagedDevices returns the set of device IDs currently staged.
func snapshotStagedDevices(inventory *devicetypes.Inventory) map[uuid.UUID]bool {
	alreadyStaged := make(map[uuid.UUID]bool)
	for id, d := range inventory.Devices {
		if strings.EqualFold(d.Status, string(devicetypes.StatusStaged)) {
			alreadyStaged[id] = true
		}
	}
	return alreadyStaged
}

// stageDevices walks registered providers and stages devices for the slug,
// stopping at the first provider that stages anything.
func stageDevices(inventory *devicetypes.Inventory, device *devicetypes.CaniDeviceType, qty int) int {
	for _, p := range provider.GetProviders() {
		if staged := stageWithProvider(p, inventory, device.Slug, qty); staged > 0 {
			return staged
		}
	}
	return 0
}

// stageWithProvider stages devices using a single provider, preferring staging
// under a new rack (cabinet) and falling back to re-staging existing devices.
func stageWithProvider(p provider.Provider, inventory *devicetypes.Inventory, slug string, qty int) int {
	if rs, ok := p.(provider.RackStager); ok {
		if staged := stageNewInRack(rs, inventory, slug, qty); staged > 0 {
			return staged
		}
	}
	if stager, ok := p.(provider.DeviceStager); ok {
		return stageExistingDevices(stager, inventory, slug, qty)
	}
	return 0
}

// stageNewInRack stages up to qty devices of slug under a new staged rack.
func stageNewInRack(rs provider.RackStager, inventory *devicetypes.Inventory, slug string, qty int) int {
	staged := 0
	for range qty {
		if rs.StageNewInRack(inventory, slug) {
			staged++
		}
	}
	return staged
}

// stageExistingDevices re-stages up to qty already-imported devices of slug.
func stageExistingDevices(stager provider.DeviceStager, inventory *devicetypes.Inventory, slug string, qty int) int {
	staged := 0
	for range qty {
		if stager.StageExisting(inventory, slug) {
			staged++
		}
	}
	return staged
}

// logStagedDevices logs detailed staging info for devices that were
// newly staged (not in alreadyStaged). It prints the display type,
// UUID, and xname components (Cabinet, Chassis, Blade) for each.
func logStagedDevices(inv *devicetypes.Inventory, alreadyStaged map[uuid.UUID]bool) {
	for id, dev := range inv.Devices {
		if alreadyStaged[id] {
			continue
		}
		if !strings.EqualFold(dev.Status, string(devicetypes.StatusStaged)) {
			continue
		}
		log.Printf("Added device %s (%s)", id, dev.Name)
		typeName := displayTypeName(dev.GetType())
		log.Printf("%s was successfully staged to be added to the system", typeName)
		log.Printf("UUID: %s", id)
		for _, p := range provider.GetProviders() {
			describer, ok := p.(provider.StagedDeviceDescriber)
			if !ok {
				continue
			}
			for _, line := range describer.DescribeStagedDevice(dev) {
				log.Printf("%s", line)
			}
		}
	}
}

// displayTypeName returns a human-friendly name for a device type.
func displayTypeName(t devicetypes.Type) string {
	switch devicetypes.Type(strings.ToLower(string(t))) {
	case devicetypes.TypeCabinet:
		return "Cabinet"
	case devicetypes.TypeChassis, devicetypes.TypeRack:
		return "Chassis"
	case devicetypes.TypeBlade:
		return "Blade"
	case devicetypes.TypeNode:
		return "Node"
	case devicetypes.TypeNodeCard:
		return "NodeBlade"
	default:
		s := string(t)
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
}
