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

	modelFilter := []string{model}
	resp, err := e.Client.DcimModuleTypesListWithResponse(ctx,
		&nautobotapi.DcimModuleTypesListParams{Model: &modelFilter})
	if err != nil {
		return nil, fmt.Errorf("module type lookup: %w", err)
	}
	if resp.StatusCode() == http.StatusOK && resp.JSON200 != nil &&
		resp.JSON200.Results != nil && len(resp.JSON200.Results) > 0 {
		moduleType := resp.JSON200.Results[0]
		return &CachedItem{
			ID: toUUID(moduleType.Id), Name: moduleType.Model,
			Display: derefString(moduleType.Display),
		}, nil
	}
	if !e.Options.CreateModuleTypes {
		return nil, fmt.Errorf("module type %q not in Nautobot (enable create_module_types)", model)
	}

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

func (e *Exporter) createModuleInterfaces(
	ctx context.Context,
	module *devicetypes.CaniModuleType,
	parentNautobotID uuid.UUID,
	result *LoadResult,
) error {
	for _, iface := range module.Interfaces {
		ifaceType := mapInterfaceType(string(iface.Type))
		if !isValidNautobotInterfaceType(ifaceType) {
			continue
		}
		existing, _ := e.Cache.GetInterfaceByDeviceAndName(parentNautobotID, iface.Name)
		if existing != nil {
			continue
		}
		role := iface.Role
		if role == "" {
			mgmtOnly := iface.MgmtOnly != nil && *iface.MgmtOnly
			role = devicetypes.InferInterfaceRole(iface.Name, iface.Type, mgmtOnly)
		}
		spec := interfaceSpec{Name: iface.Name, Type: ifaceType, Role: role}
		if err := e.createInterface(ctx, parentNautobotID, spec, result); err != nil {
			return fmt.Errorf("interface %s: %w", iface.Name, err)
		}
	}
	return nil
}

func isValidNautobotInterfaceType(ifaceType string) bool {
	switch ifaceType {
	case "100base-tx", "1000base-t", "10gbase-x-sfpp", "25gbase-x-sfp28",
		"40gbase-x-qsfpp", "100gbase-x-qsfp28", "200gbase-x-qsfp56",
		"400gbase-x-osfp", "400gbase-x-qsfpdd", "infiniband-hdr",
		"infiniband-ndr", "virtual", "lag", "other":
		return true
	default:
		return false
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
