package devicetypes

// Test coverage for all.go
//
// | Function                       | Happy-path test                                  | Failure test                                        |
// |--------------------------------|--------------------------------------------------|-----------------------------------------------------|
// | All                            | registered device returned                       | empty when no devices                               |
// | ByPartNumber                   | map keyed by part number                         | empty when no devices                               |
// | GetByPartNumber                | finds existing                                   | false for missing                                   |
// | GetBySlug                      | finds existing                                   | false for missing                                   |
// | AllCables                      | registered cables returned                       | empty when none                                     |
// | GetCableTypeByPartNumber       | finds existing                                   | false for missing                                   |
// | GetCableTypeBySlug             | finds existing                                   | false for missing                                   |
// | AllRackTypes                   | registered racks returned                        | empty when none                                     |
// | GetRackTypeBySlug              | finds existing                                   | false for missing                                   |
// | GetRackTypeByPartNumber        | finds existing                                   | false for missing                                   |
// | RegisterRackType               | adds to both maps                                | empty PN skips PN map                               |
// | AllFruTypes                    | registered FRUs returned                         | empty when none                                     |
// | GetFruTypeBySlug               | finds existing                                   | false for missing                                   |
// | GetFruTypeByPartNumber         | finds existing                                   | false for missing                                   |
// | RegisterFruType                | adds to both maps                                | empty PN skips PN map                               |
// | RegisterCableType              | adds to both maps                                | empty PN skips PN map                               |
// | RegisterDeviceType             | adds to both maps                                | empty PN skips PN map                               |
// | GetByManufacturerModel         | case-insensitive match                           | no match returns false                              |
// | (+ identifications)            | matches alternate ID                             | partial ID no match                                 |
// | AllModules                     | registered modules returned                      | empty when none                                     |
// | GetModuleBySlug                | finds existing                                   | false for missing                                   |
// | GetModuleTypeBySlug            | delegates correctly                              | false for missing                                   |
// | GetModuleTypeByPartNumber      | finds existing                                   | false for missing                                   |
// | GetModuleByManufacturerModel   | case-insensitive match                           | mismatch returns false                              |
// | RegisterModuleType             | adds to both maps                                | empty PN skips PN map                               |
// | AllTypes                       | non-empty slice                                  | returns copy (mutate safety)                        |
// | AllTypesString                 | non-empty slice                                  | contains known "rack" type                          |
// | Inventory.DevicesByType        | returns matching devices                         | nil for nil inventory                               |
// | Inventory.Exists               | true for existing                                | false for missing                                   |
// | Inventory.FindName             | returns device pointer                           | false for missing                                   |

import (
	"testing"

	"github.com/google/uuid"
)

// ---------- helpers ----------

// resetRegistries clears all package-level maps so tests are isolated.
func resetRegistries() {
	for k := range allDeviceTypes {
		delete(allDeviceTypes, k)
	}
	for k := range deviceTypesByPartNum {
		delete(deviceTypesByPartNum, k)
	}
	for k := range allModuleTypes {
		delete(allModuleTypes, k)
	}
	for k := range moduleTypesByPartNum {
		delete(moduleTypesByPartNum, k)
	}
	for k := range allCableTypes {
		delete(allCableTypes, k)
	}
	for k := range cableTypesByPartNum {
		delete(cableTypesByPartNum, k)
	}
	for k := range allRackTypes {
		delete(allRackTypes, k)
	}
	for k := range rackTypesByPartNum {
		delete(rackTypesByPartNum, k)
	}
	for k := range allFruTypes {
		delete(allFruTypes, k)
	}
	for k := range fruTypesByPartNum {
		delete(fruTypesByPartNum, k)
	}
}

// seedDevice registers a test device type and returns it.
func seedDevice(slug, partNumber, manufacturer, model string) CaniDeviceType {
	dt := CaniDeviceType{
		Slug:         slug,
		PartNumber:   partNumber,
		Manufacturer: manufacturer,
		Model:        model,
	}
	RegisterDeviceType(dt)
	return dt
}

// ---------- All ----------

func TestAllReturnsRegisteredDevices(t *testing.T) {
	resetRegistries()
	seedDevice("dev-a", "PN-A", "Acme", "ModelA")

	got := All()
	if len(got) != 1 {
		t.Fatalf("expected 1 device, got %d", len(got))
	}
	if _, ok := got["dev-a"]; !ok {
		t.Error("expected key dev-a in map")
	}
}

func TestAllEmptyWhenNoDevicesRegistered(t *testing.T) {
	resetRegistries()

	got := All()
	if len(got) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(got))
	}
}

// ---------- ByPartNumber ----------

func TestByPartNumberReturnsMapIndexedByPartNumber(t *testing.T) {
	resetRegistries()
	seedDevice("dev-a", "PN-100", "Acme", "ModelA")

	got := ByPartNumber()
	if _, ok := got["PN-100"]; !ok {
		t.Error("expected key PN-100 in part-number map")
	}
}

func TestByPartNumberEmptyWhenNoDevicesRegistered(t *testing.T) {
	resetRegistries()

	got := ByPartNumber()
	if len(got) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(got))
	}
}

// ---------- GetByPartNumber ----------

func TestGetByPartNumberFindsExistingDevice(t *testing.T) {
	resetRegistries()
	seedDevice("dev-a", "PN-200", "Acme", "ModelA")

	dt, ok := GetByPartNumber("PN-200")
	if !ok {
		t.Fatal("expected device to be found")
	}
	if dt.Slug != "dev-a" {
		t.Errorf("expected slug dev-a, got %s", dt.Slug)
	}
}

func TestGetByPartNumberReturnsFalseForMissingPartNumber(t *testing.T) {
	resetRegistries()

	_, ok := GetByPartNumber("nonexistent")
	if ok {
		t.Error("expected ok=false for missing part number")
	}
}

// ---------- GetBySlug ----------

func TestGetBySlugFindsExistingDevice(t *testing.T) {
	resetRegistries()
	seedDevice("switch-1", "PN-300", "Acme", "Switch1")

	dt, ok := GetBySlug("switch-1")
	if !ok {
		t.Fatal("expected device to be found")
	}
	if dt.PartNumber != "PN-300" {
		t.Errorf("expected part number PN-300, got %s", dt.PartNumber)
	}
}

func TestGetBySlugReturnsFalseForMissingSlug(t *testing.T) {
	resetRegistries()

	_, ok := GetBySlug("no-such-slug")
	if ok {
		t.Error("expected ok=false for missing slug")
	}
}

// ---------- AllCables ----------

func TestAllCablesReturnsRegisteredCables(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{Slug: "cat6", PartNumber: "CBL-100"})

	got := AllCables()
	if len(got) != 1 {
		t.Fatalf("expected 1 cable, got %d", len(got))
	}
}

func TestAllCablesEmptyWhenNoCablesRegistered(t *testing.T) {
	resetRegistries()

	got := AllCables()
	if len(got) != 0 {
		t.Fatalf("expected 0 cables, got %d", len(got))
	}
}

// ---------- GetCableTypeByPartNumber ----------

func TestGetCableTypeByPartNumberFindsExistingCable(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{Slug: "cat6", PartNumber: "CBL-200"})

	ct, ok := GetCableTypeByPartNumber("CBL-200")
	if !ok {
		t.Fatal("expected cable to be found")
	}
	if ct.Slug != "cat6" {
		t.Errorf("expected slug cat6, got %s", ct.Slug)
	}
}

func TestGetCableTypeByPartNumberReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetCableTypeByPartNumber("nonexistent")
	if ok {
		t.Error("expected ok=false for missing cable part number")
	}
}

// ---------- GetCableTypeBySlug ----------

func TestGetCableTypeBySlugFindsExistingCable(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{Slug: "fiber-om4", PartNumber: "CBL-300"})

	ct, ok := GetCableTypeBySlug("fiber-om4")
	if !ok {
		t.Fatal("expected cable to be found")
	}
	if ct.PartNumber != "CBL-300" {
		t.Errorf("expected part number CBL-300, got %s", ct.PartNumber)
	}
}

func TestGetCableTypeBySlugReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetCableTypeBySlug("nonexistent-cable")
	if ok {
		t.Error("expected ok=false for missing cable slug")
	}
}

// ---------- AllRackTypes ----------

func TestAllRackTypesReturnsRegisteredRacks(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{Slug: "rack-42u", PartNumber: "RK-100"})

	got := AllRackTypes()
	if len(got) != 1 {
		t.Fatalf("expected 1 rack type, got %d", len(got))
	}
}

func TestAllRackTypesEmptyWhenNoRacksRegistered(t *testing.T) {
	resetRegistries()

	got := AllRackTypes()
	if len(got) != 0 {
		t.Fatalf("expected 0 rack types, got %d", len(got))
	}
}

// ---------- GetRackTypeBySlug ----------

func TestGetRackTypeBySlugFindsExistingRack(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{Slug: "rack-42u", PartNumber: "RK-200"})

	rt, ok := GetRackTypeBySlug("rack-42u")
	if !ok {
		t.Fatal("expected rack type to be found")
	}
	if rt.PartNumber != "RK-200" {
		t.Errorf("expected part number RK-200, got %s", rt.PartNumber)
	}
}

func TestGetRackTypeBySlugReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetRackTypeBySlug("nonexistent-rack")
	if ok {
		t.Error("expected ok=false for missing rack slug")
	}
}

// ---------- GetRackTypeByPartNumber ----------

func TestGetRackTypeByPartNumberFindsExistingRack(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{Slug: "rack-48u", PartNumber: "RK-300"})

	rt, ok := GetRackTypeByPartNumber("RK-300")
	if !ok {
		t.Fatal("expected rack type to be found")
	}
	if rt.Slug != "rack-48u" {
		t.Errorf("expected slug rack-48u, got %s", rt.Slug)
	}
}

func TestGetRackTypeByPartNumberReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetRackTypeByPartNumber("nonexistent-pn")
	if ok {
		t.Error("expected ok=false for missing rack part number")
	}
}

// ---------- RegisterRackType ----------

func TestRegisterRackTypeAddsToBothMaps(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{Slug: "rack-new", PartNumber: "RK-NEW"})

	if _, ok := allRackTypes["rack-new"]; !ok {
		t.Error("expected rack-new in allRackTypes")
	}
	if _, ok := rackTypesByPartNum["RK-NEW"]; !ok {
		t.Error("expected RK-NEW in rackTypesByPartNum")
	}
}

func TestRegisterRackTypeEmptyPartNumberSkipsPartNumMap(t *testing.T) {
	resetRegistries()
	RegisterRackType(CaniRackType{Slug: "rack-no-pn"})

	if _, ok := allRackTypes["rack-no-pn"]; !ok {
		t.Error("expected rack-no-pn in allRackTypes")
	}
	if len(rackTypesByPartNum) != 0 {
		t.Error("expected rackTypesByPartNum to be empty when part number is blank")
	}
}

// ---------- AllFruTypes ----------

func TestAllFruTypesReturnsRegisteredFrus(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{Slug: "fru-psu", PartNumber: "FRU-100"})

	got := AllFruTypes()
	if len(got) != 1 {
		t.Fatalf("expected 1 FRU type, got %d", len(got))
	}
}

func TestAllFruTypesEmptyWhenNoFrusRegistered(t *testing.T) {
	resetRegistries()

	got := AllFruTypes()
	if len(got) != 0 {
		t.Fatalf("expected 0 FRU types, got %d", len(got))
	}
}

// ---------- GetFruTypeBySlug ----------

func TestGetFruTypeBySlugFindsExistingFru(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{Slug: "fru-fan", PartNumber: "FRU-200"})

	ft, ok := GetFruTypeBySlug("fru-fan")
	if !ok {
		t.Fatal("expected FRU to be found")
	}
	if ft.PartNumber != "FRU-200" {
		t.Errorf("expected part number FRU-200, got %s", ft.PartNumber)
	}
}

func TestGetFruTypeBySlugReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetFruTypeBySlug("nonexistent-fru")
	if ok {
		t.Error("expected ok=false for missing FRU slug")
	}
}

// ---------- GetFruTypeByPartNumber ----------

func TestGetFruTypeByPartNumberFindsExistingFru(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{Slug: "fru-dimm", PartNumber: "FRU-300"})

	ft, ok := GetFruTypeByPartNumber("FRU-300")
	if !ok {
		t.Fatal("expected FRU to be found")
	}
	if ft.Slug != "fru-dimm" {
		t.Errorf("expected slug fru-dimm, got %s", ft.Slug)
	}
}

func TestGetFruTypeByPartNumberReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetFruTypeByPartNumber("nonexistent-fru-pn")
	if ok {
		t.Error("expected ok=false for missing FRU part number")
	}
}

// ---------- RegisterFruType ----------

func TestRegisterFruTypeAddsToBothMaps(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{Slug: "fru-reg", PartNumber: "FRU-REG"})

	if _, ok := allFruTypes["fru-reg"]; !ok {
		t.Error("expected fru-reg in allFruTypes")
	}
	if _, ok := fruTypesByPartNum["FRU-REG"]; !ok {
		t.Error("expected FRU-REG in fruTypesByPartNum")
	}
}

func TestRegisterFruTypeEmptyPartNumberSkipsPartNumMap(t *testing.T) {
	resetRegistries()
	RegisterFruType(CaniFruType{Slug: "fru-no-pn"})

	if _, ok := allFruTypes["fru-no-pn"]; !ok {
		t.Error("expected fru-no-pn in allFruTypes")
	}
	if len(fruTypesByPartNum) != 0 {
		t.Error("expected fruTypesByPartNum to be empty when part number is blank")
	}
}

// ---------- RegisterCableType ----------

func TestRegisterCableTypeAddsToBothMaps(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{Slug: "dac-passive", PartNumber: "CBL-REG"})

	if _, ok := allCableTypes["dac-passive"]; !ok {
		t.Error("expected dac-passive in allCableTypes")
	}
	if _, ok := cableTypesByPartNum["CBL-REG"]; !ok {
		t.Error("expected CBL-REG in cableTypesByPartNum")
	}
}

func TestRegisterCableTypeEmptyPartNumberSkipsPartNumMap(t *testing.T) {
	resetRegistries()
	RegisterCableType(CaniCableType{Slug: "cable-no-pn"})

	if _, ok := allCableTypes["cable-no-pn"]; !ok {
		t.Error("expected cable-no-pn in allCableTypes")
	}
	if len(cableTypesByPartNum) != 0 {
		t.Error("expected cableTypesByPartNum to be empty when part number is blank")
	}
}

// ---------- RegisterDeviceType ----------

func TestRegisterDeviceTypeAddsToBothMaps(t *testing.T) {
	resetRegistries()
	RegisterDeviceType(CaniDeviceType{Slug: "node-1", PartNumber: "DEV-REG"})

	if _, ok := allDeviceTypes["node-1"]; !ok {
		t.Error("expected node-1 in allDeviceTypes")
	}
	if _, ok := deviceTypesByPartNum["DEV-REG"]; !ok {
		t.Error("expected DEV-REG in deviceTypesByPartNum")
	}
}

func TestRegisterDeviceTypeEmptyPartNumberSkipsPartNumMap(t *testing.T) {
	resetRegistries()
	RegisterDeviceType(CaniDeviceType{Slug: "dev-no-pn"})

	if _, ok := allDeviceTypes["dev-no-pn"]; !ok {
		t.Error("expected dev-no-pn in allDeviceTypes")
	}
	if len(deviceTypesByPartNum) != 0 {
		t.Error("expected deviceTypesByPartNum to be empty when part number is blank")
	}
}

// ---------- GetByManufacturerModel ----------

func TestGetByManufacturerModelMatchesCaseInsensitive(t *testing.T) {
	resetRegistries()
	seedDevice("dev-mm", "PN-MM", "Acme Corp", "SuperSwitch")

	dt, ok := GetByManufacturerModel("acme corp", "superswitch")
	if !ok {
		t.Fatal("expected device to be found with case-insensitive match")
	}
	if dt.Slug != "dev-mm" {
		t.Errorf("expected slug dev-mm, got %s", dt.Slug)
	}
}

func TestGetByManufacturerModelReturnsFalseForMismatch(t *testing.T) {
	resetRegistries()
	seedDevice("dev-mm", "PN-MM", "Acme Corp", "SuperSwitch")

	_, ok := GetByManufacturerModel("Unknown", "NoModel")
	if ok {
		t.Error("expected ok=false for non-matching manufacturer/model")
	}
}

// ---------- GetByManufacturerModel with Identifications ----------

func TestGetByManufacturerModelMatchesAlternateIdentification(t *testing.T) {
	resetRegistries()
	dt := CaniDeviceType{
		Slug:         "dev-alt",
		PartNumber:   "PN-ALT",
		Manufacturer: "Primary",
		Model:        "PrimaryModel",
		Identifications: []Identification{
			{Manufacturer: "Alias", Model: "AliasModel"},
		},
	}
	RegisterDeviceType(dt)

	found, ok := GetByManufacturerModel("alias", "aliasmodel")
	if !ok {
		t.Fatal("expected device to be found via alternate identification")
	}
	if found.Slug != "dev-alt" {
		t.Errorf("expected slug dev-alt, got %s", found.Slug)
	}
}

func TestGetByManufacturerModelNoMatchOnPartialIdentification(t *testing.T) {
	resetRegistries()
	dt := CaniDeviceType{
		Slug:         "dev-alt",
		Manufacturer: "Primary",
		Model:        "PrimaryModel",
		Identifications: []Identification{
			{Manufacturer: "Alias", Model: "AliasModel"},
		},
	}
	RegisterDeviceType(dt)

	_, ok := GetByManufacturerModel("Alias", "WrongModel")
	if ok {
		t.Error("expected ok=false when model does not match any identification")
	}
}

// ---------- AllModules ----------

func TestAllModulesReturnsRegisteredModules(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{Slug: "gpu-a100", PartNumber: "MOD-100"})

	got := AllModules()
	if len(got) != 1 {
		t.Fatalf("expected 1 module, got %d", len(got))
	}
}

func TestAllModulesEmptyWhenNoModulesRegistered(t *testing.T) {
	resetRegistries()

	got := AllModules()
	if len(got) != 0 {
		t.Fatalf("expected 0 modules, got %d", len(got))
	}
}

// ---------- GetModuleBySlug ----------

func TestGetModuleBySlugFindsExistingModule(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{Slug: "nic-25g", PartNumber: "MOD-200"})

	mt, ok := GetModuleBySlug("nic-25g")
	if !ok {
		t.Fatal("expected module to be found")
	}
	if mt.PartNumber != "MOD-200" {
		t.Errorf("expected part number MOD-200, got %s", mt.PartNumber)
	}
}

func TestGetModuleBySlugReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetModuleBySlug("nonexistent-module")
	if ok {
		t.Error("expected ok=false for missing module slug")
	}
}

// ---------- GetModuleTypeBySlug (alias) ----------

func TestGetModuleTypeBySlugDelegatesToGetModuleBySlug(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{Slug: "psu-1600w", PartNumber: "MOD-300"})

	mt, ok := GetModuleTypeBySlug("psu-1600w")
	if !ok {
		t.Fatal("expected module to be found via alias")
	}
	if mt.Slug != "psu-1600w" {
		t.Errorf("expected slug psu-1600w, got %s", mt.Slug)
	}
}

func TestGetModuleTypeBySlugReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetModuleTypeBySlug("no-such-module")
	if ok {
		t.Error("expected ok=false for missing module slug via alias")
	}
}

// ---------- GetModuleTypeByPartNumber ----------

func TestGetModuleTypeByPartNumberFindsExistingModule(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{Slug: "mem-ddr5", PartNumber: "MOD-400"})

	mt, ok := GetModuleTypeByPartNumber("MOD-400")
	if !ok {
		t.Fatal("expected module to be found")
	}
	if mt.Slug != "mem-ddr5" {
		t.Errorf("expected slug mem-ddr5, got %s", mt.Slug)
	}
}

func TestGetModuleTypeByPartNumberReturnsFalseForMissing(t *testing.T) {
	resetRegistries()

	_, ok := GetModuleTypeByPartNumber("nonexistent-mod-pn")
	if ok {
		t.Error("expected ok=false for missing module part number")
	}
}

// ---------- GetModuleByManufacturerModel ----------

func TestGetModuleByManufacturerModelMatchesCaseInsensitive(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{
		Slug:         "gpu-h100",
		Manufacturer: "NVIDIA",
		Model:        "H100",
	})

	mt, ok := GetModuleByManufacturerModel("nvidia", "h100")
	if !ok {
		t.Fatal("expected module to be found with case-insensitive match")
	}
	if mt.Slug != "gpu-h100" {
		t.Errorf("expected slug gpu-h100, got %s", mt.Slug)
	}
}

func TestGetModuleByManufacturerModelReturnsFalseForMismatch(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{
		Slug:         "gpu-h100",
		Manufacturer: "NVIDIA",
		Model:        "H100",
	})

	_, ok := GetModuleByManufacturerModel("AMD", "MI300")
	if ok {
		t.Error("expected ok=false for non-matching manufacturer/model")
	}
}

// ---------- RegisterModuleType ----------

func TestRegisterModuleTypeAddsToBothMaps(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{Slug: "mod-reg", PartNumber: "MOD-REG"})

	if _, ok := allModuleTypes["mod-reg"]; !ok {
		t.Error("expected mod-reg in allModuleTypes")
	}
	if _, ok := moduleTypesByPartNum["MOD-REG"]; !ok {
		t.Error("expected MOD-REG in moduleTypesByPartNum")
	}
}

func TestRegisterModuleTypeEmptyPartNumberSkipsPartNumMap(t *testing.T) {
	resetRegistries()
	RegisterModuleType(CaniModuleType{Slug: "mod-no-pn"})

	if _, ok := allModuleTypes["mod-no-pn"]; !ok {
		t.Error("expected mod-no-pn in allModuleTypes")
	}
	if len(moduleTypesByPartNum) != 0 {
		t.Error("expected moduleTypesByPartNum to be empty when part number is blank")
	}
}

// ---------- AllTypes ----------

func TestAllTypesReturnsNonEmptySlice(t *testing.T) {
	got := AllTypes()
	if len(got) == 0 {
		t.Fatal("expected non-empty slice from AllTypes")
	}
}

func TestAllTypesReturnsCopyNotOriginal(t *testing.T) {
	original := AllTypes()
	copy := AllTypes()

	// Mutating the returned slice should not affect subsequent calls.
	copy[0] = "mutated"
	fresh := AllTypes()
	if fresh[0] == "mutated" {
		t.Error("AllTypes should return a copy, not the original slice")
	}
	_ = original
}

// ---------- AllTypesString ----------

func TestAllTypesStringReturnsNonEmptySlice(t *testing.T) {
	got := AllTypesString()
	if len(got) == 0 {
		t.Fatal("expected non-empty string slice from AllTypesString")
	}
}

func TestAllTypesStringContainsKnownType(t *testing.T) {
	got := AllTypesString()
	found := false
	for _, s := range got {
		if s == "rack" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected AllTypesString to contain 'rack'")
	}
}

// ---------- Inventory.DevicesByType ----------

func TestInventoryDevicesByTypeReturnsMatchingDevices(t *testing.T) {
	id := uuid.New()
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{
			id: {Name: "node-01", Type: TypeNode},
		},
	}

	got := inv.DevicesByType("node")
	if len(got) != 1 {
		t.Fatalf("expected 1 device, got %d", len(got))
	}
	if got[0].Name != "node-01" {
		t.Errorf("expected name node-01, got %s", got[0].Name)
	}
}

func TestInventoryDevicesByTypeReturnsNilForNilInventory(t *testing.T) {
	var inv *Inventory

	got := inv.DevicesByType("node")
	if got != nil {
		t.Error("expected nil for nil inventory")
	}
}

// ---------- Inventory.Exists ----------

func TestInventoryExistsReturnsTrueForExistingDevice(t *testing.T) {
	id := uuid.New()
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{
			id: {Name: "switch-01"},
		},
	}

	if !inv.Exists("switch-01") {
		t.Error("expected Exists to return true")
	}
}

func TestInventoryExistsReturnsFalseForMissingDevice(t *testing.T) {
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{},
	}

	if inv.Exists("nonexistent") {
		t.Error("expected Exists to return false for missing device")
	}
}

// ---------- Inventory.FindName ----------

func TestInventoryFindNameReturnsDeviceWhenFound(t *testing.T) {
	id := uuid.New()
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{
			id: {Name: "blade-01", Slug: "blade-type-a"},
		},
	}

	device, ok := inv.FindName("blade-01")
	if !ok {
		t.Fatal("expected device to be found")
	}
	if device.Slug != "blade-type-a" {
		t.Errorf("expected slug blade-type-a, got %s", device.Slug)
	}
}

func TestInventoryFindNameReturnsFalseForMissing(t *testing.T) {
	inv := &Inventory{
		Devices: map[uuid.UUID]*CaniDeviceType{},
	}

	_, ok := inv.FindName("nonexistent")
	if ok {
		t.Error("expected ok=false for missing device name")
	}
}

// ---------- AllLocationTypes / GetLocationTypeBySlug ----------

// TestAllLocationTypesAndGetBySlug verifies the location-type registry returns
// the full map and resolves a known slug, while reporting false for an unknown
// slug.
//
// Why it matters: location types drive how cani builds the site hierarchy, so
// the registry accessors must expose registered definitions and clearly signal
// misses so callers can fail fast.
// Inputs: a registry seeded with one "site" definition; lookups for "site" and
// "missing". Outputs: a map containing "site"; (definition,true) for "site";
// (zero,false) for "missing".
// Data choice: a single registered slug plus one absent slug exercises both the
// hit and miss branches without depending on the embedded library's contents,
// and the registry is saved/restored to avoid leaking into other tests.
func TestAllLocationTypesAndGetBySlug(t *testing.T) {
	orig := allLocationTypes
	defer func() { allLocationTypes = orig }()
	allLocationTypes = map[string]LocationTypeDefinition{
		"site": {Name: "Site", Slug: "site", Nestable: true},
	}

	all := AllLocationTypes()
	if _, ok := all["site"]; !ok {
		t.Errorf("AllLocationTypes() missing %q, got %v", "site", all)
	}

	lt, ok := GetLocationTypeBySlug("site")
	if !ok {
		t.Fatal("GetLocationTypeBySlug(\"site\") ok = false, want true")
	}
	if lt.Name != "Site" {
		t.Errorf("GetLocationTypeBySlug name = %q, want %q", lt.Name, "Site")
	}

	if _, ok := GetLocationTypeBySlug("missing"); ok {
		t.Error("GetLocationTypeBySlug(\"missing\") ok = true, want false")
	}
}
