/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
	"errors"
	"fmt"
	stdlog "log"
	"os"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var ErrHardwareNotFound = errors.New("hardware not found")
var ErrHardwareParentNotFound = errors.New("hardware parent not found")
var ErrHardwareUUIDConflict = errors.New("hardware uuid already exists")
var ErrHardwareMissingLocationOrdinal = errors.New("hardware missing location ordinal")
var ErrEmptyLocationPath = errors.New("empty location path provided")
var ErrDatastoreValidationFailure = errors.New("datastore validation failure")

var log = stdlog.New(os.Stdout, "[datastores] ", stdlog.LstdFlags)

// DeviceStore defines the interface for device storage operations
type DeviceStore interface {
	// Create adds new devices to the store
	Create(devices map[uuid.UUID]*devicetypes.CaniDeviceType) error

	// Read retrieves devices from the store
	// If ids is empty, returns all devices
	Read(ids []uuid.UUID) (map[uuid.UUID]*devicetypes.CaniDeviceType, error)

	// Update updates existing devices in the store
	Update(devices map[uuid.UUID]*devicetypes.CaniDeviceType) error

	// Delete removes devices from the store
	Delete(ids []uuid.UUID) error

	Load() (*devicetypes.Inventory, error)
	Save(inventory *devicetypes.Inventory) error
}

// StoreType defines the type of datastore
type StoreType string

const (
	StoreTypeJSON     StoreType = "json"
	StoreTypePostgres StoreType = "postgres"
)

var Datastore DeviceStore

// SetDeviceStore returns the appropriate datastore implementation
func SetDeviceStore(cmd *cobra.Command, args []string) error {
	storeType := cmd.Root().PersistentFlags().Lookup("datastore").Value.String()
	switch StoreType(storeType) {

	case StoreTypeJSON:
		Datastore = NewJSONStore()
		return nil

	case StoreTypePostgres:
		// Get connection string from config/environment/flags
		connStr := ""
		if cmd.Root().PersistentFlags().Lookup("pg-conn") != nil {
			connStr = cmd.Root().PersistentFlags().Lookup("pg-conn").Value.String()
		}
		Datastore = NewPostgresStore(connStr)
		return nil

	default:
		return fmt.Errorf("unsupported datastore type: %s", storeType)
	}
}
