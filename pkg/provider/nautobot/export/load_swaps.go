package export

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// slotKey uniquely identifies a rack slot (rack UUID + U position + face).
type slotKey struct {
	RackID   uuid.UUID
	Position int
	Face     string
}

// pendingMove represents a device being moved to a new rack slot.
type pendingMove struct {
	DeviceID   uuid.UUID // Nautobot device UUID
	DeviceName string
	From       slotKey
	To         slotKey
}

// resolvePositionSwaps detects and resolves circular position swaps so that
// subsequent per-device PATCH calls do not hit Nautobot's unique constraint
// on (rack, position, face).
//
// The problem: Nautobot enforces a unique constraint on (rack, position, face).
// When two devices swap rack positions (e.g. compute-001 moves from U1→U2 and
// compute-002 moves from U2→U1), individual PATCH requests fail because the
// target slot is still occupied by the other device.
//
// The workaround: before running the normal per-device merge updates, this
// function scans for swap cycles.  For each cycle it temporarily clears one
// device's rack placement (rack, position, and face — all three must be
// cleared together per Nautobot validation) via a PATCH.  The freed device
// will be placed at its final position by the normal Phase 2 merge pass.
func (e *Exporter) resolvePositionSwaps(
	ctx context.Context,
	inventory *devicetypes.Inventory,
) error {
	if !e.Options.Merge {
		return nil
	}

	// Collect pending moves: local intent vs. remote current position.
	moves, occupied, err := e.collectPendingMoves(ctx, inventory)
	if err != nil {
		return err
	}
	if len(moves) == 0 {
		return nil
	}

	// Build an index: current slot → deviceID for quick cycle detection.
	// occupied is already this map.

	// For each move, check if the target slot is occupied by another device
	// that is also moving (i.e. a swap cycle).
	//
	// Both devices in the pair must be cleared because Phase 2 iterates a
	// Go map whose order is non-deterministic.  If only one device is
	// cleared, the other may be processed first and still find its target
	// slot occupied.
	cleared := make(map[uuid.UUID]bool)
	for _, m := range moves {
		if cleared[m.DeviceID] {
			continue
		}
		// Is the target slot occupied by someone else?
		blocker, ok := occupied[m.To]
		if !ok || blocker == m.DeviceID {
			continue
		}
		// The blocker is another device sitting in our target slot.
		// Check if the blocker is also moving (completing a cycle).
		blockerMoving := false
		for _, bm := range moves {
			if bm.DeviceID == blocker {
				blockerMoving = true
				break
			}
		}
		if !blockerMoving {
			continue
		}

		// Swap detected — clear both devices' positions to break the cycle.
		for _, id := range []uuid.UUID{m.DeviceID, blocker} {
			if cleared[id] {
				continue
			}
			if e.Options.DryRun {
				clog.DryRun("Would temporarily clear position of device %s to resolve swap", id)
			} else {
				if err := e.clearDevicePosition(ctx, id); err != nil {
					return fmt.Errorf("failed to clear position for swap: %w", err)
				}
			}
			cleared[id] = true
		}
		// Update occupied map so subsequent checks see the cleared slots.
		delete(occupied, m.To)
		delete(occupied, m.From)
	}

	return nil
}

// collectPendingMoves builds the list of devices whose rack position is
// changing during a merge export.  It also returns a map of currently
// occupied slots in Nautobot.
func (e *Exporter) collectPendingMoves(
	ctx context.Context,
	inventory *devicetypes.Inventory,
) ([]pendingMove, map[slotKey]uuid.UUID, error) {
	var moves []pendingMove
	occupied := make(map[slotKey]uuid.UUID)

	for _, device := range inventory.Devices {
		if device == nil || device.Name == "" {
			continue
		}
		category := devicetypes.ClassifyForNautobot(string(device.Type))
		if category != devicetypes.CategoryDevice {
			continue
		}
		nid, ok := device.ExternalIDs["nautobot"]
		if !ok || nid == uuid.Nil {
			continue
		}
		if device.RackPosition <= 0 {
			continue
		}

		// Determine the local desired slot.
		localFace := "front"
		if device.Face == "rear" {
			localFace = "rear"
		}

		rackNautobotID := e.resolveRackNautobotID(device, inventory)
		if rackNautobotID == uuid.Nil {
			continue
		}

		desiredSlot := slotKey{
			RackID:   rackNautobotID,
			Position: device.RackPosition,
			Face:     localFace,
		}

		// Fetch the current remote state.
		remote, err := e.fetchFullDeviceByID(ctx, nid)
		if err != nil {
			continue // skip devices we can't fetch
		}

		currentSlot := remoteSlotKey(remote)
		if currentSlot != nil {
			occupied[*currentSlot] = nid
		}

		// Only record a move if the position actually changed.
		if currentSlot != nil && *currentSlot == desiredSlot {
			continue
		}

		moves = append(moves, pendingMove{
			DeviceID:   nid,
			DeviceName: device.Name,
			From:       derefSlotKey(currentSlot),
			To:         desiredSlot,
		})
	}

	return moves, occupied, nil
}

// resolveRackNautobotID finds the Nautobot UUID for the rack that contains
// the given device.
func (e *Exporter) resolveRackNautobotID(
	device *devicetypes.CaniDeviceType,
	inventory *devicetypes.Inventory,
) uuid.UUID {
	if device.Parent == uuid.Nil {
		return uuid.Nil
	}
	if rack, ok := inventory.Racks[device.Parent]; ok && rack != nil {
		nb, err := e.Cache.GetRackByName(rack.Name)
		if err == nil && nb != nil {
			return nb.ID
		}
	}
	return uuid.Nil
}

// remoteSlotKey extracts the slot key from a Nautobot device response.
func remoteSlotKey(d *nautobotapi.Device) *slotKey {
	if d == nil || d.Position == nil || d.Rack == nil || d.Rack.Id == nil {
		return nil
	}
	rackID, err := d.Rack.Id.AsBulkWritableCableRequestStatusId0()
	if err != nil {
		return nil
	}

	face := "front"
	if d.Face != nil && d.Face.Value != nil && *d.Face.Value == "rear" {
		face = "rear"
	}

	return &slotKey{
		RackID:   rackID,
		Position: *d.Position,
		Face:     face,
	}
}

// derefSlotKey safely dereferences a *slotKey, returning the zero value when nil.
func derefSlotKey(sk *slotKey) slotKey {
	if sk == nil {
		return slotKey{}
	}
	return *sk
}

// clearDevicePosition PATCHes a device to remove its rack placement,
// temporarily freeing the rack slot so another device can move into it.
//
// Nautobot enforces that rack, position, and face are set or cleared together —
// sending position=null alone triggers a validation error ("Cannot select a
// rack face without assigning a rack.").  Therefore we clear all three fields
// in a single PATCH.
//
// The generated PatchedWritableDeviceRequest struct marks Face with
// `json:"face,omitempty"`, which causes a nil pointer to be omitted rather
// than serialized as JSON null.  To work around this we send a raw JSON body
// with explicit null values for all three fields.
func (e *Exporter) clearDevicePosition(ctx context.Context, deviceID uuid.UUID) error {
	body := bytes.NewReader([]byte(`{"rack":null,"position":null,"face":null}`))

	resp, err := e.Client.DcimDevicesPartialUpdateWithBodyWithResponse(
		ctx, deviceID,
		&nautobotapi.DcimDevicesPartialUpdateParams{},
		"application/json",
		body,
	)
	if err != nil {
		return fmt.Errorf("API error clearing position for %s: %w", deviceID, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status %d clearing position for %s: %s",
			resp.StatusCode(), deviceID, string(resp.Body))
	}

	clog.Detail("Temporarily cleared position of device %s to resolve swap", deviceID)
	return nil
}
