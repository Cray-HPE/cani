package devicetypes

// Test coverage for CaniModuleType methods in cani_module_types.go.
//
// | Function              | Happy-path test                                  | Failure test                                      |
// |-----------------------|--------------------------------------------------|----------------------------------------------------|
// | Validate              | TestModuleValidatePassesForValid                 | TestModuleValidateReturnsErrorForNil               |
// | GetID                 | TestModuleGetIDReturnsUUID                       | TestModuleGetIDReturnsNilForNilReceiver            |
// | GetSlug               | TestModuleGetSlugReturnsSlug                     | TestModuleGetSlugReturnsEmptyForNil                |
// | GetStatus             | TestModuleGetStatusReturnsStatus                 | TestModuleGetStatusReturnsEmptyForNil              |
// | GetVendor             | TestModuleGetVendorReturnsManufacturer            | TestModuleGetVendorReturnsEmptyForNil              |
// | GetType               | TestModuleGetTypeReturnsExplicitType             | TestModuleGetTypeReturnsDefaultForEmpty            |
// | InstantiateInterfaces | TestModuleInstantiateInterfacesCreatesInstances   | TestModuleInstantiateInterfacesReturnsNilForNil    |
// | GetInterface          | TestModuleGetInterfaceFindsMatch                 | TestModuleGetInterfaceReturnsNilForMissing         |
// | GetInterfaceByName    | TestModuleGetInterfaceByNameFindsMatch            | TestModuleGetInterfaceByNameReturnsNilForNil       |

import (
	"testing"

	"github.com/google/uuid"
)

// --- Validate ---

func TestModuleValidatePassesForValid(t *testing.T) {
	m := &CaniModuleType{Name: "gpu-a100"}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestModuleValidateReturnsErrorForNil(t *testing.T) {
	var m *CaniModuleType
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for nil receiver")
	}
}

// --- GetID ---

func TestModuleGetIDReturnsUUID(t *testing.T) {
	id := uuid.New()
	m := &CaniModuleType{ID: id}
	if got := m.GetID(); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestModuleGetIDReturnsNilForNilReceiver(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetID(); got != uuid.Nil {
		t.Fatalf("expected uuid.Nil, got %s", got)
	}
}

// --- GetSlug ---

func TestModuleGetSlugReturnsSlug(t *testing.T) {
	m := &CaniModuleType{Slug: "nvidia-a100"}
	if got := m.GetSlug(); got != "nvidia-a100" {
		t.Fatalf("expected nvidia-a100, got %s", got)
	}
}

func TestModuleGetSlugReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetSlug(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetStatus ---

func TestModuleGetStatusReturnsStatus(t *testing.T) {
	m := &CaniModuleType{Status: "active"}
	if got := m.GetStatus(); got != "active" {
		t.Fatalf("expected active, got %s", got)
	}
}

func TestModuleGetStatusReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetStatus(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetVendor ---

func TestModuleGetVendorReturnsManufacturer(t *testing.T) {
	m := &CaniModuleType{Manufacturer: "NVIDIA"}
	if got := m.GetVendor(); got != "NVIDIA" {
		t.Fatalf("expected NVIDIA, got %s", got)
	}
}

func TestModuleGetVendorReturnsEmptyForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetVendor(); got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

// --- GetType ---

func TestModuleGetTypeReturnsExplicitType(t *testing.T) {
	m := &CaniModuleType{HardwareType: string(TypeChassis)}
	if got := m.GetType(); got != TypeChassis {
		t.Fatalf("expected %s, got %s", TypeChassis, got)
	}
}

func TestModuleGetTypeReturnsDefaultForEmpty(t *testing.T) {
	m := &CaniModuleType{}
	if got := m.GetType(); got != TypeModule {
		t.Fatalf("expected %s, got %s", TypeModule, got)
	}
}

// --- InstantiateInterfaces ---

func TestModuleInstantiateInterfacesCreatesInstances(t *testing.T) {
	mgmt := true
	m := &CaniModuleType{
		ID: uuid.New(),
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT, MgmtOnly: &mgmt},
			{Name: "eth1", Type: InterfacesElemTypeA25GbaseXSfp28},
		},
	}
	instances := m.InstantiateInterfaces()
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].Name != "eth0" {
		t.Fatalf("expected eth0, got %s", instances[0].Name)
	}
	if !instances[0].MgmtOnly {
		t.Fatal("expected MgmtOnly true for eth0")
	}
	if instances[1].MgmtOnly {
		t.Fatal("expected MgmtOnly false for eth1")
	}
}

func TestModuleInstantiateInterfacesReturnsNilForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.InstantiateInterfaces(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// --- GetInterface ---

func TestModuleGetInterfaceFindsMatch(t *testing.T) {
	m := &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT},
			{Name: "eth1", Type: InterfacesElemTypeA25GbaseXSfp28},
		},
	}
	iface := m.GetInterface("eth1")
	if iface == nil {
		t.Fatal("expected non-nil interface")
	}
	if iface.Name != "eth1" {
		t.Fatalf("expected eth1, got %s", iface.Name)
	}
}

func TestModuleGetInterfaceReturnsNilForMissing(t *testing.T) {
	m := &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{Name: "eth0", Type: InterfacesElemTypeA10GbaseT},
		},
	}
	if got := m.GetInterface("eth99"); got != nil {
		t.Fatalf("expected nil for missing interface, got %v", got)
	}
}

// --- GetInterfaceByName ---

func TestModuleGetInterfaceByNameFindsMatch(t *testing.T) {
	m := &CaniModuleType{
		Interfaces: []InterfaceSpec{
			{Name: "mgmt0", Type: InterfacesElemTypeA1000BaseT},
		},
	}
	iface := m.GetInterfaceByName("mgmt0")
	if iface == nil {
		t.Fatal("expected non-nil interface")
	}
	if iface.Type != InterfacesElemTypeA1000BaseT {
		t.Fatalf("expected %s, got %s", InterfacesElemTypeA1000BaseT, iface.Type)
	}
}

func TestModuleGetInterfaceByNameReturnsNilForNil(t *testing.T) {
	var m *CaniModuleType
	if got := m.GetInterfaceByName("eth0"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
