package devicetypes

// Test coverage table for registry.go:
//
// | Function              | Happy-path test                              | Failure test                                  |
// |-----------------------|----------------------------------------------|-----------------------------------------------|
// | ClassifyForNautobot   | TestClassifyForNautobotKnownTypes            | TestClassifyForNautobotUnknownType            |
// | ListCaniDeviceTypes   | TestListCaniDeviceTypesMatchesHardwareType    | TestListCaniDeviceTypesNoMatch                |
// | ListAllAvailableTypes | TestListAllAvailableTypesReturnsAllRegistered | TestListAllAvailableTypesEmptyRegistries       |

import "testing"

// ---------- ClassifyForNautobot ----------

func TestClassifyForNautobotKnownTypes(t *testing.T) {
	cases := []struct {
		input    string
		expected Category
	}{
		{string(TypeRack), CategoryRack},
		{string(TypeCabinet), CategoryRack},
		{string(TypeCable), CategoryCable},
		{string(TypeNIC), CategoryModule},
		{string(TypeGPU), CategoryModule},
		{string(TypeCPU), CategoryModule},
		{string(TypeMemory), CategoryModule},
		{string(TypePowerSupply), CategoryModule},
		{string(TypeFru), CategoryFru},
		{string(TypeChassis), CategoryDevice},
		{string(TypeBlade), CategoryDevice},
		{string(TypeNode), CategoryDevice},
		{string(TypeSwitch), CategoryDevice},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := ClassifyForNautobot(tc.input)
			if got != tc.expected {
				t.Errorf("ClassifyForNautobot(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestClassifyForNautobotUnknownType(t *testing.T) {
	got := ClassifyForNautobot("nonexistent-type")
	if got != CategoryDevice {
		t.Errorf("ClassifyForNautobot(%q) = %q, want default %q", "nonexistent-type", got, CategoryDevice)
	}
}

// ---------- ListCaniDeviceTypes ----------

func TestListCaniDeviceTypesMatchesHardwareType(t *testing.T) {
	// Save original map and restore after test.
	orig := make(map[string]CaniDeviceType, len(allDeviceTypes))
	for k, v := range allDeviceTypes {
		orig[k] = v
	}
	defer func() {
		allDeviceTypes = orig
	}()

	allDeviceTypes = map[string]CaniDeviceType{
		"chassis-a": {Slug: "chassis-a", Model: "Chassis A", Type: TypeChassis},
		"blade-b":   {Slug: "blade-b", Model: "Blade B", Type: TypeBlade},
		"node-c":    {Slug: "node-c", Model: "Node C", Type: TypeNode},
	}

	result := ListCaniDeviceTypes(TypeChassis, TypeBlade)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if _, ok := result["chassis-a"]; !ok {
		t.Error("expected chassis-a in results")
	}
	if _, ok := result["blade-b"]; !ok {
		t.Error("expected blade-b in results")
	}
	if _, ok := result["node-c"]; ok {
		t.Error("node-c should not be in results")
	}
}

func TestListCaniDeviceTypesNoMatch(t *testing.T) {
	orig := make(map[string]CaniDeviceType, len(allDeviceTypes))
	for k, v := range allDeviceTypes {
		orig[k] = v
	}
	defer func() {
		allDeviceTypes = orig
	}()

	allDeviceTypes = map[string]CaniDeviceType{
		"blade-x": {Slug: "blade-x", Model: "Blade X", Type: TypeBlade},
	}

	result := ListCaniDeviceTypes(TypeCDU)
	if len(result) != 0 {
		t.Errorf("expected 0 results for non-matching type, got %d", len(result))
	}
}

// ---------- ListAllAvailableTypes ----------

func TestListAllAvailableTypesReturnsAllRegistered(t *testing.T) {
	// Save and restore all registries.
	origDevices := allDeviceTypes
	origModules := allModuleTypes
	origCables := allCableTypes
	origRacks := allRackTypes
	origFrus := allFruTypes
	origLocations := allLocationTypes
	defer func() {
		allDeviceTypes = origDevices
		allModuleTypes = origModules
		allCableTypes = origCables
		allRackTypes = origRacks
		allFruTypes = origFrus
		allLocationTypes = origLocations
	}()

	allDeviceTypes = map[string]CaniDeviceType{
		"dev1": {Model: "Dev1", Slug: "dev1", PartNumber: "PN-D1", Type: "chassis", Source: "test"},
	}
	allModuleTypes = map[string]CaniModuleType{
		"mod1": {Model: "Mod1", Slug: "mod1", PartNumber: "PN-M1", Type: "gpu", Source: "test"},
	}
	allCableTypes = map[string]CaniCableType{
		"cab1": {Model: "Cab1", Slug: "cab1", PartNumber: "PN-C1", Type: "cable", Source: "test"},
	}
	allRackTypes = map[string]CaniRackType{
		"rack1": {Model: "Rack1", Slug: "rack1", PartNumber: "PN-R1", Type: "rack", Source: "test"},
	}
	allFruTypes = map[string]CaniFruType{
		"fru1": {Model: "Fru1", Slug: "fru1", PartNumber: "PN-F1", Type: "fru", Source: "test"},
	}
	allLocationTypes = map[string]LocationTypeDefinition{}

	entries := ListAllAvailableTypes()
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	slugs := make(map[string]bool, len(entries))
	for _, e := range entries {
		slugs[e.Slug] = true
	}
	for _, want := range []string{"dev1", "mod1", "cab1", "rack1", "fru1"} {
		if !slugs[want] {
			t.Errorf("expected slug %q in results", want)
		}
	}
}

func TestListAllAvailableTypesEmptyRegistries(t *testing.T) {
	origDevices := allDeviceTypes
	origModules := allModuleTypes
	origCables := allCableTypes
	origRacks := allRackTypes
	origFrus := allFruTypes
	origLocations := allLocationTypes
	defer func() {
		allDeviceTypes = origDevices
		allModuleTypes = origModules
		allCableTypes = origCables
		allRackTypes = origRacks
		allFruTypes = origFrus
		allLocationTypes = origLocations
	}()

	allDeviceTypes = map[string]CaniDeviceType{}
	allModuleTypes = map[string]CaniModuleType{}
	allCableTypes = map[string]CaniCableType{}
	allRackTypes = map[string]CaniRackType{}
	allFruTypes = map[string]CaniFruType{}
	allLocationTypes = map[string]LocationTypeDefinition{}

	entries := ListAllAvailableTypes()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries from empty registries, got %d", len(entries))
	}
}
