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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	openapi_types "github.com/Cray-HPE/cani/internal/openapi/types"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// findExistingModule returns the Nautobot module that already represents
// this one, or nil, and whether it already sits where the inventory places
// it. Identity comes first: the recorded external ID names the module
// wherever Nautobot now holds it, so a module the inventory moved is found
// and moved rather than matched to whatever occupies its new bay. An
// external ID Nautobot no longer has is dropped with a warning. Without an
// identity, an installed module is found through the occupant of its bay,
// which sits there by construction and where a different known serial is a
// conflict rather than a match, and a spare through its serial. A bay that
// does not exist yet has no occupant.
func (e *Exporter) findExistingModule(
	ctx context.Context, module *devicetypes.CaniModuleType, placement modulePlacement,
) (remote *nautobotapi.Module, current bool, err error) {
	if known := module.ExternalIDs[externalIDKeyNautobot]; known != uuid.Nil {
		remote, err = e.retrieveModule(ctx, known)
		if err != nil || remote != nil {
			return remote, remote != nil && placementCurrent(remote, placement), err
		}
		clog.Warn("Stored Nautobot ID %s for module %q no longer exists; re-resolving by placement", known, module.Name)
		delete(module.ExternalIDs, externalIDKeyNautobot)
	}
	switch {
	case placement.planned:
		return nil, false, nil
	case placement.installed():
		remote, err = e.findModuleInBay(ctx, module, placement)
		return remote, remote != nil, err
	case module.Serial == "":
		return nil, false, nil
	}
	found, err := e.listModules(ctx, &nautobotapi.DcimModulesListParams{Serial: &[]string{module.Serial}})
	if err != nil || len(found) == 0 {
		return nil, false, err
	}
	return &found[0], placementCurrent(&found[0], placement), nil
}

// findModuleInBay returns the module occupying the placement's bay, treating
// an occupant with a different known serial as a conflict.
func (e *Exporter) findModuleInBay(
	ctx context.Context, module *devicetypes.CaniModuleType, placement modulePlacement,
) (*nautobotapi.Module, error) {
	occupant, err := e.bayOccupant(ctx, placement.bay.ID)
	if err != nil || occupant == nil {
		return nil, err
	}
	if module.Serial != "" && derefString(occupant.Serial) != "" && derefString(occupant.Serial) != module.Serial {
		return nil, fmt.Errorf("bay %q on %s is occupied by module serial %q, not %q",
			placement.bayName, placement.parentName, derefString(occupant.Serial), module.Serial)
	}
	return occupant, nil
}

// bayOccupant returns the module installed in the bay, or nil.
func (e *Exporter) bayOccupant(ctx context.Context, bayID uuid.UUID) (*nautobotapi.Module, error) {
	bayUUID := openapi_types.UUID(bayID)
	found, err := e.listModules(ctx, &nautobotapi.DcimModulesListParams{ParentModuleBay: &[]openapi_types.UUID{bayUUID}})
	if err != nil || len(found) == 0 {
		return nil, err
	}
	return &found[0], nil
}

// retrieveModule fetches one module by ID; a 404 returns nil, nil so a stale
// recorded ID is distinguishable from a failed request.
func (e *Exporter) retrieveModule(ctx context.Context, id uuid.UUID) (*nautobotapi.Module, error) {
	resp, err := e.Client.DcimModulesRetrieveWithResponse(ctx, openapi_types.UUID(id), &nautobotapi.DcimModulesRetrieveParams{})
	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, nil
	}
	return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode(), string(resp.Body))
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

// placementCurrent reports whether the remote module already sits where the
// inventory places it. A planned bay does not exist yet, so nothing sits in
// it.
func placementCurrent(remote *nautobotapi.Module, placement modulePlacement) bool {
	switch {
	case placement.planned:
		return false
	case placement.installed():
		return remote.ParentModuleBay != nil && refID(remote.ParentModuleBay.Id) == placement.bay.ID
	}
	return remote.Location != nil && refID(remote.Location.Id) == placement.location.ID
}

// moveModule places a module Nautobot already holds where the inventory now
// places it. Without --merge the difference is reported and counted as not
// exported as authored, as a device whose fields differ is. Nautobot holds
// one module per bay, so a bay another module still occupies is refused;
// loadModules exports spares first so the module a swap stores vacates its
// bay before the module the swap installs claims it.
func (e *Exporter) moveModule(
	ctx context.Context, module *devicetypes.CaniModuleType, remote *nautobotapi.Module,
	placement modulePlacement, result *LoadResult,
) error {
	if !e.Options.Merge {
		clog.Skipped("Skipped module (placement differs; use --merge to move): %s (%s)", module.Name, placement.describe())
		result.ModulesMisplaced++
		return nil
	}
	if e.Options.DryRun {
		clog.DryRun("Would move module: %s (%s)", module.Name, placement.describe())
		result.ModulesMoved++
		return nil
	}
	moduleID := toUUID(remote.Id)
	if err := e.requireFreeBay(ctx, moduleID, placement); err != nil {
		return err
	}
	body, err := json.Marshal(placementPatch(placement))
	if err != nil {
		return fmt.Errorf("marshal placement: %w", err)
	}
	resp, err := e.Client.DcimModulesPartialUpdateWithBodyWithResponse(ctx, openapi_types.UUID(moduleID),
		&nautobotapi.DcimModulesPartialUpdateParams{}, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("API error moving module: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d moving module: %s", resp.StatusCode(), string(resp.Body))
	}
	clog.Changed("Moved module: %s (%s)", module.Name, placement.describe())
	result.ModulesMoved++
	return nil
}

// requireFreeBay refuses a bay that holds a module other than the one moving
// into it.
func (e *Exporter) requireFreeBay(ctx context.Context, moduleID uuid.UUID, placement modulePlacement) error {
	if !placement.installed() {
		return nil
	}
	occupant, err := e.bayOccupant(ctx, placement.bay.ID)
	if err != nil {
		return fmt.Errorf("bay occupancy lookup: %w", err)
	}
	if occupant == nil || toUUID(occupant.Id) == moduleID {
		return nil
	}
	return fmt.Errorf("bay %q on %s is occupied by module %s; store or move that module first",
		placement.bayName, placement.parentName, toUUID(occupant.Id))
}

// placementPatch is the PATCH body for a placement. The generated request
// omits nil references, and Nautobot requires the placement field a module
// leaves to be cleared in the same write as the one it takes.
func placementPatch(placement modulePlacement) map[string]any {
	if placement.installed() {
		return map[string]any{"parent_module_bay": map[string]string{"id": placement.bay.ID.String()}, "location": nil}
	}
	return map[string]any{"parent_module_bay": nil, "location": map[string]string{"id": placement.location.ID.String()}}
}
