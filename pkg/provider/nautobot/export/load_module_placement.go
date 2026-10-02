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
	"errors"
	"fmt"
	"net/http"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// modulePlacement is where a module goes in Nautobot: a bay on an installed
// parent, or a location for a spare. Exactly one of bay or location is set,
// except in a dry run, where an installed module whose bay Nautobot lacks is
// planned: the bay is nil and no write may create it.
type modulePlacement struct {
	parentName string
	bayName    string
	bay        *CachedItem
	location   *CachedItem
	planned    bool
}

func (p modulePlacement) installed() bool { return p.bay != nil || p.planned }

func (p modulePlacement) describe() string {
	switch {
	case p.planned:
		return fmt.Sprintf("parent: %s, bay: %s (to be created)", p.parentName, p.bayName)
	case p.installed():
		return fmt.Sprintf("parent: %s, bay: %s", p.parentName, p.bayName)
	}
	return fmt.Sprintf("spare at location: %s", p.location.Name)
}

// setOn writes the single placement reference onto the request.
func (p modulePlacement) setOn(req *nautobotapi.ModuleRequest) error {
	if p.planned {
		return errors.New("a planned module bay has no Nautobot ID to reference")
	}
	if p.installed() {
		if err := setRefID(&req.ParentModuleBay, p.bay.ID); err != nil {
			return fmt.Errorf("set parent module bay reference: %w", err)
		}
		return nil
	}
	if err := setRefID(&req.Location, p.location.ID); err != nil {
		return fmt.Errorf("set module location reference: %w", err)
	}
	return nil
}

// resolveModulePlacement decides whether the module is installed or a spare
// and resolves the remote objects that placement needs. An installed module
// never sends its location: Nautobot derives it from the parent, and a legacy
// authored location that matches the parent must not become a second field.
// A spare's location is resolved as a device's is: the authored location, or
// its deepest descendant whose type admits modules, then the default location.
func (e *Exporter) resolveModulePlacement(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
) (modulePlacement, error) {
	if module.ParentDevice != uuid.Nil {
		return e.resolveInstalledPlacement(ctx, module, inventory, createdDeviceIDs)
	}
	if module.Location == uuid.Nil {
		return modulePlacement{}, fmt.Errorf("module has neither a parent device nor a location; set one before exporting")
	}
	if inventory.Locations[module.Location] == nil {
		return modulePlacement{}, fmt.Errorf("location %s not found in inventory", module.Location)
	}
	name := resolveContentLocation(module.Location, "module", inventory)
	if name == "" {
		name = e.Options.DefaultLocation
	}
	if name == "" {
		return modulePlacement{}, fmt.Errorf("location %q cannot hold modules and no --default-location is set",
			inventory.Locations[module.Location].Name)
	}
	locItem, err := e.Cache.GetLocation(name)
	if err != nil || locItem == nil {
		return modulePlacement{}, fmt.Errorf("location %q not found in Nautobot: %v", name, err)
	}
	return modulePlacement{location: locItem}, nil
}

// resolveInstalledPlacement finds or creates the parent's module bay for an
// installed module. A module with no named bay falls back to its own name as
// the bay only on a parent that declares no module bays; on a parent that
// declares them, the fallback would create a bay the parent's device type does
// not declare, so the module is refused with the remedy instead. A parent a
// dry run only plans has no Nautobot ID, so the placement is planned too.
func (e *Exporter) resolveInstalledPlacement(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
) (modulePlacement, error) {
	parentDevice := inventory.Devices[module.ParentDevice]
	if parentDevice == nil {
		return modulePlacement{}, fmt.Errorf("parent device %s not found in inventory", module.ParentDevice)
	}
	if module.ModuleBayName == "" && inventory.DeclaresModuleBays(parentDevice.ID) {
		return modulePlacement{}, fmt.Errorf(
			"module has no named bay in %s, which declares module bays; set one with `update module --bay`", parentDevice.Name)
	}
	placement := modulePlacement{parentName: parentDevice.Name, bayName: module.EffectiveModuleBayName()}
	if e.Options.DryRun && e.Cache.DevicePlanned(parentDevice.Name) {
		clog.DryRun("Would create module bay %q on %s", placement.bayName, parentDevice.Name)
		placement.planned = true
		return placement, nil
	}
	parentNautobotID, ok := createdDeviceIDs[parentDevice.Name]
	if !ok {
		return modulePlacement{}, fmt.Errorf("parent device %s not found in Nautobot", parentDevice.Name)
	}
	return e.resolveInstalledBay(ctx, parentNautobotID, placement)
}

// resolveInstalledBay finds the placement's bay on the parent, plans it in a
// dry run when Nautobot lacks it, and otherwise creates it.
func (e *Exporter) resolveInstalledBay(ctx context.Context, parentNautobotID uuid.UUID, placement modulePlacement) (modulePlacement, error) {
	bay, err := e.findModuleBay(ctx, parentNautobotID, placement.bayName)
	if err != nil {
		return modulePlacement{}, fmt.Errorf("module bay resolution: %w", err)
	}
	if bay == nil && e.Options.DryRun {
		clog.DryRun("Would create module bay %q on %s", placement.bayName, placement.parentName)
		placement.planned = true
		return placement, nil
	}
	if bay == nil {
		if bay, err = e.createModuleBay(ctx, parentNautobotID, placement.bayName); err != nil {
			return modulePlacement{}, fmt.Errorf("module bay resolution: %w", err)
		}
	}
	placement.bay = bay
	return placement, nil
}

// findExistingModule returns the Nautobot ID of the module that already
// represents this one, or uuid.Nil. An installed module is found through its
// bay; a different module in that bay (by serial, when both are known) is a
// conflict rather than a match. A bay that does not exist yet has no
// occupant. A spare is found through its recorded external ID, then its
// serial, because it has no bay to anchor it.
func (e *Exporter) findExistingModule(
	ctx context.Context, module *devicetypes.CaniModuleType, placement modulePlacement,
) (uuid.UUID, error) {
	if placement.planned {
		return uuid.Nil, nil
	}
	if placement.installed() {
		return e.findModuleInBay(ctx, module, placement)
	}
	if known := module.ExternalIDs[externalIDKeyNautobot]; known != uuid.Nil {
		if found, err := e.listModules(ctx, &nautobotapi.DcimModulesListParams{Id: &[]openapi_types.UUID{openapi_types.UUID(known)}}); err != nil || len(found) > 0 {
			return known, err
		}
	}
	if module.Serial == "" {
		return uuid.Nil, nil
	}
	found, err := e.listModules(ctx, &nautobotapi.DcimModulesListParams{Serial: &[]string{module.Serial}})
	if err != nil || len(found) == 0 {
		return uuid.Nil, err
	}
	return toUUID(found[0].Id), nil
}

// findModuleInBay returns the module occupying the placement's bay, treating
// an occupant with a different known serial as a conflict.
func (e *Exporter) findModuleInBay(
	ctx context.Context, module *devicetypes.CaniModuleType, placement modulePlacement,
) (uuid.UUID, error) {
	bayUUID := openapi_types.UUID(placement.bay.ID)
	found, err := e.listModules(ctx, &nautobotapi.DcimModulesListParams{ParentModuleBay: &[]openapi_types.UUID{bayUUID}})
	if err != nil || len(found) == 0 {
		return uuid.Nil, err
	}
	occupant := found[0]
	if module.Serial != "" && derefString(occupant.Serial) != "" && derefString(occupant.Serial) != module.Serial {
		return uuid.Nil, fmt.Errorf("bay %q on %s is occupied by module serial %q, not %q",
			placement.bayName, placement.parentName, derefString(occupant.Serial), module.Serial)
	}
	return toUUID(occupant.Id), nil
}

// listModules runs one module list query and returns its results.
func (e *Exporter) listModules(ctx context.Context, params *nautobotapi.DcimModulesListParams) ([]nautobotapi.Module, error) {
	resp, err := e.Client.DcimModulesListWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	if resp.JSON200 == nil {
		return nil, nil
	}
	return resp.JSON200.Results, nil
}
