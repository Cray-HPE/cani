package transform

import "testing"

func TestValidTypeForClass_River(t *testing.T) {
	tests := []struct {
		typeString string
		want       bool
	}{
		{XnameTypeCabinet, true},
		{XnameTypeMgmtSwitch, true},
		{XnameTypeMgmtHLSwitch, true},
		{XnameTypeMgmtSwitchConnector, true},
		{XnameTypeComputeModule, true},
		{XnameTypeNodeBMC, true},
		{XnameTypeNode, true},
		{XnameTypeCabinetPDUController, true},
		// Mountain-only types should not be valid in River
		{XnameTypeChassisBMC, false},
		{XnameTypeRouterModule, false},
		{XnameTypeRouterBMC, false},
		{XnameTypeHSNBoard, false},
		{XnameTypeMgmtCDUSwitch, false},
	}
	for _, tt := range tests {
		got := validTypeForClass(tt.typeString, ClassRiver)
		if got != tt.want {
			t.Errorf("validTypeForClass(%q, River) = %v, want %v",
				tt.typeString, got, tt.want)
		}
	}
}

func TestValidTypeForClass_Mountain(t *testing.T) {
	tests := []struct {
		typeString string
		want       bool
	}{
		{XnameTypeCabinet, true},
		{XnameTypeChassis, true},
		{XnameTypeChassisBMC, true},
		{XnameTypeRouterModule, true},
		{XnameTypeRouterBMC, true},
		{XnameTypeHSNBoard, true},
		{XnameTypeMgmtCDUSwitch, true},
		// River-only types should not be valid in Mountain
		{XnameTypeMgmtSwitch, false},
		{XnameTypeMgmtHLSwitch, false},
		{XnameTypeMgmtSwitchConnector, false},
		{XnameTypeCabinetPDUController, false},
	}
	for _, tt := range tests {
		got := validTypeForClass(tt.typeString, ClassMountain)
		if got != tt.want {
			t.Errorf("validTypeForClass(%q, Mountain) = %v, want %v",
				tt.typeString, got, tt.want)
		}
	}
}

func TestValidTypeForClass_Hill(t *testing.T) {
	// Hill accepts the union of River and Mountain types.
	types := []string{
		XnameTypeMgmtSwitch,
		XnameTypeRouterModule,
		XnameTypeMgmtCDUSwitch,
		XnameTypeCabinetPDUController,
	}
	for _, ts := range types {
		if !validTypeForClass(ts, ClassHill) {
			t.Errorf("validTypeForClass(%q, Hill) = false, want true", ts)
		}
	}
}

func TestValidTypeForClass_UnknownClass(t *testing.T) {
	// Unknown class is permissive — everything valid.
	if !validTypeForClass("MadeUpType", "FutureClass") {
		t.Error("unknown class should be permissive")
	}
	if !validTypeForClass(XnameTypeMgmtSwitch, "") {
		t.Error("empty class should be permissive")
	}
}

func TestClassForCabinetNumber(t *testing.T) {
	tests := []struct {
		cabinet int
		want    string
	}{
		{1000, ClassMountain},
		{2999, ClassMountain},
		{1500, ClassMountain},
		{3000, ClassRiver},
		{3999, ClassRiver},
		{3001, ClassRiver},
		{0, ""},
		{999, ""},
		{4000, ""},
		{9000, ClassHill},
		{9999, ClassHill},
	}
	for _, tt := range tests {
		got := classForCabinetNumber(tt.cabinet)
		if got != tt.want {
			t.Errorf("classForCabinetNumber(%d) = %q, want %q",
				tt.cabinet, got, tt.want)
		}
	}
}
