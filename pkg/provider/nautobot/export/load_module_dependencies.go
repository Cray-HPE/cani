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
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// moduleTypeModel is the model the module phase looks a module type up by.
func moduleTypeModel(module *devicetypes.CaniModuleType) string {
	return firstSet(module.Model, module.Slug, module.Name)
}

// findModuleType looks a module type up by model; nil means Nautobot has
// none, which is not an error.
func (e *Exporter) findModuleType(ctx context.Context, model string) (*CachedItem, error) {
	modelFilter := []string{model}
	resp, err := e.Client.DcimModuleTypesListWithResponse(ctx,
		&nautobotapi.DcimModuleTypesListParams{Model: &modelFilter})
	if err != nil {
		return nil, fmt.Errorf("module type lookup: %w", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil ||
		resp.JSON200.Results == nil || len(resp.JSON200.Results) == 0 {
		return nil, nil
	}
	moduleType := resp.JSON200.Results[0]
	return &CachedItem{
		ID: toUUID(moduleType.Id), Name: moduleType.Model,
		Display: derefString(moduleType.Display),
	}, nil
}

// errModuleTypeNotFound is the module phase's refusal when a module type is
// absent and the export may not create it.
func errModuleTypeNotFound(model string) error {
	return fmt.Errorf("module type %q not in Nautobot (enable create_module_types)", model)
}

func (e *Exporter) getOrCreateModuleType(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
) (*CachedItem, error) {
	model := moduleTypeModel(module)
	moduleType, err := e.findModuleType(ctx, model)
	if err != nil {
		return nil, err
	}
	if moduleType != nil {
		return moduleType, nil
	}
	if !e.Options.CreateModuleTypes {
		return nil, errModuleTypeNotFound(model)
	}
	return e.createModuleType(ctx, module, model)
}

// createModuleType creates the module type under the module's manufacturer,
// creating that too when Nautobot lacks it.
func (e *Exporter) createModuleType(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	model string,
) (*CachedItem, error) {
	manufacturer, err := e.Cache.GetOrCreateManufacturer(module.Manufacturer)
	if err != nil {
		return nil, fmt.Errorf("manufacturer resolution: %w", err)
	}
	request := nautobotapi.ModuleTypeRequest{Model: model}
	if err := setRefID(&request.Manufacturer, manufacturer.ID); err != nil {
		return nil, fmt.Errorf("set module type manufacturer reference: %w", err)
	}
	if module.PartNumber != "" {
		request.PartNumber = &module.PartNumber
	}
	if module.Comments != "" {
		request.Comments = &module.Comments
	}

	clog.Detail("[nautobot] Creating module type: %s (manufacturer: %s)", model, module.Manufacturer)
	createResp, err := e.Client.DcimModuleTypesCreateWithResponse(ctx,
		&nautobotapi.DcimModuleTypesCreateParams{}, request)
	if err != nil {
		return nil, fmt.Errorf("module type create: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("module type create: status %d: %s",
			createResp.StatusCode(), string(createResp.Body))
	}
	if createResp.JSON201 == nil {
		return nil, fmt.Errorf("module type create: no response body")
	}
	return &CachedItem{
		ID: toUUID(createResp.JSON201.Id), Name: createResp.JSON201.Model,
		Display: derefString(createResp.JSON201.Display),
	}, nil
}

func (e *Exporter) getOrCreateModuleBay(
	ctx context.Context,
	deviceNautobotID uuid.UUID,
	bayName string,
) (*CachedItem, error) {
	nameFilter := []string{bayName}
	deviceFilter := []string{deviceNautobotID.String()}
	resp, err := e.Client.DcimModuleBaysListWithResponse(ctx,
		&nautobotapi.DcimModuleBaysListParams{Name: &nameFilter, ParentDevice: &deviceFilter})
	if err != nil {
		return nil, fmt.Errorf("module bay lookup: %w", err)
	}
	if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil &&
		resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		moduleBay := resp.JSON200.Results[0]
		return &CachedItem{
			ID: toUUID(moduleBay.Id), Name: moduleBay.Name,
			Display: derefString(moduleBay.Display),
		}, nil
	}

	request := nautobotapi.ModuleBayRequest{Name: bayName}
	if err := setRefID(&request.ParentDevice, deviceNautobotID); err != nil {
		return nil, fmt.Errorf("set module bay parent device reference: %w", err)
	}
	createResp, err := e.Client.DcimModuleBaysCreateWithResponse(ctx,
		&nautobotapi.DcimModuleBaysCreateParams{}, request)
	if err != nil {
		return nil, fmt.Errorf("module bay create: %w", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("module bay create: status %d: %s",
			createResp.StatusCode(), string(createResp.Body))
	}
	if createResp.JSON201 == nil {
		return nil, fmt.Errorf("module bay create: no response body")
	}
	return &CachedItem{
		ID: toUUID(createResp.JSON201.Id), Name: createResp.JSON201.Name,
		Display: derefString(createResp.JSON201.Display),
	}, nil
}

// createModuleInterfaces creates the module's ports on the parent device. A
// port whose type Nautobot lacks is skipped and counted by
// supportedInterfaceSpecs; a port the parent already has is accounted for
// rather than silently dropped.
func (e *Exporter) createModuleInterfaces(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	parent *devicetypes.CaniDeviceType,
	parentNautobotID uuid.UUID,
	result *LoadResult,
) error {
	for _, spec := range supportedInterfaceSpecs(moduleInterfaceSpecs(module), module.Name, result) {
		existing, _ := e.Cache.GetInterfaceByDeviceAndName(parentNautobotID, spec.Name)
		if existing != nil {
			noteExistingModuleInterface(result, module, parent, spec.Name)
			continue
		}
		if err := e.createInterface(ctx, parentNautobotID, spec, result); err != nil {
			return fmt.Errorf("interface %s: %w", spec.Name, err)
		}
	}
	return nil
}

// moduleInterfaceSpecs maps a module's library interfaces to export specs with
// the same attribute set device-level interfaces carry.
func moduleInterfaceSpecs(module *devicetypes.CaniModuleType) []interfaceSpec {
	specs := make([]interfaceSpec, 0, len(module.Interfaces))
	for _, iface := range module.Interfaces {
		specs = append(specs, specFromInterface(iface))
	}
	return specs
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
