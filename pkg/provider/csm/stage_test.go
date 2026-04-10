package csm

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func makeDev(id uuid.UUID, slug string, parent uuid.UUID, xn string) *devicetypes.CaniDeviceType {
	dev, _ := devicetypes.NewDeviceFromSlug(slug)
	if dev == nil {
		dev = &devicetypes.CaniDeviceType{Slug: slug}
	}
	dev.ID = id
	dev.Parent = parent
	dev.Status = string(devicetypes.StatusActive)
	dev.SetProviderMeta("csm", "xname", xn)
	return dev
}

func TestStageExisting(t *testing.T) {
	bladeSlug := "hpe-crayex-ex235a-compute-blade"
	ncSlug := "hpe-crayex-ex235a-compute-blade-bard-peak-node-card"
	nodeSlug := "hpe-crayex-ex235a-compute-node"

	bladeID := uuid.New()
	nc0ID := uuid.New()
	nc1ID := uuid.New()
	n0ID := uuid.New()
	n1ID := uuid.New()
	n2ID := uuid.New()
	n3ID := uuid.New()

	blade := makeDev(bladeID, bladeSlug, uuid.Nil, "x9000c1s0")
	nc0 := makeDev(nc0ID, ncSlug, bladeID, "x9000c1s0b0")
	nc1 := makeDev(nc1ID, ncSlug, bladeID, "x9000c1s0b1")
	n0 := makeDev(n0ID, nodeSlug, nc0ID, "x9000c1s0b0n0")
	n1 := makeDev(n1ID, nodeSlug, nc0ID, "x9000c1s0b0n1")
	n2 := makeDev(n2ID, nodeSlug, nc1ID, "x9000c1s0b1n0")
	n3 := makeDev(n3ID, nodeSlug, nc1ID, "x9000c1s0b1n1")

	blade.Children = []uuid.UUID{nc0ID, nc1ID}
	nc0.Children = []uuid.UUID{n0ID, n1ID}
	nc1.Children = []uuid.UUID{n2ID, n3ID}

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			bladeID: blade,
			nc0ID:   nc0,
			nc1ID:   nc1,
			n0ID:    n0,
			n1ID:    n1,
			n2ID:    n2,
			n3ID:    n3,
		},
	}

	ok := StageExisting(inv, bladeSlug, devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug)
	if !ok {
		t.Fatal("StageExisting returned false, expected true")
	}

	if !strings.EqualFold(blade.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("blade status = %q, want %q", blade.Status, devicetypes.StatusStaged)
	}

	for _, nc := range []*devicetypes.CaniDeviceType{nc0, nc1} {
		if !strings.EqualFold(nc.Status, string(devicetypes.StatusStaged)) {
			t.Errorf("node card %s status = %q, want %q", xname(nc), nc.Status, devicetypes.StatusStaged)
		}
	}

	for _, n := range []*devicetypes.CaniDeviceType{n0, n2} {
		if !strings.EqualFold(n.Status, string(devicetypes.StatusStaged)) {
			t.Errorf("node %s status = %q, want %q", xname(n), n.Status, devicetypes.StatusStaged)
		}
	}

	for _, n := range []*devicetypes.CaniDeviceType{n1, n3} {
		if !strings.EqualFold(n.Status, string(devicetypes.StatusActive)) {
			t.Errorf("node %s status = %q, want %q", xname(n), n.Status, devicetypes.StatusActive)
		}
	}
}

func TestStageExistingPicksLowestXname(t *testing.T) {
	bladeSlug := "hpe-crayex-ex235a-compute-blade"

	blade0ID := uuid.New()
	blade6ID := uuid.New()

	blade0 := makeDev(blade0ID, bladeSlug, uuid.Nil, "x9000c1s0")
	blade6 := makeDev(blade6ID, bladeSlug, uuid.Nil, "x9000c1s6")

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			blade0ID: blade0,
			blade6ID: blade6,
		},
	}

	ok := StageExisting(inv, bladeSlug, devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug)
	if !ok {
		t.Fatal("StageExisting returned false")
	}

	if !strings.EqualFold(blade0.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("blade0 status = %q, want %q", blade0.Status, devicetypes.StatusStaged)
	}
	if !strings.EqualFold(blade6.Status, string(devicetypes.StatusActive)) {
		t.Errorf("blade6 status = %q, want %q (should not have been staged)", blade6.Status, devicetypes.StatusActive)
	}
}

func TestStageExistingCrossSlug(t *testing.T) {
	importedSlug := "hpe-crayex-ex235a-compute-blade"
	addSlug := "hpe-crayex-ex420-compute-blade"

	ncSlug := "hpe-crayex-ex235a-compute-blade-bard-peak-node-card"
	nodeSlug := "hpe-crayex-ex235a-compute-node"

	bladeID := uuid.New()
	nc0ID := uuid.New()
	nc1ID := uuid.New()
	n0ID := uuid.New()
	n1ID := uuid.New()
	n2ID := uuid.New()
	n3ID := uuid.New()

	blade := makeDev(bladeID, importedSlug, uuid.Nil, "x9000c1s0")
	nc0 := makeDev(nc0ID, ncSlug, bladeID, "x9000c1s0b0")
	nc1 := makeDev(nc1ID, ncSlug, bladeID, "x9000c1s0b1")
	n0 := makeDev(n0ID, nodeSlug, nc0ID, "x9000c1s0b0n0")
	n1 := makeDev(n1ID, nodeSlug, nc0ID, "x9000c1s0b0n1")
	n2 := makeDev(n2ID, nodeSlug, nc1ID, "x9000c1s0b1n0")
	n3 := makeDev(n3ID, nodeSlug, nc1ID, "x9000c1s0b1n1")

	blade.Children = []uuid.UUID{nc0ID, nc1ID}
	nc0.Children = []uuid.UUID{n0ID, n1ID}
	nc1.Children = []uuid.UUID{n2ID, n3ID}

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			bladeID: blade,
			nc0ID:   nc0,
			nc1ID:   nc1,
			n0ID:    n0,
			n1ID:    n1,
			n2ID:    n2,
			n3ID:    n3,
		},
	}

	ok := StageExisting(inv, addSlug, devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug)
	if !ok {
		t.Fatal("StageExisting returned false, expected true")
	}

	if blade.Slug != addSlug {
		t.Errorf("blade slug = %q, want %q", blade.Slug, addSlug)
	}

	for _, n := range []*devicetypes.CaniDeviceType{n0, n1, n2, n3} {
		if !strings.EqualFold(n.Status, string(devicetypes.StatusStaged)) {
			t.Errorf("node %s status = %q, want %q", xname(n), n.Status, devicetypes.StatusStaged)
		}
	}
}

// TestStageExistingEX235N mimics the integration scenario: devices imported
// with Mountain-default slugs (EX235A), then adding an EX235N blade.
// EX235N blade has 1 node card bay (2 nodes), so b0 should be staged
// and both b0n0 and b0n1 should be staged.
func TestStageExistingEX235N(t *testing.T) {
	// Imported slugs (Mountain defaults from SLS import).
	importedBladeSlug := "hpe-crayex-ex235a-compute-blade"
	importedNCSlug := "hpe-crayex-ex235a-compute-blade-bard-peak-node-card"
	importedNodeSlug := "hpe-crayex-ex235a-compute-node"

	// Adding this slug.
	addSlug := "hpe-crayex-ex235n-compute-blade"

	bladeID := uuid.New()
	nc0ID := uuid.New()
	nc1ID := uuid.New()
	n0ID := uuid.New()
	n1ID := uuid.New()
	n2ID := uuid.New()
	n3ID := uuid.New()

	blade := makeDev(bladeID, importedBladeSlug, uuid.Nil, "x9000c1s0")
	nc0 := makeDev(nc0ID, importedNCSlug, bladeID, "x9000c1s0b0")
	nc1 := makeDev(nc1ID, importedNCSlug, bladeID, "x9000c1s0b1")
	n0 := makeDev(n0ID, importedNodeSlug, nc0ID, "x9000c1s0b0n0")
	n1 := makeDev(n1ID, importedNodeSlug, nc0ID, "x9000c1s0b0n1")
	n2 := makeDev(n2ID, importedNodeSlug, nc1ID, "x9000c1s0b1n0")
	n3 := makeDev(n3ID, importedNodeSlug, nc1ID, "x9000c1s0b1n1")

	blade.Children = []uuid.UUID{nc0ID, nc1ID}
	nc0.Children = []uuid.UUID{n0ID, n1ID}
	nc1.Children = []uuid.UUID{n2ID, n3ID}

	// Set Hill class metadata.
	blade.SetProviderMeta("csm", "class", "Hill")
	nc0.SetProviderMeta("csm", "class", "Hill")
	nc1.SetProviderMeta("csm", "class", "Hill")
	n0.SetProviderMeta("csm", "class", "Hill")
	n1.SetProviderMeta("csm", "class", "Hill")
	n2.SetProviderMeta("csm", "class", "Hill")
	n3.SetProviderMeta("csm", "class", "Hill")

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			bladeID: blade,
			nc0ID:   nc0,
			nc1ID:   nc1,
			n0ID:    n0,
			n1ID:    n1,
			n2ID:    n2,
			n3ID:    n3,
		},
	}

	ok := StageExisting(inv, addSlug, devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug)
	if !ok {
		t.Fatal("StageExisting returned false, expected true")
	}

	// Blade should be staged.
	if !strings.EqualFold(blade.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("blade status = %q, want %q", blade.Status, devicetypes.StatusStaged)
	}

	// EX235N has 1 node card bay, so only b0 should be staged.
	if !strings.EqualFold(nc0.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("nc0 (%s) status = %q, want %q", xname(nc0), nc0.Status, devicetypes.StatusStaged)
	}
	if !strings.EqualFold(nc1.Status, string(devicetypes.StatusActive)) {
		t.Errorf("nc1 (%s) status = %q, want %q (should NOT be staged)", xname(nc1), nc1.Status, devicetypes.StatusActive)
	}

	// EX235N node card has 2 node bays, so both n0 and n1 under b0 should be staged.
	for _, n := range []*devicetypes.CaniDeviceType{n0, n1} {
		if !strings.EqualFold(n.Status, string(devicetypes.StatusStaged)) {
			t.Errorf("node %s status = %q, want %q", xname(n), n.Status, devicetypes.StatusStaged)
		}
	}

	// Nodes under b1 should remain Active.
	for _, n := range []*devicetypes.CaniDeviceType{n2, n3} {
		if !strings.EqualFold(n.Status, string(devicetypes.StatusActive)) {
			t.Errorf("node %s status = %q, want %q", xname(n), n.Status, devicetypes.StatusActive)
		}
	}
}

// TestStageExistingEX4252 tests that when the template has more default
// children than exist (EX4252: 4 nodes, but only 2 imported from SLS),
// the missing children are created with derived xnames.
func TestStageExistingEX4252(t *testing.T) {
	importedBladeSlug := "hpe-crayex-ex235a-compute-blade"
	importedNCSlug := "hpe-crayex-ex235a-compute-blade-bard-peak-node-card"
	importedNodeSlug := "hpe-crayex-ex235a-compute-node"

	addSlug := "hpe-crayex-ex4252-compute-blade"

	bladeID := uuid.New()
	nc0ID := uuid.New()
	n0ID := uuid.New()
	n1ID := uuid.New()

	blade := makeDev(bladeID, importedBladeSlug, uuid.Nil, "x9000c1s0")
	nc0 := makeDev(nc0ID, importedNCSlug, bladeID, "x9000c1s0b0")
	n0 := makeDev(n0ID, importedNodeSlug, nc0ID, "x9000c1s0b0n0")
	n1 := makeDev(n1ID, importedNodeSlug, nc0ID, "x9000c1s0b0n1")

	blade.Children = []uuid.UUID{nc0ID}
	nc0.Children = []uuid.UUID{n0ID, n1ID}

	blade.SetProviderMeta("csm", "class", "Hill")
	nc0.SetProviderMeta("csm", "class", "Hill")
	n0.SetProviderMeta("csm", "class", "Hill")
	n1.SetProviderMeta("csm", "class", "Hill")

	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			bladeID: blade,
			nc0ID:   nc0,
			n0ID:    n0,
			n1ID:    n1,
		},
	}

	ok := StageExisting(inv, addSlug, devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug)
	if !ok {
		t.Fatal("StageExisting returned false, expected true")
	}

	// Blade and nc0 should be staged.
	if !strings.EqualFold(blade.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("blade status = %q, want Staged", blade.Status)
	}
	if !strings.EqualFold(nc0.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("nc0 status = %q, want Staged", nc0.Status)
	}

	// Existing nodes n0 and n1 should be staged.
	if !strings.EqualFold(n0.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("n0 (%s) status = %q, want Staged", xname(n0), n0.Status)
	}
	if !strings.EqualFold(n1.Status, string(devicetypes.StatusStaged)) {
		t.Errorf("n1 (%s) status = %q, want Staged", xname(n1), n1.Status)
	}

	// EX4252 has 4 node bays. 2 new nodes should have been created (n2, n3).
	newNodes := 0
	for _, child := range nc0.Children {
		dev, ok := inv.Devices[child]
		if !ok {
			continue
		}
		if normalizeType(dev.GetType()) == devicetypes.TypeNode && dev.ID != n0ID && dev.ID != n1ID {
			newNodes++
			if !strings.EqualFold(dev.Status, string(devicetypes.StatusStaged)) {
				t.Errorf("new node %s status = %q, want Staged", xname(dev), dev.Status)
			}
			if deviceClass(dev) != "Hill" {
				t.Errorf("new node %s class = %q, want Hill", xname(dev), deviceClass(dev))
			}
		}
	}
	if newNodes != 2 {
		t.Errorf("expected 2 new nodes created, got %d", newNodes)
	}
}

func TestStageExistingNoMatch(t *testing.T) {
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			uuid.New(): {
				ID:         uuid.New(),
				Slug:       "some-other-slug",
				ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusActive)},
			},
		},
	}
	if StageExisting(inv, "hpe-crayex-ex235a-compute-blade", devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug) {
		t.Error("expected false when no matching type")
	}
}

func TestStageExistingAlreadyStaged(t *testing.T) {
	id := uuid.New()
	inv := &devicetypes.Inventory{
		Devices: map[uuid.UUID]*devicetypes.CaniDeviceType{
			id: {
				ID:         id,
				Slug:       "hpe-crayex-ex235a-compute-blade",
				ObjectMeta: devicetypes.ObjectMeta{Status: string(devicetypes.StatusStaged)},
			},
		},
	}
	if StageExisting(inv, "hpe-crayex-ex235a-compute-blade", devicetypes.GetBySlug, devicetypes.ApplyDeviceType, devicetypes.NewDeviceFromSlug) {
		t.Error("expected false when device already Staged")
	}
}
