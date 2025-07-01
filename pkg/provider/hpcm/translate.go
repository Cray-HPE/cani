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
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hpcm_client "github.com/Cray-HPE/cani/pkg/hpcm-client"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (hpcm *Hpcm) Translate(cmd *cobra.Command, args []string, ds inventory.Datastore) (err error) {
	var translated = map[uuid.UUID]*inventory.Hardware{}

	// check if input is coming from a cm config file
	if cmd.Flags().Changed("cm-config") {
		translated, err = hpcm.TranslateCmConfig()
		if err != nil {
			return err
		}
		// by default, translate information from the cmdb
	} else {
		translated, err = hpcm.TranslateCmdb()
		if err != nil {
			return err
		}
	}

	// add the translated hardware to the datastore
	for _, hw := range translated {
		err = ds.Add(hw)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractProviderMetadata is a catchall function for HPCM data that doesn't fit elsewhere
// it is all meant to be mapped to the provider properties
func extractProviderMetadata(node hpcm_client.Node) (md map[string]interface{}, err error) {
	md = make(map[string]interface{}, 0)
	md["AdministrativeStatus"] = node.AdministrativeStatus
	md["Aliases"] = node.Aliases
	md["Attributes"] = node.Attributes
	md["BiosBootMode"] = node.BiosBootMode
	md["Controller"] = make(map[string]interface{}, 0)
	// password should come from elsewhere
	cntrlr := md["Controller"].(map[string]interface{})
	cntrlr["channel"] = node.Controller.Channel
	cntrlr["ipAddress"] = node.Controller.IpAddress
	cntrlr["macAddress"] = node.Controller.MacAddress
	cntrlr["protocol"] = node.Controller.Protocol
	cntrlr["type"] = node.Controller.Type_
	cntrlr["username"] = node.Controller.Username
	md["CreationTime"] = node.CreationTime
	md["DeletionTime"] = node.DeletionTime
	md["Etag"] = node.Etag
	md["Image"] = node.Image
	md["ImagePending"] = node.ImagePending
	md["ImageTransport"] = node.ImageTransport
	md["InternalName"] = node.InternalName
	md["Inventory"] = make(map[string]interface{}, 0)
	mdInv := md["Inventory"].(map[string]interface{})
	if node.Inventory != nil {
		inv := *node.Inventory
		for k, v := range inv.(map[string]interface{}) {
			mdInv[k] = v
		}
	}
	md["IscsiRoot"] = node.IscsiRoot
	md["Links"] = node.Links
	md["Managed"] = node.Managed
	md["Management"] = make(map[string]interface{}, 0)
	// password should come from elsewhere
	mgmt := md["Management"].(map[string]interface{})
	mgmt["cardIpAddress"] = node.Management.CardIpAddress
	mgmt["cardMacAddress"] = node.Management.CardMacAddress
	mgmt["cardType"] = node.Management.CardType
	mgmt["channel"] = node.Management.Channel
	mgmt["protocol"] = node.Management.Protocol
	mgmt["username"] = node.Management.Username
	md["ModificationTime"] = node.ModificationTime
	md["Monitoring"] = node.Monitoring
	md["Network"] = node.Network
	md["NodeController"] = node.NodeController
	md["AdministrativeStatus"] = node.OperationalStatus
	md["Platform"] = node.Platform
	md["RootFs"] = node.RootFs
	md["RootSlot"] = node.RootSlot
	md["TemplateName"] = node.TemplateName

	return md, nil
}

// getVendor messily gets the vendor from an interface
func getVendor(node hpcm_client.Node) (vendor string) {
	md := make(map[string]interface{}, 0)
	md["Inventory"] = make(map[string]interface{}, 0)
	mdInv := md["Inventory"].(map[string]interface{})
	if node.Inventory != nil {
		inv := *node.Inventory
		for k, v := range inv.(map[string]interface{}) {
			if k == "fru.Manufacturer" {
				vendor = v.(string)
			}
			if vendor == "" {
				if k == "bios.Vendor" {
					vendor = v.(string)
				}
			}
			mdInv[k] = v
		}
	}
	return vendor
}

// getModel messily gets the model from an interface
func getModel(node hpcm_client.Node) (vendor string) {
	md := make(map[string]interface{}, 0)
	md["Inventory"] = make(map[string]interface{}, 0)
	mdInv := md["Inventory"].(map[string]interface{})
	if node.Inventory != nil {
		inv := *node.Inventory
		for k, v := range inv.(map[string]interface{}) {
			if k == "fru.Model" {
				vendor = v.(string)
			}
			mdInv[k] = v
		}
	}
	return vendor
}

// xnameToLocationPath
func xnameToLocationPath(x string) (lp inventory.LocationPath, err error) {
	xname := xnames.FromString(x)
	if xname != nil {
		lp, err = csm.FromXname(xname)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to build location path for xname (%v)", xname), err)
		}
		log.Debug().Msgf("Parsed LocationPath via xname %s: %+v", x, lp)
	}
	return lp, nil
}

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

// hpcmLocToCaniLoc translates hpcm location keys to a cani location path
func hpcmLocToCaniLoc(caniHwType hardwaretypes.HardwareType, hpcmLoc *hpcm_client.LocationSettings) (caniLoc inventory.LocationPath, err error) {
	// HPCM        --->   CANI
	// ------------------------------
	// Rack        --->   Cabinet
	// Chassis     --->   Chassis
	// Tray        --->   NodeBlade/SwitchBlade
	// Controller  --->   NodeController
	// Node        --->   Node
	var system, cabinet, chassis, blade, node inventory.LocationToken
	// rack and chassis map to cabinet and chassis
	system = inventory.LocationToken{HardwareType: hardwaretypes.System, Ordinal: 0}
	cabinet = inventory.LocationToken{HardwareType: hardwaretypes.Cabinet, Ordinal: int(hpcmLoc.Rack)}
	chassis = inventory.LocationToken{HardwareType: hardwaretypes.Chassis, Ordinal: int(hpcmLoc.Chassis)}

	// HPCM's Tray could be one of NodeBlade, ManagementSwitchEnclosure, or HighSpeedSwitchEnclosure
	switch caniHwType {
	case hardwaretypes.System:
		log.Debug().Msgf("LocationPath for %+v is currently limited to a single system", hardwaretypes.System)
		caniLoc = inventory.LocationPath{system}
	case hardwaretypes.Cabinet:
		caniLoc = inventory.LocationPath{system, cabinet}
	case hardwaretypes.Chassis:
		caniLoc = inventory.LocationPath{system, cabinet, chassis}
	case hardwaretypes.NodeBlade:
		blade = inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: int(hpcmLoc.Tray)}
		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade}
	case hardwaretypes.ManagementSwitchEnclosure:
		blade = inventory.LocationToken{HardwareType: hardwaretypes.ManagementSwitchEnclosure, Ordinal: int(hpcmLoc.Tray)}
		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade}
	case hardwaretypes.HighSpeedSwitchEnclosure:
		blade = inventory.LocationToken{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: int(hpcmLoc.Tray)}
		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade}
	case hardwaretypes.Node:
		blade = inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: int(hpcmLoc.Tray)}
		node = inventory.LocationToken{HardwareType: hardwaretypes.Node, Ordinal: int(hpcmLoc.Node)}
		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade, node}
	default:
		log.Warn().Msgf("Unable to get LocationPath from hardware type: %v", caniHwType)
	}

	log.Debug().Msgf("Set LocationPath via HPCM geo location values: %+v -> %v", hpcmLoc, caniLoc)

	return caniLoc, nil
}

// hpcmTypeToCaniHardwareType converts an HPCM type/group into a CANI hardwaretype
func hpcmTypeToCaniHardwareType(hpcmType string) (t hardwaretypes.HardwareType, err error) {
	switch hpcmType {
	case "admin":
		t = hardwaretypes.Cabinet
	case "chassis":
		t = hardwaretypes.Chassis
	case "cmc":
		t = hardwaretypes.ChassisManagementModule
	case "compute":
		t = hardwaretypes.Node
	case "cooldev":
		t = hardwaretypes.CoolingDistributionUnit
	case "ib_switch":
		t = hardwaretypes.HighSpeedSwitch
	case "leader", "leader_alias", "ice_compute":
		t = hardwaretypes.Node
	case "leaf", "spine", "mgmt_switch":
		t = hardwaretypes.ManagementSwitch
	case "pdu":
		t = hardwaretypes.CabinetPDU
	case "switch_blade":
		t = hardwaretypes.ManagementSwitchEnclosure
	default:
		err = fmt.Errorf("unabled to map HPCM type to CANI hardwaretype: %v", hpcmType)
	}
	if err != nil {
		return t, err
	}

	return t, nil
}
