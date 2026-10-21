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
package export

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

const deviceBayPageSize = 100

type deviceBayTarget struct {
	parent uuid.UUID
	name   string
}

func (e *Exporter) reconcileDeviceBays(ctx context.Context, inventory *devicetypes.Inventory, ids map[string]uuid.UUID, result *LoadResult) error {
	targets, err := desiredDeviceBays(inventory, ids)
	if err != nil {
		return err
	}
	bays, err := e.managedDeviceBays(ctx, targets)
	if err != nil {
		return err
	}
	moves, err := deviceBayMoves(bays, targets)
	if err != nil {
		return err
	}
	if err := e.checkDeviceBayDestinations(ctx, targets, moves); err != nil {
		return err
	}
	for _, bay := range moves {
		if err := e.clearDeviceBay(ctx, bay); err != nil {
			return err
		}
		result.DeviceBaysCleared++
	}
	return nil
}

func desiredDeviceBays(inventory *devicetypes.Inventory, ids map[string]uuid.UUID) (map[uuid.UUID]deviceBayTarget, error) {
	targets := make(map[uuid.UUID]deviceBayTarget)
	for _, device := range inventory.Devices {
		if device == nil || ids[device.Name] == uuid.Nil {
			continue
		}
		target, err := desiredDeviceBay(device, inventory, ids)
		if err != nil {
			return nil, err
		}
		targets[ids[device.Name]] = target
	}
	return targets, nil
}

func desiredDeviceBay(device *devicetypes.CaniDeviceType, inventory *devicetypes.Inventory, ids map[string]uuid.UUID) (deviceBayTarget, error) {
	if device.ParentBay == "" {
		return deviceBayTarget{}, nil
	}
	parent := inventory.Devices[device.Parent]
	if parent == nil || devicetypes.ClassifyForNautobot(string(parent.Type)) != devicetypes.CategoryDevice {
		return deviceBayTarget{}, fmt.Errorf("%s: device bay parent is not an exportable device", device.Name)
	}
	return deviceBayTarget{parent: ids[parent.Name], name: device.ParentBay}, nil
}

func (e *Exporter) managedDeviceBays(ctx context.Context, targets map[uuid.UUID]deviceBayTarget) ([]remoteDeviceBay, error) {
	ids := make([]string, 0, len(targets))
	for id := range targets {
		ids = append(ids, id.String())
	}
	slices.Sort(ids)
	var bays []remoteDeviceBay
	for batch := range slices.Chunk(ids, deviceBayPageSize) {
		page, err := e.listDeviceBays(ctx, nautobotapi.DcimDeviceBaysListParams{InstalledDevice: &batch})
		if err != nil {
			return nil, err
		}
		bays = append(bays, page...)
	}
	return bays, nil
}

func deviceBayMoves(bays []remoteDeviceBay, targets map[uuid.UUID]deviceBayTarget) ([]remoteDeviceBay, error) {
	var moves []remoteDeviceBay
	for _, bay := range bays {
		target, managed := targets[bay.installed]
		if !managed || bay.id == uuid.Nil || bay.parent == uuid.Nil || bay.name == "" {
			return nil, fmt.Errorf("invalid managed device bay response for bay %s", bay.id)
		}
		if target != (deviceBayTarget{parent: bay.parent, name: bay.name}) {
			moves = append(moves, bay)
		}
	}
	slices.SortFunc(moves, func(a, b remoteDeviceBay) int {
		return strings.Compare(a.id.String(), b.id.String())
	})
	return moves, nil
}

func (e *Exporter) checkDeviceBayDestinations(ctx context.Context, targets map[uuid.UUID]deviceBayTarget, moves []remoteDeviceBay) error {
	leaving := make(map[uuid.UUID]uuid.UUID)
	for _, bay := range moves {
		leaving[bay.installed] = bay.id
	}
	for _, bay := range moves {
		target := targets[bay.installed]
		if target.parent == uuid.Nil || target.name == "" {
			continue
		}
		if err := e.checkDeviceBayDestination(ctx, target, bay.installed, leaving); err != nil {
			return err
		}
	}
	return nil
}

func (e *Exporter) checkDeviceBayDestination(ctx context.Context, target deviceBayTarget, childID uuid.UUID, leaving map[uuid.UUID]uuid.UUID) error {
	names, parents := []string{target.name}, []string{target.parent.String()}
	destinations, err := e.listDeviceBays(ctx, nautobotapi.DcimDeviceBaysListParams{Name: &names, Device: &parents})
	if err != nil {
		return err
	}
	for _, destination := range destinations {
		if destination.installed != uuid.Nil && destination.installed != childID && leaving[destination.installed] != destination.id {
			return fmt.Errorf("bay %q of %s already holds device %s, which is not moving", target.name, target.parent, destination.installed)
		}
	}
	return nil
}

func (e *Exporter) listDeviceBays(ctx context.Context, params nautobotapi.DcimDeviceBaysListParams) ([]remoteDeviceBay, error) {
	limit, offset := deviceBayPageSize, 0
	params.Limit, params.Offset = &limit, &offset
	var bays []remoteDeviceBay
	for {
		page, err := e.deviceBayPage(ctx, &params)
		if err != nil {
			return nil, err
		}
		for _, bay := range page.Results {
			bays = append(bays, toRemoteDeviceBay(bay))
		}
		if page.Next == nil || *page.Next == "" {
			return bays, nil
		}
		offset += len(page.Results)
	}
}

func (e *Exporter) deviceBayPage(ctx context.Context, params *nautobotapi.DcimDeviceBaysListParams) (*nautobotapi.PaginatedDeviceBayList, error) {
	resp, err := e.Client.DcimDeviceBaysListWithResponse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("device bay listing: %w", err)
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return nil, fmt.Errorf("device bay listing: status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	if err := validateDeviceBayPage(resp.JSON200); err != nil {
		return nil, fmt.Errorf("device bay listing: %w", err)
	}
	return resp.JSON200, nil
}

func validateDeviceBayPage(page *nautobotapi.PaginatedDeviceBayList) error {
	if len(page.Results) == 0 && page.Next != nil && *page.Next != "" {
		return fmt.Errorf("empty page has a next page")
	}
	if page.Results == nil {
		return fmt.Errorf("missing results array")
	}
	if len(page.Results) == 0 && page.Count > 0 {
		return fmt.Errorf("empty results contradict count %d", page.Count)
	}
	for _, bay := range page.Results {
		if err := validateDeviceBay(bay); err != nil {
			return err
		}
	}
	return nil
}

func (e *Exporter) clearDeviceBay(ctx context.Context, bay remoteDeviceBay) error {
	if e.Options.DryRun {
		clog.DryRun("Would uninstall device %s from bay %q of %s", bay.installed, bay.name, bay.parent)
		return nil
	}
	// The generated request omits nil installed_device values rather than sending null.
	body := bytes.NewReader([]byte(`{"installed_device":null}`))
	resp, err := e.Client.DcimDeviceBaysPartialUpdateWithBodyWithResponse(ctx, bay.id,
		&nautobotapi.DcimDeviceBaysPartialUpdateParams{}, "application/json", body)
	if err != nil {
		return fmt.Errorf("clear device bay %s: %w", bay.id, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("clear device bay %s: status %d: %s", bay.id, resp.StatusCode(), string(resp.Body))
	}
	clog.Changed("Uninstalled device %s from bay %q of %s", bay.installed, bay.name, bay.parent)
	return nil
}
