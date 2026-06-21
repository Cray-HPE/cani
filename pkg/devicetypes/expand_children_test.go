package devicetypes

import (
	"testing"

	"github.com/google/uuid"
)

func TestExpandChildrenEX235A(t *testing.T) {
	d, err := NewDeviceFromSlug("hpe-crayex-ex235a-compute-blade")
	if err != nil {
		t.Fatalf("NewDeviceFromSlug failed: %v", err)
	}
	if len(d.DeviceBays) == 0 {
		t.Fatal("expected device-bays on blade, got none")
	}
	for i, bay := range d.DeviceBays {
		t.Logf("bay %d: Name=%q Default=%v Extra=%v", i, bay.Name, bay.Default, bay.Extra)
		if bay.Default == nil {
			t.Errorf("bay %d (%s): Default is nil", i, bay.Name)
			continue
		}
		slugs := bay.Default.Slugs()
		if len(slugs) == 0 {
			t.Errorf("bay %d (%s): Default.Slugs() is empty", i, bay.Name)
		}
		for _, s := range slugs {
			if _, ok := GetBySlug(s); !ok {
				t.Errorf("bay %d (%s): default slug %q not found in registry", i, bay.Name, s)
			} else {
				t.Logf("bay %d (%s): default slug %q found in registry", i, bay.Name, s)
			}
		}
	}

	children := ExpandChildren(d)
	if len(children) == 0 {
		t.Fatal("ExpandChildren returned no children")
	}

	// EX235A hierarchy: blade → 2 node cards → each has 1 BMC + 1 node = 6 total
	foundNode := false
	for _, c := range children {
		if c.Slug == "hpe-crayex-ex235a-compute-node" {
			foundNode = true
		}
		if c.Status != string(StatusStaged) {
			t.Errorf("child %s status = %q, want %q", c.Slug, c.Status, StatusStaged)
		}
		if c.Parent == d.ID {
			// direct child
		}
	}
	if !foundNode {
		t.Error("expected at least one hpe-crayex-ex235a-compute-node child")
	}

	t.Logf("Created %d children from blade", len(children))
	for _, c := range children {
		t.Logf("  %s slug=%s type=%s parent=%s", c.ID, c.Slug, c.Type, c.Parent)
	}
}

// TestExpandBaysSkipBranches verifies expandBays ignores bays with no default,
// an empty slug list, or an unresolvable slug, and expands only the valid bay.
//
// Why it matters: device-bay defaults are author-supplied and frequently sparse
// or stale; expansion must tolerate every malformed shape and still wire up the
// bays that do resolve, or whole device trees would fail to materialize.
// Inputs: a parent device with four bays — nil default, empty slug, unknown
// slug, and one referencing a freshly registered child slug. Outputs: exactly
// one accumulated child whose Parent points back at the device. Data choice:
// one bay per skip branch plus a single valid bay isolates each continue path
// and the successful append+recurse in a single call.
func TestExpandBaysSkipBranches(t *testing.T) {
	RegisterDeviceType(CaniDeviceType{Slug: "expand-bay-child", Model: "Child", Manufacturer: "TestCo"})
	t.Cleanup(func() { delete(allDeviceTypes, "expand-bay-child") })

	parent := &CaniDeviceType{
		ID:   uuid.New(),
		Name: "parent",
		DeviceBays: []DeviceBaySpec{
			{Name: "nil-default", Default: nil},
			{Name: "empty-slug", Default: &DeviceBaySlugRef{Slug: nil}},
			{Name: "bad-slug", Default: &DeviceBaySlugRef{Slug: "no-such-slug-zzz-9999"}},
			{Name: "good", Default: &DeviceBaySlugRef{Slug: "expand-bay-child"}},
		},
	}

	acc := make(map[uuid.UUID]*CaniDeviceType)
	expandBays(parent, acc)

	if len(acc) != 1 {
		t.Fatalf("acc size = %d, want 1 (only the valid bay expands)", len(acc))
	}
	if len(parent.Children) != 1 {
		t.Errorf("parent.Children = %d, want 1", len(parent.Children))
	}
	for _, child := range acc {
		if child.Parent != parent.ID {
			t.Errorf("child.Parent = %s, want %s", child.Parent, parent.ID)
		}
	}
}
