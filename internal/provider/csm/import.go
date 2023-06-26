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
package csm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

//
// Other thoughts
//
//
// 1. Create a top level system object in SLS, this will serve at the main place to store the CANI metadata
//		- Last time was imported by CANI
// 		- Version of CANI when the last import occured
//		- SLS CANI Schema Level
// 2. For missing hardware in SLS like Mountain Cabinet RouterBMCs, add them from HSM state
// 3. Each hardware object should have the UUID of the assoicated CANI Hardware UUID that it is assoicated to. Hopefullt this will be 1-to-1
// 4. If hardware is added to SLS without the special CANI metadata it can detected as being added outside the normal process
// 5. For hardware that doesn't exist in mountain cabinets (phantom nodes) either we mark thinks as absent as a CANI state (here is the logical data, but no physical data)
//		or out right remove them, but that will break existing procedures.

func loadJSON(path string, dest interface{}) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, dest)
}

func (csm *CSM) Import(ctx context.Context, datastore inventory.Datastore) error {

	//
	// Retrieve current state from the system
	//
	slsDumpstate, _, err := csm.slsClient.DumpstateApi.DumpstateGet(ctx)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to perform SLS dumpstate"), err)
	}

	hsmStateComponents, _, err := csm.hsmClient.ComponentApi.DoComponentsGet(ctx, nil)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to retrieve HSM State Components"), err)
	}

	hsmHardwareInventory, _, err := csm.hsmClient.HWInventoryByLocationApi.DoHWInvByLocationGetAll(ctx, nil)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to retrieve HSM State Components"), err)
	}

	//
	// HSM lookup tables
	//
	hsmStateComponentsMap := map[string]hsm_client.Component100Component{}
	for _, hsmComponent := range hsmStateComponents.Components {
		hsmStateComponentsMap[hsmComponent.ID] = hsmComponent
	}
	hsmHardwareInventoryMap := map[string]hsm_client.HwInventory100HwInventoryByLocation{}
	for _, hsmHardware := range hsmHardwareInventory {
		hsmHardwareInventoryMap[hsmHardware.ID] = hsmHardware
	}

	tempDatastore, err := datastore.Clone()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to clone datastore"), err)
	}

	// Prune non-mountain hardware
	slsDumpstate.Hardware, _ = sls.FilterHardware(slsDumpstate.Hardware, func(hardware sls_client.Hardware) (bool, error) {
		return hardware.Class != sls_client.HardwareClassRiver, nil
	})

	// Get the system UUID
	cSystem, err := tempDatastore.GetSystemZero()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to get system:0 ID"), err)
	}

	// Import related changes for SLS
	// slsHardwareToAdd := map[string]sls_client.Hardware{}
	slsHardwareToModify := map[string]sls_client.Hardware{}
	// slsHardwareExists := map[string]sls_client.Hardware{}

	// CANI Hardware changes
	// TODO

	// TODO Unable to POST this into SLS
	// Check to see if a system object exists in the SLS dumpstate
	// slsSystem, exists := slsDumpstate.Hardware["s0"]
	// if !exists {
	// 	log.Warn().Msgf("SLS does not contain a system object, creating one")
	//
	// 	slsSystem = sls.NewHardware(xnames.System{}, sls_client.HardwareClassRiver, sls_client.HardwareExtraPropertiesSystem{
	// 		CaniId:               cSystem.ID.String(),
	// 		CaniSlsSchemaVersion: "v1alpha1", // TODO make this a enum
	// 		CaniLastModified:     time.Now().UTC().String(),
	// 	})

	// 	slsHardwareToAdd[slsSystem.Xname] = slsSystem
	// }

	// log.Info().Msgf("System: %v", slsSystem)

	//
	// Import Cabinets and Chassis
	//
	allCabinets, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Cabinet)
	allChassis, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Chassis)
	allChassisBMCs, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.ChassisBMC)

	// Find all cabinets and what chassis they have
	cabinetChassisCounts := map[string][]int{}
	for _, chassis := range allChassis {
		chassisXname := xnames.FromStringToStruct[xnames.Chassis](chassis.Xname)
		cabinet := chassisXname.Parent()

		cabinetChassisCounts[cabinet.String()] = append(cabinetChassisCounts[cabinet.String()], chassisXname.Chassis)
	}
	for cabinet, chassisOrdinals := range cabinetChassisCounts {
		sort.Ints(chassisOrdinals)
		log.Debug().Msgf("%s: %v - %v", cabinet, len(chassisOrdinals), chassisOrdinals)
	}

	// Find all cabinets and build up HMN VLAN Mappings
	cabinetHMNVlans := map[string]int{}
	cabinetSubnetRegex := regexp.MustCompile(`cabinet_(\d+)`)

	// The networking data in SLS should be considered the source of truth for networking information
	// instead of looking at the SLS hardware part of SLS
	for _, networkName := range []string{"HMN_MTN", "HMN_RVR"} {
		network, exists := slsDumpstate.Networks[networkName]
		if !exists {
			log.Warn().Msgf("SLS Network (%s) does not exist", networkName)
			continue
		}

		for _, subnet := range network.ExtraProperties.Subnets {
			matches := cabinetSubnetRegex.FindStringSubmatch(subnet.Name)
			if len(matches) != 2 {
				log.Warn().Msgf("Skipping subnet (%s) in network (%s) for cabinet HMN Vlan lookup", subnet.Name, networkName)
				continue
			}

			cabinetXname := xnames.Cabinet{}
			cabinetXname.Cabinet, err = strconv.Atoi(matches[1])
			if err != nil {
				return errors.Join(fmt.Errorf("failed to extract cabinet number from subnet (%s)", subnet.Name), err)
			}

			cabinetHMNVlans[cabinetXname.String()] = int(subnet.VlanID)
		}
	}

	for cabinet, vlan := range cabinetHMNVlans {
		log.Debug().Msgf("Cabinet (%s) has HMN VLAN (%d)", cabinet, vlan)
	}

	for _, slsCabinet := range allCabinets {
		cabinetXname := xnames.FromStringToStruct[xnames.Cabinet](slsCabinet.Xname)
		if cabinetXname == nil {
			return fmt.Errorf("failed to parse cabinet xname (%s)", slsCabinet.Xname)
		}

		locationPath, err := FromXname(cabinetXname)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to build location path for xname (%v)", cabinetXname), err)
		}

		//
		// Stage 1: Determine of the cabinet is new or currently exists. If they don't exist push the required hardware data into the database
		//
		cCabinet, err := tempDatastore.GetAtLocation(locationPath)
		if err == nil {
			// Cabinet exists
			log.Info().Msgf("Cabinet %s (%v) exists in datastore with ID (%s)", cabinetXname, locationPath, cCabinet.ID)

			// TODO Build metadata from sls data
		} else if errors.Is(err, inventory.ErrHardwareNotFound) {
			// Cabinet does not exist, which means it needs to be added
			// TODO When reconstituting the CANI inventory (say it was lost), should we reuse existing IDs?
			log.Info().Msgf("Cabinet %s does not exist in datastore at %s", cabinetXname, locationPath)

			deviceTypeSlug := ""

			switch slsCabinet.Class {
			case sls_client.HardwareClassRiver:
				deviceTypeSlug = "hpe-eia-cabinet"
			case sls_client.HardwareClassHill:
				if reflect.DeepEqual(cabinetChassisCounts[cabinetXname.String()], []int{1, 3}) {
					deviceTypeSlug = "hpe-ex2000"
				} else if reflect.DeepEqual(cabinetChassisCounts[cabinetXname.String()], []int{0}) {
					deviceTypeSlug = "hpe-ex2500-1-liquid-cooled-chassis"
				} else if reflect.DeepEqual(cabinetChassisCounts[cabinetXname.String()], []int{0, 1}) {
					deviceTypeSlug = "hpe-ex2500-2-liquid-cooled-chassis"
				} else if reflect.DeepEqual(cabinetChassisCounts[cabinetXname.String()], []int{0, 1, 2}) {
					deviceTypeSlug = "hpe-ex2500-3-liquid-cooled-chassis"
				}
			case sls_client.HardwareClassMountain:
				if reflect.DeepEqual(cabinetChassisCounts[cabinetXname.String()], []int{0, 1, 2, 3, 4, 5, 6, 7}) {
					deviceTypeSlug = "hpe-ex4000" // TODO This is ambiguous with the EX3000 cabinet, for right now assume
				}
			default:
				return fmt.Errorf("cabinet (%s) has unknown class (%s)", cabinetXname, slsCabinet.Class)
			}

			if deviceTypeSlug == "" {
				log.Warn().Msgf("Cabinet %s device type slug is unknown, ignoring", cabinetXname.String())
				continue
			} else {
				log.Info().Msgf("Cabinet %s device type slug is %s", cabinetXname.String(), deviceTypeSlug)
			}

			// Now its time to build up what the hardware looks like
			newHardware, err := csm.buildInventoryHardware(deviceTypeSlug, cabinetXname.Cabinet, cSystem.ID, inventory.HardwareStatusProvisioned)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to build hardware for cabinet (%s)", cabinetXname.String()), err)
			}

			// Push the new hardware into the datastore
			for _, cHardware := range newHardware {
				log.Info().Msgf("Hardware from cabinet %s: %s", cabinetXname.String(), cHardware.ID)
				if err := tempDatastore.Add(&cHardware); err != nil {
					return fmt.Errorf("failed to add hardware (%s) to in memory datastore", cHardware.ID)
				}
			}

			// Set cabinet metadata
			cabinetMetadata := CabinetMetadata{}
			if vlan, exists := cabinetHMNVlans[slsCabinet.Xname]; exists {
				cabinetMetadata.HMNVlan = IntPtr(vlan)
			}

			cCabinet, err = tempDatastore.GetAtLocation(locationPath)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to query datastore for %s", locationPath), err)
			}

			cCabinet.ProviderProperties = map[string]interface{}{
				"csm": cabinetMetadata,
			}

			if err := tempDatastore.Update(&cCabinet); err != nil {
				return fmt.Errorf("failed to update hardware (%s) in memory datastore", cCabinet.ID)
			}

		} else {
			// Error occurred
			return errors.Join(fmt.Errorf("failed to query datastore"), err)
		}

		// Update SLS metadata
		slsCabinetEP, err := sls.DecodeExtraProperties[sls_client.HardwareExtraPropertiesCabinet](slsCabinet)
		if err != nil {
			return fmt.Errorf("failed to decode SLS hardware extra properties (%s)", slsCabinet.Xname)
		}

		if slsCabinetEP.CaniId != cCabinet.ID.String() {
			if len(slsCabinetEP.CaniId) != 0 {
				log.Warn().Msgf("Detected CANI hardware ID change from %s to %s for SLS Hardware %s", slsCabinetEP.CaniId, cCabinet.ID, slsCabinet.Xname)
			}

			// Add in CANI properties
			slsCabinetEP.CaniId = cCabinet.ID.String()
			slsCabinetEP.CaniSlsSchemaVersion = "v1alpha1" // TODO make this a enum
			slsCabinetEP.CaniLastModified = time.Now().UTC().String()

			log.Info().Msgf("SLS extra properties changed for %s", slsCabinet.Xname)

			slsCabinet.ExtraProperties = slsCabinetEP
			slsHardwareToModify[slsCabinet.Xname] = slsCabinet
		}
	}

	//
	// Fix up Chassis SLS metadata
	//
	for _, slsHardware := range allChassis {
		xname := xnames.FromString(slsHardware.Xname)
		if xname == nil {
			return fmt.Errorf("failed to parse xname (%s)", slsHardware.Xname)
		}

		locationPath, err := FromXname(xname)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to build location path for xname (%v)", xname), err)
		}

		cHardware, err := tempDatastore.GetAtLocation(locationPath)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to query datastore for %s", locationPath), err)
		}

		// Update SLS metadata
		slsEP, err := sls.DecodeExtraProperties[sls_client.HardwareExtraPropertiesChassis](slsHardware)
		if err != nil {
			return fmt.Errorf("failed to decode SLS hardware extra properties (%s)", slsHardware.Xname)
		}

		if slsEP.CaniId != cHardware.ID.String() {
			if len(slsEP.CaniId) != 0 {
				log.Warn().Msgf("Detected CANI hardware ID change from %s to %s for SLS Hardware %s", slsEP.CaniId, cHardware.ID, slsHardware.Xname)
			}

			// Add in CANI properties
			slsEP.CaniId = cHardware.ID.String()
			slsEP.CaniSlsSchemaVersion = "v1alpha1" // TODO make this a enum
			slsEP.CaniLastModified = time.Now().UTC().String()

			log.Info().Msgf("SLS extra properties changed for %s", slsHardware.Xname)

			slsHardware.ExtraProperties = slsEP
			slsHardwareToModify[slsHardware.Xname] = slsHardware
		}
	}

	//
	// Fix up ChassisBMC SLS Metadata
	//
	for _, slsHardware := range allChassisBMCs {
		xname := xnames.FromString(slsHardware.Xname)
		if xname == nil {
			return fmt.Errorf("failed to parse xname (%s)", slsHardware.Xname)
		}

		locationPath, err := FromXname(xname)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to build location path for xname (%v)", xname), err)
		}

		cHardware, err := tempDatastore.GetAtLocation(locationPath)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to query datastore for %s", locationPath), err)
		}

		// Update SLS metadata
		slsEP, err := sls.DecodeExtraProperties[sls_client.HardwareExtraPropertiesChassisBmc](slsHardware)
		if err != nil {
			return fmt.Errorf("failed to decode SLS hardware extra properties (%s)", slsHardware.Xname)
		}

		if slsEP.CaniId != cHardware.ID.String() {
			if len(slsEP.CaniId) != 0 {
				log.Warn().Msgf("Detected CANI hardware ID change from %s to %s for SLS Hardware %s", slsEP.CaniId, cHardware.ID, slsHardware.Xname)
			}

			// Add in CANI properties
			slsEP.CaniId = cHardware.ID.String()
			slsEP.CaniSlsSchemaVersion = "v1alpha1" // TODO make this a enum
			slsEP.CaniLastModified = time.Now().UTC().String()

			log.Info().Msgf("SLS extra properties changed for %s", slsHardware.Xname)

			slsHardware.ExtraProperties = slsEP
			slsHardwareToModify[slsHardware.Xname] = slsHardware
		}
	}

	//
	// Import Nodes
	//
	allNodes, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Node)

	// 1. Find all slots holding blades (either currently populated or could be populated) from SLS
	slsNodeBladeXnames := []xnames.ComputeModule{}
	slsNodeBladesFound := map[xnames.ComputeModule][]xnames.NodeBMC{}
	slsNodeBMCFound := map[xnames.NodeBMC]bool{}
	for _, slsNode := range allNodes {
		nodeXname := xnames.FromStringToStruct[xnames.Node](slsNode.Xname)
		if nodeXname == nil {
			return fmt.Errorf("failed to parse node xname (%s)", slsNode.Xname)
		}

		// Node -> NodeBMC (Node Card) -> ComputeModule (Node Blade)
		nodeBMCXname := nodeXname.Parent()
		nodeBladeXname := nodeBMCXname.Parent()

		if slsNodeBMCFound[nodeBMCXname] {
			// We have already discovered this node BMC, and we don't need to add it again
			continue
		}

		// Keep track that we have seem this BMC
		slsNodeBMCFound[nodeBMCXname] = true

		if _, exists := slsNodeBladesFound[nodeBladeXname]; !exists {
			// This is the first time we have seem this blade, lets add it to our list of node blade xnames
			slsNodeBladeXnames = append(slsNodeBladeXnames, nodeBladeXname)
		}

		// Keep track that we found this node BMC on this blade
		slsNodeBladesFound[nodeBladeXname] = append(slsNodeBladesFound[nodeBladeXname], nodeBMCXname)
	}

	// 1.1 Sort the found node blade xnames, so the output is nice to look at
	for _, nodeBMCs := range slsNodeBladesFound {
		sort.Slice(nodeBMCs, func(i, j int) bool {
			return nodeBMCs[i].String() < nodeBMCs[j].String()
		})
	}
	sort.Slice(slsNodeBladeXnames, func(i, j int) bool {
		return slsNodeBladeXnames[i].String() < slsNodeBladeXnames[j].String()
	})

	// 2. Find all slots holding blades from HSM, and identify hardware
	nodeBladeDeviceSlugs := map[xnames.ComputeModule]string{}
	for _, nodeBladeXname := range slsNodeBladeXnames {
		hsmComponent, exists := hsmStateComponentsMap[nodeBladeXname.String()]
		if !exists {
			log.Debug().Msgf("%s exists in SLS, but not HSM", nodeBladeXname)

			continue
		}

		if hsmComponent.State != nil {
			log.Debug().Msgf("%s exists in HSM with state %s", nodeBladeXname, *hsmComponent.State)
		}
		for _, nodeBMCXname := range slsNodeBladesFound[nodeBladeXname] {
			// Don't need to do this if we already identified the blade
			if _, exists := nodeBladeDeviceSlugs[nodeBladeXname]; exists {
				continue
			}

			// For every BMC in HSM there is a NodeEnclosure. The NodeEnclosure ordinal matches
			// the BMC ordinal
			nodeEnclosureXname := nodeBladeXname.NodeEnclosure(nodeBMCXname.NodeBMC)

			nodeEnclosure, exists := hsmHardwareInventoryMap[nodeEnclosureXname.String()]
			if !exists {
				log.Warn().Msgf("%s is missing from HSM hardware inventory, possible phantom hardware", nodeEnclosureXname)
				continue // TODO what should happen here?
			}

			if nodeEnclosure.PopulatedFRU == nil {
				log.Warn().Msgf("%s is missing PopulatedFRU data", nodeEnclosureXname)
				continue // TODO what should happen here?
			}

			if nodeEnclosure.PopulatedFRU.HMSNodeEnclosureFRUInfo == nil {
				log.Warn().Msgf("%s is missing PopulatedFRU node enclosure data", nodeEnclosureXname)
				continue // TODO what should happen here?
			}
			nodeEnclosureFru := nodeEnclosure.PopulatedFRU.HMSNodeEnclosureFRUInfo

			log.Debug().Msgf("%s has manufacturer %s and model %s", nodeEnclosureXname, nodeEnclosureFru.Manufacturer, nodeEnclosureFru.Model)

			bladeDeviceSlug, err := csm.identifyDeviceSlug(nodeEnclosureFru.Manufacturer, nodeEnclosureFru.Model)
			if err != nil {
				log.Warn().Msgf("%s unable to determine blade device slug from Node Enclosure FRU data: %s", nodeEnclosureXname, err)
				continue
			}

			nodeBladeDeviceSlugs[nodeBladeXname] = bladeDeviceSlug

			log.Debug().Msgf("%s has blade device slug: %s", nodeBladeXname, bladeDeviceSlug)
		}

	}

	// 3.
	for nodeBladeXname, deviceSlug := range nodeBladeDeviceSlugs {
		// Check to see if the node blade exists

		nodeBladeLocationPath, err := FromXname(nodeBladeXname)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to build location path for xname (%v)", nodeBladeXname), err)
		}
		cNodeBlade, err := tempDatastore.GetAtLocation(nodeBladeLocationPath)
		if err == nil {
			// Blade currently exists
			log.Debug().Msgf("Node blade %s (%v) exists in datastore with ID (%s)", nodeBladeXname, nodeBladeLocationPath, cNodeBlade.ID)

			// TODO Build metadata from sls data for merging

		} else if errors.Is(err, inventory.ErrHardwareNotFound) {
			// Node blade does not exist

			// Determine the chassis ID
			chassisLocationPath, err := FromXname(nodeBladeXname.Parent())
			if err != nil {
				return errors.Join(fmt.Errorf("failed to build location path for xname (%v)", nodeBladeXname), err)
			}
			cChassis, err := tempDatastore.GetAtLocation(chassisLocationPath)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to get datastore ID for %v", chassisLocationPath), err)
			}

			// Now its time to build up what the hardware looks like
			newHardware, err := csm.buildInventoryHardware(deviceSlug, nodeBladeXname.ComputeModule, cChassis.ID, inventory.HardwareStatusProvisioned)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to build hardware for node blade (%s)", nodeBladeXname.String()), err)
			}

			// Push the new hardware into the datastore
			for _, cHardware := range newHardware {
				log.Debug().Msgf("Hardware from node blade %s: %s", nodeBladeXname.String(), cHardware.ID)
				if err := tempDatastore.Add(&cHardware); err != nil {
					return fmt.Errorf("failed to add hardware (%s) to in memory datastore", cHardware.ID)
				}
			}

		} else {
			// Error occurred
			return errors.Join(fmt.Errorf("failed to query datastore"), err)
		}
	}

	// Update node metadata in CANI and SLS
	for _, slsNode := range allNodes {
		nodeXname := xnames.FromString(slsNode.Xname)
		if nodeXname == nil {
			return fmt.Errorf("failed to parse node xname (%s)", slsNode.Xname)
		}

		nodeLocationPath, err := FromXname(nodeXname)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to build location path for xname (%v)", nodeXname), err)
		}

		//
		// Build up node extra properties for CANI
		//
		slsNodeEP, err := sls.DecodeExtraProperties[sls_client.HardwareExtraPropertiesNode](slsNode)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to decode hardware extra properties for (%s)", slsNode.Xname), err)
		}

		nodeMetadata := NodeMetadata{}
		if slsNodeEP.Role != "" {
			nodeMetadata.Role = StringPtr(slsNodeEP.Role)
		}

		if slsNodeEP.SubRole != "" {
			nodeMetadata.Role = StringPtr(slsNodeEP.SubRole)
		}

		if slsNodeEP.NID != 0 {
			nodeMetadata.Nid = IntPtr(int(slsNodeEP.NID))
		}

		if len(slsNodeEP.Aliases) != 0 {
			nodeMetadata.Alias = slsNodeEP.Aliases
		}

		cNode, err := tempDatastore.GetAtLocation(nodeLocationPath)
		if errors.Is(err, inventory.ErrHardwareNotFound) {
			log.Warn().Msgf("Hardware does not exist (possible phantom hardware): %s", nodeLocationPath)
			// This is a phantom node, and we need to push this into the inventory to preserve the logical information
			// of the node
			// TODO an interesting scenario to test with this would be the an bard peak blade in the location that that SLS assumes to be a windom blade

			// The cabinet and chassis should exist
			// cChassis, err := tempDatastore.GetAtLocation(nodeLocationPath[0:3])
			// if errors.Is(err, inventory.ErrHardwareNotFound) {
			// 	return errors.Join(fmt.Errorf("failed to query datastore for %s", nodeLocationPath), err)
			// } else if err != nil {
			// 	return errors.Join(fmt.Errorf("chassis of phantom node (%s) does not exist in datastore", nodeLocationPath), err)
			// }

			// // The Node Blade may not exist
			// cNodeBlade, err := tempDatastore.GetAtLocation(nodeLocationPath[0:4])
			// if errors.Is(err, inventory.ErrHardwareNotFound) {
			// 	// It doesn't exist, so lets create an empty one
			// 	cNodeBlade = inventory.Hardware{
			// 		Parent: cChassis.ID,
			// 	}
			// 	tempDatastore.Add()
			// } else if err != nil {
			// 	return errors.Join(fmt.Errorf("failed to query datastore for %s", nodeLocationPath), err)
			// }

			// // The Node Card may not exist
			// nodeCardExists, err := nodeLocationPath[0:5].Exists(tempDatastore)
			// if err != nil {
			// 	return errors.Join(fmt.Errorf("failed to query datastore for %s", nodeLocationPath), err)
			// }

			// log.Fatal().Msg("Panic!")
			continue
		} else if err != nil {
			return errors.Join(fmt.Errorf("failed to query datastore for %s", nodeLocationPath), err)
		}

		// Initialize the properties map if not done already
		if cNode.ProviderProperties == nil {
			cNode.ProviderProperties = map[string]interface{}{}
		}
		cNode.ProviderProperties["csm"] = nodeMetadata

		// Push updates into the datastore
		if err := tempDatastore.Update(&cNode); err != nil {
			return fmt.Errorf("failed to update hardware (%s) in memory datastore", cNode.ID)
		}

		//
		// Update SLS Extra Properties
		//
		if slsNodeEP.CaniId != cNode.ID.String() {
			if len(slsNodeEP.CaniId) != 0 {
				log.Warn().Msgf("Detected CANI hardware ID change from %s to %s for SLS Hardware %s", slsNodeEP.CaniId, cNode.ID, slsNode.Xname)
			}

			// Update it if it has changed
			slsNodeEP.CaniId = cNode.ID.String()
			slsNodeEP.CaniSlsSchemaVersion = "v1alpha1" // TODO make this a enum
			slsNodeEP.CaniLastModified = time.Now().UTC().String()

			slsNode.ExtraProperties = slsNodeEP
			slsHardwareToModify[slsNode.Xname] = slsNode

			log.Debug().Msgf("SLS extra properties changed for %s", slsNode.Xname)
		}
	}

	//
	// Import Router BMCs
	//
	// TODO

	//
	// Handle phantom mountain/hill nodes
	//
	// TODO this might be better handled in the some code above

	//
	// Push updates to SLS
	//
	if err := sls.HardwareUpdate(csm.slsClient, ctx, slsHardwareToModify, 10); err != nil {
		return errors.Join(fmt.Errorf("failed to update hardware in SLS"), err)
	}

	// TODO need a sls.HardwareCreate function
	// for _, slsHardware := range slsHardwareToAdd {
	// 	// Perform a POST against SLS
	//
	// 	_, r, err := csm.slsClient.HardwareApi.HardwarePost(ctx, sls.NewHardwarePostOpts(
	// 		slsHardware,
	// 	))
	// 	if err != nil {
	// 		return errors.Join(
	// 			fmt.Errorf("failed to add hardware (%s) to SLS", slsHardware.Xname),
	// 			err,
	// 		)
	// 	}
	// 	log.Info().Int("status", r.StatusCode).Msg("Added hardware to SLS")
	// }

	// Commit changes!
	if err := datastore.Merge(tempDatastore); err != nil {
		return errors.Join(fmt.Errorf("failed to merge temporary datastore with actual datastore"), err)
	}

	return datastore.Flush()
}

func (csm *CSM) buildInventoryHardware(deviceTypeSlug string, ordinal int, parentID uuid.UUID, status inventory.HardwareStatus) ([]inventory.Hardware, error) {
	if csm.hardwareLibrary == nil {
		panic("Hardware type library is nil")
	}

	// Build up the expected hardware
	// Generate a hardware build out using the system as a parent
	hardwareBuildOutItems, err := csm.hardwareLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, ordinal, parentID)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
			err,
		)
	}

	var allHardware []inventory.Hardware
	for _, hardwareBuildOut := range hardwareBuildOutItems {
		locationOrdinal := hardwareBuildOut.OrdinalPath[len(hardwareBuildOut.OrdinalPath)-1]

		allHardware = append(allHardware, inventory.Hardware{
			ID:             hardwareBuildOut.ID,
			Parent:         hardwareBuildOut.ParentID,
			Type:           hardwareBuildOut.DeviceType.HardwareType,
			DeviceTypeSlug: hardwareBuildOut.DeviceType.Slug,
			Vendor:         hardwareBuildOut.DeviceType.Manufacturer,
			Model:          hardwareBuildOut.DeviceType.Model,

			LocationOrdinal: &locationOrdinal,

			Status: inventory.HardwareStatusProvisioned,
		})

	}

	return allHardware, nil
}

func FromXname(xnameRaw xnames.Xname) (inventory.LocationPath, error) {
	// TODO Look into go generating this

	switch xname := xnameRaw.(type) {
	// System
	case xnames.System:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
		}, nil
	case *xnames.System:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
		}, nil
	// Cabinet
	case xnames.Cabinet:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
		}, nil
	case *xnames.Cabinet:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
		}, nil
	// Chassis
	case xnames.Chassis:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
		}, nil
	case *xnames.Chassis:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
		}, nil
	// Chassis BMC
	case xnames.ChassisBMC:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.ChassisManagementModule, Ordinal: xname.ChassisBMC},
		}, nil
	case *xnames.ChassisBMC:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.ChassisManagementModule, Ordinal: xname.ChassisBMC},
		}, nil
	// Compute Module
	case xnames.ComputeModule:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.NodeBlade, Ordinal: xname.ComputeModule},
		}, nil
	case *xnames.ComputeModule:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.NodeBlade, Ordinal: xname.ComputeModule},
		}, nil
	// Node BMC
	case xnames.NodeBMC:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.NodeBlade, Ordinal: xname.ComputeModule},
			{HardwareType: hardwaretypes.NodeCard, Ordinal: xname.NodeBMC},
			{HardwareType: hardwaretypes.NodeController, Ordinal: 0}, // Assumes one Node BMC per node card, For all supported CSM hardware this is true
		}, nil
	case *xnames.NodeBMC:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.NodeBlade, Ordinal: xname.ComputeModule},
			{HardwareType: hardwaretypes.NodeCard, Ordinal: xname.NodeBMC},
			{HardwareType: hardwaretypes.NodeController, Ordinal: 0}, // Assumes one Node BMC per node card, For all supported CSM hardware this is true
		}, nil
	// Node
	case xnames.Node:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.NodeBlade, Ordinal: xname.ComputeModule},
			{HardwareType: hardwaretypes.NodeCard, Ordinal: xname.NodeBMC},
			{HardwareType: hardwaretypes.Node, Ordinal: xname.Node},
		}, nil
	case *xnames.Node:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.NodeBlade, Ordinal: xname.ComputeModule},
			{HardwareType: hardwaretypes.NodeCard, Ordinal: xname.NodeBMC},
			{HardwareType: hardwaretypes.Node, Ordinal: xname.Node},
		}, nil
	// Router Module
	case xnames.RouterModule:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: xname.RouterModule},
		}, nil
	case *xnames.RouterModule:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: xname.RouterModule},
		}, nil
	// Router BMC
	case xnames.RouterBMC:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: xname.RouterModule},
			{HardwareType: hardwaretypes.HighSpeedSwitchController, Ordinal: xname.RouterBMC},
		}, nil
	case *xnames.RouterBMC:
		return inventory.LocationPath{
			{HardwareType: hardwaretypes.System, Ordinal: 0},
			{HardwareType: hardwaretypes.Cabinet, Ordinal: xname.Cabinet},
			{HardwareType: hardwaretypes.Chassis, Ordinal: xname.Chassis},
			{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: xname.RouterModule},
			{HardwareType: hardwaretypes.HighSpeedSwitchController, Ordinal: xname.RouterBMC},
		}, nil
	}

	return nil, fmt.Errorf("unable to convert xname type (%s)", xnameRaw.Type())
}

func (csm *CSM) identifyDeviceSlug(manufacturer, model string) (string, error) {
	for deviceSlug, deviceType := range csm.hardwareLibrary.DeviceTypes {
		for _, identification := range deviceType.Identifications {
			// log.Info().Msgf("Checking %v against [%s, %s]", identification, manufacturer, model)
			if identification.Manufacturer == manufacturer && identification.Model == model {
				return deviceSlug, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find corrensponding device slug for manufacturer (%s) and model (%s)", manufacturer, model)
}
