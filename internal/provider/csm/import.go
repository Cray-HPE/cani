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

// Import external inventory data into CANI's inventory format
// func (csm *CSM) Import2(ctx context.Context, datastore inventory.Datastore) error {
// 	// Get the system UUID
// 	cSystem, err := datastore.GetSystemZero()
// 	if err != nil {
// 		return errors.Join(fmt.Errorf("failed to get system:0 ID"), err)
// 	}

// 	// Get hardware contents of SLS
// 	slsDumpstate, _, err := csm.slsClient.DumpstateApi.DumpstateGet(ctx)
// 	if err != nil {
// 		return errors.Join(fmt.Errorf("failed to perform dumpstate from SLS"), err)
// 	}

// 	// Sort hardware by type
// 	allCabinets, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Cabinet)
// 	allChassis, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Chassis)
// 	// chassisBMC, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.ChassisBMC)
// 	// routerBMCs, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.RouterBMC)
// 	// mgmtSwitches, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.MgmtSwitch, xnametypes.MgmtHLSwitch, xnametypes.CDUMgmtSwitch)
// 	// mgmtSwitchConnectors, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.MgmtSwitchConnector)
// 	// nodes, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Node)
// 	// gigabyteCMCs := map[string]sls_client.Hardware{}

// 	//
// 	// Attempt to identify cabinets
// 	//

// 	for cabinet, chassisOrdinals := range cabinetChassisCounts {
// 		sort.Ints(chassisOrdinals)
// 		log.Debug().Msgf("%s: %v - %v", cabinet, len(chassisOrdinals), chassisOrdinals)
// 	}

// 	// Apply some heuristics to infer the type of cabinet this is
// 	for _, cabinet := range allCabinets {
// 		cabinetXname := xnames.FromStringToStruct[xnames.Cabinet](cabinet.Xname)
// 		if cabinetXname == nil {
// 			return fmt.Errorf("failed to parse cabinet xname (%s)", cabinet.Xname)
// 		}

// 		deviceTypeSlug := ""
// 		switch cabinet.Class {
// 		case sls_client.HardwareClassRiver:
// 			deviceTypeSlug = "hpe-eia-cabinet"
// 		case sls_client.HardwareClassHill:
// 			if reflect.DeepEqual(cabinetChassisCounts[cabinet.Xname], []int{1, 3}) {
// 				deviceTypeSlug = "hpe-ex2000"
// 			} else if reflect.DeepEqual(cabinetChassisCounts[cabinet.Xname], []int{0}) {
// 				deviceTypeSlug = "hpe-ex2500-1-liquid-cooled-chassis"
// 			} else if reflect.DeepEqual(cabinetChassisCounts[cabinet.Xname], []int{0, 1}) {
// 				deviceTypeSlug = "hpe-ex2500-2-liquid-cooled-chassis"
// 			} else if reflect.DeepEqual(cabinetChassisCounts[cabinet.Xname], []int{0, 1, 2}) {
// 				deviceTypeSlug = "hpe-ex2500-3-liquid-cooled-chassis"
// 			}
// 		case sls_client.HardwareClassMountain:
// 			if reflect.DeepEqual(cabinetChassisCounts[cabinet.Xname], []int{0, 1, 2, 3, 4, 5, 6, 7}) {
// 				deviceTypeSlug = "hpe-ex4000" // TODO This is ambiguous with the EX3000 cabinet, for right now assume
// 			}
// 		default:
// 			return fmt.Errorf("cabinet (%s) has unknown class (%s)", cabinet.Xname, cabinet.Class)
// 		}

// 		if deviceTypeSlug == "" {
// 			log.Warn().Msgf("Cabinet %s device type slug is unknown, ignoring", cabinet.Xname)
// 			continue
// 		} else {
// 			log.Info().Msgf("Cabinet %s device type slug is %s", cabinet.Xname, deviceTypeSlug)
// 		}

// 		// Build up the expected hardware
// 		// Generate a hardware build out using the system as a parent
// 		hardwareBuildOutItems, err := csm.hardwareLibrary.GetDefaultHardwareBuildOut(deviceTypeSlug, cabinetXname.Cabinet, cSystem.ID)
// 		if err != nil {
// 			return errors.Join(
// 				fmt.Errorf("unable to build default hardware build out for %s", deviceTypeSlug),
// 				err,
// 			)
// 		}

// 		for _, hardwareBuildOut := range hardwareBuildOutItems {
// 			locationOrdinal := hardwareBuildOut.OrdinalPath[len(hardwareBuildOut.OrdinalPath)-1]

// 			hardware := inventory.Hardware{
// 				ID:             hardwareBuildOut.ID,
// 				Parent:         hardwareBuildOut.ParentID,
// 				Type:           hardwareBuildOut.DeviceType.HardwareType,
// 				DeviceTypeSlug: hardwareBuildOut.DeviceType.Slug,
// 				Vendor:         hardwareBuildOut.DeviceType.Manufacturer,
// 				Model:          hardwareBuildOut.DeviceType.Model,

// 				LocationOrdinal: &locationOrdinal,

// 				Status: inventory.HardwareStatusProvisioned,
// 			}

// 			log.Info().Msgf("Hardware from cabinet %s: %v", cabinetXname.String(), hardware)
// 		}

// 	}

// 	//
// 	// Comparison
// 	//

// 	// Build up location paths, and determine if the hardware currently exists in the inventory

// 	//
// 	// Other thoughts
// 	//

// 	// 1. Create a top level system object in SLS, this will serve at the main place to store the CANI metadata
// 	//		- Last time was imported by CANI
// 	// 		- Version of CANI when the last import occured
// 	//		- SLS CANI Schema Level
// 	// 2. For missing hardware in SLS like Mountain Cabinet RouterBMCs, add them from HSM state
// 	// 3. Each hardware object should have the UUID of the assoicated CANI Hardware UUID that it is assoicated to. Hopefullt this will be 1-to-1
// 	// 4. If hardware is added to SLS without the special CANI metadata it can detected as being added outside the normal process
// 	// 5. For hardware that doesn't exist in mountain cabinets (phantom nodes) either we mark thinks as absent as a CANI state (here is the logical data, but no physical data)
// 	//		or out right remove them, but that will break existing procedures.

// 	return nil
// }

func loadJSON(path string, dest interface{}) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, dest)
}

func (csm *CSM) Import(ctx context.Context, datastore inventory.Datastore) error {
	var slsDumpstate sls_client.SlsState
	var hsmStateComponents hsm_client.ComponentArrayComponentArray
	var hsmHardwareInventory []hsm_client.HwInventory100HwInventoryByLocation

	// Load in data from test data directories for right now
	if err := loadJSON("./testdata/system/shandy/sls_dump.json", &slsDumpstate); err != nil {
		return err
	}
	if err := loadJSON("./testdata/system/shandy/hsm_state_components.json", &hsmStateComponents); err != nil {
		return err
	}
	if err := loadJSON("testdata/system/shandy/hsm_inventory_hardware.json", &hsmHardwareInventory); err != nil {
		return err
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

	//
	// THIS IS TEMP stuff for local testing
	//

	// For right now dump the contents of the inventory struct to disk
	importInventory, err := tempDatastore.List()
	if err != nil {
		panic(err)
	}
	importInventoryRaw, _ := json.MarshalIndent(importInventory, "", "  ")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("import_inventory_clone.json", importInventoryRaw, 0600)

	// Get the system UUID
	cSystem, err := tempDatastore.GetSystemZero()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to get system:0 ID"), err)
	}

	// Import related changes for SLS
	slsHardwareToAdd := map[string]sls_client.Hardware{}
	slsHardwareToModify := map[string]sls_client.Hardware{}
	slsHardwareExists := map[string]sls_client.Hardware{}

	// CANI Hardware changes
	// TODO

	// Check to see if a system object exists in the SLS dumpstate
	slsSystem, exists := slsDumpstate.Hardware["s0"]
	if !exists {
		log.Warn().Msgf("SLS does not contain a system object, creating one")

		slsSystem = sls.NewHardware(xnames.System{}, sls_client.HardwareClassRiver, sls_client.HardwareExtraPropertiesSystem{
			CaniId:               cSystem.ID.String(),
			CaniSlsSchemaVersion: "v1alpha1", // TODO make this a enum
			CaniLastModified:     time.Now().UTC().String(),
		})

		slsHardwareToAdd[slsSystem.Xname] = slsSystem
	}

	log.Info().Msgf("System: %v", slsSystem)

	//
	// Import Cabinets and Chassis
	//
	allCabinets, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Cabinet)
	allChassis, _ := sls.FilterHardwareByType(slsDumpstate.Hardware, xnametypes.Chassis)

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

		// slsCabinetEP, err := sls.DecodeExtraProperties[sls_client.HardwareExtraPropertiesCabinet](slsCabinet)
		// if err != nil {
		// 	return fmt.Errorf("failed to decode SLS hardware extra properties (%s)", slsCabinet.Xname)
		// }

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

			// For right now dump the contents of the inventory struct to disk
			importInventory, err = tempDatastore.List()
			if err != nil {
				panic(err)
			}
			importInventoryRaw, _ = json.MarshalIndent(importInventory, "", "  ")
			if err != nil {
				panic(err)
			}

			ioutil.WriteFile("import_inventory.json", importInventoryRaw, 0600)
			// END

			// Set cabinet metadata
			cabinetMetadata := CabinetMetadata{}
			if vlan, exists := cabinetHMNVlans[slsCabinet.Xname]; exists {
				cabinetMetadata.HMNVlan = IntPtr(vlan)
			}

			cCabinet, err = tempDatastore.GetAtLocation(locationPath)
			if err != nil {
				return errors.Join(fmt.Errorf("failed to query datastore for %s", locationPath), err)
			}

			cCabinet.Properties = map[string]interface{}{
				"csm": cabinetMetadata,
			}

			if err := tempDatastore.Update(&cCabinet); err != nil {
				return fmt.Errorf("failed to update hardware (%s) in memory datastore", cCabinet.ID)
			}

		} else {
			// Error occurred
			return errors.Join(fmt.Errorf("failed to query datastore"), err)
		}

		// if slsCabinetEP.CaniId != cCabinet.ID.String() {
		// 	if len(slsCabinetEP.CaniId) != 0 {
		// 		log.Warn().Msgf("Detected CANI hardware ID change from %s to %s for SLS Hardware %s", slsCabinetEP.CaniId, cCabinet.ID, slsCabinet.Xname)
		// 	}
		// 	slsCabinetEP.CaniId = cCabinet.ID.String()

		// 	log.Info().Msgf("SLS extra properties changed for %s", slsCabinet.Xname)

		// 	slsCabinet.ExtraProperties = slsCabinetEP
		// 	slsHardwareToModify[slsCabinet.Xname] = slsCabinet
		// }

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

	// 2. Find all slots holding blades from HSM, and inventory data
	for _, nodeBladeXname := range slsNodeBladeXnames {
		hsmComponent, exists := hsmStateComponentsMap[nodeBladeXname.String()]
		if !exists {
			log.Info().Msgf("%s exists in SLS, but not HSM", nodeBladeXname)
			continue
		}

		log.Info().Msgf("%s exists in HSM with state %s", nodeBladeXname, *hsmComponent.State)
		for _, nodeBMCXname := range slsNodeBladesFound[nodeBladeXname] {
			// For every BMC in HSM there is a NodeEnclosure. The NodeEnclosure ordinal matches
			// the BMC ordinal
			nodeEnclosureXname := nodeBladeXname.NodeEnclosure(nodeBMCXname.NodeBMC)

			nodeEnclosure, exists := hsmHardwareInventoryMap[nodeEnclosureXname.String()]
			if !exists {
				log.Warn().Msgf("%s is missing from HSM hardware inventory", nodeEnclosureXname)
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

			log.Info().Msgf("%s has manufacturer %s and model %s", nodeEnclosureXname, nodeEnclosureFru.Manufacturer, nodeEnclosureFru.Model)

			// TODO this needs to live in the hardware type library
			// Instead of hard coded here
			// deviceSlugMapping := map[string]string{
			// 	"WNC": "hpe-crayex-ex420-compute-blade",
			// 	"WindomNodeCard": "hpe-crayex-ex420-compute-blade"
			// }
		}
	}

	// // 2. Iterate through all node blades and try to identify the hardware
	// for nodeBladeXname, nodeCards := range nodeBlades {
	// 	log.Info().Msgf("%s: %v", nodeBladeXname.String(), nodeCards)
	// }

	//
	// Import Router BMCs
	//

	//
	// Handle phantom mountain/hill nodes
	//

	_ = slsHardwareToModify
	_ = slsHardwareExists

	//
	// THIS IS TEMP stuff for local testing
	//

	// For right now dump the contents of the inventory struct to disk
	importInventory, err = tempDatastore.List()
	if err != nil {
		panic(err)
	}
	importInventoryRaw, _ = json.MarshalIndent(importInventory, "", "  ")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("import_inventory.json", importInventoryRaw, 0600)

	slsHardwareToModifyRaw, _ := json.MarshalIndent(slsHardwareToModify, "", "  ")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("slsHardwareToModifyRaw.json", slsHardwareToModifyRaw, 0600)

	return nil
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

type Importer struct {
}

func (im *Importer) handleCabinets() {

}
