/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024, 2026 Hewlett Packard Enterprise Development LP
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
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// loadModules exports CaniModuleType records to Nautobot as Module objects.
// For each module it:
//  1. Creates or looks up the ModuleType (template) from the library.
//  2. Creates or looks up the ModuleBay on the parent device.
//  3. Creates the Module instance.
func (e *Exporter) loadModules(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	result *LoadResult,
) error {
	if len(inventory.Modules) == 0 {
		return nil
	}

	for _, module := range inventory.Modules {
		if module == nil || module.Name == "" {
			continue
		}

		if err := e.createModuleFromCani(ctx, module, inventory, createdDeviceIDs, result); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("module %s: create error: %v", module.Name, err))
		}
	}
	return nil
}

// createModuleFromCani exports a single CaniModuleType as a Nautobot Module.
// An installed module is sent with only its parent module bay (its location
// is inherited from the parent); a spare is sent with only its location. A
// module with neither is reported. The module's own ports are created on the
// module so same-named ports on sibling modules stay distinct.
func (e *Exporter) createModuleFromCani(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	result *LoadResult,
) error {
	placement, err := e.resolveModulePlacement(ctx, module, inventory, createdDeviceIDs)
	if err != nil {
		return err
	}
	moduleNautobotID, err := e.findExistingModule(ctx, module, placement)
	if err != nil {
		return fmt.Errorf("existing module lookup: %w", err)
	}
	switch {
	case moduleNautobotID != uuid.Nil:
		clog.Skipped("Skipped module (already exists): %s (%s)", module.Name, placement.describe())
		result.ModulesSkipped++
	case e.Options.DryRun:
		if err := e.previewModuleCreate(ctx, module, placement); err != nil {
			return err
		}
		result.ModulesCreated++
		e.Cache.PlanModule(module.ID)
		e.exportModulePorts(ctx, module, uuid.Nil, result)
		return nil
	default:
		req, err := e.buildModuleRequest(ctx, module, placement)
		if err != nil {
			return err
		}
		if moduleNautobotID, err = e.postModule(ctx, req, module, placement, result); err != nil {
			return err
		}
	}
	setExternalID(&module.ExternalIDs, externalIDKeyNautobot, moduleNautobotID)
	e.Cache.RecordModule(module.ID, moduleNautobotID)
	e.exportModulePorts(ctx, module, moduleNautobotID, result)
	return nil
}

// previewModuleCreate performs the lookups a create would need without
// writing anything: the module type must exist or be creatable, and the
// status must resolve. It then reports the planned module.
func (e *Exporter) previewModuleCreate(ctx context.Context, module *devicetypes.CaniModuleType, placement modulePlacement) error {
	model := moduleTypeModel(module)
	moduleType, err := e.findModuleType(ctx, model)
	if err != nil {
		return fmt.Errorf("module type resolution: %w", err)
	}
	switch {
	case moduleType != nil:
	case !e.Options.CreateModuleTypes:
		return fmt.Errorf("module type resolution: %w", errModuleTypeNotFound(model))
	default:
		clog.DryRun("Would create module type: %s (manufacturer: %s)", model, module.Manufacturer)
	}
	if _, err := e.Cache.GetStatus(e.moduleStatusName(module)); err != nil {
		return fmt.Errorf("status resolution: %w", err)
	}
	clog.DryRun("Would create module: %s (%s)", module.Name, placement.describe())
	return nil
}

// exportModulePorts creates the module's ports on the module, or plans them in
// a dry run, where the module has no Nautobot ID yet. Nautobot does not
// auto-create interfaces for module types created via the API. A failure is
// recorded on the result so the export reports it rather than only logging it.
func (e *Exporter) exportModulePorts(ctx context.Context, module *devicetypes.CaniModuleType, moduleNautobotID uuid.UUID, result *LoadResult) {
	if len(module.Interfaces) == 0 || (moduleNautobotID == uuid.Nil && !e.Options.DryRun) {
		return
	}
	if err := e.createModuleInterfaces(ctx, module, moduleNautobotID, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("module %s: interfaces: %v", module.Name, err))
	}
}

// buildModuleRequest assembles the Module create request with exactly one
// placement reference.
func (e *Exporter) buildModuleRequest(
	ctx context.Context, module *devicetypes.CaniModuleType, placement modulePlacement,
) (nautobotapi.ModuleRequest, error) {
	req := nautobotapi.ModuleRequest{}
	moduleTypeItem, err := e.getOrCreateModuleType(ctx, module)
	if err != nil {
		return req, fmt.Errorf("module type resolution: %w", err)
	}
	status, err := e.Cache.GetStatus(e.moduleStatusName(module))
	if err != nil {
		return req, fmt.Errorf("status resolution: %w", err)
	}
	if err := setRefID(&req.ModuleType, moduleTypeItem.ID); err != nil {
		return req, fmt.Errorf("set module type reference: %w", err)
	}
	if err := setRefID(&req.Status, status.ID); err != nil {
		return req, fmt.Errorf("set module status reference: %w", err)
	}
	if err := placement.setOn(&req); err != nil {
		return req, err
	}
	return req, e.applyModuleOptionalFields(&req, module)
}

// applyModuleOptionalFields maps serial, asset tag, and role when present.
func (e *Exporter) applyModuleOptionalFields(req *nautobotapi.ModuleRequest, module *devicetypes.CaniModuleType) error {
	if module.Serial != "" {
		req.Serial = &module.Serial
	}
	if module.AssetTag != "" {
		req.AssetTag = &module.AssetTag
	}
	if module.Role == "" {
		return nil
	}
	role, err := e.Cache.GetRole(module.Role)
	if err != nil || role == nil {
		return nil
	}
	if err := setRefID(&req.Role, role.ID); err != nil {
		return fmt.Errorf("set module role reference: %w", err)
	}
	return nil
}

// moduleStatusName returns the module's status, the provider default, or Active.
func (e *Exporter) moduleStatusName(module *devicetypes.CaniModuleType) string {
	if module.Status != "" {
		return module.Status
	}
	if e.Options.DefaultStatus != "" {
		return e.Options.DefaultStatus
	}
	return "Active"
}

// postModule sends the module create request and returns the new Nautobot ID.
func (e *Exporter) postModule(
	ctx context.Context,
	req nautobotapi.ModuleRequest,
	module *devicetypes.CaniModuleType,
	placement modulePlacement,
	result *LoadResult,
) (uuid.UUID, error) {
	resp, err := e.Client.DcimModulesCreateWithResponse(ctx,
		&nautobotapi.DcimModulesCreateParams{}, req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return uuid.Nil, fmt.Errorf("unexpected status %d: %s",
			resp.StatusCode(), string(resp.Body))
	}
	var created uuid.UUID
	if resp.JSON201 != nil {
		created = toUUID(resp.JSON201.Id)
	}
	clog.Created("Created module: %s (%s)", module.Name, placement.describe())
	result.ModulesCreated++
	return created, nil
}

// printModuleSummary reports module create/skip counts.
func printModuleSummary(result *LoadResult) {
	if result.ModulesCreated > 0 {
		clog.Created("Created modules: %d", result.ModulesCreated)
	}
	if result.ModulesSkipped > 0 {
		clog.Skipped("Skipped modules (already exist): %d", result.ModulesSkipped)
	}
}
