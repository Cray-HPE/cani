package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"
)

func TestClassifyNode_Admin(t *testing.T) {
	node := import_.Node{Name: "headnode", Type: "admin"}
	cl := classifyNode(node)
	if cl.Category != CategoryLocation {
		t.Errorf("admin category = %q, want %q", cl.Category, CategoryLocation)
	}
}

func TestClassifyNode_Chassis(t *testing.T) {
	node := import_.Node{Name: "x9000c1", Type: "chassis"}
	cl := classifyNode(node)
	if cl.Category != CategoryDevice {
		t.Errorf("chassis category = %q, want %q", cl.Category, CategoryDevice)
	}
	if cl.DeviceTypeHint != devicetypes.TypeChassis {
		t.Errorf("chassis hint = %v, want %v", cl.DeviceTypeHint, devicetypes.TypeChassis)
	}
}

func TestClassifyNode_MgmtSwitch(t *testing.T) {
	node := import_.Node{Name: "sw1", Type: "mgmt_switch"}
	cl := classifyNode(node)
	if cl.Category != CategoryDevice {
		t.Errorf("mgmt_switch category = %q, want %q", cl.Category, CategoryDevice)
	}
	if cl.DeviceTypeHint != devicetypes.TypeMgmtSwitch {
		t.Errorf("mgmt_switch hint = %v, want %v", cl.DeviceTypeHint, devicetypes.TypeMgmtSwitch)
	}
}

func TestClassifyNode_PDU(t *testing.T) {
	node := import_.Node{Name: "pdu1", Type: "pdu"}
	cl := classifyNode(node)
	if cl.Category != CategoryDevice {
		t.Errorf("pdu category = %q, want %q", cl.Category, CategoryDevice)
	}
	if cl.DeviceTypeHint != devicetypes.TypeCabinetPDU {
		t.Errorf("pdu hint = %v, want %v", cl.DeviceTypeHint, devicetypes.TypeCabinetPDU)
	}
}

func TestClassifyNode_ComputeStandalone(t *testing.T) {
	// Compute with nil location → device.
	node := import_.Node{Name: "node1", Type: "compute", Location: nil}
	cl := classifyNode(node)
	if cl.Category != CategoryDevice {
		t.Errorf("standalone category = %q, want %q", cl.Category, CategoryDevice)
	}
	if cl.DeviceTypeHint != devicetypes.TypeNode {
		t.Errorf("standalone hint = %v, want %v", cl.DeviceTypeHint, devicetypes.TypeNode)
	}
}

func TestClassifyNode_ComputeChassisOnly(t *testing.T) {
	// Compute with only rack+chassis set → device (sits in chassis slot).
	node := import_.Node{
		Name: "blade1",
		Type: "compute",
		Location: &import_.LocationSettings{
			Rack:    import_.Int32Ptr(9000),
			Chassis: import_.Int32Ptr(1),
		},
	}
	cl := classifyNode(node)
	if cl.Category != CategoryDevice {
		t.Errorf("chassis-only category = %q, want %q", cl.Category, CategoryDevice)
	}
}

func TestClassifyNode_ComputeWithTray(t *testing.T) {
	// Compute with tray set → module.
	node := import_.Node{
		Name: "tray-node",
		Type: "compute",
		Location: &import_.LocationSettings{
			Rack:    import_.Int32Ptr(9000),
			Chassis: import_.Int32Ptr(1),
			Tray:    import_.Int32Ptr(7),
			Node:    import_.Int32Ptr(0),
		},
	}
	cl := classifyNode(node)
	if cl.Category != CategoryModule {
		t.Errorf("tray category = %q, want %q", cl.Category, CategoryModule)
	}
}

func TestClassifyNode_ComputeWithNode(t *testing.T) {
	// Compute with node set (tray nil) → module.
	node := import_.Node{
		Name: "node-only",
		Type: "compute",
		Location: &import_.LocationSettings{
			Rack:    import_.Int32Ptr(9000),
			Chassis: import_.Int32Ptr(1),
			Node:    import_.Int32Ptr(2),
		},
	}
	cl := classifyNode(node)
	if cl.Category != CategoryModule {
		t.Errorf("node-only category = %q, want %q", cl.Category, CategoryModule)
	}
}

func TestClassifyNode_ComputeWithController(t *testing.T) {
	// Compute with controller set → module.
	node := import_.Node{
		Name: "ctrl-node",
		Type: "compute",
		Location: &import_.LocationSettings{
			Rack:       import_.Int32Ptr(9000),
			Chassis:    import_.Int32Ptr(1),
			Controller: import_.Int32Ptr(1),
		},
	}
	cl := classifyNode(node)
	if cl.Category != CategoryModule {
		t.Errorf("controller category = %q, want %q", cl.Category, CategoryModule)
	}
}

func TestClassifyNode_LookupQueries(t *testing.T) {
	node := import_.Node{
		Name:    "dl360gen11",
		Type:    "compute",
		Aliases: map[string]string{"product": "DL_360_gen11"},
		Inventory: map[string]string{
			"fru.system.SKU":   "P38578-B21",
			"sys.Product Name": "ProLiant DL360 Gen11",
			"fru.Model":        "ProLiant DL360 Gen11",
		},
	}
	cl := classifyNode(node)
	if len(cl.LookupQueries) == 0 {
		t.Fatal("expected lookup queries, got none")
	}
	// SKU should appear before the node name.
	skuIdx := -1
	nameIdx := -1
	for i, q := range cl.LookupQueries {
		if q == "P38578-B21" {
			skuIdx = i
		}
		if q == "dl360gen11" {
			nameIdx = i
		}
	}
	if skuIdx < 0 {
		t.Error("SKU P38578-B21 not found in queries")
	}
	if nameIdx < 0 {
		t.Error("node name dl360gen11 not found in queries")
	}
	if skuIdx >= 0 && nameIdx >= 0 && skuIdx > nameIdx {
		t.Errorf("SKU (idx %d) should rank before name (idx %d)", skuIdx, nameIdx)
	}
}

func TestCollectQueries_SentinelFiltered(t *testing.T) {
	node := import_.Node{
		Name:    "dl380gen11",
		Type:    "compute",
		Aliases: map[string]string{"product": "DL_380_gen11"},
		Inventory: map[string]string{
			"fru.system.SKU":     "NA",
			"sys.SKU Number":     "NA",
			"fru.SKU":            "NA",
			"fru.system.Model":   "ProLiant DL380 Gen11",
			"sys.Product Name":   "ProLiant DL380 Gen11",
			"board.Product Name": "ProLiant DL380 Gen11",
			"fru.Model":          "ProLiant DL380 Gen11",
		},
	}
	queries := collectQueries(nil, node)

	// "NA" must be filtered out.
	for _, q := range queries {
		if q == "NA" {
			t.Error("sentinel value 'NA' should be filtered from queries")
		}
	}
	// Model should still be present.
	found := false
	for _, q := range queries {
		if q == "ProLiant DL380 Gen11" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'ProLiant DL380 Gen11' in queries")
	}
}

func TestCollectQueries_CommonSentinels(t *testing.T) {
	tests := []string{
		"NA", "N/A", "Not Specified", "UNKNOWN", "Unknown",
		"Default string", "Unspecified", "None", "Not Available",
	}
	for _, val := range tests {
		node := import_.Node{
			Name: "test",
			Inventory: map[string]string{
				"fru.system.SKU": val,
			},
		}
		queries := collectQueries(nil, node)
		for _, q := range queries {
			if q == val {
				t.Errorf("sentinel %q should be filtered from queries", val)
			}
		}
	}
}

func TestIsSentinel(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"NA", true},
		{"na", true},
		{"N/A", true},
		{"Not Specified", true},
		{"UNKNOWN", true},
		{"Default string", true},
		{"AB", true},          // too short (< 3)
		{"00000000", true},    // all zeros
		{"P38578-B21", false}, // real part number
		{"ProLiant DL380 Gen11", false},
		{"DL_380_gen11", false},
		{"1395A3306301", false}, // hex but not all-zero
	}
	for _, tt := range tests {
		got := isSentinel(tt.input)
		if got != tt.want {
			t.Errorf("isSentinel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestClassifyNode_RackOnlyLocation(t *testing.T) {
	node := import_.Node{
		Name: "standalone",
		Type: "compute",
		Location: &import_.LocationSettings{
			Rack: import_.Int32Ptr(4006),
		},
	}
	cl := classifyNode(node)
	if cl.Category != CategoryDevice {
		t.Errorf("rack-only category = %q, want %q", cl.Category, CategoryDevice)
	}
}

func TestClassifyByLocation_Nil(t *testing.T) {
	cat, hint := classifyByLocation(nil)
	if cat != CategoryDevice {
		t.Errorf("nil location category = %q, want %q", cat, CategoryDevice)
	}
	if hint != devicetypes.TypeNode {
		t.Errorf("nil location hint = %v, want %v", hint, devicetypes.TypeNode)
	}
}

func TestModuleBayName(t *testing.T) {
	tests := []struct {
		name string
		loc  *import_.LocationSettings
		want string
	}{
		{"nil", nil, ""},
		{"tray-node", &import_.LocationSettings{
			Tray: import_.Int32Ptr(7),
			Node: import_.Int32Ptr(0),
		}, "tray-7-node-0"},
		{"controller", &import_.LocationSettings{
			Tray:       import_.Int32Ptr(0),
			Node:       import_.Int32Ptr(0),
			Controller: import_.Int32Ptr(1),
		}, "tray-0-node-0-ctrl-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := moduleBayName(tt.loc)
			if got != tt.want {
				t.Errorf("moduleBayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollectQueries_AllAliases(t *testing.T) {
	node := import_.Node{
		Name: "fmn2",
		Type: "compute",
		Aliases: map[string]string{
			"product": "XL225n_Gen10Plus",
			"a2000":   "XL225n_Gen10_Plus",
			"fmn":     "fmn-standby",
		},
		Inventory: map[string]string{
			"fru.system.Model": "ProLiant XL225n Gen10 Plus",
			"fru.system.SKU":   "P21163-B21",
		},
	}
	queries := collectQueries(nil, node)

	// product alias should be present.
	found := make(map[string]bool)
	for _, q := range queries {
		found[q] = true
	}

	if !found["XL225n_Gen10Plus"] {
		t.Error("expected product alias 'XL225n_Gen10Plus' in queries")
	}
	// a2000 alias should also be present (all aliases extracted).
	if !found["XL225n_Gen10_Plus"] {
		t.Error("expected a2000 alias 'XL225n_Gen10_Plus' in queries")
	}
	// fmn alias should be present too.
	if !found["fmn-standby"] {
		t.Error("expected fmn alias 'fmn-standby' in queries")
	}
	// SKU should come before aliases.
	skuIdx := -1
	aliasIdx := -1
	for i, q := range queries {
		if q == "P21163-B21" {
			skuIdx = i
		}
		if q == "XL225n_Gen10_Plus" {
			aliasIdx = i
		}
	}
	if skuIdx >= 0 && aliasIdx >= 0 && skuIdx > aliasIdx {
		t.Errorf("SKU (idx %d) should come before non-product alias (idx %d)", skuIdx, aliasIdx)
	}
}

func TestCollectQueries_CtrlModel(t *testing.T) {
	// cm.config node with ctrl_model: should appear after product alias, before name.
	node := import_.Node{
		Name: "antero001",
		Type: "compute",
		Aliases: map[string]string{
			"ctrl_model":    "EX425",
			"template_name": "genoa",
			"card_type":     "iLO",
		},
	}
	queries := collectQueries(nil, node)

	found := make(map[string]bool)
	for _, q := range queries {
		found[q] = true
	}

	if !found["EX425"] {
		t.Error("expected ctrl_model 'EX425' in queries")
	}
	if !found["genoa"] {
		t.Error("expected template_name 'genoa' in queries")
	}

	// ctrl_model (step 3b) should appear before node name (step 4).
	ctrlIdx := -1
	nameIdx := -1
	for i, q := range queries {
		if q == "EX425" {
			ctrlIdx = i
		}
		if q == "antero001" {
			nameIdx = i
		}
	}
	if ctrlIdx >= 0 && nameIdx >= 0 && ctrlIdx > nameIdx {
		t.Errorf("ctrl_model (idx %d) should rank before name (idx %d)", ctrlIdx, nameIdx)
	}

	// template_name (step 4b) should appear after name (step 4).
	tplIdx := -1
	for i, q := range queries {
		if q == "genoa" {
			tplIdx = i
		}
	}
	if tplIdx >= 0 && nameIdx >= 0 && tplIdx < nameIdx {
		t.Errorf("template_name (idx %d) should rank after name (idx %d)", tplIdx, nameIdx)
	}
}

func TestClassifyNode_CmConfigCompute(t *testing.T) {
	// cm.config compute node with location and ctrl_model alias.
	node := import_.Node{
		Name: "antero001",
		Type: "compute",
		Aliases: map[string]string{
			"ctrl_model":    "EX425",
			"template_name": "genoa",
			"cm-geo-name":   "x9000c1s7b0n0",
		},
		Location: &import_.LocationSettings{
			Rack:    import_.Int32Ptr(9000),
			Chassis: import_.Int32Ptr(1),
			Tray:    import_.Int32Ptr(7),
			Node:    import_.Int32Ptr(0),
		},
	}
	cl := classifyNode(node)
	if cl.Category != CategoryModule {
		t.Errorf("category = %q, want %q", cl.Category, CategoryModule)
	}
	// ctrl_model should be in lookup queries.
	found := false
	for _, q := range cl.LookupQueries {
		if q == "EX425" {
			found = true
		}
	}
	if !found {
		t.Error("expected ctrl_model 'EX425' in lookup queries")
	}
}
