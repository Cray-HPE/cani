package datastores

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

// PostgresStore implements DeviceStore interface using PostgreSQL
type PostgresStore struct {
	connStr string
	db      *sql.DB
}

// NewPostgresStore creates a new PostgreSQL-based device store
func NewPostgresStore(connStr string) *PostgresStore {
	return &PostgresStore{
		connStr: connStr,
	}
}

// connect establishes a database connection if not already connected
func (s *PostgresStore) connect() error {
	if s.db != nil {
		// Test if connection is still valid
		err := s.db.Ping()
		if err == nil {
			log.Printf("Reusing existing PostgreSQL connection")
			return nil // Connection is good
		}
		log.Printf("Lost connection to PostgreSQL: %v", err)
		s.db.Close() // Explicitly close the broken connection
	}

	// // Fix connection string
	// if s.connStr == "postgres://admin:adminpass@localhost/cani?sslmode=disable" {
	// 	log.Printf("Fixing default connection string to match Docker Compose configuration")
	// 	s.connStr = "postgres://admin:cani@localhost:5432/cani?sslmode=disable"
	// }

	// Connection retry logic
	maxRetries := 5
	var lastError error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("PostgreSQL connection attempt %d/%d", attempt, maxRetries)

		db, err := sql.Open("pgx", s.connStr)
		if err != nil {
			log.Printf("Failed to open DB connection: %v", err)
			lastError = err
			time.Sleep(2 * time.Second)
			continue
		}

		// Set connection parameters before ping
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Verify connection with ping
		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping PostgreSQL: %v", err)
			db.Close() // Close the failed connection
			lastError = err
			time.Sleep(2 * time.Second)
			continue
		}

		// Connection successful
		log.Printf("PostgreSQL connection successful on attempt %d", attempt)
		s.db = db

		// Create schema
		if err := s.ensureSchema(); err != nil {
			log.Printf("Schema creation failed: %v", err)
			return err
		}

		return nil // Successfully connected and schema created
	}

	return fmt.Errorf("failed to connect to PostgreSQL after %d attempts: %v", maxRetries, lastError)
}

// ensureSchema creates the necessary tables if they don't exist
func (s *PostgresStore) ensureSchema() error {
	log.Printf("Creating PostgreSQL schema if needed...")

	// First check if old schema exists and needs migration
	var tableExists bool
	err := s.db.QueryRow(`
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'devices'
        )
    `).Scan(&tableExists)

	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	// Create new schema with individual columns
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS devices (
        id UUID PRIMARY KEY,
        name TEXT,
        type TEXT,
        device_type_slug TEXT,
        vendor TEXT,
        architecture TEXT,
        model TEXT,
        status TEXT DEFAULT 'Staged',
        properties JSONB,
        provider_metadata JSONB,
        parent UUID,
        children JSONB,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

    CREATE INDEX IF NOT EXISTS devices_type_idx ON devices (type);
    CREATE INDEX IF NOT EXISTS devices_name_idx ON devices (name);
    CREATE INDEX IF NOT EXISTS devices_parent_idx ON devices (parent);`

	log.Println("Ensuring PostgreSQL schema exists")
	_, err = s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Printf("Schema creation completed successfully")
	return nil
}

func (s *PostgresStore) Load() (*devicetypes.Inventory, error) {
	if err := s.connect(); err != nil {
		return nil, err
	}

	inventory := &devicetypes.Inventory{
		Devices: make(map[uuid.UUID]*devicetypes.CaniDeviceType),
	}

	query := `
    SELECT 
        id, name, type, device_type_slug, vendor, architecture,
        model, status, properties, provider_metadata, parent, children
    FROM devices`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	// Scan results into inventory
	for rows.Next() {
		var (
			id             uuid.UUID
			name           sql.NullString
			deviceType     sql.NullString
			deviceTypeSlug sql.NullString
			vendor         sql.NullString
			architecture   sql.NullString
			model          sql.NullString
			status         sql.NullString
			propertiesJSON []byte
			metadataJSON   []byte
			parent         uuid.NullUUID
			childrenJSON   []byte
		)

		if err := rows.Scan(
			&id, &name, &deviceType, &deviceTypeSlug, &vendor, &architecture,
			&model, &status, &propertiesJSON, &metadataJSON, &parent, &childrenJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}

		// Create device and populate fields
		device := devicetypes.CaniDeviceType{
			ID: id,
		}

		if name.Valid {
			device.Name = name.String
		}
		if deviceType.Valid {
			device.Type = devicetypes.Type(deviceType.String)
		}
		if deviceTypeSlug.Valid {
			device.DeviceTypeSlug = deviceTypeSlug.String
		}
		if vendor.Valid {
			device.Vendor = vendor.String
		}
		if architecture.Valid {
			device.Architecture = architecture.String
		}
		if model.Valid {
			device.Model = model.String
		}
		if status.Valid {
			device.Status = status.String
		}
		if parent.Valid {
			device.Parent = parent.UUID
		}

		// Unmarshal JSON fields
		if len(propertiesJSON) > 0 {
			var props map[string]any
			if err := json.Unmarshal(propertiesJSON, &props); err != nil {
				return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
			}
			device.Properties = props
		}

		if len(metadataJSON) > 0 {
			var metadata map[string]any
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			device.ProviderMetadata = metadata
		}

		if len(childrenJSON) > 0 {
			var children []uuid.UUID
			if err := json.Unmarshal(childrenJSON, &children); err != nil {
				return nil, fmt.Errorf("failed to unmarshal children: %w", err)
			}
			device.Children = children
		}

		inventory.Devices[id] = &device
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	// If no devices found, create a default system
	if len(inventory.Devices) == 0 {
		log.Printf("No devices found in database, creating default system")
		systemID := uuid.New()
		system := devicetypes.CaniDeviceType{
			ID:   systemID,
			Name: "SystemZero",
			Type: "system",
		}

		// Insert directly with individual columns
		_, err = s.db.Exec(
			"INSERT INTO devices (id, name, type) VALUES ($1, $2, $3)",
			systemID, system.Name, system.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert default system: %w", err)
		}

		inventory.Devices[systemID] = &system
		log.Printf("Default system created with ID %s", systemID)
	}

	log.Printf("Loaded inventory from %s", "PostgreSQL")
	return inventory, nil
}

func (s *PostgresStore) Save(inventory *devicetypes.Inventory) error {
	log.Printf("Saving inventory with %d devices", len(inventory.Devices))

	if err := s.connect(); err != nil {
		return err
	}

	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	var txErr error
	defer func() {
		if txErr != nil {
			log.Printf("Rolling back due to error: %v", txErr)
			tx.Rollback()
		}
	}()

	// Delete all existing devices
	_, txErr = tx.Exec(`DELETE FROM devices`)
	if txErr != nil {
		return fmt.Errorf("failed to clear devices table: %w", txErr)
	}

	// Prepare statement for insert
	stmt, txErr := tx.Prepare(`
    INSERT INTO devices (
        id, name, type, device_type_slug, vendor, architecture,
        model, status, properties, provider_metadata, parent, children
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
    )`)
	if txErr != nil {
		return fmt.Errorf("failed to prepare statement: %w", txErr)
	}
	defer stmt.Close()

	deviceCount := 0
	for id, device := range inventory.Devices {
		// Marshal the map fields to JSON
		propertiesJSON, err := json.Marshal(device.Properties)
		if err != nil {
			txErr = fmt.Errorf("failed to marshal properties: %w", err)
			return txErr
		}

		providerMetadataJSON, err := json.Marshal(device.ProviderMetadata)
		if err != nil {
			txErr = fmt.Errorf("failed to marshal provider metadata: %w", err)
			return txErr
		}

		childrenJSON, err := json.Marshal(device.Children)
		if err != nil {
			txErr = fmt.Errorf("failed to marshal children: %w", err)
			return txErr
		}

		// Insert the device with individual columns
		_, err = stmt.Exec(
			id,
			device.Name,
			device.Type,
			device.DeviceTypeSlug,
			device.Vendor,
			device.Architecture,
			device.Model,
			device.Status,
			propertiesJSON,
			providerMetadataJSON,
			device.Parent,
			childrenJSON,
		)
		if err != nil {
			txErr = fmt.Errorf("failed to insert device %s: %w", id, err)
			return txErr
		}
		deviceCount++
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", txErr)
	}

	log.Printf("Successfully saved %d devices to PostgreSQL", deviceCount)
	return nil
}

// Create adds new devices to the store
func (s *PostgresStore) Create(devices map[uuid.UUID]*devicetypes.CaniDeviceType) error {
	// Load existing inventory
	inventory, err := s.Load()
	if err != nil {
		return err
	}

	// Ensure devices have parents assigned
	shouldPrompt := false
	for _, device := range devices {
		if device.Type != "system" && device.Parent == uuid.Nil {
			shouldPrompt = true
			break
		}
	}

	if shouldPrompt {
		// This is where we'd implement parent selection logic, similar to the JSON store
		// For now, let's assign to the first system we find
		systems := inventory.Systems()
		if len(systems) > 0 {
			selectedSystem := systems[0]
			for _, device := range devices {
				if device.Type != "system" && device.Parent == uuid.Nil {
					device.Parent = selectedSystem.ID
					log.Printf("Set parent of %s to system %s", device.Name, selectedSystem.Name)
				}
			}
		}
	}

	// Add new devices
	for id, device := range devices {
		if _, exists := inventory.Devices[id]; exists {
			return fmt.Errorf("device with ID %s already exists", id)
		}
		inventory.Devices[id] = device
	}

	// Update parent-child relationships
	inventory.VerifyParentChildRelationships()

	// Save updated inventory
	return s.Save(inventory)
}

// Read retrieves devices from the store
func (s *PostgresStore) Read(ids []uuid.UUID) (map[uuid.UUID]*devicetypes.CaniDeviceType, error) {
	if err := s.connect(); err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*devicetypes.CaniDeviceType)

	// If no IDs provided, return all devices
	if len(ids) == 0 {
		inventory, err := s.Load()
		if err != nil {
			return nil, err
		}
		return inventory.Devices, nil
	}

	// Build a query for the specific devices
	query := `SELECT id, device_data FROM devices WHERE id = ANY($1)`

	// Convert UUID slice to string slice for the query
	idArgs := make([]interface{}, len(ids))
	for i, id := range ids {
		idArgs[i] = id
	}

	rows, err := s.db.Query(query, idArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var deviceJSON []byte

		if err := rows.Scan(&id, &deviceJSON); err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}

		var device devicetypes.CaniDeviceType
		if err := json.Unmarshal(deviceJSON, &device); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device data: %w", err)
		}

		result[id] = &device
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return result, nil
}

// Update updates existing devices in the store
func (s *PostgresStore) Update(devices map[uuid.UUID]*devicetypes.CaniDeviceType) error {
	// Load existing inventory
	inventory, err := s.Load()
	if err != nil {
		return err
	}

	// Update devices
	for id, device := range devices {
		if _, exists := inventory.Devices[id]; !exists {
			return fmt.Errorf("device with ID %s does not exist", id)
		}
		inventory.Devices[id] = device
	}

	// Update parent-child relationships
	inventory.VerifyParentChildRelationships()

	// Save updated inventory
	return s.Save(inventory)
}

// Delete removes devices from the store
func (s *PostgresStore) Delete(ids []uuid.UUID) error {
	// Load existing inventory
	inventory, err := s.Load()
	if err != nil {
		return err
	}

	// Check if deleting systems with children
	systemsWithChildren := make(map[uuid.UUID][]string)
	for _, id := range ids {
		if device, exists := inventory.Devices[id]; exists && device.Type == "system" {
			// Find children of this system
			childNames := []string{}
			for _, child := range inventory.Devices {
				if child.Parent == id {
					childNames = append(childNames, child.Name)
				}
			}

			if len(childNames) > 0 {
				systemsWithChildren[id] = childNames
			}
		}
	}

	// Warn about orphaned children if there are any
	// For a real implementation, add prompting here similar to JSON store

	// Delete devices
	for _, id := range ids {
		delete(inventory.Devices, id)
	}

	// Ensure we still have at least one system
	if len(inventory.Systems()) == 0 {
		log.Printf("Adding a system since all were deleted")
		system := devicetypes.CaniDeviceType{
			ID:   uuid.New(),
			Name: "SystemZero",
			Type: "system",
		}
		inventory.Devices[system.ID] = &system
	}

	// Update parent-child relationships
	inventory.VerifyParentChildRelationships()

	// Save updated inventory
	return s.Save(inventory)
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
