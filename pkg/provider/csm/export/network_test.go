package export

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/google/uuid"
)

func TestDeriveNMNVlan(t *testing.T) {
	tests := []struct {
		name     string
		hmnVlan  int
		hmn      *import_.SlsNetwork
		nmn      *import_.SlsNetwork
		expected int
	}{
		{
			name:    "passing test computes NMN VLAN from offset",
			hmnVlan: 3002,
			hmn: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					VlanRange: []int{3000},
				},
			},
			nmn: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					VlanRange: []int{2000},
				},
			},
			expected: 2002,
		},
		{
			name:    "failing test with nil extra properties returns zero offset",
			hmnVlan: 5,
			hmn: &import_.SlsNetwork{
				ExtraProperties: nil,
			},
			nmn: &import_.SlsNetwork{
				ExtraProperties: nil,
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveNMNVlan(tt.hmnVlan, tt.hmn, tt.nmn)
			if got != tt.expected {
				t.Errorf("deriveNMNVlan() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestComputeSubnet(t *testing.T) {
	tests := []struct {
		name        string
		network     *import_.SlsNetwork
		vlan        int
		xname       string
		expectedErr bool
	}{
		{
			name: "passing test computes valid subnet",
			network: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					CIDR:      "10.100.0.0/17",
					VlanRange: []int{3000},
				},
			},
			vlan:        3001,
			xname:       "x9000",
			expectedErr: false,
		},
		{
			name: "failing test with invalid CIDR",
			network: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					CIDR:      "not-a-cidr",
					VlanRange: []int{3000},
				},
			},
			vlan:        3001,
			xname:       "x9000",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := computeSubnet(tt.network, tt.vlan, tt.xname)

			if tt.expectedErr && err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !tt.expectedErr {
				if info.CIDR == "" {
					t.Error("CIDR should not be empty")
				}
				if info.Gateway == "" {
					t.Error("Gateway should not be empty")
				}
			}
		})
	}
}

func TestSetNetworkMetadata(t *testing.T) {
	tests := []struct {
		name    string
		hw      import_.SlsHardware
		hmn     subnetInfo
		nmn     subnetInfo
		hmnVlan int
		nmnVlan int
	}{
		{
			name: "passing test sets Networks in ExtraProperties",
			hw: import_.SlsHardware{
				ExtraProperties: map[string]any{},
			},
			hmn:     subnetInfo{CIDR: "10.100.0.0/22", Gateway: "10.100.0.1"},
			nmn:     subnetInfo{CIDR: "10.104.0.0/22", Gateway: "10.104.0.1"},
			hmnVlan: 3000,
			nmnVlan: 2000,
		},
		{
			name: "failing test with nil ExtraProperties initializes map",
			hw: import_.SlsHardware{
				ExtraProperties: nil,
			},
			hmn:     subnetInfo{CIDR: "10.100.0.0/22", Gateway: "10.100.0.1"},
			nmn:     subnetInfo{CIDR: "10.104.0.0/22", Gateway: "10.104.0.1"},
			hmnVlan: 3000,
			nmnVlan: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setNetworkMetadata(&tt.hw, tt.hmn, tt.nmn, tt.hmnVlan, tt.nmnVlan)

			if tt.hw.ExtraProperties == nil {
				t.Error("ExtraProperties should not be nil after setNetworkMetadata")
				return
			}
			if _, ok := tt.hw.ExtraProperties["Networks"]; !ok {
				t.Error("ExtraProperties should contain Networks key")
			}
		})
	}
}

func TestAppendSubnet(t *testing.T) {
	tests := []struct {
		name     string
		network  *import_.SlsNetwork
		subName  string
		info     subnetInfo
		vlan     int
		expected bool
	}{
		{
			name: "passing test appends new subnet",
			network: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					Subnets: []import_.SlsSubnet{},
				},
			},
			subName:  "cabinet_9000",
			info:     subnetInfo{CIDR: "10.100.0.0/22", Gateway: "10.100.0.1"},
			vlan:     3000,
			expected: true,
		},
		{
			name: "failing test skips duplicate subnet",
			network: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					Subnets: []import_.SlsSubnet{
						{Name: "cabinet_9000"},
					},
				},
			},
			subName:  "cabinet_9000",
			info:     subnetInfo{CIDR: "10.100.0.0/22", Gateway: "10.100.0.1"},
			vlan:     3000,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendSubnet(tt.network, tt.subName, tt.info, tt.vlan)
			if got != tt.expected {
				t.Errorf("appendSubnet() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseVlan(t *testing.T) {
	tests := []struct {
		name     string
		net      *import_.SlsNetwork
		expected int
	}{
		{
			name: "passing test returns first VLAN",
			net: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					VlanRange: []int{3000, 3999},
				},
			},
			expected: 3000,
		},
		{
			name: "failing test with nil ExtraProperties returns zero",
			net: &import_.SlsNetwork{
				ExtraProperties: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := baseVlan(tt.net)
			if got != tt.expected {
				t.Errorf("baseVlan() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestNetworkBaseCIDR(t *testing.T) {
	tests := []struct {
		name     string
		net      *import_.SlsNetwork
		expected string
	}{
		{
			name: "passing test returns CIDR",
			net: &import_.SlsNetwork{
				ExtraProperties: &import_.SlsNetworkExtraProperties{
					CIDR: "10.100.0.0/17",
				},
			},
			expected: "10.100.0.0/17",
		},
		{
			name: "failing test with nil ExtraProperties returns empty",
			net: &import_.SlsNetwork{
				ExtraProperties: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := networkBaseCIDR(tt.net)
			if got != tt.expected {
				t.Errorf("networkBaseCIDR() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCabinetSubnetName(t *testing.T) {
	tests := []struct {
		name     string
		ordinal  int
		expected string
	}{
		{
			name:     "passing test returns formatted name",
			ordinal:  9000,
			expected: "cabinet_9000",
		},
		{
			name:     "failing test with zero ordinal",
			ordinal:  0,
			expected: "cabinet_0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cabinetSubnetName(tt.ordinal)
			if got != tt.expected {
				t.Errorf("cabinetSubnetName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCabinetOrdinalFromXname(t *testing.T) {
	tests := []struct {
		name     string
		xname    string
		expected int
	}{
		{
			name:     "passing test extracts ordinal",
			xname:    "x9000",
			expected: 9000,
		},
		{
			name:     "failing test with invalid xname returns zero",
			xname:    "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cabinetOrdinalFromXname(tt.xname)
			if got != tt.expected {
				t.Errorf("cabinetOrdinalFromXname() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestFindDeviceByXname(t *testing.T) {
	deviceID := uuid.New()

	tests := []struct {
		name    string
		xname   string
		inv     devicetypes.Inventory
		wantNil bool
	}{
		{
			name:  "passing test finds device by xname",
			xname: "x9000",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
					deviceID: {
						ID: deviceID,
						ProviderMetadata: map[string]any{
							"csm": map[string]any{
								"xname": "x9000",
							},
						},
					},
				},
			},
			wantNil: false,
		},
		{
			name:  "failing test device not found",
			xname: "x1234",
			inv: devicetypes.Inventory{
				Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDeviceByXname(tt.xname, tt.inv)
			if tt.wantNil && got != nil {
				t.Error("expected nil but got a device")
			}
			if !tt.wantNil && got == nil {
				t.Error("expected a device but got nil")
			}
		})
	}
}

func TestIntFromMeta(t *testing.T) {
	tests := []struct {
		name     string
		meta     map[string]any
		key      string
		expected int
	}{
		{
			name:     "passing test with int value",
			meta:     map[string]any{"hmnVlan": 3000},
			key:      "hmnVlan",
			expected: 3000,
		},
		{
			name:     "failing test with missing key returns zero",
			meta:     map[string]any{},
			key:      "hmnVlan",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intFromMeta(tt.meta, tt.key)
			if got != tt.expected {
				t.Errorf("intFromMeta() = %d, want %d", got, tt.expected)
			}
		})
	}
}
