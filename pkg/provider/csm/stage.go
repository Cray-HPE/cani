package csm

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/csm/transform"
	"github.com/google/uuid"
)

// TemplateResolver looks up a device type template by slug.
// The default implementation is devicetypes.GetBySlug.
type TemplateResolver func(slug string) (devicetypes.CaniDeviceType, bool)

// DeviceApplier applies a device type template to a device.
// The default implementation is devicetypes.ApplyDeviceType.
type DeviceApplier func(device *devicetypes.CaniDeviceType, slug string) error

// DeviceCreator creates a new device from a slug. Returns nil if the slug is unknown.
// The default implementation is devicetypes.NewDeviceFromSlug.
type DeviceCreator func(slug string) (*devicetypes.CaniDeviceType, error)

// StageExisting finds the first Active device (by xname order) whose
// hardware type matches the given slug's type, sets its status to Staged,
// applies the new slug, and recursively stages children according to
// the device-bay defaults in the device type template.
//
// The selection is xname-ordered so that the lowest-numbered slot is
// always chosen first, giving deterministic results across runs.
func StageExisting(
	inv *devicetypes.Inventory,
	slug string,
	resolve TemplateResolver,
	apply DeviceApplier,
	create DeviceCreator,
) bool {
	if inv == nil || slug == "" {
		return false
	}

	tmpl, ok := resolve(slug)
	if !ok {
		return false
	}

	// Collect all Active devices that match by hardware type.
	// Also filter by CSM class so River blades don't match Mountain slugs.
	validClasses := transform.ClassesForSlug(slug)
	type candidate struct {
		id    uuid.UUID
		xname string
	}
	var candidates []candidate
	for id, dev := range inv.Devices {
		if normalizeType(dev.GetType()) != normalizeType(tmpl.GetType()) {
			continue
		}
		if !strings.EqualFold(dev.Status, string(devicetypes.StatusActive)) {
			continue
		}
		if len(validClasses) > 0 {
			devClass := deviceClass(dev)
			if devClass != "" && !validClasses[devClass] {
				continue
			}
		}
		candidates = append(candidates, candidate{id: id, xname: xname(dev)})
	}
	if len(candidates) == 0 {
		return false
	}

	// Pick the device with the lowest xname.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].xname < candidates[j].xname
	})

	dev := inv.Devices[candidates[0].id]
	dev.Status = string(devicetypes.StatusStaged)
	_ = apply(dev, slug)
	stageDefaultChildren(inv, dev, &tmpl, resolve, apply, create)
	return true
}

// stageDefaultChildren walks the device-bay defaults in tmpl and
// stages matching children of parent in the inventory.  For each
// default slug it resolves the child type and stages that many
// children (sorted by xname) whose type matches.  When the template
// specifies more defaults than exist in the inventory, new child
// devices are created with derived xnames.
func stageDefaultChildren(
	inv *devicetypes.Inventory,
	parent *devicetypes.CaniDeviceType,
	tmpl *devicetypes.CaniDeviceType,
	resolve TemplateResolver,
	apply DeviceApplier,
	create DeviceCreator,
) {
	type bayInfo struct {
		slug      string
		childTmpl devicetypes.CaniDeviceType
	}
	countByType := make(map[devicetypes.Type]int)
	templateByType := make(map[devicetypes.Type]bayInfo)
	for _, bay := range tmpl.DeviceBays {
		if bay.Default == nil {
			continue
		}
		slugs := bay.Default.Slugs()
		if len(slugs) == 0 {
			continue
		}
		childTmpl, childOk := resolve(slugs[0])
		if !childOk {
			continue
		}
		ct := normalizeType(childTmpl.GetType())
		countByType[ct]++
		if _, exists := templateByType[ct]; !exists {
			templateByType[ct] = bayInfo{slug: slugs[0], childTmpl: childTmpl}
		}
	}
	if len(countByType) == 0 {
		return
	}

	children := sortedChildDevices(inv, parent)

	childrenByType := make(map[devicetypes.Type][]*devicetypes.CaniDeviceType)
	for _, child := range children {
		childrenByType[normalizeType(child.GetType())] = append(childrenByType[normalizeType(child.GetType())], child)
	}

	parentXn := xname(parent)
	parentClass := deviceClass(parent)

	for ct, count := range countByType {
		matching := childrenByType[ct]
		info := templateByType[ct]
		staged := 0
		for i := 0; i < count && i < len(matching); i++ {
			child := matching[i]
			if !strings.EqualFold(child.Status, string(devicetypes.StatusActive)) {
				continue
			}
			child.Status = string(devicetypes.StatusStaged)
			_ = apply(child, info.slug)
			stageDefaultChildren(inv, child, &info.childTmpl, resolve, apply, create)
			staged++
		}
		// Create missing children when the template has more defaults
		// than already exist in the inventory.
		for i := len(matching); i < count && create != nil; i++ {
			child := createChild(inv, parent, info.slug, parentXn, parentClass, ct, i, create)
			if child == nil {
				continue
			}
			stageDefaultChildren(inv, child, &info.childTmpl, resolve, apply, create)
		}
	}
}

// sortedChildDevices returns children of parent sorted by xname.
func sortedChildDevices(inv *devicetypes.Inventory, parent *devicetypes.CaniDeviceType) []*devicetypes.CaniDeviceType {
	children := make([]*devicetypes.CaniDeviceType, 0, len(parent.Children))
	for _, cid := range parent.Children {
		child, ok := inv.Devices[cid]
		if !ok {
			continue
		}
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		return xname(children[i]) < xname(children[j])
	})
	return children
}

// createChild creates a new child device and adds it to the inventory.
// The child receives an xname derived from the parent's xname plus an
// ordinal suffix appropriate for the child type.
func createChild(
	inv *devicetypes.Inventory,
	parent *devicetypes.CaniDeviceType,
	slug string,
	parentXn string,
	parentClass string,
	childType devicetypes.Type,
	ordinal int,
	create DeviceCreator,
) *devicetypes.CaniDeviceType {
	child, err := create(slug)
	if err != nil || child == nil {
		return nil
	}
	child.Parent = parent.ID
	parent.Children = append(parent.Children, child.ID)
	inv.Devices[child.ID] = child

	if parentXn != "" {
		childXn := parentXn + xnameSuffix(childType, ordinal)
		child.SetProviderMeta("csm", "xname", childXn)
		child.Name = childXn
	}
	if parentClass != "" {
		child.SetProviderMeta("csm", "class", parentClass)
	}
	return child
}

// xnameSuffix returns the xname segment for a child type and ordinal.
func xnameSuffix(t devicetypes.Type, ordinal int) string {
	switch devicetypes.Type(strings.ToLower(string(t))) {
	case devicetypes.TypeNodeCard:
		return fmt.Sprintf("b%d", ordinal)
	case devicetypes.TypeNode:
		return fmt.Sprintf("n%d", ordinal)
	case devicetypes.TypeBlade:
		return fmt.Sprintf("s%d", ordinal)
	case devicetypes.TypeChassis, devicetypes.TypeRack:
		return fmt.Sprintf("c%d", ordinal)
	default:
		return fmt.Sprintf("x%d", ordinal)
	}
}

// xname extracts the CSM xname from a device's provider metadata.
func xname(dev *devicetypes.CaniDeviceType) string {
	sub, ok := dev.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	x, _ := sub["xname"].(string)
	return x
}

// normalizeType returns a lowercased Type for case-insensitive matching.
func normalizeType(t devicetypes.Type) devicetypes.Type {
	return devicetypes.Type(strings.ToLower(string(t)))
}

// deviceClass returns the CSM class (River, Mountain, Hill) from provider metadata.
func deviceClass(dev *devicetypes.CaniDeviceType) string {
	sub, ok := dev.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	c, _ := sub["class"].(string)
	return c
}

// rackXname extracts the CSM xname from a rack's provider metadata.
func rackXname(r *devicetypes.CaniRackType) string {
	sub, ok := r.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	x, _ := sub["xname"].(string)
	return x
}

// rackClass returns the CSM class from a rack's provider metadata.
func rackClass(r *devicetypes.CaniRackType) string {
	sub, ok := r.GetProviderSubMap("csm")
	if !ok {
		return ""
	}
	c, _ := sub["class"].(string)
	return c
}

// StageNewInRack creates a new device hierarchy under the first staged
// rack whose CSM class matches the given slug. It walks the rack's
// device-bay defaults to build a chassis→blade→nodecard→node chain and
// places the requested slug at the first matching level.
// Returns true if a device was created.
func StageNewInRack(
	inv *devicetypes.Inventory,
	slug string,
	resolve TemplateResolver,
	create DeviceCreator,
) bool {
	if inv == nil || slug == "" {
		return false
	}

	tmpl, ok := resolve(slug)
	if !ok {
		return false
	}

	validClasses := transform.ClassesForSlug(slug)

	// Find the first staged rack that matches the slug's class.
	type rackCandidate struct {
		id    uuid.UUID
		xn    string
		class string
	}
	var candidates []rackCandidate
	for id, rack := range inv.Racks {
		if !strings.EqualFold(rack.Status, string(devicetypes.StatusStaged)) {
			continue
		}
		rc := rackClass(rack)
		if rc == "" || (len(validClasses) > 0 && !validClasses[rc]) {
			continue
		}
		candidates = append(candidates, rackCandidate{id: id, xn: rackXname(rack), class: rc})
	}
	if len(candidates) == 0 {
		return false
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].xn < candidates[j].xn
	})
	chosen := candidates[0]

	// Walk the rack's device-bay hierarchy to create the path from
	// rack → chassis → blade, placing the requested slug at the right level.
	rack := inv.Racks[chosen.id]
	return createDeviceInRack(inv, rack, chosen.xn, chosen.class, slug, tmpl.GetType(), resolve, create)
}

// createDeviceInRack walks the rack's device-bay defaults and creates
// the device hierarchy needed to place a device of targetType.
func createDeviceInRack(
	inv *devicetypes.Inventory,
	rack *devicetypes.CaniRackType,
	rackXn string,
	rackClass string,
	slug string,
	targetType devicetypes.Type,
	resolve TemplateResolver,
	create DeviceCreator,
) bool {
	// Walk device bays to find the path from rack to targetType.
	// For a blade: rack → chassis (device-bay default) → blade.
	for i, bay := range rack.DeviceBays {
		if bay.Default == nil {
			continue
		}
		slugs := bay.Default.Slugs()
		if len(slugs) == 0 {
			continue
		}
		bayTmpl, bayOk := resolve(slugs[0])
		if !bayOk {
			continue
		}

		ord := bayOrdinal(bay, i)

		// If the bay's default type matches what we need, create directly.
		if normalizeType(bayTmpl.GetType()) == normalizeType(targetType) {
			return createStagedDevice(inv, uuid.Nil, rackXn, rackClass, slug, ord, create) != nil
		}

		// Otherwise, create this intermediate device and look for
		// matching allowed types or defaults in its bays.
		if !hasMatchingBay(bayTmpl, targetType) {
			continue
		}

		intermediate := createStagedDevice(inv, uuid.Nil, rackXn, rackClass, slugs[0], ord, create)
		if intermediate == nil {
			continue
		}

		intXn := xname(intermediate)
		target := createStagedDevice(inv, intermediate.ID, intXn, rackClass, slug, 0, create)
		if target != nil {
			expandStagedChildren(inv, target, intXn, rackClass)
			return true
		}
	}
	return false
}

// bayOrdinal returns the ordinal from the bay's Extra map, falling
// back to the loop index when the key is absent.
func bayOrdinal(bay devicetypes.DeviceBaySpec, fallback int) int {
	v, ok := bay.Extra["ordinal"]
	if !ok {
		return fallback
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return fallback
}

// hasMatchingBay returns true if tmpl has any device-bay whose allowed
// types include the target type.
func hasMatchingBay(tmpl devicetypes.CaniDeviceType, targetType devicetypes.Type) bool {
	for _, bay := range tmpl.DeviceBays {
		if bay.Allowed == nil {
			continue
		}
		for _, t := range bay.Allowed.AllowedTypes() {
			if strings.EqualFold(t, string(targetType)) {
				return true
			}
		}
	}
	return false
}

// expandStagedChildren expands device-bay children of a staged device
// and assigns them CSM metadata.
func expandStagedChildren(inv *devicetypes.Inventory, dev *devicetypes.CaniDeviceType, parentXn string, class string) {
	for cid, child := range devicetypes.ExpandChildren(dev) {
		inv.Devices[cid] = child
		child.Status = string(devicetypes.StatusStaged)
		if parentXn != "" {
			childXn := xname(dev) + xnameSuffix(child.GetType(), 0)
			child.SetProviderMeta("csm", "xname", childXn)
			child.Name = childXn
		}
		if class != "" {
			child.SetProviderMeta("csm", "class", class)
		}
	}
}

// createStagedDevice creates a new staged device and adds it to the inventory.
func createStagedDevice(
	inv *devicetypes.Inventory,
	parentID uuid.UUID,
	parentXn string,
	class string,
	slug string,
	ordinal int,
	create DeviceCreator,
) *devicetypes.CaniDeviceType {
	dev, err := create(slug)
	if err != nil || dev == nil {
		return nil
	}
	dev.Parent = parentID
	if parentID != uuid.Nil {
		if parent, ok := inv.Devices[parentID]; ok {
			parent.Children = append(parent.Children, dev.ID)
		}
	}
	dev.Status = string(devicetypes.StatusStaged)
	inv.Devices[dev.ID] = dev

	devType := dev.GetType()

	if parentXn != "" {
		devXn := parentXn + xnameSuffix(devType, ordinal)
		dev.SetProviderMeta("csm", "xname", devXn)
		dev.Name = devXn
	}
	if class != "" {
		dev.SetProviderMeta("csm", "class", class)
	}

	log.Printf("%s was successfully staged to be added to the system", displayType(devType))
	log.Printf("UUID: %s", dev.ID)
	if parentXn != "" {
		info := transform.ParseXname(xname(dev))
		log.Printf("Cabinet: %d", info.Cabinet)
		log.Printf("Chassis: %d", info.Chassis)
		log.Printf("Blade: %d", info.Slot)
	}

	return dev
}

// displayType returns a human-friendly name for a device type.
func displayType(t devicetypes.Type) string {
	switch devicetypes.Type(strings.ToLower(string(t))) {
	case devicetypes.TypeCabinet:
		return "Cabinet"
	case devicetypes.TypeChassis, devicetypes.TypeRack:
		return "Chassis"
	case devicetypes.TypeBlade:
		return "Blade"
	case devicetypes.TypeNode:
		return "Node"
	case devicetypes.TypeNodeCard:
		return "NodeBlade"
	default:
		s := string(t)
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
}
