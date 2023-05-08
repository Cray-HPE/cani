package inventory

import (
	"net"

	"github.com/google/uuid"
)

// // Inventory table
// type PhysicalInventory struct {
// 	Cabinets map[uuid.UUID]Hardware

// 	Chassis     map[uuid.UUID]Hardware
// 	ChassisBMCs map[uuid.UUID]Hardware

// 	// A node Blade is compose of node BMC(s), node(s), and CMC(s)
// 	// and is treated as a single physical unit when added or removed.
// 	// Ie when removing a node, you are also removing its BMC.
// 	NodeBlades  map[uuid.UUID]Hardware
// 	NodeBMCs    map[uuid.UUID]Hardware
// 	NodeBMCNICs map[uuid.UUID]Hardware
// 	Nodes       map[uuid.UUID]Hardware
// 	NodeNICs    map[uuid.UUID]Hardware

// 	// HSN Switches like Slingshot. Should these be treated like normal management switches?
// 	HighSpeedNetworkSwitches   map[uuid.UUID]Hardware
// 	HighSpeedNetworkSwitchBMCs map[uuid.UUID]Hardware

// 	ManagementSwitches map[uuid.UUID]Hardware
// }

// type Hardware struct {
// 	ID uuid.UUID
// }

// // Question if we have the concept of a system, could all for inventory to be aware of multiple systems/sites
// // Like move node from one system to another system
// type PhysicalSystem struct {
// 	ID uuid.UUID
// }

// type PhysicalCabinet struct {
// 	ID uuid.UUID

// 	HardwareType string
// }

// // type LogicalCabinet struct {
// // 	ID              uuid.UUID
// // 	PhysicalCabinet uuid.UUID

// // 	Subnet net.IPNet
// // }

// // type PhysicalChassis struct {
// // }

// // type PhysicalChassisBMC struct {
// // }

// // type PhysicalNodeBlade struct {
// // }

// // type PhysicalNodeBMC struct {
// // }

// // type PhysicalNode struct {
// // }

// // type PhysicalNodeNICs struct {
// // }

// TODO better name for locations

type ArchitectureType string

const (
	TypeSystem                 = ArchitectureType("System")
	TypeCDU                    = ArchitectureType("CDU")
	TypeCabinet                = ArchitectureType("Cabinet")
	TypeCEC                    = ArchitectureType("CEC")
	TypeChassis                = ArchitectureType("Chassis")
	TypeChassisBMC             = ArchitectureType("ChassisBMC")
	TypeCabinetPDUController   = ArchitectureType("CabinetPDUController")
	TypePDU                    = ArchitectureType("PDU")
	TypeManagementSwitch       = ArchitectureType("ManagementSwitch")
	TypeHighSpeedNetworkSwitch = ArchitectureType("HighSpeedNetworkSwitch")
	TypeNodeBlade              = ArchitectureType("NodeBlade")
	TypeNodeBMC                = ArchitectureType("NodeBMC")
	TypeNode                   = ArchitectureType("Node")
)

type AllowedHierarchyNode struct {
	Parents []ArchitectureType
}

var allowedHierarchy = map[ArchitectureType]AllowedHierarchyNode{
	TypeSystem:                 {},
	TypeCDU:                    {[]ArchitectureType{TypeSystem}},
	TypeCabinet:                {[]ArchitectureType{TypeSystem}},
	TypeCEC:                    {[]ArchitectureType{TypeCabinet}},
	TypeChassis:                {[]ArchitectureType{TypeCabinet}},
	TypeChassisBMC:             {[]ArchitectureType{TypeChassis}},
	TypeCabinetPDUController:   {[]ArchitectureType{TypeCabinet}},
	TypePDU:                    {[]ArchitectureType{TypeCabinetPDUController}},
	TypeManagementSwitch:       {[]ArchitectureType{TypeChassis, TypeCDU}},
	TypeHighSpeedNetworkSwitch: {[]ArchitectureType{TypeChassis}},
	TypeNodeBlade:              {[]ArchitectureType{TypeChassis}},
	TypeNodeBMC:                {[]ArchitectureType{TypeNodeBlade}},
	TypeNode:                   {[]ArchitectureType{TypeNodeBMC}},
}

type Hardware struct {
	ID       uuid.UUID
	ParentID uuid.UUID

	ArchitectureType ArchitectureType
	Model            string

	Properties interface{} // Could do it this way, or do a pointer for each of the properties.
}

func NewHardware(architectureType ArchitectureType, model string, properties interface{}) *Hardware {
	return &Hardware{
		ID:               uuid.New(),
		ArchitectureType: architectureType,
		Model:            model,
		Properties:       properties,
	}
}

type SystemProperties struct {
	Name string
}

type CabinetProperties struct {
	// Physical information
	CabinetNumber int

	// Logical information
	VlanHMN int
	VlanNMN int
	Subnet  net.IPNet
}

type CECProperties struct {
	// Physical
	Location int
}

type ChassisProperties struct {
	// Physical Location
	Location int
}

type ChassisBMCProperties struct {
}

type CabinetPDUControllerProperties struct {
	// Physical Location
	Location int

	// Topology Information
	// This is starting to get into CANU territory with how the system is cabled up
	Switch     *uuid.UUID
	SwitchPort *int
}

type PDUProperties struct {
	// Physical Location
	Location int
}

type HighSpeedNetworkSwitch struct {
	// Physical Location
	Location int
}

type HighSpeedNetworkSwitchBMC struct {
	// Physical
	// Location int

	// Topology Information
	// This is starting to get into CANU territory with how the system is cabled up
	Switch     *uuid.UUID
	SwitchPort *int
}

type ManagementSwitchProperties struct {
	// Physical Location
	Location int
}

type NodeBladeProperties struct {
	// Physical Location
	Location int
}

type NodeBMCProperties struct {
	// Physical Location
	Location int

	MACAddress      *net.HardwareAddr
	DefaultUserName *string
	DefaultPassword *string

	// Topology Information
	// This is starting to get into CANU territory with how the system is cabled up
	Switch     *uuid.UUID
	SwitchPort *int
}

type NodeProperties struct {
	// Physical Location
	Location int

	// Logical Information/System Role information
	// This right here is a lot of CSM-isms, and might be good to find a way to make this
	// not baked in at this level.
	Role    *string
	SubRole *string
	NID     *int
	Aliases []string
}
