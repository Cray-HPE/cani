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
package datastores

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// JSONStore handles inventory persistence to a JSON file.
type JSONStore struct {
	Path string
}

// NewJSONStore creates a new store with the config path.
// If the configured datastore path is absolute it is used as-is;
// otherwise it is resolved relative to the config file directory.
func NewJSONStore() *JSONStore {
	ds := config.Cfg.Datastore
	if !filepath.IsAbs(ds) {
		ds = filepath.Join(filepath.Dir(config.Cfg.Path), filepath.Base(ds))
	}
	return &JSONStore{Path: ds}
}

// Load reads the inventory from disk.
// Returns an empty inventory when the file does not exist yet.
// If a v1alpha1 (legacy) datastore is detected, it is automatically
// migrated to v1alpha2 and the original is backed up to .canisave.
func (s *JSONStore) Load() (*devicetypes.Inventory, error) {
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return devicetypes.NewInventory(), nil
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, fmt.Errorf("reading inventory file: %w", err)
	}

	// Detect and migrate legacy (v1alpha1) datastores.
	if isLegacyDatastore(data) {
		if err := backupDatastore(s.Path); err != nil {
			return nil, fmt.Errorf("backing up legacy datastore: %w", err)
		}

		inventory, err := migrateV1Alpha1(data)
		if err != nil {
			return nil, fmt.Errorf("migrating v1alpha1 datastore: %w", err)
		}

		if err := s.Save(inventory); err != nil {
			return nil, fmt.Errorf("saving migrated datastore: %w", err)
		}

		log.Printf("Migrated datastore from v1alpha1 to v1alpha2; backup at %s.canisave", s.Path)
		return inventory, nil
	}

	inventory := devicetypes.NewInventory()
	if err := json.Unmarshal(data, inventory); err != nil {
		return nil, fmt.Errorf("parsing inventory: %w", err)
	}

	// Default schema version for files written before version tracking.
	if inventory.SchemaVersion == "" {
		inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha2
	}

	// Migrate legacy inventory-level providerMetadata to the typed
	// Metadata field if the old JSON key is present.
	if migrateInventoryMetadata(data, inventory) {
		if err := s.Save(inventory); err != nil {
			return nil, fmt.Errorf("saving metadata-migrated datastore: %w", err)
		}
		log.Printf("Migrated inventory-level providerMetadata to metadata")
	}

	inventory.RebuildProviderKeyIndex()

	return inventory, nil
}

// Save writes the inventory to disk, creating directories as needed.
//
// The write is atomic: the inventory is marshalled to a temporary file in the
// destination directory, flushed to stable storage, and renamed into place.
// A crash or power loss mid-write therefore leaves the previous inventory
// intact rather than a partially written, corrupt file.
func (s *JSONStore) Save(inventory *devicetypes.Inventory) error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating inventory directory: %w", err)
	}

	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding inventory: %w", err)
	}

	return s.writeAtomic(dir, data)
}

// writeAtomic writes data to a temporary file in dir and renames it onto
// s.Path. The temporary file is removed on any failure before the rename so
// no partial files are left behind. Placing the temporary file in the same
// directory as the destination keeps the rename on a single filesystem, which
// is what makes it atomic.
func (s *JSONStore) writeAtomic(dir string, data []byte) error {
	tmp, err := os.CreateTemp(dir, ".inventory-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary inventory file: %w", err)
	}
	tmpPath := tmp.Name()
	// Best-effort cleanup; a successful rename removes tmpPath first, making
	// this a no-op in the happy path.
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing inventory file: %w", err)
	}

	// Flush file contents to disk before the rename so the rename cannot
	// expose a truncated file after a crash.
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing inventory file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing inventory file: %w", err)
	}

	if err := os.Rename(tmpPath, s.Path); err != nil {
		return fmt.Errorf("replacing inventory file: %w", err)
	}

	return nil
}
