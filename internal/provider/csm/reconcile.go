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
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/ipam"
	"github.com/Cray-HPE/cani/internal/provider/csm/sls"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Reconcile CANI's inventory state with the external inventory state and apply required changes
// TODO perhaps Reconcile should return a ReconcileResult struct, that can contain what the provider wants to do
// This would enable these two things
//   - Provide a way to pass downwards the result, and allow for a custom string/Presentation function to
//     format the results in a human readable way
//   - Allow for a process like the following:
//     1. Figure out what has changed
//     2. Validate the changes
//     3. Display what changed
//     4. Make changes
func (csm *CSM) Reconcile(ctx context.Context, datastore inventory.Datastore, dryrun bool) (err error) {
	// TODO should we have a presentation callback to confirm the removal of hardware?

	log.Info().Msg("Starting CSM reconcile process")

	// TODO grab the allowed HSM Roles and SubRoles from HSM
	// This is for data validation

	//
	// Retrieve the current SLS state
	//
	currentSLSState, _, err := csm.slsClient.DumpstateApi.DumpstateGet(ctx)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to perform SLS dumpstate"),
			err,
		)
	}

	//
	// Retrieve the current BSS state
	//

	// bssGlobalBootParameters, err := csm.bssClient.GetBSSBootparametersByName("Global")
	// if err != nil {
	// 	return errors.Join(
	// 		fmt.Errorf("failed to retrieve Global Boot parameters"),
	// 		err,
	// 	)
	// }
	// managementNCNBootParams := map[string]*bssTypes.BootParams{}
	// for _, managementNCN := range managementNCNs {
	// 	log.Printf("Retrieving boot parameters for %s from BSS\n", managementNCN.Xname)
	// 	bootParams, err := bssClient.GetBSSBootparametersByName(managementNCN.Xname)
	// 	if err != nil {
	// 		log.Fatal("Error: ", err)
	// 	}

	// 	managementNCNBootParams[managementNCN.Xname] = bootParams

	// }

	//
	// Reconcile Network changes
	//
	networkChanges, err := reconcileNetworkChanges(datastore, *csm.hardwareLibrary, currentSLSState.Networks)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to reconcile network changes"), err)
	}

	//
	// Reconcile Hardware changes
	//
	hardwareChanges, err := reconcileHardwareChanges(*csm.hardwareLibrary, datastore, currentSLSState, networkChanges)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to reconcile hardware changes"), err)
	}

	//
	// Simulate and validate SLS actions
	//
	modifiedState, err := sls.CopyState(currentSLSState)
	if err != nil {
		return errors.Join(fmt.Errorf("unable to copy SLS state"), err)
	}
	for _, hardware := range hardwareChanges.Removed {
		delete(modifiedState.Hardware, hardware.Xname)
	}
	for _, hardware := range hardwareChanges.Added {
		modifiedState.Hardware[hardware.Xname] = hardware
	}
	for _, hardwarePair := range hardwareChanges.Changed {
		updatedHardware := hardwarePair.HardwareA
		modifiedState.Hardware[updatedHardware.Xname] = updatedHardware
	}
	for _, network := range networkChanges.ModifiedNetworks {
		modifiedState.Networks[network.Name] = network
	}

	_, err = validate.Validate(&modifiedState)
	if err != nil {
		return fmt.Errorf("Validation failed. %v\n", err)
	}

	//
	// Determine changes requires to downstream services from SLS. Like HSM and BSS
	//

	// TODO For right now lets just always push the host records, unless the reflect.DeepEqual
	// says they are equal.
	// Because the logic to compare the expected BSS host records with the current ones
	// is kind of hard. If anything is different just recalculate it.
	// This should be harmless just the order of records will shift around.

	// Recalculate the systems host recorded
	// TODO the following should be required for adding UAN/Management types of nodes and management switches
	// modifiedGlobalBootParameters := false
	// expectedGlobalHostRecords := bss.GetBSSGlobalHostRecords(managementNCNs, sls.Networks(currentSLSState))
	// var currentGlobalHostRecords bss.HostRecords
	// if err := mapstructure.Decode(bssGlobalBootParameters.CloudInit.MetaData["host_records"], &currentGlobalHostRecords); err != nil {
	// 	log.Fatal("Error: ", err)
	// }

	// if !reflect.DeepEqual(currentGlobalHostRecords, expectedGlobalHostRecords) {
	// 	log.Println("Host records in BSS Global boot parameters are out of date")
	// 	bssGlobalBootParameters.CloudInit.MetaData["host_records"] = expectedGlobalHostRecords
	// 	modifiedGlobalBootParameters = true
	// }

	// Recalculate cabinet routes
	// TODO NOTE this is the list of the managementNCNs before the topology of SLS changed.
	// modifiedManagementNCNBootParams := map[string]bool{}
	// for _, managementNCN := range managementNCNs {

	// 	// The following was stolen from CSI
	// 	extraNets := []string{}
	// 	var foundCAN = false
	// 	var foundCHN = false

	// 	for _, net := range sls.Networks(currentSLSState) {
	// 		if strings.ToLower(net.Name) == "can" {
	// 			extraNets = append(extraNets, "can")
	// 			foundCAN = true
	// 		}
	// 		if strings.ToLower(net.Name) == "chn" {
	// 			foundCHN = true
	// 		}
	// 	}
	// 	if !foundCAN && !foundCHN {
	// 		log.Fatal("Error no CAN or CHN network defined in SLS networks")
	// 	}

	// 	// IPAM
	// 	ipamNetworks := bss.GetIPAMForNCN(managementNCN, sls.Networks(currentSLSState), extraNets...)
	// 	expectedWriteFiles := bss.GetWriteFiles(sls.Networks(currentSLSState), ipamNetworks)

	// 	var currentWriteFiles []bss.WriteFile
	// 	if err := mapstructure.Decode(managementNCNBootParams[managementNCN.Xname].CloudInit.UserData["write_files"], &currentWriteFiles); err != nil {
	// 		panic(err)
	// 	}

	// 	// TODO For right now lets just always push the writefiles, unless the reflect.DeepEqual
	// 	// says they are equal.
	// 	// This should be harmless, the cabinet routes may be in a different order. This is due to cabinet routes do not overlap with each other.
	// 	if !reflect.DeepEqual(expectedWriteFiles, currentWriteFiles) {
	// 		log.Printf("Cabinet routes for %s in BSS Global boot parameters are out of date\n", managementNCN.Xname)
	// 		managementNCNBootParams[managementNCN.Xname].CloudInit.UserData["write_files"] = expectedWriteFiles
	// 		modifiedManagementNCNBootParams[managementNCN.Xname] = true
	// 	}

	// }

	// TODO determine changes to BSS NTP data to ensure that the {HMN,NMN}_{RVR,MTN} networks are present.

	if !dryrun {
		//
		// Modify the System's SLS instance
		//

		// Sort hardware so children are deleted before their parents
		sls.SortHardwareReverse(hardwareChanges.Removed)

		// Remove hardware that no longer exists
		for _, hardware := range hardwareChanges.Removed {
			log.Info().Str("xname", hardware.Xname).Msg("Removing")
			// Put into transaction log with old and new value
			// TODO

			// Perform a DELETE against SLS
			r, err := csm.slsClient.HardwareApi.HardwareXnameDelete(ctx, hardware.Xname)
			if err != nil {
				return errors.Join(
					fmt.Errorf("failed to delete hardware (%s) from SLS", hardware.Xname),
					err,
				)
			}
			log.Info().Int("status", r.StatusCode).Msg("Deleted hardware from SLS")
		}

		// Add hardware new hardware
		for _, hardware := range hardwareChanges.Added {
			log.Info().Str("xname", hardware.Xname).Msg("Adding")
			// Put into transaction log with old and new value
			// TODO

			// Perform a POST against SLS
			_, r, err := csm.slsClient.HardwareApi.HardwarePost(ctx, sls.NewHardwarePostOpts(hardware))
			if err != nil {
				return errors.Join(
					fmt.Errorf("failed to add hardware (%s) to SLS", hardware.Xname),
					err,
				)
			}
			log.Info().Int("status", r.StatusCode).Msg("Added hardware to SLS")
		}

		// Update existing hardware
		for _, hardwarePair := range hardwareChanges.Changed {
			updatedHardware := hardwarePair.HardwareA // A is expected, B is actual
			log.Info().Str("xname", updatedHardware.Xname).Msg("Updating")
			// Put into transaction log with old and new value
			// TODO

			// Perform a PUT against SLS
			_, r, err := csm.slsClient.HardwareApi.HardwareXnamePut(ctx, updatedHardware.Xname, sls.NewHardwareXnamePutOpts(updatedHardware))
			if err != nil {
				return errors.Join(
					fmt.Errorf("failed to update hardware (%s) from SLS", updatedHardware.Xname),
					err,
				)
			}
			log.Info().Int("status", r.StatusCode).Msg("Updated hardware to SLS")
		}

		// Update modified networks
		for _, network := range networkChanges.ModifiedNetworks {
			log.Info().Msgf("Updating SLS network %s", network.Name)

			// Perform a PUT against SLS
			_, r, err := csm.slsClient.NetworkApi.NetworksNetworkPut(ctx, network.Name, sls.NewNetworkApiNetworksNetworkPutOpts(network))
			if err != nil {
				return errors.Join(
					fmt.Errorf("failed to update hardware (%s) from SLS", network.Name),
					err,
				)
			}
			log.Info().Int("status", r.StatusCode).Msgf("Updated network %s in SLS", network.Name)
		}

		//
		// Modify the System's BSS instance
		//

		// if !modifiedGlobalBootParameters {
		// 	log.Println("No BSS Global boot parameters changes required")
		// } else {
		// 	log.Println("Updating BSS Global boot parameters")

		// 	if dryRun {
		// 		log.Println("  Dry run enabled not modifying BSS")
		// 	} else {
		// 		_, err := bssClient.UploadEntryToBSS(*bssGlobalBootParameters, http.MethodPut)
		// 		if err != nil {
		// 			log.Fatal("Error: ", err)
		// 		}
		// 	}
		// }

		// // Update per NCN BSS Boot parameters
		// for _, managementNCN := range managementNCNs {
		// 	xname := managementNCN.Xname

		// 	if !modifiedManagementNCNBootParams[xname] {
		// 		log.Printf("No changes to BSS boot parameters for %s\n", xname)
		// 		continue
		// 	}
		// 	log.Printf("Updating BSS boot parameters for %s\n", xname)

		// 	if dryRun {
		// 		log.Println("  Dry run enabled not modifying BSS")
		// 	} else {
		// 		_, err := bssClient.UploadEntryToBSS(*managementNCNBootParams[xname], http.MethodPut)
		// 		if err != nil {
		// 			log.Fatal("Error: ", err)
		// 		}

		// 	}
		// }
	} else {
		log.Warn().Msgf("Dryrun enabled, no changes performed!")
	}

	return nil
}

//
// Hardware Changes
//

type HardwareChanges struct {
	Removed   []sls_client.Hardware
	Added     []sls_client.Hardware
	Changed   []sls.HardwarePair
	Identical []sls_client.Hardware
}

func reconcileHardwareChanges(hardwareTypeLibrary hardwaretypes.Library, datastore inventory.Datastore, currentSLSState sls_client.SlsState, networkChanges *NetworkChanges) (*HardwareChanges, error) {
	//
	// Build up the expected SLS state
	//

	// First merge in any network changes
	expectedSLSNetworks, err := sls.CopyNetworks(currentSLSState.Networks)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("unable to copy SLS networks"), err)
	}
	for _, network := range networkChanges.ModifiedNetworks {
		expectedSLSNetworks[network.Name] = network
	}

	// Secondly generate the expected SLS state
	expectedSLSState, hardwareMapping, err := BuildExpectedHardwareState(hardwareTypeLibrary, datastore, expectedSLSNetworks)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to build expected SLS state"),
			err,
		)
	}

	// HACK Prune non-supported hardware from the current SLS state.
	// For right only remove all objects from the current SLS state that the SLS generator
	// has no idea about
	currentSLSState.Hardware, _ = sls.FilterHardware(currentSLSState.Hardware, func(hardware sls_client.Hardware) (bool, error) {
		_, exists := hardwareMapping[hardware.Xname]
		return exists, nil
	})

	//
	// Compare the current hardware state with the expected hardware state
	//

	hardwareRemoved, err := sls.HardwareSubtract(currentSLSState, expectedSLSState)
	if err != nil {
		return nil, err
	}

	hardwareAdded, err := sls.HardwareSubtract(expectedSLSState, currentSLSState)
	if err != nil {
		return nil, err
	}

	// Identify hardware present in both states
	// Does not take into account differences in Class/ExtraProperties, just by the primary key of xname
	identicalHardware, hardwareWithDifferingValues, err := sls.HardwareUnion(expectedSLSState, currentSLSState)
	if err != nil {
		return nil, err
	}

	if err := displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware, hardwareWithDifferingValues); err != nil {
		return nil, err
	}

	//
	// Verify expected hardware actions are taking place.
	// This can detect drift from when hardware was removed/added outside of CANI after the session was started
	//
	unexpectedHardwareRemoval := []sls_client.Hardware{}
	for _, hardware := range hardwareRemoved {
		if hardwareMapping[hardware.Xname].Status != inventory.HardwareStatusDecommissioned {
			// This piece of hardware wasn't flagged for removal from the system, but a
			// the reconcile logic wants to remove it and this is bad
			unexpectedHardwareRemoval = append(unexpectedHardwareRemoval, hardware)
		}

		// This piece of hardware is allowed to be removed from the system, as it has the right
		// inventory status
	}

	unexpectedHardwareAdditions := []sls_client.Hardware{}
	for _, hardware := range hardwareAdded {
		if hardwareMapping[hardware.Xname].Status != inventory.HardwareStatusStaged {
			// This piece of hardware wasn't flagged to be added to the system, but a
			// the reconcile logic wants to remove it and this is bad
			unexpectedHardwareAdditions = append(unexpectedHardwareAdditions, hardware)
		}
		// This piece of hardware is allowed to be added from the system, as it has the right
		// inventory status
	}

	// TODO need a good way to signal in the inventory structure that we need to support
	// update metadata for a piece of hardware, for now just not handle it
	// for _, hardware := range hardwareWithDifferingValues {
	// }

	if len(unexpectedHardwareRemoval) != 0 || len(unexpectedHardwareAdditions) != 0 {
		displayUnwantedChanges(unexpectedHardwareRemoval, unexpectedHardwareAdditions)
		return nil, fmt.Errorf("detected unexpected hardware changes between current and expected system states")
	}

	return &HardwareChanges{
		Added:     (hardwareAdded),
		Removed:   (hardwareRemoved),
		Identical: (identicalHardware),
		Changed:   (hardwareWithDifferingValues),
	}, nil
}

//
// IPAM/Network Changes Handling
//

type SubnetChange struct {
	NetworkName string
	Subnet      sls_client.NetworkIpv4Subnet
}

type IPReservationChange struct {
	NetworkName   string
	SubnetName    string
	IPReservation sls_client.NetworkIpReservation

	// TODO have a better description of what caused the changed

	// This is the hardware object that triggered the change
	// If empty, then this was not changed by hardware
	ChangedByXname string
}

type NetworkChanges struct {
	ModifiedNetworks map[string]sls_client.Network

	// The following fields are for book keeping to trigger other events
	SubnetsAdded        []SubnetChange
	IPReservationsAdded []IPReservationChange

	// TODO Add in HSM EthernetEthernetInterface information
	// This is needed if the state IP address range for a network needs to be expanded
	// so we can check to see if the IP has been allocated.
	// These issues need to be recorded, as the subnets DHCP range needs to be expanded.
}

func sortByLocationPath(allHardware map[uuid.UUID]inventory.Hardware) []inventory.Hardware {
	result := []inventory.Hardware{}

	// Convert to a slice
	for _, hardware := range allHardware {
		result = append(result, hardware)
	}

	// Perform the sort
	sort.Slice(result, func(i, j int) bool {
		// Simple way
		return result[i].LocationPath.String() < result[j].LocationPath.String()
	})

	return result
}

func reconcileNetworkChanges(datastore inventory.Datastore, hardwareTypeLibrary hardwaretypes.Library, networks map[string]sls_client.Network) (*NetworkChanges, error) {
	// Create lookup maps for hardware
	allHardware, err := datastore.List()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to list hardware from the datastore"), err)
	}

	// Create lookup maps for network extra properties for easier modified networks
	modifiedNetworks := map[string]bool{}
	networkExtraProperties := map[string]*sls_client.NetworkExtraProperties{}
	for networkName, slsNetwork := range networks {
		networkExtraProperties[networkName] = slsNetwork.ExtraProperties
	}

	// More bookkeeping to keep track of what network items have changed at a more granular level
	subnetsAdded := []SubnetChange{}
	ipReservationsAdded := []IPReservationChange{}

	// Allocate Cabinet Subnets
	for _, cabinet := range sortByLocationPath(allHardware.FilterHardwareByTypeStatus(inventory.HardwareStatusStaged, hardwaretypes.Cabinet)) {
		// Determine the xname of the cabinet
		locationPath, err := datastore.GetLocation(cabinet)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to get location path for cabinet (%v)", cabinet.ID), err)
		}

		xname, err := BuildXnameT[xnames.Cabinet](cabinet, locationPath)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to build xname for cabinet (%v)", cabinet.ID), err)
		}

		// Determine cabinet class
		class, err := DetermineHardwareClass(cabinet, allHardware, hardwareTypeLibrary)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to determine class of cabinet (%v)", xname), err)
		}

		// Allocation of the Cabinet Subnets
		log.Info().Msgf("Attempting to allocate cabinet subnets for %s", xname.String())
		for _, networkPrefix := range []string{"HMN", "NMN"} {
			networkName, err := determineCabinetNetwork(networkPrefix, class)
			if err != nil {
				return nil, err
			}

			// Retrieve the network
			networkExtraProperties, present := networkExtraProperties[networkName]
			if !present {
				return nil, fmt.Errorf("unable to allocate cabinet subnet network does not exist (%s)", networkName)
			}

			// Check to see if an subnet already exists
			if subnet, _, err := sls.LookupSubnetInEP(networkExtraProperties, fmt.Sprintf("cabinet_%d", xname.Cabinet)); err == nil {
				// Subnet Already exists
				log.Info().Msgf("Found existing subnet in network %s for cabinet %s with CIDR %v", networkName, xname, subnet.CIDR)
				continue
			} else if !errors.Is(err, sls.ErrSubnetNotFound) {
				return nil, errors.Join(fmt.Errorf("failed to lookup subnet for %d", xname))
			}

			// Find an available subnet
			subnet, err := ipam.AllocateCabinetSubnet(networkName, *networkExtraProperties, *xname, nil)
			if err != nil {
				return nil, fmt.Errorf("unable to allocate subnet for cabinet (%s) in network (%s)", xname, networkName)
			}

			log.Info().Msgf("Allocated subnet in network %s for cabinet %s with CIDR %v", networkName, xname, subnet.CIDR)

			// TODO Verify subnet VLAN is unique

			log.Printf("Allocated cabinet subnet %s with vlan %d in network %s for %s\n", subnet.CIDR, subnet.VlanID, networkName, xname)
			subnetsAdded = append(subnetsAdded, SubnetChange{
				NetworkName: networkName,
				Subnet:      subnet,
			})

			// Push in the newly created subnet into the SLS network
			networkExtraProperties.Subnets = append(networkExtraProperties.Subnets, subnet)
			modifiedNetworks[networkName] = true
		}
	}

	// Deallocate Cabinet Subnets
	for _, cabinet := range allHardware.FilterHardwareByTypeStatus(inventory.HardwareStatusDecommissioned, hardwaretypes.Cabinet) {
		// Determine the xname of the cabinet
		locationPath, err := datastore.GetLocation(cabinet)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to get location path for cabinet (%v)", cabinet.ID), err)
		}

		xname, err := BuildXnameT[xnames.Cabinet](cabinet, locationPath)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to build xname for cabinet (%v)", cabinet.ID), err)
		}

		return nil, fmt.Errorf("de-allocating subnets for cabinet (%s) is not currently supported", xname)
	}

	// Allocate Management Switch IPs
	for _, mgmtSwitch := range sortByLocationPath(allHardware.FilterHardwareByTypeStatus(inventory.HardwareStatusStaged, hardwaretypes.ManagementSwitch)) {
		// TODO in the future the code from here can be adapted: https://github.com/Cray-HPE/hardware-topology-assistant/blob/main/internal/engine/engine.go#L292-L392
		return nil, fmt.Errorf("allocating IP addresses for ManagementSwitch (%s) is not currently supported", mgmtSwitch.ID)
	}

	// Deallocate Management Switch IPs
	for _, mgmtSwitch := range allHardware.FilterHardwareByTypeStatus(inventory.HardwareStatusDecommissioned, hardwaretypes.ManagementSwitch) {
		return nil, fmt.Errorf("de-allocating IP addresses for switch (%s) is not currently supported", mgmtSwitch.ID)
	}

	// Allocate Node IPs
	for _, node := range sortByLocationPath(allHardware.FilterHardwareByTypeStatus(inventory.HardwareStatusStaged, hardwaretypes.Node)) {
		providerProperties, err := GetProviderMetadataT[NodeMetadata](node)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to get provider properties for node (%s)", node.ID), err)
		}

		var role, subRole string
		if providerProperties.Role != nil {
			role = *providerProperties.Role
		}
		if providerProperties.SubRole != nil {
			subRole = *providerProperties.SubRole
		}

		switch role {
		case "Application":
			if subRole != "UAN" {
				continue
			}

			// TODO the code from here can be adapted: https://github.com/Cray-HPE/hardware-topology-assistant/blob/main/internal/engine/engine.go#L394-L525

			return nil, fmt.Errorf("allocating IP addresses for UAN node (%s) is not currently supported", node.ID)
		case "Management":

			// TODO the logic from here can be adapted: https://github.com/Cray-HPE/docs-csm/blob/main/scripts/operations/node_management/Add_Remove_Replace_NCNs/add_management_ncn.py#L734

			return nil, fmt.Errorf("allocating IP addresses for Management node (%s) is not currently supported", node.ID)
		default:
			// Nothing to do here
		}
	}

	// Deallocate  Node IPs
	for _, node := range allHardware.FilterHardwareByTypeStatus(inventory.HardwareStatusStaged, hardwaretypes.Node) {
		providerProperties, err := GetProviderMetadataT[NodeMetadata](node)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to get provider properties for node (%s)", node.ID), err)
		}

		var role, subRole string
		if providerProperties.Role != nil {
			role = *providerProperties.Role
		}
		if providerProperties.SubRole != nil {
			subRole = *providerProperties.SubRole
		}

		switch role {
		case "Application":
			if subRole != "UAN" {
				continue
			}

			return nil, fmt.Errorf("de-allocating IP addresses for UAN node (%s) is not currently supported", node.ID)
		case "Management":
			return nil, fmt.Errorf("de-allocating IP addresses for Management node (%s) is not currently supported", node.ID)
		default:
			// Nothing to do here
		}
	}

	// Filter NetworkExtraProperties to include only the modified networks
	modifiedNetworksSet := map[string]sls_client.Network{}
	for networkName, networkExtraProperties := range networkExtraProperties {
		if !modifiedNetworks[networkName] {
			continue
		}

		// Merge extra properties with the top level network with SLS
		slsNetwork := networks[networkName]
		slsNetwork.ExtraProperties = networkExtraProperties

		// TODO update vlan range.

		modifiedNetworksSet[networkName] = slsNetwork
	}

	// TODO pretty print network changes

	return &NetworkChanges{
		ModifiedNetworks:    modifiedNetworksSet,
		SubnetsAdded:        subnetsAdded,
		IPReservationsAdded: ipReservationsAdded,
	}, nil

}

func determineCabinetNetwork(networkPrefix string, class sls_client.HardwareClass) (string, error) {
	var suffix string
	switch class {
	case sls_client.HardwareClassRiver:
		suffix = "_RVR"
	case sls_client.HardwareClassHill:
		fallthrough
	case sls_client.HardwareClassMountain:
		suffix = "_MTN"
	default:
		return "", fmt.Errorf("unknown cabinet class (%s)", class)
	}

	return networkPrefix + suffix, nil
}

//
// The following is taken from: https://github.com/Cray-HPE/hardware-topology-assistant/blob/main/internal/engine/engine.go
//

func displayHardwareComparisonReport(hardwareRemoved, hardwareAdded, identicalHardware []sls_client.Hardware, hardwareWithDifferingValues []sls.HardwarePair) error {
	log.Info().Msgf("")
	log.Info().Msgf("Identical hardware between current and expected states")
	if len(identicalHardware) == 0 {
		log.Info().Msgf("  None")
	}
	for _, hardware := range identicalHardware {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
	}

	log.Info().Msgf("")
	log.Info().Msgf("Common hardware between current and expected states with differing class or extra properties")
	if len(hardwareWithDifferingValues) == 0 {
		log.Info().Msg("  None")
	}
	for _, pair := range hardwareWithDifferingValues {
		log.Info().Msgf("  %s", pair.Xname)

		// Expected Hardware json
		pair.HardwareA.LastUpdated = 0
		pair.HardwareA.LastUpdatedTime = ""
		hardwareRaw, err := buildHardwareString(pair.HardwareA)
		if err != nil {
			return err
		}
		log.Info().Msgf("  - Expected: %-16s", hardwareRaw)

		// Actual Hardware json
		pair.HardwareB.LastUpdated = 0
		pair.HardwareB.LastUpdatedTime = ""
		hardwareRaw, err = buildHardwareString(pair.HardwareB)
		if err != nil {
			return err
		}
		log.Info().Msgf("  - Actual:   %-16s", hardwareRaw)
	}

	log.Info().Msgf("")
	log.Info().Msgf("Hardware added to the system")
	if len(hardwareAdded) == 0 {
		log.Info().Msgf("  None")
	}
	for _, hardware := range hardwareAdded {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
	}

	log.Info().Msgf("")
	log.Info().Msgf("Hardware removed from system")
	if len(hardwareRemoved) == 0 {
		log.Info().Msgf("  None")
	}
	for _, hardware := range hardwareRemoved {
		hardwareRaw, err := buildHardwareString(hardware)
		if err != nil {
			return err
		}

		log.Info().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
	}

	log.Info().Msgf("")
	return nil
}

func displayUnwantedChanges(unwantedHardwareRemoved, unwantedHardwareAdded []sls_client.Hardware) error {
	if len(unwantedHardwareAdded) != 0 {
		log.Error().Msgf("")
		log.Error().Msgf("Unexpected Hardware detected added to the system")
		for _, hardware := range unwantedHardwareAdded {
			hardwareRaw, err := buildHardwareString(hardware)
			if err != nil {
				return err
			}

			log.Error().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
		}
	}

	if len(unwantedHardwareRemoved) != 0 {
		log.Error().Msgf("")
		log.Error().Msgf("Unexpected Hardware detected removed from the system")
		for _, hardware := range unwantedHardwareRemoved {
			hardwareRaw, err := buildHardwareString(hardware)
			if err != nil {
				return err
			}

			log.Error().Msgf("  %-16s - %s", hardware.Xname, hardwareRaw)
		}
	}

	log.Info().Msgf("")
	return nil
}

func buildHardwareString(hardware sls_client.Hardware) (string, error) {
	// TODO include CANU UUID

	extraPropertiesRaw, err := hardware.DecodeExtraProperties()
	if err != nil {
		return "", err
	}

	var tokens []string
	tokens = append(tokens, fmt.Sprintf("Type: %s", hardware.TypeString))
	tokens = append(tokens, fmt.Sprintf("Class: %s", hardware.Class))

	switch hardware.TypeString {
	case xnametypes.Cabinet:
		// If we don't know how to pretty print it, lets just do the raw JSON
		// extraPropertiesRaw, err := json.Marshal(hardware.ExtraProperties)
		// if err != nil {
		// 	return "", err
		// }
		// tokens = append(tokens, string(extraPropertiesRaw))
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesCabinet); ok {
			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.DHCPRelaySwitches != nil {
				tokens = append(tokens, fmt.Sprintf("DHCPRelaySwitches: %s", strings.Join(extraProperties.DHCPRelaySwitches, ",")))
			}
			if extraProperties.Networks != nil {
				networksRaw, err := json.Marshal(extraProperties.Networks)
				if err != nil {
					return "", err
				}
				tokens = append(tokens, fmt.Sprintf("Networks: %s", networksRaw))
			}
		}

	case xnametypes.Chassis:
		// Nothing to do
	case xnametypes.ChassisBMC:
		// Nothing to do
	case xnametypes.CabinetPDUController:
		// Nothing to do
	case xnametypes.RouterBMC:
		// Nothing to do
	case xnametypes.NodeBMC:
		// Nothing to do
	case xnametypes.Node:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesNode); ok {
			tokens = append(tokens, fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")))
			if extraProperties.Role != "" {
				tokens = append(tokens, fmt.Sprintf("Role: %s", extraProperties.Role))
			}
			if extraProperties.SubRole != "" {
				tokens = append(tokens, fmt.Sprintf("SubRole: %s", extraProperties.SubRole))
			}
			if extraProperties.NID != 0 {
				tokens = append(tokens, fmt.Sprintf("NID: %d", extraProperties.NID))
			}
		}
	case xnametypes.MgmtSwitch:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesMgmtSwitch); ok {
			tokens = append(tokens,
				fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")),
				fmt.Sprintf("Brand: %s", extraProperties.Brand),
			)

			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.IP4addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP4addr: %s", extraProperties.IP4addr))
			}
			if extraProperties.IP6addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP6addr: %s", extraProperties.IP6addr))
			}

			tokens = append(tokens,
				fmt.Sprintf("SNMPUsername: %s", extraProperties.SNMPUsername),
				fmt.Sprintf("SNMPAuthProtocol: %s", extraProperties.SNMPAuthProtocol),
				fmt.Sprintf("SNMPAuthPassword: %s", extraProperties.SNMPAuthPassword),
				fmt.Sprintf("SNMPPrivProtocol: %s", extraProperties.SNMPPrivProtocol),
				fmt.Sprintf("SNMPPrivPassword: %s", extraProperties.SNMPPrivPassword),
			)
		}
	case xnametypes.MgmtHLSwitch:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesMgmtHlSwitch); ok {
			tokens = append(tokens,
				fmt.Sprintf("Aliases: [%s]", strings.Join(extraProperties.Aliases, ",")),
				fmt.Sprintf("Brand: %s", extraProperties.Brand),
			)

			if extraProperties.Model != "" {
				tokens = append(tokens, fmt.Sprintf("Model: %s", extraProperties.Model))
			}
			if extraProperties.IP4addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP4addr: %s", extraProperties.IP4addr))
			}
			if extraProperties.IP6addr != "" {
				tokens = append(tokens, fmt.Sprintf("IP6addr: %s", extraProperties.IP6addr))
			}
		}
	case xnametypes.MgmtSwitchConnector:
		if extraProperties, ok := extraPropertiesRaw.(sls_client.HardwareExtraPropertiesMgmtSwitchConnector); ok {
			tokens = append(tokens,
				fmt.Sprintf("VendorName: %s", extraProperties.VendorName),
				fmt.Sprintf("NodeNics: [%s]", strings.Join(extraProperties.NodeNics, ",")),
			)
		}
	default:
		// If we don't know how to pretty print it, lets just do the raw JSON
		hardwareRaw, err := json.Marshal(hardware)
		if err != nil {
			return "", err
		}
		tokens = append(tokens, string(hardwareRaw))
	}

	return strings.Join(tokens, ", "), nil
}
