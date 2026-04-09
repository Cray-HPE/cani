package devicetypes

import (
	"testing"
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
