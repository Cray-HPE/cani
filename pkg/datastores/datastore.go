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
)

var ErrHardwareNotFound = errors.New("hardware not found")
var ErrHardwareParentNotFound = errors.New("hardware parent not found")
var ErrHardwareUUIDConflict = errors.New("hardware uuid already exists")
var ErrHardwareMissingLocationOrdinal = errors.New("hardware missing location ordinal")
var ErrEmptyLocationPath = errors.New("empty location path provided")
var ErrDatastoreValidationFailure = errors.New("datastore validation failure")

var log = stdlog.New(os.Stderr, "[datastores] ", stdlog.LstdFlags)

// DeviceStore defines the interface for inventory persistence.
// Domain logic (parent assignment, system creation, etc.) lives on
// devicetypes.Inventory methods; the store is pure read/write.
type DeviceStore interface {
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

// override, when non-nil, is selected by SetDeviceStore in place of a
// storeType-derived backend. It is unexported so no production path can reach
// it; SetDeviceStoreForTest is the only way in.
var override DeviceStore

// TestingTB is the subset of *testing.T that SetDeviceStoreForTest needs.
// Declaring it here keeps this package from importing testing, which would
// register the test flags into the shipped binary.
type TestingTB interface {
	Helper()
	Cleanup(func())
}

// SetDeviceStoreForTest routes SetDeviceStore to store for the duration of t,
// restoring the previous selection when t finishes.
//
// Commands re-select the backend themselves, so a test driving the command
// layer cannot simply assign Datastore. Registering the restore here rather
// than leaving it to the caller means a test cannot leak its in-memory store
// into the next one.
//
// Selection is process-global, so a test calling this must not call t.Parallel.
func SetDeviceStoreForTest(t TestingTB, store DeviceStore) {
	t.Helper()

	previousOverride, previousStore := override, Datastore
	override, Datastore = store, store
	t.Cleanup(func() {
		override, Datastore = previousOverride, previousStore
	})
}

// SetDeviceStore selects the datastore implementation for storeType and assigns
// it to the package-level Datastore. Resolving storeType from CLI flags or
// configuration is the command layer's responsibility, which keeps this
// persistence package free of any CLI dependency.
//
// A store installed by SetDeviceStoreForTest takes precedence and makes
// storeType irrelevant for as long as it is in effect.
func SetDeviceStore(storeType string) error {
	if override != nil {
		Datastore = override
		return nil
	}

	switch StoreType(storeType) {

	case StoreTypeJSON:
		Datastore = NewJSONStore()
		return nil

	// TODO: Implement Postgres datastore
	// case StoreTypePostgres:
	// 	Datastore = NewPostgresStore(connStr)
	// 	return nil

	default:
		return fmt.Errorf("unsupported datastore type: %s", storeType)
	}
}
