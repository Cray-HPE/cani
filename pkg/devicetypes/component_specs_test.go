package devicetypes

// Test coverage for component_specs.go
//
// | Function           | Happy-path test                            | Failure test                                  |
// |--------------------|--------------------------------------------|-----------------------------------------------|
// | InterfaceSpec      | TestInterfaceSpecJSONRoundTrip              | TestInterfaceSpecUnmarshalInvalidID           |
// | CaniInterface      | TestCaniInterfaceJSONRoundTrip              | TestCaniInterfaceUnmarshalInvalidID           |
// | ConsolePortSpec    | TestConsolePortSpecJSONRoundTrip            | TestConsolePortSpecUnmarshalTypeMismatch      |
// | PowerPortSpec      | TestPowerPortSpecJSONRoundTrip              | TestPowerPortSpecUnmarshalTypeMismatch        |
// | ModuleBaySpec      | TestModuleBaySpecJSONRoundTrip              | TestModuleBaySpecUnmarshalTypeMismatch        |
// | DeviceBaySpec      | TestDeviceBaySpecJSONRoundTrip              | TestDeviceBaySpecUnmarshalTypeMismatch        |
// | Identification     | TestIdentificationJSONRoundTrip            | TestIdentificationUnmarshalTypeMismatch       |
// | InterfacesElemType | TestInterfacesElemTypeValidConstants        | TestInterfacesElemTypeInvalidValue             |

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func TestInterfaceSpecJSONRoundTrip(t *testing.T) {
	mgmt := true
	cable := uuid.New()
	orig := InterfaceSpec{
		ID:             uuid.New(),
		Name:           "eth0",
		Type:           InterfacesElemTypeA10GbaseT,
		Label:          "Ethernet 0",
		MgmtOnly:       &mgmt,
		ConnectedCable: &cable,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got InterfaceSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID = %v, want %v", got.ID, orig.ID)
	}
	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Type != orig.Type {
		t.Errorf("Type = %q, want %q", got.Type, orig.Type)
	}
	if got.Label != orig.Label {
		t.Errorf("Label = %q, want %q", got.Label, orig.Label)
	}
	if got.MgmtOnly == nil || *got.MgmtOnly != mgmt {
		t.Errorf("MgmtOnly = %v, want %v", got.MgmtOnly, mgmt)
	}
	if got.ConnectedCable == nil || *got.ConnectedCable != cable {
		t.Errorf("ConnectedCable = %v, want %v", got.ConnectedCable, cable)
	}
}

func TestInterfaceSpecUnmarshalInvalidID(t *testing.T) {
	data := []byte(`{"id":"not-a-valid-uuid","name":"eth0","type":"10gbase-t"}`)
	var spec InterfaceSpec
	if err := json.Unmarshal(data, &spec); err == nil {
		t.Fatal("expected error for invalid UUID in id field, got nil")
	}
}

func TestCaniInterfaceJSONRoundTrip(t *testing.T) {
	cable := uuid.New()
	orig := CaniInterface{
		ID:             uuid.New(),
		Name:           "eth1",
		InterfaceType:  InterfacesElemTypeA25GbaseXSfp28,
		DeviceID:       uuid.New(),
		ObjectMeta:     ObjectMeta{Status: "Active"},
		MgmtOnly:       false,
		Label:          "Management",
		ConnectedCable: &cable,
		ContentType:    "dcim.interface",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got CaniInterface
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID = %v, want %v", got.ID, orig.ID)
	}
	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.InterfaceType != orig.InterfaceType {
		t.Errorf("InterfaceType = %q, want %q", got.InterfaceType, orig.InterfaceType)
	}
	if got.DeviceID != orig.DeviceID {
		t.Errorf("DeviceID = %v, want %v", got.DeviceID, orig.DeviceID)
	}
	if got.Status != orig.Status {
		t.Errorf("Status = %q, want %q", got.Status, orig.Status)
	}
	if got.ContentType != orig.ContentType {
		t.Errorf("ContentType = %q, want %q", got.ContentType, orig.ContentType)
	}
	if got.ConnectedCable == nil || *got.ConnectedCable != cable {
		t.Errorf("ConnectedCable = %v, want %v", got.ConnectedCable, cable)
	}
}

func TestCaniInterfaceUnmarshalInvalidID(t *testing.T) {
	data := []byte(`{"id":"bad-uuid","name":"eth1","interfaceType":"10gbase-t","deviceId":"also-bad","status":"Active"}`)
	var inst CaniInterface
	if err := json.Unmarshal(data, &inst); err == nil {
		t.Fatal("expected error for invalid UUID in id field, got nil")
	}
}

func TestConsolePortSpecJSONRoundTrip(t *testing.T) {
	orig := ConsolePortSpec{
		Name: "Console Port 1",
		Type: "rj-45",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got ConsolePortSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Type != orig.Type {
		t.Errorf("Type = %q, want %q", got.Type, orig.Type)
	}
}

func TestConsolePortSpecUnmarshalTypeMismatch(t *testing.T) {
	data := []byte(`{"name":123,"type":"rj-45"}`)
	var spec ConsolePortSpec
	if err := json.Unmarshal(data, &spec); err == nil {
		t.Fatal("expected error for numeric name field, got nil")
	}
}

func TestPowerPortSpecJSONRoundTrip(t *testing.T) {
	orig := PowerPortSpec{
		Name:          "PSU 1",
		Type:          "iec-60320-c14",
		MaximumDraw:   750,
		AllocatedDraw: 500,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got PowerPortSpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Type != orig.Type {
		t.Errorf("Type = %q, want %q", got.Type, orig.Type)
	}
	if got.MaximumDraw != orig.MaximumDraw {
		t.Errorf("MaximumDraw = %d, want %d", got.MaximumDraw, orig.MaximumDraw)
	}
	if got.AllocatedDraw != orig.AllocatedDraw {
		t.Errorf("AllocatedDraw = %d, want %d", got.AllocatedDraw, orig.AllocatedDraw)
	}
}

func TestPowerPortSpecUnmarshalTypeMismatch(t *testing.T) {
	data := []byte(`{"name":"PSU","type":"iec-60320-c14","maximum_draw":"not-a-number"}`)
	var spec PowerPortSpec
	if err := json.Unmarshal(data, &spec); err == nil {
		t.Fatal("expected error for string in maximum_draw field, got nil")
	}
}

func TestModuleBaySpecJSONRoundTrip(t *testing.T) {
	orig := ModuleBaySpec{
		Name:     "Module Bay 1",
		Position: "1",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got ModuleBaySpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Position != orig.Position {
		t.Errorf("Position = %q, want %q", got.Position, orig.Position)
	}
}

func TestModuleBaySpecUnmarshalTypeMismatch(t *testing.T) {
	data := []byte(`{"name":true,"position":"1"}`)
	var spec ModuleBaySpec
	if err := json.Unmarshal(data, &spec); err == nil {
		t.Fatal("expected error for boolean name field, got nil")
	}
}

func TestDeviceBaySpecJSONRoundTrip(t *testing.T) {
	orig := DeviceBaySpec{
		Name:     "U1",
		Position: "1",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got DeviceBaySpec
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Name != orig.Name {
		t.Errorf("Name = %q, want %q", got.Name, orig.Name)
	}
	if got.Position != orig.Position {
		t.Errorf("Position = %q, want %q", got.Position, orig.Position)
	}
}

func TestDeviceBaySpecUnmarshalTypeMismatch(t *testing.T) {
	data := []byte(`{"name":["array"],"position":"1"}`)
	var spec DeviceBaySpec
	if err := json.Unmarshal(data, &spec); err == nil {
		t.Fatal("expected error for array in name field, got nil")
	}
}

func TestIdentificationJSONRoundTrip(t *testing.T) {
	orig := Identification{
		Manufacturer: "Cray",
		Model:        "EX4252",
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got Identification
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Manufacturer != orig.Manufacturer {
		t.Errorf("Manufacturer = %q, want %q", got.Manufacturer, orig.Manufacturer)
	}
	if got.Model != orig.Model {
		t.Errorf("Model = %q, want %q", got.Model, orig.Model)
	}
}

func TestIdentificationUnmarshalTypeMismatch(t *testing.T) {
	data := []byte(`{"manufacturer":999,"model":"EX4252"}`)
	var id Identification
	if err := json.Unmarshal(data, &id); err == nil {
		t.Fatal("expected error for numeric manufacturer field, got nil")
	}
}

func TestInterfacesElemTypeValidConstants(t *testing.T) {
	expected := map[InterfacesElemType]string{
		InterfacesElemTypeA1000BaseT:       "1000base-t",
		InterfacesElemTypeA1000BaseKx:      "1000base-kx",
		InterfacesElemTypeA10GbaseT:        "10gbase-t",
		InterfacesElemTypeA10GbaseXSfpp:    "10gbase-x-sfpp",
		InterfacesElemTypeA25GbaseXSfp28:   "25gbase-x-sfp28",
		InterfacesElemTypeA40GbaseXQsfpp:   "40gbase-x-qsfpp",
		InterfacesElemTypeA100GbaseXQsfp28: "100gbase-x-qsfp28",
		InterfacesElemTypeA200GbaseXQsfp56: "200gbase-x-qsfp56",
		InterfacesElemTypeA400GbaseXQsfpdd: "400gbase-x-qsfpdd",
		InterfacesElemTypeA400GbaseXOsfp:   "400gbase-x-osfp",
		InterfacesElemTypeVirtual:          "virtual",
		InterfacesElemTypeLag:              "lag",
	}

	for constant, want := range expected {
		if string(constant) != want {
			t.Errorf("constant %q = %q, want %q", constant, string(constant), want)
		}
	}
}

func TestInterfacesElemTypeInvalidValue(t *testing.T) {
	invalid := InterfacesElemType("nonexistent-interface-type")
	valid := []InterfacesElemType{
		InterfacesElemTypeA1000BaseT,
		InterfacesElemTypeA1000BaseKx,
		InterfacesElemTypeA10GbaseT,
		InterfacesElemTypeA10GbaseXSfpp,
		InterfacesElemTypeA25GbaseXSfp28,
		InterfacesElemTypeA40GbaseXQsfpp,
		InterfacesElemTypeA100GbaseXQsfp28,
		InterfacesElemTypeA200GbaseXQsfp56,
		InterfacesElemTypeA400GbaseXQsfpdd,
		InterfacesElemTypeA400GbaseXOsfp,
		InterfacesElemTypeVirtual,
		InterfacesElemTypeLag,
	}

	for _, v := range valid {
		if invalid == v {
			t.Fatalf("invalid type %q should not match valid constant %q", invalid, v)
		}
	}
}

func TestDeviceBaySpecYAMLExtra(t *testing.T) {
	input := "name: Chassis 0\nordinal: 3\n"
	var bay DeviceBaySpec
	if err := yaml.Unmarshal([]byte(input), &bay); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if bay.Name != "Chassis 0" {
		t.Errorf("Name = %q, want %q", bay.Name, "Chassis 0")
	}
	v, ok := bay.Extra["ordinal"]
	if !ok {
		t.Fatal("Extra[\"ordinal\"] not present")
	}
	n, ok := v.(int)
	if !ok {
		t.Fatalf("Extra[\"ordinal\"] type = %T, want int", v)
	}
	if n != 3 {
		t.Errorf("Extra[\"ordinal\"] = %d, want 3", n)
	}
}
