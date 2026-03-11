package devicetypes

// Debug controls verbose logging during type loading.
var Debug bool

// Package-level registries for device, module, cable, rack, and FRU types loaded from YAML files.
var (
	allDeviceTypes       = make(map[string]CaniDeviceType)
	deviceTypesByPartNum = make(map[string]CaniDeviceType)
	allModuleTypes       = make(map[string]CaniModuleType)
	moduleTypesByPartNum = make(map[string]CaniModuleType)
	allCableTypes        = make(map[string]CaniCableType)
	cableTypesByPartNum  = make(map[string]CaniCableType)
	allRackTypes         = make(map[string]CaniRackType)
	rackTypesByPartNum   = make(map[string]CaniRackType)
	allFruTypes          = make(map[string]CaniFruType)
	fruTypesByPartNum    = make(map[string]CaniFruType)
	allTypes             []Type
)

// Type represents a hardware classification type.
type Type string

const (
	TypeRack        Type = "rack"
	TypeCabinet     Type = "cabinet"
	TypeChassis     Type = "chassis"
	TypeBlade       Type = "blade"
	TypeNode        Type = "node"
	TypeNodeCard    Type = "nodecard"
	TypeSwitch      Type = "switch"
	TypeMgmtSwitch  Type = "mgmt-switch"
	TypeHsnSwitch   Type = "hsn-switch"
	TypeCabinetPDU  Type = "cabinet-pdu"
	TypeCDU         Type = "cdu"
	TypeModule      Type = "module"
	TypeNIC         Type = "nic"
	TypeGPU         Type = "gpu"
	TypeCPU         Type = "cpu"
	TypeMemory      Type = "memory"
	TypePowerSupply Type = "power-supply"
	TypeCable       Type = "cable"
	TypeFru         Type = "fru"
)

// Hardware type aliases for backwards compatibility.
const (
	Rack        = TypeRack
	Cabinet     = TypeCabinet
	Chassis     = TypeChassis
	Blade       = TypeBlade
	Node        = TypeNode
	NodeCard    = TypeNodeCard
	Switch      = TypeSwitch
	MgmtSwitch  = TypeMgmtSwitch
	HSNSwitch   = TypeHsnSwitch
	CabinetPDU  = TypeCabinetPDU
	CDU         = TypeCDU
	Module      = TypeModule
	NIC         = TypeNIC
	GPU         = TypeGPU
	CPU         = TypeCPU
	Memory      = TypeMemory
	PowerSupply = TypePowerSupply
	Cable       = TypeCable
	Fru         = TypeFru
)

// Category represents a classification category.
type Category string

const (
	CategoryDevice Category = "device"
	CategoryModule Category = "module"
	CategoryCable  Category = "cable"
	CategoryRack   Category = "rack"
	CategoryFru    Category = "fru"
)

// CableHardware is a list of cable/transceiver hardware types.
var CableHardware = []Type{TypeCable}

func init() {
	allTypes = []Type{
		TypeRack,
		TypeCabinet,
		TypeChassis,
		TypeBlade,
		TypeNode,
		TypeNodeCard,
		TypeSwitch,
		TypeMgmtSwitch,
		TypeHsnSwitch,
		TypeCabinetPDU,
		TypeCDU,
		TypeModule,
		TypeNIC,
		TypeGPU,
		TypeCPU,
		TypeMemory,
		TypePowerSupply,
		TypeCable,
		TypeFru,
	}
}

// ClassifyForNautobot returns the category for a hardware type.
func ClassifyForNautobot(hardwareType string) Category {
	switch Type(hardwareType) {
	case TypeRack, TypeCabinet:
		return CategoryRack
	case TypeCable:
		return CategoryCable
	case TypeNIC, TypeGPU, TypeCPU, TypeMemory, TypePowerSupply:
		return CategoryModule
	case TypeFru:
		return CategoryFru
	default:
		return CategoryDevice
	}
}

// ListCaniDeviceTypes returns all device types whose HardwareType matches
// any of the given types.
func ListCaniDeviceTypes(types ...Type) map[string]CaniDeviceType {
	result := make(map[string]CaniDeviceType)
	for slug, dt := range allDeviceTypes {
		for _, t := range types {
			if dt.HardwareType == string(t) {
				result[slug] = dt
				break
			}
		}
	}
	return result
}

// TypeEntry is a flat representation of any hardware type for listing purposes.
type TypeEntry struct {
	Name       string
	Slug       string
	PartNumber string
	Category   string
	Source     string
}

// ListAllAvailableTypes returns a flat list of every registered type across
// all registries (devices, modules, cables, racks, FRUs).
func ListAllAvailableTypes() []TypeEntry {
	var entries []TypeEntry

	for _, dt := range allDeviceTypes {
		entries = append(entries, TypeEntry{
			Name:       dt.Model,
			Slug:       dt.Slug,
			PartNumber: dt.PartNumber,
			Category:   dt.HardwareType,
			Source:     dt.Source,
		})
	}
	for _, mt := range allModuleTypes {
		entries = append(entries, TypeEntry{
			Name:       mt.Model,
			Slug:       mt.Slug,
			PartNumber: mt.PartNumber,
			Category:   mt.HardwareType,
			Source:     mt.Source,
		})
	}
	for _, ct := range allCableTypes {
		entries = append(entries, TypeEntry{
			Name:       ct.Model,
			Slug:       ct.Slug,
			PartNumber: ct.PartNumber,
			Category:   ct.HardwareType,
			Source:     ct.Source,
		})
	}
	for _, rt := range allRackTypes {
		entries = append(entries, TypeEntry{
			Name:       rt.Model,
			Slug:       rt.Slug,
			PartNumber: rt.PartNumber,
			Category:   rt.HardwareType,
			Source:     rt.Source,
		})
	}
	for _, ft := range allFruTypes {
		entries = append(entries, TypeEntry{
			Name:       ft.Model,
			Slug:       ft.Slug,
			PartNumber: ft.PartNumber,
			Category:   ft.HardwareType,
			Source:     ft.Source,
		})
	}
	return entries
}
