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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// ErrUnsupportedSchemaVersion marks an inventory generation that this cani
// release cannot safely read or write without risking data loss.
var ErrUnsupportedSchemaVersion = errors.New("unsupported inventory schema version")

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
// to .canisave; later generations are migrated sequentially the same way.
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

	schemaVersion, err := inventorySchemaVersion(data)
	if err != nil {
		return nil, err
	}

	switch schemaVersion {
	case devicetypes.SchemaVersionV1Alpha1:
		if !isLegacyDatastore(data) {
			return nil, fmt.Errorf("%w %q: the document does not match the legacy inventory shape",
				ErrUnsupportedSchemaVersion, schemaVersion)
		}
		return s.loadLegacy(data)
	case devicetypes.SchemaVersionV1Alpha2, devicetypes.SchemaVersionV1Alpha3,
		devicetypes.SchemaVersionV1Alpha4, devicetypes.SchemaVersionV1Alpha5,
		devicetypes.CurrentSchemaVersion:
		return s.loadCurrent(data, schemaVersion)
	default:
		return nil, fmt.Errorf("%w %q: this cani release supports %q through %q and will not rewrite the file",
			ErrUnsupportedSchemaVersion, schemaVersion,
			devicetypes.SchemaVersionV1Alpha1, devicetypes.CurrentSchemaVersion)
	}
}

// inventorySchemaVersion reads only the version discriminator needed to route
// decoding. Missing versions retain the historical defaults: the legacy
// Hardware shape is v1alpha1 and the current lowercase shape is v1alpha2.
func inventorySchemaVersion(data []byte) (string, error) {
	var probe struct {
		SchemaVersion       string          `json:"schemaVersion"`
		LegacySchemaVersion string          `json:"SchemaVersion"`
		Hardware            json.RawMessage `json:"Hardware"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return "", fmt.Errorf("parsing inventory: %w", err)
	}
	if probe.SchemaVersion != "" {
		return probe.SchemaVersion, nil
	}
	if probe.LegacySchemaVersion != "" {
		return probe.LegacySchemaVersion, nil
	}
	if len(probe.Hardware) > 0 {
		return devicetypes.SchemaVersionV1Alpha1, nil
	}
	return devicetypes.SchemaVersionV1Alpha2, nil
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
	migrateV1Alpha3(inventory)
	migrateV1Alpha4(inventory)
	migrateV1Alpha5(inventory)
	relationships := inventory.RebuildDerivedState()
	if err := relationships.Err(); err != nil {
		log.Printf("Skipped saving migrated datastore with relationship errors: %v", err)
		return inventory, nil
	}

	if err := s.Save(inventory); err != nil {
		return nil, fmt.Errorf("saving migrated datastore: %w", err)
	}

	log.Printf("Migrated datastore from v1alpha1 to %s; backup at %s.canisave",
		inventory.SchemaVersion, s.Path)
	return inventory, nil
}

// loadCurrent parses a v1alpha2-or-newer datastore, applies each generation's
// migration in order, rebuilds derived state, and persists when needed.
func (s *JSONStore) loadCurrent(data []byte, schemaVersion string) (*devicetypes.Inventory, error) {
	inventory := devicetypes.NewInventory()
	if err := json.Unmarshal(data, inventory); err != nil {
		return nil, fmt.Errorf("parsing inventory: %w", err)
	}
	inventory.SchemaVersion = schemaVersion

	originalVersion := inventory.SchemaVersion
	metaMigrated := migrateInventoryMetadata(data, inventory)
	schemaMigrated := inventory.SchemaVersion != devicetypes.CurrentSchemaVersion
	if metaMigrated || schemaMigrated {
		if err := backupDatastore(s.Path); err != nil {
			return nil, fmt.Errorf("backing up datastore: %w", err)
		}
	}
	migrateSchemaToCurrent(data, inventory)

	inventory.RebuildProviderKeyIndex()
	relationships := inventory.RebuildDerivedState()

	if metaMigrated || schemaMigrated {
		if err := relationships.Err(); err != nil {
			log.Printf("Skipped saving migrated datastore with relationship errors: %v", err)
			return inventory, nil
		}
		if err := s.Save(inventory); err != nil {
			return nil, fmt.Errorf("saving migrated datastore: %w", err)
		}
		if schemaMigrated {
			log.Printf("Migrated datastore from %s to %s; backup at %s.canisave",
				originalVersion, inventory.SchemaVersion, s.Path)
		} else {
			log.Printf("Migrated inventory-level providerMetadata to metadata; backup at %s.canisave", s.Path)
		}
	}

	return inventory, nil
}

// migrateSchemaToCurrent applies each schema migration in generation order.
func migrateSchemaToCurrent(data []byte, inventory *devicetypes.Inventory) {
	if inventory.SchemaVersion == devicetypes.SchemaVersionV1Alpha2 {
		migrateV1Alpha2(data, inventory)
	}
	if inventory.SchemaVersion == devicetypes.SchemaVersionV1Alpha3 {
		migrateV1Alpha3(inventory)
	}
	if inventory.SchemaVersion == devicetypes.SchemaVersionV1Alpha4 {
		migrateV1Alpha4(inventory)
	}
	if inventory.SchemaVersion == devicetypes.SchemaVersionV1Alpha5 {
		migrateV1Alpha5(inventory)
	}
}

// Save writes the inventory to disk, creating directories as needed.
//
// The write is atomic: the inventory is marshalled to a temporary file in the
// destination directory, flushed to stable storage, and renamed into place.
// A crash or power loss mid-write therefore leaves the previous inventory
// intact rather than a partially written, corrupt file.
func (s *JSONStore) Save(inventory *devicetypes.Inventory) error {
	if inventory == nil {
		return fmt.Errorf("encoding inventory: inventory is nil")
	}
	if inventory.SchemaVersion != devicetypes.CurrentSchemaVersion {
		return fmt.Errorf("%w %q: this cani release writes only %q",
			ErrUnsupportedSchemaVersion, inventory.SchemaVersion, devicetypes.CurrentSchemaVersion)
	}

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
