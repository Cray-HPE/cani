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
package hpcm

import (
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	hpcm_client "github.com/Cray-HPE/cani/pkg/hpcm-client"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TranslateCmConfigToCaniHw translates data from Hpcm.CmConfig to CANI format
func (cm HpcmConfig) TranslateCmConfig() (translated map[uuid.UUID]*inventory.Hardware, err error) {
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
	for _, d := range cm.Discover {
		log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "HPCM", taxonomy.App, d.Hostname1)

		_ = d.getModel()
		// 	// translate the hpcm fields to cani fields
		// 	t, err := TranslateCmHardwareToCaniHw(d, cm)
		// 	if err != nil {
		// 		return translated, err
		// 	}

		// 	// add the hardware to the map if it does not exist
		// 	_, exists := translated[t.ID]
		// 	if exists {
		// 		return translated, fmt.Errorf("Hardware already exists: %s", hw.ID)
		// 	}
		// 	translated[hw.ID] = hw
	}

	// return the map of translated hpcm --> cani hardware
	return translated, nil
}

func (cm HpcmConfig) TranslateCmHardwareToCaniHw() (translated map[uuid.UUID]*inventory.Hardware, err error) {
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)

	for _, d := range cm.Discover {
		hw := &inventory.Hardware{}
		hw.ID = uuid.New()

		// Convert HPCM type to cani hardwaretypes
		t, err := hpcmTypeToCaniHardwareType(d.Type)
		if err != nil {
			return translated, err
		}
		hw.Type = t

		// Convert HPCM location to cani location
		cmloc := &hpcm_client.LocationSettings{
			Rack:       int32(d.RackNr),
			Chassis:    int32(d.Chassis),
			Tray:       int32(d.Tray),
			Controller: int32(d.ControllerNr),
			Node:       int32(d.NodeNr),
		}

		lp, err := hpcmLocToCaniLoc(hw.Type, cmloc)
		if err != nil {
			return translated, err
		}

		hw.LocationPath = lp
		hw.Architecture = d.Architecture
		hw.Model = d.TemplateName
		hw.Name = d.Hostname1

		_, exists := translated[hw.ID]
		if !exists {
			translated[hw.ID] = hw
		}
	}
	return translated, nil
}

// getModel messily gets the model from a Discover object
func (d Discover) getModel() (model string) {
	log.Info().Msgf("%+v", d)
	// md := make(map[string]interface{}, 0)
	// md["Inventory"] = make(map[string]interface{}, 0)
	// mdInv := md["Inventory"].(map[string]interface{})
	// if node.Inventory != nil {
	// 	inv := *node.Inventory
	// 	for k, v := range inv.(map[string]interface{}) {
	// 		if k == "fru.Model" {
	// 			vendor = v.(string)
	// 		}
	// 		mdInv[k] = v
	// 	}
	// }
	return model
}

// hpcmCmTemplateNameToCaniSlug
func hpcmCmTemplateNameToCaniSlug(d Discover, cm HpcmConfig) (t string) {
	switch d.Type {
	case "":
		// tpl := getDiscoverTemplate(d, cm)
		// t = d.TemplateName
	default:
		t = d.TemplateName
	}

	return t
}
