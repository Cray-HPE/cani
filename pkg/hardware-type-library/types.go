package hardware_type_library

type HardwareType string

const (
	HardwareTypeCabinet                        HardwareType = "Cabinet"
	HardwareTypeChassis                        HardwareType = "Chassis"
	HardwareTypeChassisManagementModule        HardwareType = "ChassisManagementModule"
	HardwareTypeCabinetEnvironmentalController HardwareType = "CabinetEnvironmentalController"
	HardwareTypeNodeBlade                      HardwareType = "NodeBlade"
	HardwareTypeNodeBMC                        HardwareType = "NodeBMC"
	HardwareTypeNodeCard                       HardwareType = "NodeCard"
	HardwareTypeNode                           HardwareType = "Node"
	HardwareTypeManagementSwitch               HardwareType = "ManagementSwitch"
	HardwareTypeHighSpeedSwitchBlade           HardwareType = "HardwareTypeHighSpeedSwitch"
	HardwareTypeCabinetPDUController           HardwareType = "CabinetPDUController"
	HardwareTypePDU                            HardwareType = "PDU"
)

type Airflow string

const (
	AirflowFrontToRear Airflow = "front-to-rear"
	AirflowRearToFront Airflow = "rear-to-front"
	AirflowLeftToRight Airflow = "left-to-right"
	AirflowRightToLeft Airflow = "right-to-left"
	AirflowSideToRear  Airflow = "side-to-rear"
	AirflowPassive     Airflow = "passive"
)

type WeightUnit string

const (
	WeightUnitKiloGram WeightUnit = "kg"
	WeightUnitGram     WeightUnit = "g"
	WeightUnitPound    WeightUnit = "lb"
	WeightUnitOunce    WeightUnit = "oz"
)

type SubDeviceRole string

const (
	SubDeviceRoleParent SubDeviceRole = "parent"
	SubDeviceRoleChild  SubDeviceRole = "child"
)

type DeviceType struct {
	Manufacturer string       `yaml:"manufacturer"`
	Model        string       `yaml:"model"`
	HardwareType HardwareType `yaml:"hardware-type"`
	Slug         string       `yaml:"slug"`

	PartNumber  *string     `yaml:"part_number"`
	UHeight     *float64    `yaml:"u_height"`
	IsFullDepth *bool       `yaml:"is_full_depth"`
	Weight      *float64    `yaml:"weight"`
	WeightUnit  *WeightUnit `yaml:"weight_unit"`

	FrontImage bool `yaml:"front_image"`
	RearImage  bool `yaml:"rear_image"`

	SubDeviceRole SubDeviceRole `yaml:"subdevice_role"`

	// TODO
	// ConsolePorts       []ConsolePort       `yaml:"console-ports"`
	// ConsoleServerPorts []ConsoleServerPort `yaml:"console-server-ports"`
	// PowerPowers        []PowerPower        `yaml:"power-ports"`
	// PowerOutlets       []PowerOutlets      `yaml:"power-outlets"`

	DeviceBays []DeviceBay `yaml:"device-bays"`
}

type DeviceBay struct {
	Name    string           `yaml:"name"`
	Allowed *AllowedHardware `yaml:"allowed"`
	Default *DefaultHardware `yaml:"default"`
}

type AllowedHardware struct {
	HardwareTypes []HardwareType `yaml:"hardware-type"`
	Slug          []string       `yaml:"slug"`
}

type DefaultHardware struct {
	Slug string `yaml:"slug"`
}

// TODO

// type ConsolePort struct {
// }

// type ConsoleServerPort struct {
// }

// type PowerPower struct {
// }
// type PowerOutlets struct {
// }

// type Interface struct {
// }

// type FrontPort struct {
// }

// type RearPort struct {
// }

// type ModuleBay struct {
// }

// type InventoryItem struct {
// }