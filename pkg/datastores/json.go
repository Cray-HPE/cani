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
// Legacy (v1alpha1) datastores are migrated to the current schema and backed up
// to .canisave; v1alpha2 datastores are migrated to v1alpha3 the same way.
// Derived reverse indices and FK fields are rebuilt from the authoritative
// forward FKs on every load, so persisted derived values are never trusted.
func (s *JSONStore) Load() (*devicetypes.Inventory, error) {
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return devicetypes.NewInventory(), nil
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, fmt.Errorf("reading inventory file: %w", err)
	}

	if isLegacyDatastore(data) {
		return s.loadLegacy(data)
	}
	return s.loadCurrent(data)
}

// loadLegacy migrates a v1alpha1 datastore through to the current schema,
// rebuilds derived state, and persists the result after backing up the original.
func (s *JSONStore) loadLegacy(data []byte) (*devicetypes.Inventory, error) {
	if err := backupDatastore(s.Path); err != nil {
		return nil, fmt.Errorf("backing up legacy datastore: %w", err)
	}

	inventory, err := migrateV1Alpha1(data)
	if err != nil {
		return nil, fmt.Errorf("migrating v1alpha1 datastore: %w", err)
	}

	// migrateV1Alpha1 sets Parent on every device, so the v1alpha2->v1alpha3
	// back-fill is a no-op here; it only advances the schema version.
	migrateV1Alpha2(data, inventory)
	inventory.RebuildDerivedState()

	if err := s.Save(inventory); err != nil {
		return nil, fmt.Errorf("saving migrated datastore: %w", err)
	}

	log.Printf("Migrated datastore from v1alpha1 to %s; backup at %s.canisave",
		inventory.SchemaVersion, s.Path)
	return inventory, nil
}

// loadCurrent parses a v1alpha2-or-newer datastore, applies the metadata and
// relationship migrations as needed, rebuilds derived state, and persists when a
// migration occurred.
func (s *JSONStore) loadCurrent(data []byte) (*devicetypes.Inventory, error) {
	inventory := devicetypes.NewInventory()
	if err := json.Unmarshal(data, inventory); err != nil {
		return nil, fmt.Errorf("parsing inventory: %w", err)
	}

	// Default schema version for files written before version tracking.
	if inventory.SchemaVersion == "" {
		inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha2
	}

	metaMigrated := migrateInventoryMetadata(data, inventory)
	relMigrated := inventory.SchemaVersion == devicetypes.SchemaVersionV1Alpha2
	if relMigrated {
		if err := backupDatastore(s.Path); err != nil {
			return nil, fmt.Errorf("backing up datastore: %w", err)
		}
		migrateV1Alpha2(data, inventory)
	}

	inventory.RebuildProviderKeyIndex()
	inventory.RebuildDerivedState()

	if metaMigrated || relMigrated {
		if err := s.Save(inventory); err != nil {
			return nil, fmt.Errorf("saving migrated datastore: %w", err)
		}
		if relMigrated {
			log.Printf("Migrated datastore from v1alpha2 to v1alpha3; backup at %s.canisave", s.Path)
		} else {
			log.Printf("Migrated inventory-level providerMetadata to metadata")
		}
	}

	return inventory, nil
}

// Save writes the inventory to disk, creating directories as needed.
//
// The write is atomic with respect to partial writes: the inventory is
// marshalled to a temporary file in the destination directory and renamed into
// place, so a crash mid-write leaves the previous inventory intact rather than
// a partially written, corrupt file. The temporary file is not fsync'd before
// the rename — the datastore is a local, regenerable cache, and skipping the
// flush keeps each write fast for scripts that add inventory in a tight loop.
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
// is what makes it atomic. The contents are not fsync'd before the rename; see
// Save for the rationale.
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

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing inventory file: %w", err)
	}

	if err := os.Rename(tmpPath, s.Path); err != nil {
		return fmt.Errorf("replacing inventory file: %w", err)
	}

	return nil
}
