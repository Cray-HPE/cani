package csm

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/rs/zerolog/log"
)

func BuildExpectedHardwareState(data inventory.Inventory) (sls_common.SLSState, error) {
	// Iterate over the CANI inventory data to build SLS data
	allHardware := map[string]sls_common.GenericHardware{}
	for _, cHardware := range data.Hardware {
		//
		// Build the SLS hardware representation
		//
		log.Debug().Any("cHardware", cHardware).Msg("Processing")
		hardware, err := BuildSLSHardware(cHardware)
		// if err != nil && ignoreUnknownCANUHardwareArchitectures && strings.Contains(err.Error(), "unknown architecture type") {
		// 	log.Printf("WARNING %s", err.Error())
		// } else if err != nil {
		if err != nil {
			return sls_common.SLSState{}, err
		}

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

func BuildSLSHardware(cHardware inventory.Hardware) (sls_common.GenericHardware, error) {
	// TODO use CANU files for lookup
	// ALso look at using type

	// switch topologyNode.Architecture {
	// case "kvm":
	// 	// TODO SLS does not know anything about KVM, because HMS software doesn't support them.
	// 	fallthrough
	// case "cec":
	// 	// TODO SLS does not know anything about CEC, because HMS software doesn't support them.
	// 	return sls_common.GenericHardware{}, nil
	// case "cmm":
	// 	return buildSLSChassisBMC(topologyNode.Location, cabinetLookup)
	// case "subrack":
	// 	return buildSLSCMC(topologyNode.Location)
	// case "pdu":
	// 	return buildSLSPDUController(topologyNode.Location)
	// case "slingshot_hsn_switch":
	// 	return buildSLSSlingshotHSNSwitch(topologyNode.Location)
	// case "mountain_compute_leaf": // CDUMgmtSwitch
	// 	if strings.HasPrefix(topologyNode.Location.Rack, "x") {
	// 		// This CDU MgmtSwitch is present in a river cabinet.
	// 		// This is normally seen on newer TDS/Hill cabinet systems
	// 		return buildSLSMgmtHLSwitch(topologyNode, switchAliasesOverrides)
	// 	} else {
	// 		// Otherwise the switch is in a CDU cabinet
	// 		return buildSLSCDUMgmtSwitch(topologyNode, switchAliasesOverrides)
	// 	}
	// case "customer_edge_router":
	// 	fallthrough
	// case "spine":
	// 	fallthrough
	// case "river_ncn_leaf":
	// 	return buildSLSMgmtHLSwitch(topologyNode, switchAliasesOverrides)
	// case "river_bmc_leaf":
	// 	return buildSLSMgmtSwitch(topologyNode, switchAliasesOverrides)
	// default:
	// 	// There are a lot of architecture types that can be a node, but for SLS we just need to know that it is a server
	// 	// of some sort.
	// 	if topologyNode.Type == "node" || topologyNode.Type == "server" {
	// 		// All node architecture needs to go through this function
	// 		return buildSLSNode(topologyNode, paddle, applicationNodeMetadata)
	// 	}
	// }
	//
	// return sls_common.GenericHardware{}, fmt.Errorf("unknown architecture type %s for CANU common name %s", topologyNode.Architecture, topologyNode.CommonName)

	switch cHardware.Type {
	case hardwaretypes.HardwareTypeNodeBlade:
	case hardwaretypes.HardwareTypeNodeCard:
	case hardwaretypes.HardwareTypeNode:
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

// func buildSLSNode(topologyNode TopologyNode, paddle Paddle, applicationNodeMetadata configs.ApplicationNodeMetadataMap) (sls_common.GenericHardware, error) {
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
