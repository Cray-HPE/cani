/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package hpengi

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/hpcm"
	"github.com/Cray-HPE/cani/pkg/canu"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ImportInit imports the external inventory data into CANI's inventory format
func (hpengi *Hpengi) ImportInit(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	log.Warn().Msgf("ImportInit partially implemented")
	// copy the datastore and add set provider metadata
	ds, err := setupTempDatastore(datastore)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("cm-config") {
		f, _ := cmd.Flags().GetString("cm-config")
		cm, err := hpcm.LoadCmConfig(f)
		if err != nil {
			return err
		}

		translated, err = cm.TranslateCmHardwareToCaniHw()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("sls-config") {
		// get a map of translated hardware from the hpcm config
		translated, err = hpengi.translateSlsDumpstateToCaniHw(cmd, args)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("paddle") {
		// get a map of translated hardware from the hpcm config
		translated, err = hpengi.translatePaddleToCaniHw(cmd, args)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("cmdb") {
		// translate external inventory data to cani hardware entries
		translated, err = hpcmObj.Translate(cmd, args)
		if err != nil {
			return err
		}
	}

	// all flags will return a map of translated hardware
	// loop through that, and add it each to the datastore
	for _, hw := range translated {
		err = ds.Add(hw)
		if err != nil {
			return err
		}
	}

	// merge the datastore if all is successful
	err = datastore.Merge(ds)
	if err != nil {
		return err
	}

	return datastore.Flush()
}

// Import
func (hpengi *Hpengi) Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	log.Warn().Msgf("Import not yet implemented")
	return nil
}

// translateSlsDumpstateToCaniHw
func (hpengi *Hpengi) translateSlsDumpstateToCaniHw(cmd *cobra.Command, args []string) (translated map[uuid.UUID]*inventory.Hardware, err error) {
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
	for _, s := range hpengi.SlsInput.Hardware {
		// log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "CANU", taxonomy.App, *n.CommonName)
		hw := &inventory.Hardware{}
		hw.Name = s.Xname
		// // translate the hpcm fields to cani fields
		// err = translatePaddleHardwareToCaniHw(n, *hpengi.Paddle, hw)
		// if err != nil {
		// 	return translated, err
		// }

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

// translatePaddleToCaniHw
func (hpengi *Hpengi) translatePaddleToCaniHw(cmd *cobra.Command, args []string) (translated map[uuid.UUID]*inventory.Hardware, err error) {
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
	for _, n := range hpengi.Paddle.Topology {
		log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "CANU", taxonomy.App, *n.CommonName)
		hw := &inventory.Hardware{}

		// translate the hpcm fields to cani fields
		err = translatePaddleHardwareToCaniHw(n, *hpengi.Paddle, hw)
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

// func (hpengi *Hpengi) translateCmConfigToCaniHw(cmd *cobra.Command, args []string) (translated map[uuid.UUID]*inventory.Hardware, err error) {
// 	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
// 	for _, d := range hpengi.CmConfig.Discover {
// 		log.Debug().Msgf("Translating %s hardware to %s format (%+v)", "HPCM", taxonomy.App, d.Hostname1)
// 		hw := &inventory.Hardware{}

// 		// translate the hpcm fields to cani fields
// 		err = translateCmHardwareToCaniHw(d, hpengi.CmConfig, hw)
// 		if err != nil {
// 			return translated, err
// 		}

// 		// add the hardware to the map if it does not exist
// 		_, exists := translated[hw.ID]
// 		if exists {
// 			return translated, fmt.Errorf("Hardware already exists: %s", hw.ID)
// 		}
// 		translated[hw.ID] = hw
// 	}

// 	// return the map of translated hpcm --> cani hardware
// 	return translated, nil
// }

func translatePaddleHardwareToCaniHw(n canu.PaddleTopologyElem, ccj canu.Paddle, hw *inventory.Hardware) (err error) {
	// create a uuid for the new hardware
	u := uuid.New()
	log.Debug().Msgf("  Unique Identifier:  --> %s: %+v", "ID", u)
	hw.ID = u

	hw.Architecture = *n.Architecture
	log.Debug().Msgf("  %s: %s %s %s: %s", "Architecture", *n.Architecture, "-->", "Architecture", hw.Architecture)
	hw.Model = *n.Model
	log.Debug().Msgf("  %s: %s %s %s: %s", "Model", *n.Model, "-->", "Model", hw.Model)
	hw.Name = *n.CommonName
	log.Debug().Msgf("  %s: %s %s %s: %s", "CommonName", *n.CommonName, "-->", "Name", hw.Name)
	hw.Vendor = *n.Vendor
	log.Debug().Msgf("  %s: %s %s %s: %s", "Vendor", *n.Vendor, "-->", "Name", hw.Vendor)
	log.Debug().Msgf("")

	return nil
}

// func translateCmHardwareToCaniHw(d hpcm.Discover, cm hpcm.HpcmConfig, hw *inventory.Hardware) (err error) {
// 	// create a uuid for the new hardware
// 	u := uuid.New()
// 	log.Debug().Msgf("  Unique Identifier:  --> %s: %+v", "ID", u)
// 	hw.ID = u

// 	// Convert HPCM type to cani hardwaretypes
// 	t := hpcmTypeToCaniType(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.Type = t
// 	log.Debug().Msgf("  type: %s --> %s: %s", d.Type, "Type", t)

// 	// Convert HPCM template name to cani device type slug
// 	s := hpcmTemplateNameToCaniSlug(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.DeviceTypeSlug = s
// 	log.Debug().Msgf("  template_name: %s --> %s: %s", d.TemplateName, "DeviceTypeSlug", s)

// 	// Convert HPCM card type to cani vendor
// 	v := hpcmCardTypeToCaniVendor(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.Vendor = v
// 	log.Debug().Msgf("  card_type: %s --> %s: %s", d.CardType, "Vendor", v)

// 	// Convert HPCM card type to cani vendor
// 	lp := hpcmGeoCaniLocationPath(d, cm)
// 	if err != nil {
// 		return err
// 	}
// 	hw.LocationPath = lp
// 	log.Debug().Msgf("  *_nr: %d->%d->%d->%d --> %s: %s",
// 		d.RackNr,
// 		d.Chassis,
// 		d.Tray,
// 		d.NodeNr,
// 		"LocationPath", lp.String())

// 	// these fields map 1:1 and are not necessarily required, so just fill them
// 	hw.LocationOrdinal = &d.NodeNr
// 	log.Debug().Msgf("  node_nr: %d --> %s: %d", d.NodeNr, "LocationOrdinal", *hw.LocationOrdinal)
// 	hw.Architecture = d.Architecture
// 	log.Debug().Msgf("  %s: %s %s %s: %s", "architecture", d.Architecture, "-->", "Architecture", hw.Architecture)
// 	hw.Model = d.TemplateName
// 	// log.Debug().Msgf("  %s: %s %s %s: %s", "template_name", d.TemplateName, "-->", "Model", hw.Model)
// 	hw.Name = d.Hostname1
// 	log.Debug().Msgf("  %s: %s %s %s: %s", "hostname1", d.Hostname1, "-->", "Name", hw.Name)
// 	log.Debug().Msgf("")

// 	return nil
// }

// // hpcmGeoCaniLocationPath
// func hpcmGeoCaniLocationPath(d hpcm.Discover, cm hpcm.HpcmConfig) (lp inventory.LocationPath) {
// 	lp = inventory.LocationPath{
// 		inventory.LocationToken{HardwareType: hardwaretypes.Cabinet, Ordinal: d.RackNr},
// 		inventory.LocationToken{HardwareType: hardwaretypes.Chassis, Ordinal: d.Chassis},
// 		inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: d.Tray},
// 		inventory.LocationToken{HardwareType: hardwaretypes.Node, Ordinal: d.NodeNr},
// 	}
// 	return lp
// }

// // hpcmCardTypeToCaniVendor
// func hpcmCardTypeToCaniVendor(d hpcm.Discover, cm hpcm.HpcmConfig) (v string) {
// 	switch d.CardType {
// 	case "iLo":
// 		v = "HPE"
// 	case "Intel":
// 		v = "Intel"
// 	case "IPMI":
// 		v = "IPMI"
// 	default:
// 		v = "Unknown"
// 	}

// 	return v
// }

// // hpcmTemplateNameToCaniSlug
// func hpcmTemplateNameToCaniSlug(d hpcm.Discover, cm hpcm.HpcmConfig) (t string) {
// 	switch d.Type {
// 	case "":
// 		// tpl := getDiscoverTemplate(d, cm)
// 		// t = d.TemplateName
// 	default:
// 		t = d.TemplateName
// 	}

// 	return t
// }

// // hpcmTypeToCaniType
// func hpcmTypeToCaniType(d hpcm.Discover, cm hpcm.HpcmConfig) (t hardwaretypes.HardwareType) {
// 	switch d.Type {
// 	case "":
// 		// tpl := getDiscoverTemplate(d, cm)
// 		t = hardwaretypes.NodeBlade
// 	case "leaf", "spine":
// 		t = hardwaretypes.ManagementSwitch
// 	}

// 	return t
// }

// // getDiscoverTemplate
// func getDiscoverTemplate(d hpcm.Discover, cm hpcm.HpcmConfig) (tpl hpcm.Template) {
// 	val, exists := cm.Templates[d.TemplateName]
// 	if exists {
// 		tpl = val
// 	}

// 	return tpl
// }

// setupTempDatastore
func setupTempDatastore(datastore inventory.Datastore) (temp inventory.Datastore, err error) {
	temp, err = datastore.Clone()
	if err != nil {
		return temp, errors.Join(fmt.Errorf("failed to clone datastore"), err)
	}

	// Get the parent system
	sys, err := datastore.GetSystemZero()
	if err != nil {
		return temp, err
	}
	// Set additional metadata
	p, err := datastore.InventoryProvider()
	if err != nil {
		return temp, err
	}
	// Set top-level meta to the "system"
	sysMeta := inventory.ProviderMetadataRaw{}
	sys.ProviderMetadata = make(map[inventory.Provider]inventory.ProviderMetadataRaw)
	sys.ProviderMetadata[p] = sysMeta

	// Add it to the datastore
	err = temp.Update(&sys)
	if err != nil {
		return temp, err
	}
	return temp, nil
}
