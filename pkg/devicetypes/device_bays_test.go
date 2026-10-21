/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package devicetypes

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// bayInventory returns an inventory with a two-bay chassis whose Bay 2 only
// allows "sw-b", plus the chassis itself.
func bayInventory() (*Inventory, *CaniDeviceType) {
	inv := NewInventory()
	chassis := &CaniDeviceType{
		ID:   uuid.New(),
		Name: "chassis-15",
		DeviceBays: []DeviceBaySpec{
			{Name: "Bay 1"},
			{Name: "Bay 2", Allowed: &DeviceBaySlugRef{Slug: []interface{}{"sw-b"}}},
		},
	}
	inv.Devices[chassis.ID] = chassis
	return inv, chassis
}

// addChild installs a fresh device with slug into bay and registers it.
func addChild(t *testing.T, inv *Inventory, parentID uuid.UUID, slug, bay string) *CaniDeviceType {
	t.Helper()
	child := &CaniDeviceType{ID: uuid.New(), Name: slug + "-dev", Slug: slug, RackPosition: 15, Face: "front"}
	if err := inv.InstallInBay(child, parentID, bay, nil); err != nil {
		t.Fatalf("InstallInBay(%s, %q): %v", slug, bay, err)
	}
	inv.Devices[child.ID] = child
	return child
}

// TestInstallInBayAutoAssigns verifies children fill bays in order, respect
// allowed slugs, and clear their own rack placement.
//
// Why it matters: scripts add half-width switches without naming bays, and
// each must land in a distinct bay that permits it. Inputs: an unrestricted
// "sw-a" and a restricted-bay "sw-b". Outputs: Bay 1 then Bay 2, parent set,
// rack position and face cleared.
func TestInstallInBayAutoAssigns(t *testing.T) {
	inv, chassis := bayInventory()
	a := addChild(t, inv, chassis.ID, "sw-a", "")
	b := addChild(t, inv, chassis.ID, "sw-b", "")
	if a.ParentBay != "Bay 1" || b.ParentBay != "Bay 2" {
		t.Errorf("bays = %q, %q; want Bay 1, Bay 2", a.ParentBay, b.ParentBay)
	}
	if a.Parent != chassis.ID || a.RackPosition != 0 || a.Face != "" {
		t.Errorf("child placement = parent %s pos %d face %q; want chassis, 0, empty", a.Parent, a.RackPosition, a.Face)
	}
}

// TestInstallInBayErrors verifies the rejections: unknown bay, occupied bay,
// disallowed slug, full chassis, and a parent without bays.
//
// Why it matters: these are the mistakes Nautobot would otherwise reject at
// export time, far from the command that caused them.
func TestInstallInBayErrors(t *testing.T) {
	inv, chassis := bayInventory()
	addChild(t, inv, chassis.ID, "sw-a", "bay 1") // case-insensitive match

	plain := &CaniDeviceType{ID: uuid.New(), Name: "server"}
	inv.Devices[plain.ID] = plain

	cases := []struct {
		name, slug, bay, want string
		parent                uuid.UUID
	}{
		{"unknown bay", "sw-a", "Bay 9", "has no bay", chassis.ID},
		{"occupied", "sw-a", "Bay 1", "occupied by", chassis.ID},
		{"disallowed", "sw-a", "Bay 2", "does not allow", chassis.ID},
		{"full", "sw-a", "", "no free bay", chassis.ID},
		{"no bays", "sw-a", "", "has no device bays", plain.ID},
		{"missing parent", "sw-a", "", "not found", uuid.New()},
	}
	for _, c := range cases {
		child := &CaniDeviceType{ID: uuid.New(), Slug: c.slug}
		err := inv.InstallInBay(child, c.parent, c.bay, nil)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want containing %q", c.name, err, c.want)
		}
	}
}

// TestInstallInBayRespectsPending verifies devices not yet in the inventory
// still reserve their bays.
//
// Why it matters: "add device --qty 2" builds both children before saving, so
// the second must not be assigned the first one's bay. Inputs: a pending child
// in Bay 1. Outputs: the next child gets Bay 2.
func TestInstallInBayRespectsPending(t *testing.T) {
	inv, chassis := bayInventory()
	first := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	if err := inv.InstallInBay(first, chassis.ID, "", nil); err != nil {
		t.Fatal(err)
	}
	pending := map[uuid.UUID]*CaniDeviceType{first.ID: first}
	second := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	if err := inv.InstallInBay(second, chassis.ID, "", pending); err != nil || second.ParentBay != "Bay 2" {
		t.Errorf("second child bay = %q, err = %v; want Bay 2", second.ParentBay, err)
	}
}

func loadedBayOccupant(parentID uuid.UUID, bay string) *CaniDeviceType {
	return &CaniDeviceType{
		ID: uuid.New(), Name: "occupant", Slug: "sw-b", Parent: parentID, ParentBay: bay,
	}
}

func TestInstallInBayRejectsCaseVariantLoadedOccupant(t *testing.T) {
	inv, chassis := bayInventory()
	occupant := loadedBayOccupant(chassis.ID, "bAy 1")
	inv.Devices[occupant.ID] = occupant
	child := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b", RackPosition: 9, Face: "rear"}
	err := inv.InstallInBay(child, chassis.ID, "BAY 1", nil)
	if err == nil || !strings.Contains(err.Error(), "occupied by") {
		t.Fatalf("err=%v, want occupied bay", err)
	}
	if child.Parent != uuid.Nil || child.ParentBay != "" || child.RackPosition != 9 || child.Face != "rear" {
		t.Fatalf("rejected installation changed child placement: %+v", child)
	}
	if occupant.ParentBay != "bAy 1" {
		t.Errorf("existing bay name changed to %q", occupant.ParentBay)
	}
}

func TestInstallInBayAutoSkipsCaseVariantLoadedOccupant(t *testing.T) {
	inv, chassis := bayInventory()
	occupant := loadedBayOccupant(chassis.ID, "bay 1")
	inv.Devices[occupant.ID] = occupant
	child := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	if err := inv.InstallInBay(child, chassis.ID, "", nil); err != nil {
		t.Fatal(err)
	}
	if child.ParentBay != "Bay 2" || occupant.ParentBay != "bay 1" {
		t.Errorf("new bay=%q, existing bay=%q; want Bay 2 and bay 1", child.ParentBay, occupant.ParentBay)
	}
}

func TestInstallInBayRejectsCaseVariantPendingOccupant(t *testing.T) {
	inv, chassis := bayInventory()
	occupant := loadedBayOccupant(chassis.ID, "bAy 1")
	pending := map[uuid.UUID]*CaniDeviceType{occupant.ID: occupant}
	child := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	err := inv.InstallInBay(child, chassis.ID, "Bay 1", pending)
	if err == nil || !strings.Contains(err.Error(), "occupied by") {
		t.Fatalf("err=%v, want pending occupant to reserve bay", err)
	}
}

func TestInstallInBayAutoSkipsCaseVariantPendingOccupant(t *testing.T) {
	inv, chassis := bayInventory()
	occupant := loadedBayOccupant(chassis.ID, "bAy 1")
	pending := map[uuid.UUID]*CaniDeviceType{occupant.ID: occupant}
	child := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	if err := inv.InstallInBay(child, chassis.ID, "", pending); err != nil {
		t.Fatal(err)
	}
	if child.ParentBay != "Bay 2" {
		t.Errorf("new bay=%q, want Bay 2", child.ParentBay)
	}
}

func TestInstallInBayReinstallsCaseVariantOccupant(t *testing.T) {
	inv, chassis := bayInventory()
	child := loadedBayOccupant(chassis.ID, "bay 1")
	inv.Devices[child.ID] = child
	if err := inv.InstallInBay(child, chassis.ID, "BAY 1", nil); err != nil {
		t.Fatal(err)
	}
	if child.ParentBay != "Bay 1" {
		t.Errorf("reinstalled bay=%q, want declared Bay 1", child.ParentBay)
	}
}

func TestInstallInBayAutoReinstallsCaseVariantOccupant(t *testing.T) {
	inv, chassis := bayInventory()
	child := loadedBayOccupant(chassis.ID, "bay 1")
	inv.Devices[child.ID] = child
	if err := inv.InstallInBay(child, chassis.ID, "", nil); err != nil {
		t.Fatal(err)
	}
	if child.ParentBay != "Bay 1" {
		t.Errorf("reinstalled bay=%q, want declared Bay 1", child.ParentBay)
	}
}

func TestInstallInBayOwnAssignmentDoesNotHideCaseVariantConflict(t *testing.T) {
	inv, chassis := bayInventory()
	child := loadedBayOccupant(chassis.ID, "Bay 1")
	inv.Devices[child.ID] = child
	other := loadedBayOccupant(chassis.ID, "bay 1")
	err := inv.InstallInBay(child, chassis.ID, "BAY 1", map[uuid.UUID]*CaniDeviceType{other.ID: other})
	if err == nil || !strings.Contains(err.Error(), "occupied by") {
		t.Fatalf("err=%v, want own assignment not to hide another occupant", err)
	}
}

func TestInstallInBayUnicodeOccupancyMatchesBaySelection(t *testing.T) {
	inv, chassis := bayInventory()
	chassis.DeviceBays[0].Name = "\u03a3lot 1"
	occupant := loadedBayOccupant(chassis.ID, "\u03c2lot 1")
	inv.Devices[occupant.ID] = occupant
	child := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	err := inv.InstallInBay(child, chassis.ID, "\u03c3lot 1", nil)
	if err == nil || !strings.Contains(err.Error(), "occupied by") {
		t.Fatalf("err=%v, want Unicode case-folded occupant to reserve bay", err)
	}
}

func TestInstallInBayCaseVariantOccupancyIsParentScoped(t *testing.T) {
	inv, chassis := bayInventory()
	otherParent := &CaniDeviceType{ID: uuid.New(), Name: "other-chassis", DeviceBays: chassis.DeviceBays}
	inv.Devices[otherParent.ID] = otherParent
	occupant := loadedBayOccupant(otherParent.ID, "bay 1")
	inv.Devices[occupant.ID] = occupant
	child := &CaniDeviceType{ID: uuid.New(), Slug: "sw-b"}
	if err := inv.InstallInBay(child, chassis.ID, "Bay 1", nil); err != nil {
		t.Fatalf("another parent's occupant blocked placement: %v", err)
	}
}

// TestExpandChildrenSetsParentBay verifies default bay children record the
// bay they were created for.
func TestExpandChildrenSetsParentBay(t *testing.T) {
	slug := firstEmbeddedDeviceSlug(t)
	parent := &CaniDeviceType{ID: uuid.New(), DeviceBays: []DeviceBaySpec{
		{Name: "Slot 1", Default: &DeviceBaySlugRef{Slug: slug}},
	}}
	children := ExpandChildren(parent)
	if len(children) != 1 {
		t.Fatalf("ExpandChildren created %d children, want 1", len(children))
	}
	for _, child := range children {
		if child.ParentBay != "Slot 1" {
			t.Errorf("ParentBay = %q, want Slot 1", child.ParentBay)
		}
	}
}

// TestMergePropertiesParentBay verifies a merge takes a non-empty incoming
// bay but keeps the existing bay when the incoming device has none.
//
// Why it matters: Nautobot import supplies bays and must update them, while
// providers that know nothing about bays must not wipe a user's assignment.
func TestMergePropertiesParentBay(t *testing.T) {
	existing := &CaniDeviceType{ParentBay: "Bay 1"}
	if existing.MergeProperties(&CaniDeviceType{}) || existing.ParentBay != "Bay 1" {
		t.Errorf("empty incoming bay changed existing bay to %q", existing.ParentBay)
	}
	if !existing.MergeProperties(&CaniDeviceType{ParentBay: "Bay 2"}) || existing.ParentBay != "Bay 2" {
		t.Errorf("incoming Bay 2 not merged, got %q", existing.ParentBay)
	}
}

// TestInstallInBayUsesLibraryBays verifies a parent instance saved without
// bay specs falls back to its library device type's bays.
//
// Why it matters: inventories created before bays were tracked, or imported
// from providers that don't copy bay specs, still need bay validation.
// Inputs: a parent carrying only the slug of a library type with bays, and
// one with an unknown slug. Outputs: the library's first bay is assigned and
// HasDeviceBays is true; the unknown slug has no bays.
func TestInstallInBayUsesLibraryBays(t *testing.T) {
	slug, bay := libraryTypeWithFreeBay(t)
	inv := NewInventory()
	parent := &CaniDeviceType{ID: uuid.New(), Name: "lib-parent", Slug: slug}
	inv.Devices[parent.ID] = parent
	if !parent.HasDeviceBays() {
		t.Fatalf("HasDeviceBays() = false for library slug %s", slug)
	}
	child := &CaniDeviceType{ID: uuid.New()}
	if err := inv.InstallInBay(child, parent.ID, bay, nil); err != nil || child.ParentBay != bay {
		t.Errorf("InstallInBay into %s %q: bay %q err %v", slug, bay, child.ParentBay, err)
	}
	if (&CaniDeviceType{Slug: "no-such-slug"}).HasDeviceBays() {
		t.Error("HasDeviceBays() = true for an unknown slug")
	}
}

// libraryTypeWithFreeBay returns a registered device type with a bay that
// accepts any slug, and that bay's name.
func libraryTypeWithFreeBay(t *testing.T) (string, string) {
	t.Helper()
	for slug, dt := range allDeviceTypes {
		for _, b := range dt.DeviceBays {
			if b.Allowed == nil || len(b.Allowed.Slugs()) == 0 {
				return slug, b.Name
			}
		}
	}
	t.Skip("no library device type with an unrestricted bay")
	return "", ""
}

// firstEmbeddedDeviceSlug returns any registered device slug without bays.
func firstEmbeddedDeviceSlug(t *testing.T) string {
	t.Helper()
	for slug, dt := range allDeviceTypes {
		if len(dt.DeviceBays) == 0 {
			return slug
		}
	}
	t.Skip("no embedded device types loaded")
	return ""
}
