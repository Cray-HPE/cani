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
	"fmt"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	hpcm_client "github.com/Cray-HPE/cani/pkg/hpcm-client"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TranslateCmConfigToCaniHw translates data from Hpcm.CmConfig to CANI format
func (hpcm *Hpcm) TranslateCmConfig() (translated map[uuid.UUID]*inventory.Hardware, err error) {
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
	for _, d := range hpcm.CmConfig.Discover {
		log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "HPCM", taxonomy.App, d.Hostname1)
		hw := &inventory.Hardware{}

		// translate the hpcm fields to cani fields
		err = TranslateCmHardwareToCaniHw(d, hpcm.CmConfig, hw)
		if err != nil {
			return translated, err
		}

		// add the hardware to the map if it does not exist
		_, exists := translated[hw.ID]
		if exists {
			return translated, fmt.Errorf("Hardware already exists: %s", hw.ID)
		}
		translated[hw.ID] = hw
	}

	// return the map of translated hpcm --> cani hardware
	return translated, nil
}

func TranslateCmHardwareToCaniHw(d Discover, cm HpcmConfig, hw *inventory.Hardware) (err error) {
	// create a uuid for the new hardware
	u := uuid.New()
	log.Debug().Msgf("  Unique Identifier:  --> %s: %+v", "ID", u)
	hw.ID = u

	// Convert HPCM type to cani hardwaretypes
	t, err := hpcmTypeToCaniHardwareType(d.Type)
	if err != nil {
		return err
	}
	hw.Type = t
	log.Debug().Msgf("  type: %s --> %s: %s", d.Type, "Type", t)

	// Convert HPCM template name to cani device type slug
	s := hpcmCmTemplateNameToCaniSlug(d, cm)
	if err != nil {
		return err
	}
	hw.DeviceTypeSlug = s
	log.Debug().Msgf("  template_name: %s --> %s: %s", d.TemplateName, "DeviceTypeSlug", s)

	// Convert HPCM card type to cani vendor
	v := hpcmCmCardTypeToCaniVendor(d, cm)
	if err != nil {
		return err
	}
	hw.Vendor = v
	log.Debug().Msgf("  card_type: %s --> %s: %s", d.CardType, "Vendor", v)

	// Convert HPCM card type to cani vendor
	cmloc := &hpcm_client.LocationSettings{
		Rack:       int32(d.RackNr),
		Chassis:    int32(d.Chassis),
		Tray:       int32(d.Tray),
		Node:       int32(d.NodeNr),
		Controller: int32(d.ControllerNr),
	}
	lp, err := hpcmLocToCaniLoc(hw.Type, cmloc)
	if err != nil {
		return err
	}
	hw.LocationPath = lp
	log.Debug().Msgf("  *_nr: %d->%d->%d->%d --> %s: %s",
		d.RackNr,
		d.Chassis,
		d.Tray,
		d.NodeNr,
		"LocationPath", lp.String())

	// these fields map 1:1 and are not necessarily required, so just fill them
	hw.LocationOrdinal = &d.NodeNr
	log.Debug().Msgf("  node_nr: %d --> %s: %d", d.NodeNr, "LocationOrdinal", *hw.LocationOrdinal)
	hw.Architecture = d.Architecture
	log.Debug().Msgf("  %s: %s %s %s: %s", "architecture", d.Architecture, "-->", "Architecture", hw.Architecture)
	hw.Model = d.TemplateName
	// log.Debug().Msgf("  %s: %s %s %s: %s", "template_name", d.TemplateName, "-->", "Model", hw.Model)
	hw.Name = d.Hostname1
	log.Debug().Msgf("  %s: %s %s %s: %s", "hostname1", d.Hostname1, "-->", "Name", hw.Name)
	log.Debug().Msgf("")

	return nil
}

// hpcmCmCardTypeToCaniVendor
func hpcmCmCardTypeToCaniVendor(d Discover, cm HpcmConfig) (v string) {
	switch d.CardType {
	case "iLo":
		v = "HPE"
	case "Intel":
		v = "Intel"
	case "IPMI":
		v = "IPMI"
	default:
		v = "Unknown"
	}

	return v
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
