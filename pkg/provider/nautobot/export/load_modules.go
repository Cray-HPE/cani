/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	openapi_types "github.com/oapi-codegen/runtime/types"
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

// createModuleFromCani creates a single Nautobot Module from a CaniModuleType.
func (e *Exporter) createModuleFromCani(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	result *LoadResult,
) error {
	// Resolve parent device in Nautobot.
	parentDevice := inventory.Devices[module.ParentDevice]
	if parentDevice == nil {
		return fmt.Errorf("parent device %s not found in inventory", module.ParentDevice)
	}
	parentNautobotID, ok := createdDeviceIDs[parentDevice.Name]
	if !ok {
		return fmt.Errorf("parent device %s not found in Nautobot", parentDevice.Name)
	}

	// Create or look up ModuleType.
	moduleTypeItem, err := e.getOrCreateModuleType(ctx, module)
	if err != nil {
		return fmt.Errorf("module type resolution: %w", err)
	}

	// Create or look up ModuleBay on parent device.
	moduleBayName := module.ModuleBayName
	if moduleBayName == "" {
		moduleBayName = module.Name
	}
	moduleBayItem, err := e.getOrCreateModuleBay(ctx, parentNautobotID, moduleBayName)
	if err != nil {
		return fmt.Errorf("module bay resolution: %w", err)
	}

	// Resolve status.
	statusName := module.Status
	if statusName == "" {
		statusName = e.Options.DefaultStatus
	}
	if statusName == "" {
		statusName = "Active"
	}
	status, err := e.Cache.GetStatus(statusName)
	if err != nil {
		return fmt.Errorf("status resolution: %w", err)
	}

	// Build the Module request.
	req := nautobotapi.ModuleRequest{
		ModuleType:      makeStatusRef(moduleTypeItem.ID),
		ParentModuleBay: makeTenantRef(moduleBayItem.ID),
		Status:          makeStatusRef(status.ID),
	}

	// Optional fields.
	if module.Serial != "" {
		req.Serial = &module.Serial
	}
	if module.AssetTag != "" {
		req.AssetTag = &module.AssetTag
	}
	if module.Role != "" {
		role, err := e.Cache.GetRole(module.Role)
		if err == nil && role != nil {
			req.Role = makeTenantRef(role.ID)
		}
	}
	if module.Location != uuid.Nil {
		// Try to resolve location by looking it up in the inventory.
		if loc, ok := inventory.Locations[module.Location]; ok && loc != nil {
			locItem, err := e.Cache.GetLocation(loc.Name)
			if err == nil && locItem != nil {
				req.Location = makeTenantRef(locItem.ID)
			}
		}
	}

	// Idempotency: check if a module already occupies this bay.
	bayUUID := openapi_types.UUID(moduleBayItem.ID)
	existResp, err := e.Client.DcimModulesListWithResponse(ctx, &nautobotapi.DcimModulesListParams{
		ParentModuleBay: &[]openapi_types.UUID{bayUUID},
	})
	if err == nil && existResp.StatusCode() == http.StatusOK &&
		existResp.JSON200 != nil && existResp.JSON200.Count > 0 {
		clog.Skipped("Module already exists in bay %s on %s — skipping %s",
			moduleBayName, parentDevice.Name, module.Name)
		return nil
	}

	if e.Options.DryRun {
		clog.DryRun("Would create module: %s (parent: %s, bay: %s)",
			module.Name, parentDevice.Name, moduleBayName)
		result.ModulesCreated++
		return nil
	}

	resp, err := e.Client.DcimModulesCreateWithResponse(ctx,
		&nautobotapi.DcimModulesCreateParams{}, req)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s",
			resp.StatusCode(), string(resp.Body))
	}

	clog.Created("Created module: %s (parent: %s, bay: %s)",
		module.Name, parentDevice.Name, moduleBayName)
	result.ModulesCreated++
	return nil
}

// getOrCreateModuleType looks up a ModuleType by model name in Nautobot,
// creating it if not found and create_device_types is enabled.
func (e *Exporter) getOrCreateModuleType(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
) (*CachedItem, error) {
	model := module.Model
	if model == "" {
		model = module.Slug
	}
	if model == "" {
		model = module.Name
	}

	// Search for existing ModuleType by model.
	modelFilter := []string{model}
	resp, err := e.Client.DcimModuleTypesListWithResponse(ctx,
		&nautobotapi.DcimModuleTypesListParams{Model: &modelFilter})
	if err != nil {
		return nil, fmt.Errorf("module type lookup: %w", err)
	}
	if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil &&
		resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		mt := resp.JSON200.Results[0]
		return &CachedItem{
			ID:      toUUID(mt.Id),
			Name:    mt.Model,
			Display: derefString(mt.Display),
		}, nil
	}

	// Not found — create if allowed.
	if !e.Options.CreateModuleTypes {
		return nil, fmt.Errorf("module type %q not in Nautobot (enable create_module_types)", model)
	}

	manufacturer, err := e.Cache.GetOrCreateManufacturer(module.Manufacturer)
	if err != nil {
		return nil, fmt.Errorf("manufacturer resolution: %w", err)
	}

	mfrRef := makeStatusRef(manufacturer.ID)
	createReq := nautobotapi.ModuleTypeRequest{
		Model:        model,
		Manufacturer: mfrRef,
	}
	if module.PartNumber != "" {
		createReq.PartNumber = &module.PartNumber
	}
	if module.Comments != "" {
		createReq.Comments = &module.Comments
	}

	clog.Detail("[nautobot] Creating module type: %s (manufacturer: %s)", model, module.Manufacturer)
	createResp, err := e.Client.DcimModuleTypesCreateWithResponse(ctx,
		&nautobotapi.DcimModuleTypesCreateParams{}, createReq)
	if err != nil {
		return nil, fmt.Errorf("module type create: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("module type create: status %d: %s",
			createResp.StatusCode(), string(createResp.Body))
	}
	if createResp.JSON201 != nil {
		return &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Model,
			Display: derefString(createResp.JSON201.Display),
		}, nil
	}
	return nil, fmt.Errorf("module type create: no response body")
}

// getOrCreateModuleBay looks up a ModuleBay by name on a device, creating it
// if not found.
func (e *Exporter) getOrCreateModuleBay(
	ctx context.Context,
	deviceNautobotID uuid.UUID,
	bayName string,
) (*CachedItem, error) {
	// Search for existing ModuleBay.
	nameFilter := []string{bayName}
	deviceFilter := []string{deviceNautobotID.String()}
	resp, err := e.Client.DcimModuleBaysListWithResponse(ctx,
		&nautobotapi.DcimModuleBaysListParams{
			Name:         &nameFilter,
			ParentDevice: &deviceFilter,
		})
	if err != nil {
		return nil, fmt.Errorf("module bay lookup: %w", err)
	}
	if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil &&
		resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		mb := resp.JSON200.Results[0]
		return &CachedItem{
			ID:      toUUID(mb.Id),
			Name:    mb.Name,
			Display: derefString(mb.Display),
		}, nil
	}

	// Create the module bay.
	createReq := nautobotapi.ModuleBayRequest{
		Name:         bayName,
		ParentDevice: makeTenantRef(deviceNautobotID),
	}

	clog.Detail("[nautobot] Creating module bay: %s on device %s", bayName, deviceNautobotID)
	createResp, err := e.Client.DcimModuleBaysCreateWithResponse(ctx,
		&nautobotapi.DcimModuleBaysCreateParams{}, createReq)
	if err != nil {
		return nil, fmt.Errorf("module bay create: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("module bay create: status %d: %s",
			createResp.StatusCode(), string(createResp.Body))
	}
	if createResp.JSON201 != nil {
		return &CachedItem{
			ID:      toUUID(createResp.JSON201.Id),
			Name:    createResp.JSON201.Name,
			Display: derefString(createResp.JSON201.Display),
		}, nil
	}
	return nil, fmt.Errorf("module bay create: no response body")
}

// derefString safely dereferences a *string, returning "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
