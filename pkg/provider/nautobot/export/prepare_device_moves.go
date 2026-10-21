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
	"context"
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func (e *Exporter) prepareDeviceMoves(ctx context.Context, inventory *devicetypes.Inventory, result *LoadResult) error {
	if !e.Options.Merge {
		return nil
	}
	ids, err := e.existingBayDeviceIDs(ctx, inventory)
	if err != nil {
		return fmt.Errorf("device bay reconciliation: %w", err)
	}
	if err := e.reconcileDeviceBays(ctx, inventory, ids, result); err != nil {
		return fmt.Errorf("device bay reconciliation: %w", err)
	}
	if err := e.resolvePositionSwaps(ctx, inventory); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("position swap resolution: %v", err))
	}
	return nil
}

func (e *Exporter) existingBayDeviceIDs(ctx context.Context, inventory *devicetypes.Inventory) (map[string]uuid.UUID, error) {
	ids := make(map[string]uuid.UUID)
	for _, device := range inventory.Devices {
		if device == nil || device.Name == "" || devicetypes.ClassifyForNautobot(string(device.Type)) != devicetypes.CategoryDevice {
			continue
		}
		id, err := e.existingBayDeviceID(ctx, device)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", device.Name, err)
		}
		if id != uuid.Nil {
			ids[device.Name] = id
		}
	}
	return ids, nil
}

func (e *Exporter) existingBayDeviceID(ctx context.Context, device *devicetypes.CaniDeviceType) (uuid.UUID, error) {
	if id := device.ExternalIDs[externalIDKeyNautobot]; id != uuid.Nil {
		exists, err := e.deviceExistsByID(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		if exists {
			return id, nil
		}
	}
	existing, err := e.Cache.GetDeviceByName(device.Name)
	if err != nil {
		return uuid.Nil, err
	}
	if existing == nil {
		return uuid.Nil, nil
	}
	return existing.ID, nil
}
