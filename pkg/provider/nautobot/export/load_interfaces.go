/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

const interfaceBatchSize = 100

// bulkInterfaceItem pairs a device ID with the interface spec to create.
type bulkInterfaceItem struct {
	DeviceID   uuid.UUID
	DeviceName string
	Spec       interfaceSpec
}

// loadInterfaces is the Phase 3 entry point.  It collects every interface
// that needs to be created, then sends them to Nautobot in batches via the
// bulk POST endpoint (JSON array to POST /dcim/interfaces/).
// Existing interfaces are handled individually (update when --merge is set).
func (e *Exporter) loadInterfaces(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	result *LoadResult,
) error {
	clog.Info("Gathering interfaces for %d devices — this may take a moment", len(createdDeviceIDs))

	toCreate, err := e.collectNewInterfaces(ctx, inventory, createdDeviceIDs, result)
	if err != nil {
		return err
	}

	clog.Detail("Collected %d interfaces to create across %d devices", len(toCreate), len(createdDeviceIDs))
	if len(toCreate) > 0 {
		clog.Detail("Creating %d interfaces in batches of %d — this may take a while", len(toCreate), interfaceBatchSize)
	}

	if e.Options.DryRun {
		for _, item := range toCreate {
			clog.DryRun("Would create interface: %s on %s", item.Spec.Name, item.DeviceName)
		}
		result.IfacesCreated += len(toCreate)
		return nil
	}

	return e.createInterfacesBulk(ctx, toCreate, result)
}

// collectNewInterfaces iterates every device and its interface specs,
// returning only the interfaces that need to be created.  Existing
// interfaces are updated in-place when --merge is set.
func (e *Exporter) collectNewInterfaces(
	ctx context.Context,
	inventory *devicetypes.Inventory,
	createdDeviceIDs map[string]uuid.UUID,
	result *LoadResult,
) ([]bulkInterfaceItem, error) {
	var toCreate []bulkInterfaceItem

	for deviceName, nautobotID := range createdDeviceIDs {
		var device *devicetypes.CaniDeviceType
		for _, d := range inventory.Devices {
			if d.Name == deviceName {
				device = d
				break
			}
		}
		if device == nil {
			continue
		}

		// Pre-fetch all interfaces for this device in one API call.
		// This avoids per-interface queries whose "name" filter can
		// fail with 400 when interface names contain "/" characters
		// (e.g. "1/1/14" on Aruba switches).
		if err := e.Cache.PrefetchInterfacesForDevice(nautobotID); err != nil {
			clog.Warn("Warning: failed to prefetch interfaces for %s: %v", deviceName, err)
		}

		specs := getDeviceInterfaceSpecs(device)
		for _, spec := range specs {
			existing, err := e.Cache.GetInterfaceByDeviceAndName(nautobotID, spec.Name)
			if err != nil {
				clog.Warn("Warning: failed to lookup interface %s on %s: %v", spec.Name, deviceName, err)
			}

			if existing != nil {
				if e.Options.Merge {
					if err := e.updateInterface(ctx, existing.ID, nautobotID, spec, result); err != nil {
						clog.Warn("Warning: failed to update interface %s on %s: %v", spec.Name, deviceName, err)
					}
				}
				continue
			}

			toCreate = append(toCreate, bulkInterfaceItem{
				DeviceID:   nautobotID,
				DeviceName: deviceName,
				Spec:       spec,
			})
		}
	}

	return toCreate, nil
}

// createInterfacesBulk sends collected interfaces to Nautobot in batches
// of interfaceBatchSize using POST /dcim/interfaces/ with a JSON array body.
func (e *Exporter) createInterfacesBulk(ctx context.Context, items []bulkInterfaceItem, result *LoadResult) error {
	statusItem, err := e.Cache.GetStatus("Active")
	if err != nil {
		return fmt.Errorf("failed to get Active status: %w", err)
	}

	var statusIDUnion nautobotapi.BulkWritableCableRequestStatusId
	if err := statusIDUnion.FromBulkWritableCableRequestStatusId0(statusItem.ID); err != nil {
		return fmt.Errorf("failed to create status ID: %w", err)
	}
	status := nautobotapi.BulkWritableCableRequestStatus{Id: &statusIDUnion}

	var mu sync.Mutex
	var errs []string

	totalBatches := (len(items) + interfaceBatchSize - 1) / interfaceBatchSize
	batchNum := 0

	for start := 0; start < len(items); start += interfaceBatchSize {
		end := start + interfaceBatchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[start:end]
		batchNum++

		clog.Detail("Sending interface batch %d/%d (%d interfaces)", batchNum, totalBatches, len(batch))

		created, batchErr := e.sendInterfaceBatch(ctx, batch, status)
		if batchErr != nil {
			// Batch failed — fall back to individual creates for this batch
			clog.Warn("Bulk batch failed (%d items), falling back to individual creates: %v", len(batch), batchErr)
			for _, item := range batch {
				if err := e.createInterface(ctx, item.DeviceID, item.Spec, result); err != nil {
					// Skip "already exists" errors — interface was created in a prior run.
					// Fetch and cache the existing interface so cable creation can find it.
					if strings.Contains(err.Error(), "must make a unique set") {
						existing, lookupErr := e.Cache.GetInterfaceByDeviceAndName(item.DeviceID, item.Spec.Name)
						if lookupErr == nil && existing != nil {
							e.Cache.CacheInterface(item.DeviceID, item.Spec.Name, existing)
						}
						continue
					}
					mu.Lock()
					errs = append(errs, fmt.Sprintf("%s/%s: %v", item.DeviceName, item.Spec.Name, err))
					mu.Unlock()
				}
			}
			continue
		}

		// Cache results and update count
		e.cacheCreatedInterfaces(batch, created)
		mu.Lock()
		result.IfacesCreated += len(created)
		mu.Unlock()

		clog.Created("Batch %d/%d complete: %d interfaces created", batchNum, totalBatches, len(created))
	}

	if len(errs) > 0 {
		return fmt.Errorf("interface creation errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// sendInterfaceBatch POSTs a JSON array of WritableInterfaceRequest objects
// to /dcim/interfaces/ and returns the created Interface objects.
func (e *Exporter) sendInterfaceBatch(
	ctx context.Context,
	batch []bulkInterfaceItem,
	status nautobotapi.BulkWritableCableRequestStatus,
) ([]nautobotapi.Interface, error) {
	reqs := make([]nautobotapi.WritableInterfaceRequest, 0, len(batch))
	for _, item := range batch {
		var deviceIDUnion nautobotapi.BulkWritableCableRequestStatusId
		if err := deviceIDUnion.FromBulkWritableCableRequestStatusId0(item.DeviceID); err != nil {
			return nil, fmt.Errorf("failed to create device ID for %s: %w", item.DeviceName, err)
		}

		ifaceType := nautobotapi.InterfaceTypeChoices(item.Spec.Type)
		reqs = append(reqs, nautobotapi.WritableInterfaceRequest{
			Device: &nautobotapi.BulkWritableCircuitRequestTenant{Id: &deviceIDUnion},
			Name:   item.Spec.Name,
			Type:   ifaceType,
			Status: status,
		})
	}

	body, err := json.Marshal(reqs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %w", err)
	}

	resp, err := e.Client.DcimInterfacesCreateWithBody(
		ctx,
		&nautobotapi.DcimInterfacesCreateParams{},
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	// Nautobot returns an array of Interface objects for bulk create
	var created []nautobotapi.Interface
	if err := json.Unmarshal(respBody, &created); err != nil {
		// Maybe it returned a single object (non-bulk mode) — try that
		var single nautobotapi.Interface
		if err2 := json.Unmarshal(respBody, &single); err2 == nil && single.Id != nil {
			created = []nautobotapi.Interface{single}
		} else {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return created, nil
}

// cacheCreatedInterfaces stores newly created interfaces in the lookup cache
// so that cable creation (Phase 6) can resolve them.
func (e *Exporter) cacheCreatedInterfaces(batch []bulkInterfaceItem, created []nautobotapi.Interface) {
	for i, iface := range created {
		if iface.Id == nil {
			continue
		}
		// Match by position — Nautobot returns results in request order
		var deviceID uuid.UUID
		var name string
		if i < len(batch) {
			deviceID = batch[i].DeviceID
			name = batch[i].Spec.Name
		} else {
			// Fallback: use response data
			name = iface.Name
		}

		cachedItem := &CachedItem{
			ID:      uuid.UUID(*iface.Id),
			Name:    name,
			Display: name,
		}
		e.Cache.CacheInterface(deviceID, name, cachedItem)
	}
}
