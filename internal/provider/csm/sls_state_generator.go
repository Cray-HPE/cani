package csm

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/rs/zerolog/log"
)

func BuildExpectedHardwareState(datastore inventory.Datastore) (sls_common.SLSState, error) {
	// Retrieve the CANI inventory data
	data, err := datastore.List()
	if err != nil {
		return sls_common.SLSState{}, errors.Join(
			fmt.Errorf("failed to list hardware from the datastore"),
			err,
		)
	}

	// Iterate over the CANI inventory data to build SLS data
	allHardware := map[string]sls_common.GenericHardware{}
	for _, cHardware := range data.Hardware {
		//
		// Build the SLS hardware representation
		//
		log.Debug().Any("cHardware", cHardware).Msg("Processing")
		locationPath, err := datastore.GetLocation(cHardware)
		if err != nil {
			return sls_common.SLSState{}, errors.Join(
				fmt.Errorf("failed to get location of hardware (%s) from the datastore", cHardware.ID),
				err,
			)
		}

		hardware, err := BuildSLSHardware(cHardware, locationPath)
		// if err != nil && ignoreUnknownCANUHardwareArchitectures && strings.Contains(err.Error(), "unknown architecture type") {
		// 	log.Printf("WARNING %s", err.Error())
		// } else if err != nil {
		if err != nil {
			return sls_common.SLSState{}, err
		}

		log.Debug().Any("hardware", hardware).Msg("Generated SLS hardware")

		// Ignore empty hardware
		if hardware.Xname == "" {
			continue
		}

		// Verify cabinet exists (ignore CDUs)
		// TODO
		// if strings.HasPrefix(hardware.Xname, "x") {
		// 	cabinetXname, err := csi.CabinetForXname(hardware.Xname)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	if !cabinetLookup.CabinetExists(cabinetXname) {
		// 		err := fmt.Errorf("unknown cabinet (%s)", cabinetXname)
		// 		panic(err)
		// 	}
		// }

		// Verify new hardware
		if _, present := allHardware[hardware.Xname]; present {
			err := fmt.Errorf("found duplicate xname %v", hardware.Xname)
			panic(err)
		}

		allHardware[hardware.Xname] = hardware

		//
		// Build up derived hardware
		//
		// TODO This is probably not needed as the CANI Inventory will have all that we need
		// if hardware.TypeString == xnametypes.ChassisBMC {
		// 	allHardware[hardware.Xname] = sls_common.NewGenericHardware(hardware.Parent, hardware.Class, nil)
		// }

		//
		// Build the MgmtSwitchConnector for the hardware
		//
		mgmtSwtichConnector, err := BuildSLSMgmtSwitchConnector(hardware, cHardware)
		if err != nil {
			panic(err)
		}

		// Ignore empty mgmtSwtichConnectors
		if mgmtSwtichConnector.Xname == "" {
			continue
		}

		if _, present := allHardware[mgmtSwtichConnector.Xname]; present {
			err := fmt.Errorf("found duplicate xname %v", mgmtSwtichConnector.Xname)
			panic(err)
		}

		allHardware[mgmtSwtichConnector.Xname] = mgmtSwtichConnector

	}

	// Generate Cabinet Objects
	// TODO this will be handled in the code above ^
	// for cabinetKind, cabinets := range cabinetLookup {
	// 	for _, cabinet := range cabinets {
	// 		class, err := cabinetKind.Class()
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		extraProperties := sls_common.ComptypeCabinet{
	// 			Networks: map[string]map[string]sls_common.CabinetNetworks{}, // TODO this should be outright removed. MEDS and KEA no longer look here for network info, but MEDS still needs this key to exist.
	// 		}

	// 		if cabinetKind.IsModel() {
	// 			extraProperties.Model = string(cabinetKind)
	// 		}

	// 		hardware := sls_common.NewGenericHardware(cabinet, class, extraProperties)

	// 		// Verify new hardware
	// 		if _, present := allHardware[hardware.Xname]; present {
	// 			err := fmt.Errorf("found duplicate xname %v", hardware.Xname)
	// 			panic(err)
	// 		}

	// 		allHardware[hardware.Xname] = hardware
	// 	}
	// }

	// Build up and the SLS state
	return sls_common.SLSState{
		Hardware: allHardware,
	}, nil
}

func BuildSLSHardware(cHardware inventory.Hardware, locationPath inventory.LocationPath) (sls_common.GenericHardware, error) {
	log.Debug().Any("locationPath", locationPath).Msg("Location path")

	// Get the physical location for the hardware
	xname, err := BuildXname(cHardware, locationPath)
	log.Debug().Any("xname", xname).Err(err).Msg("Build xname")
	if err != nil {
		return sls_common.GenericHardware{}, err
	} else if xname == nil {
		// This means that this piece of the hardware inventory can't be represented in SLS, so just skip it
		return sls_common.GenericHardware{}, nil
	}

	// Get the class of the piece of hardware
	// Generally this will match the class of the containing cabinet, the exception is river hardware within a EX2500 cabinet.
	// TODO
	class := sls_common.ClassMountain

	switch cHardware.Type {
	case hardwaretypes.HardwareTypeCabinet:
		return sls_common.GenericHardware{}, nil
	case hardwaretypes.HardwareTypeChassis:
		return sls_common.GenericHardware{}, nil
	case hardwaretypes.HardwareTypeNodeBlade:
		return sls_common.GenericHardware{}, nil
	case hardwaretypes.HardwareTypeNodeCard:
		return sls_common.GenericHardware{}, nil
	case hardwaretypes.HardwareTypeNodeController:
		return sls_common.GenericHardware{}, nil
	case hardwaretypes.HardwareTypeNode:
		ep := sls_common.NetworkExtraProperties{}

		// In order to properly populate SLS several bits of information are required.
		// This information should have been collected when hardware was added to the inventory
		// - NID
		// - Alias/Common Name
		// - Role
		// - SubRole

		return sls_common.NewGenericHardware(xname.String(), class, ep), nil
	}

	return sls_common.GenericHardware{}, fmt.Errorf("unknown hardware type '%s'", cHardware.Type)
}

// func buildSLSPDUController(location Location) (sls_common.GenericHardware, error) {
// }

// func buildSLSSlingshotHSNSwitch(location Location) (sls_common.GenericHardware, error) {
// }

// func buildSLSCMC(location Location) (sls_common.GenericHardware, error) {
// 	// TODO what should be done if if the CMC does not have a bmc connection? Ie the Intel CMC that doesn't really exist
// 	// Right now we are emulating the current behavior of CSI, where the fake CMC exists in SLS and no MgmtSwitchConnector exists.

// }

// // BuildNodeExtraProperties will attempt to build up all of the known extra properties form a Node present in a CCJ.
// // Limiitations the following information is not populated:
// // - Management NCN NID
// // - Application Node Subrole and Alias

// func BuildNodeExtraProperties(topologyNode TopologyNode) (extraProperties sls_common.ComptypeNode, err error) {
// }

// func buildSLSNode(xname) (sls_common.GenericHardware, error) {
// }

// func buildSLSMgmtSwitch(topologyNode TopologyNode, switchAliasesOverrides map[string][]string) (sls_common.GenericHardware, error) {
// }

// func buildSLSMgmtHLSwitch(topologyNode TopologyNode, switchAliasesOverrides map[string][]string) (sls_common.GenericHardware, error) {
// }

// func buildSLSCDUMgmtSwitch(topologyNode TopologyNode, switchAliasesOverrides map[string][]string) (sls_common.GenericHardware, error) {
// }

func BuildSLSMgmtSwitchConnector(hardware sls_common.GenericHardware, cHardware inventory.Hardware) (sls_common.GenericHardware, error) {
	return sls_common.GenericHardware{}, nil
}

// func buildSLSChassisBMC(location Location, cl configs.CabinetLookup) (sls_common.GenericHardware, error) {
// }
