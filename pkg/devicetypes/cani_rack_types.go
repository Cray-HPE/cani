package devicetypes

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Face constants for rack positioning
const (
	FaceFront = "front"
	FaceRear  = "rear"
	FaceFull  = "full" // marks both faces blocked (full-depth device)
)

// SlotOccupancy tracks which device occupies a slot at a given face.
type SlotOccupancy struct {
	DeviceID uuid.UUID `json:"deviceId" yaml:"DeviceID"`
	Face     string    `json:"face" yaml:"Face"` // "front", "rear", or "full"
}

// CaniRackType represents a physical rack, both as hardware-library template
// and inventory instance.  Tracks devices and their U-slot positions.
type CaniRackType struct {
	// Identity
	ID           uuid.UUID `json:"id" yaml:"id,omitempty"`
	Name         string    `json:"name" yaml:"name,omitempty"`
	Slug         string    `json:"slug" yaml:"slug,omitempty"`
	PartNumber   string    `json:"partNumber,omitempty" yaml:"part_number,omitempty"`
	Manufacturer string    `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`
	Model        string    `json:"model,omitempty" yaml:"model,omitempty"`
	Description  string    `json:"description,omitempty" yaml:"description,omitempty"`
	Type         Type      `json:"type,omitempty" yaml:"type,omitempty"`

	// Physical
	UHeight          int             `json:"uHeight" yaml:"u_height,omitempty"`
	OuterWidth       int             `json:"outerWidth,omitempty" yaml:"outer_width,omitempty"`
	OuterDepth       int             `json:"outerDepth,omitempty" yaml:"outer_depth,omitempty"`
	OuterUnit        string          `json:"outerUnit,omitempty" yaml:"outer_unit,omitempty"` // mm or in
	Width            string          `json:"width,omitempty" yaml:"width,omitempty"`          // Nautobot WidthEnum (10/19/21/23 inch)
	Weight           float64         `json:"weight,omitempty" yaml:"weight,omitempty"`
	WeightUnit       string          `json:"weightUnit,omitempty" yaml:"weight_unit,omitempty"`
	DeviceBays       []DeviceBaySpec `json:"deviceBays,omitempty" yaml:"device-bays,omitempty"`
	ModuleBays       []ModuleBaySpec `json:"moduleBays,omitempty" yaml:"module-bays,omitempty"`
	TopZoneHeight    int             `json:"topZoneHeight,omitempty" yaml:"top_zone_height,omitempty"`
	BottomZoneHeight int             `json:"bottomZoneHeight,omitempty" yaml:"bottom_zone_height,omitempty"`

	// Shared metadata (status, role, tags, tenant, custom fields, external IDs, provider metadata)
	ObjectMeta `yaml:",inline"`

	// Inventory state
	Location      uuid.UUID                    `json:"location,omitempty" yaml:"location,omitempty"`
	RackType      string                       `json:"rackType,omitempty" yaml:"rack_type,omitempty"` // Nautobot enum: 2-post-frame, 4-post-cabinet, etc.
	Serial        string                       `json:"serial,omitempty" yaml:"serial,omitempty"`
	AssetTag      string                       `json:"assetTag,omitempty" yaml:"asset_tag,omitempty"`
	FacilityId    string                       `json:"facilityId,omitempty" yaml:"facility_id,omitempty"`
	DescUnits     bool                         `json:"descUnits,omitempty" yaml:"desc_units,omitempty"` // Descending unit numbering
	Comments      string                       `json:"comments,omitempty" yaml:"comments,omitempty"`
	Devices       []uuid.UUID                  `json:"devices,omitempty" yaml:"devices,omitempty"`              // rebuilt from CaniDeviceType.Rack at load time
	OccupiedSlots map[int]map[string]uuid.UUID `json:"occupiedSlots,omitempty" yaml:"occupied_slots,omitempty"` // rebuilt from CaniDeviceType.RackPosition + .Face at load time

	// ProviderDefaults holds provider-specific defaults from the hardware
	// type YAML (e.g. CSM class, starting ordinal, VLAN ranges).
	// Each key is a provider name ("csm"), and the value is a map of
	// provider-specific settings decoded by the provider package.
	ProviderDefaults map[string]any `json:"providerDefaults,omitempty" yaml:"provider_defaults,omitempty"`

	// Source tracks where this type was loaded from (e.g. "builtin", "local:/path", "git:url").
	Source string `json:"-" yaml:"-"`
}

// Validate checks the rack for internal consistency.
func (r *CaniRackType) Validate() error {
	if r == nil {
		return errors.New("cannot validate nil CaniRackType")
	}
	if r.UHeight < 1 {
		return errors.New("rack UHeight must be at least 1")
	}
	return nil
}

// GetID returns the unique identifier.
func (r *CaniRackType) GetID() uuid.UUID {
	if r == nil {
		return uuid.Nil
	}
	return r.ID
}

// GetSlug returns the rack type slug.
func (r *CaniRackType) GetSlug() string {
	if r == nil {
		return ""
	}
	return r.Slug
}

// GetVendor returns the manufacturer name.
func (r *CaniRackType) GetVendor() string {
	if r == nil {
		return ""
	}
	return r.Manufacturer
}

// GetType returns the hardware type as a Type constant.
func (r *CaniRackType) GetType() Type {
	if r == nil {
		return ""
	}
	if r.Type != "" {
		return r.Type
	}
	return TypeRack
}

// GetStatus returns the current status.
func (r *CaniRackType) GetStatus() string {
	if r == nil {
		return ""
	}
	return r.Status
}

// isSlotBlocked returns true if the given U-slot conflicts with a device
// placement on the specified face. Full-depth devices check both faces.
func (r *CaniRackType) isSlotBlocked(u int, face string, isFullDepth bool) bool {
	slot := r.OccupiedSlots[u]
	if slot == nil {
		return false
	}
	if _, hasFullDepth := slot[FaceFull]; hasFullDepth {
		return true
	}
	if isFullDepth {
		return r.hasFaceOccupant(slot, FaceFront) || r.hasFaceOccupant(slot, FaceRear)
	}
	return r.hasFaceOccupant(slot, face)
}

// hasFaceOccupant returns true if the slot has an occupant on the given face.
func (r *CaniRackType) hasFaceOccupant(slot map[string]uuid.UUID, face string) bool {
	_, occupied := slot[face]
	return occupied
}

// CanFitDevice checks if a device can fit at startU with given height and face.
// startU is 1-based (U1 is the bottom of the rack).
// isFullDepth devices block both front and rear faces.
func (r *CaniRackType) CanFitDevice(startU, height int, face string, isFullDepth bool) bool {
	if r == nil || startU < 1 || height < 1 {
		return false
	}
	if face == "" {
		face = FaceFront
	}
	endU := startU + height - 1
	if endU > r.UHeight {
		return false
	}
	for u := startU; u <= endU; u++ {
		if r.isSlotBlocked(u, face, isFullDepth) {
			return false
		}
	}
	return true
}

// PlaceDevice places a device in the rack starting at startU with the given height and face.
// Returns false if the device cannot fit (slots occupied or out of bounds).
// isFullDepth devices occupy both front and rear faces.
func (r *CaniRackType) PlaceDevice(deviceID uuid.UUID, startU, height int, face string, isFullDepth bool) bool {
	if !r.CanFitDevice(startU, height, face, isFullDepth) {
		return false
	}
	if face == "" {
		face = FaceFront
	}
	if r.OccupiedSlots == nil {
		r.OccupiedSlots = make(map[int]map[string]uuid.UUID)
	}

	slotFace := face
	if isFullDepth {
		slotFace = FaceFull // marks both faces blocked
	}

	for u := startU; u <= startU+height-1; u++ {
		if r.OccupiedSlots[u] == nil {
			r.OccupiedSlots[u] = make(map[string]uuid.UUID)
		}
		r.OccupiedSlots[u][slotFace] = deviceID
	}
	r.addDevice(deviceID)
	return true
}

// RemoveDevice removes a device from the rack and frees its U-slots.
func (r *CaniRackType) RemoveDevice(deviceID uuid.UUID) {
	if r == nil {
		return
	}
	// Remove from slot map
	for u, faces := range r.OccupiedSlots {
		for face, id := range faces {
			if id == deviceID {
				delete(r.OccupiedSlots[u], face)
			}
		}
		// Clean up empty U entries
		if len(r.OccupiedSlots[u]) == 0 {
			delete(r.OccupiedSlots, u)
		}
	}
	// Remove from devices list
	for i, id := range r.Devices {
		if id == deviceID {
			r.Devices = append(r.Devices[:i], r.Devices[i+1:]...)
			break
		}
	}
}

// SwapDevices atomically swaps the rack positions of two devices.
// Each device is placed at the other's former startU, preserving each device's
// own height and face. Returns an error if either placement would fail.
func (r *CaniRackType) SwapDevices(idA, idB uuid.UUID) error {
	if r == nil {
		return errors.New("cannot swap on nil rack")
	}
	// Capture current placement for both devices.
	startA := r.GetDeviceStartU(idA)
	heightA := r.GetDeviceHeight(idA)
	faceA := r.GetDeviceFace(idA)

	startB := r.GetDeviceStartU(idB)
	heightB := r.GetDeviceHeight(idB)
	faceB := r.GetDeviceFace(idB)

	if startA == 0 || startB == 0 {
		return errors.New("both devices must be placed in the rack to swap")
	}

	isFullA := faceA == FaceFull
	isFullB := faceB == FaceFull

	// Remove both devices so their slots are free.
	r.RemoveDevice(idA)
	r.RemoveDevice(idB)

	// Place each device at the other's former position.
	if !r.PlaceDevice(idA, startB, heightA, faceA, isFullA) {
		// Rollback: restore both to original positions.
		r.PlaceDevice(idA, startA, heightA, faceA, isFullA)
		r.PlaceDevice(idB, startB, heightB, faceB, isFullB)
		return fmt.Errorf("cannot place device A at U%d after swap", startB)
	}
	if !r.PlaceDevice(idB, startA, heightB, faceB, isFullB) {
		// Rollback: undo A's placement, restore originals.
		r.RemoveDevice(idA)
		r.PlaceDevice(idA, startA, heightA, faceA, isFullA)
		r.PlaceDevice(idB, startB, heightB, faceB, isFullB)
		return fmt.Errorf("cannot place device B at U%d after swap", startA)
	}
	return nil
}

// GetSlotOccupant returns the device UUID occupying the given U-position and face,
// or uuid.Nil if the slot is empty. Also matches full-depth occupants.
func (r *CaniRackType) GetSlotOccupant(u int, face string) uuid.UUID {
	if r == nil {
		return uuid.Nil
	}
	slot := r.OccupiedSlots[u]
	if slot == nil {
		return uuid.Nil
	}
	if face == "" {
		face = FaceFront
	}
	if id, ok := slot[FaceFull]; ok {
		return id
	}
	if id, ok := slot[face]; ok {
		return id
	}
	return uuid.Nil
}

// GetDeviceFace returns the face string stored for a device, or "" if not found.
func (r *CaniRackType) GetDeviceFace(deviceID uuid.UUID) string {
	if r == nil {
		return ""
	}
	for _, faces := range r.OccupiedSlots {
		for face, id := range faces {
			if id == deviceID {
				return face
			}
		}
	}
	return ""
}

// GetDeviceStartU returns the starting U-position for a device, or 0 if not found.
func (r *CaniRackType) GetDeviceStartU(deviceID uuid.UUID) int {
	if r == nil {
		return 0
	}
	minU := 0
	for u, faces := range r.OccupiedSlots {
		for _, id := range faces {
			if id == deviceID {
				if minU == 0 || u < minU {
					minU = u
				}
			}
		}
	}
	return minU
}

// GetDeviceHeight returns the number of U-slots occupied by a device, or 0 if not found.
func (r *CaniRackType) GetDeviceHeight(deviceID uuid.UUID) int {
	if r == nil {
		return 0
	}
	count := 0
	for _, faces := range r.OccupiedSlots {
		for _, id := range faces {
			if id == deviceID {
				count++
			}
		}
	}
	return count
}

// addDevice adds a device UUID to the devices list if not already present.
func (r *CaniRackType) addDevice(deviceID uuid.UUID) {
	for _, id := range r.Devices {
		if id == deviceID {
			return
		}
	}
	r.Devices = append(r.Devices, deviceID)
}

// FindNextAvailableSlot finds the next available starting U-position
// that can fit a device of the given height and face. It scans from the
// top of the rack downward so devices populate top-to-bottom.
// Returns 0 if no space is available.
func (r *CaniRackType) FindNextAvailableSlot(height int, face string, isFullDepth bool) int {
	if r == nil || height < 1 {
		return 0
	}
	for startU := r.UHeight - height + 1; startU >= 1; startU-- {
		if r.CanFitDevice(startU, height, face, isFullDepth) {
			return startU
		}
	}
	return 0
}

// MigrateLegacySlots converts legacy map[int]uuid.UUID to face-aware format.
// Call this when loading old inventory data. Defaults all devices to front face.
func (r *CaniRackType) MigrateLegacySlots(legacy map[int]uuid.UUID) {
	if r.OccupiedSlots == nil {
		r.OccupiedSlots = make(map[int]map[string]uuid.UUID)
	}
	for u, deviceID := range legacy {
		if r.OccupiedSlots[u] == nil {
			r.OccupiedSlots[u] = make(map[string]uuid.UUID)
		}
		r.OccupiedSlots[u][FaceFront] = deviceID
	}
}
