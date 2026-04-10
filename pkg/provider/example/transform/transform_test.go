package transform

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestDeviceTypePriority(t *testing.T) {
	tests := []struct {
		hwType string
		want   int
	}{
		{"pdu", 0},
		{"cdu", 0},
		{"node", 1},
		{"blade", 1},
		{"chassis", 1},
		{"switch", 2},
		{"mgmt-switch", 2},
		{"hsn-switch", 2},
		{"", 1},
	}

	for _, tt := range tests {
		t.Run(tt.hwType, func(t *testing.T) {
			got := deviceTypePriority(tt.hwType)
			if got != tt.want {
				t.Errorf("deviceTypePriority(%q) = %d, want %d", tt.hwType, got, tt.want)
			}
		})
	}
}

func TestSortDevicesByRackPriority(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	switchID := uuid.New()
	nodeID := uuid.New()
	pduID := uuid.New()
	bladeID := uuid.New()
	cduID := uuid.New()

	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[bladeID] = &devicetypes.CaniDeviceType{ID: bladeID, Type: devicetypes.Type("blade")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}

	ids := []uuid.UUID{switchID, nodeID, pduID, bladeID, cduID}
	sortDevicesByRackPriority(inv, ids)

	// PDUs/CDUs should come first, then nodes/blades, then switches
	for i, id := range ids {
		dev := inv.Devices[id]
		pri := deviceTypePriority(string(dev.Type))
		if i > 0 {
			prevDev := inv.Devices[ids[i-1]]
			prevPri := deviceTypePriority(string(prevDev.Type))
			if pri < prevPri {
				t.Errorf("device at index %d (%s, priority %d) sorted before index %d (%s, priority %d)",
					i, dev.Type, pri, i-1, prevDev.Type, prevPri)
			}
		}
	}

	// Verify first two are pdu/cdu (priority 0)
	for _, id := range ids[:2] {
		dev := inv.Devices[id]
		if deviceTypePriority(string(dev.Type)) != 0 {
			t.Errorf("expected pdu/cdu in first two positions, got %q", dev.Type)
		}
	}

	// Verify last is switch (priority 2)
	lastDev := inv.Devices[ids[len(ids)-1]]
	if deviceTypePriority(string(lastDev.Type)) != 2 {
		t.Errorf("expected switch in last position, got %q", lastDev.Type)
	}
}

func TestGroupDevicesByZone(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	pduID := uuid.New()
	cduID := uuid.New()
	nodeID := uuid.New()
	switchID := uuid.New()

	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu")}
	inv.Devices[cduID] = &devicetypes.CaniDeviceType{ID: cduID, Type: devicetypes.Type("cdu")}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node")}
	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch")}

	ids := []uuid.UUID{pduID, cduID, nodeID, switchID}
	bottom, middle, top := groupDevicesByZone(inv, ids)

	if len(bottom) != 2 {
		t.Errorf("expected 2 bottom devices, got %d", len(bottom))
	}
	if len(middle) != 1 {
		t.Errorf("expected 1 middle device, got %d", len(middle))
	}
	if len(top) != 1 {
		t.Errorf("expected 1 top device, got %d", len(top))
	}
}

func TestRackPositionOrdering(t *testing.T) {
	// Reset global state
	resetRackPositionStates()

	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{
		ID:      rackID,
		Name:    "Rack-001",
		UHeight: 48,
		Devices: []uuid.UUID{},
	}

	switchID := uuid.New()
	nodeID := uuid.New()
	pduID := uuid.New()

	inv.Devices[switchID] = &devicetypes.CaniDeviceType{ID: switchID, Type: devicetypes.Type("switch"), Slug: "switch-1u"}
	inv.Devices[nodeID] = &devicetypes.CaniDeviceType{ID: nodeID, Type: devicetypes.Type("node"), Slug: "node-1u"}
	inv.Devices[pduID] = &devicetypes.CaniDeviceType{ID: pduID, Type: devicetypes.Type("pdu"), Slug: "pdu-1u"}

	racksByGroup := map[string][]uuid.UUID{
		"0100": {rackID},
	}
	devicesByGroup := map[string][]uuid.UUID{
		"0200": {switchID, nodeID, pduID},
	}

	assignConfigGroupParenting(inv, racksByGroup, devicesByGroup)

	pdu := inv.Devices[pduID]
	node := inv.Devices[nodeID]
	sw := inv.Devices[switchID]

	// PDU should be at the bottom (low U)
	if pdu.RackPosition != 1 {
		t.Errorf("PDU position = %d, want 1 (bottom)", pdu.RackPosition)
	}

	// Switch should be at the top (U48)
	if sw.RackPosition != 48 {
		t.Errorf("switch position = %d, want 48 (top)", sw.RackPosition)
	}

	// Node should be just below switch (U47, filling downward)
	if node.RackPosition != 47 {
		t.Errorf("node position = %d, want 47 (below switch)", node.RackPosition)
	}
}

func TestRackZonesFillCorrectly(t *testing.T) {
	resetRackPositionStates()

	inv := &devicetypes.Inventory{
		Racks:   make(map[uuid.UUID]*devicetypes.CaniRackType),
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	rackID := uuid.New()
	inv.Racks[rackID] = &devicetypes.CaniRackType{
		ID:      rackID,
		Name:    "Rack-001",
		UHeight: 10,
		Devices: []uuid.UUID{},
	}

	// 2 PDUs (1U each) at bottom, 2 switches (1U each) at top, 1 node in middle
	pdu1 := uuid.New()
	pdu2 := uuid.New()
	sw1 := uuid.New()
	sw2 := uuid.New()
	node1 := uuid.New()

	inv.Devices[pdu1] = &devicetypes.CaniDeviceType{ID: pdu1, Type: devicetypes.Type("pdu")}
	inv.Devices[pdu2] = &devicetypes.CaniDeviceType{ID: pdu2, Type: devicetypes.Type("cdu")}
	inv.Devices[sw1] = &devicetypes.CaniDeviceType{ID: sw1, Type: devicetypes.Type("switch")}
	inv.Devices[sw2] = &devicetypes.CaniDeviceType{ID: sw2, Type: devicetypes.Type("mgmt-switch")}
	inv.Devices[node1] = &devicetypes.CaniDeviceType{ID: node1, Type: devicetypes.Type("node")}

	racksByGroup := map[string][]uuid.UUID{"0100": {rackID}}
	devicesByGroup := map[string][]uuid.UUID{
		"0200": {pdu1, pdu2, sw1, sw2, node1},
	}

	assignConfigGroupParenting(inv, racksByGroup, devicesByGroup)

	// Bottom zone: U1, U2
	if inv.Devices[pdu1].RackPosition != 1 {
		t.Errorf("pdu1 position = %d, want 1", inv.Devices[pdu1].RackPosition)
	}
	if inv.Devices[pdu2].RackPosition != 2 {
		t.Errorf("pdu2 position = %d, want 2", inv.Devices[pdu2].RackPosition)
	}

	// Middle zone: node fills downward from below switches → U8
	if inv.Devices[node1].RackPosition != 8 {
		t.Errorf("node position = %d, want 8", inv.Devices[node1].RackPosition)
	}

	// Top zone: U10 (first switch), U9 (second switch, fills downward)
	if inv.Devices[sw1].RackPosition != 10 {
		t.Errorf("sw1 position = %d, want 10", inv.Devices[sw1].RackPosition)
	}
	if inv.Devices[sw2].RackPosition != 9 {
		t.Errorf("sw2 position = %d, want 9", inv.Devices[sw2].RackPosition)
	}
}
