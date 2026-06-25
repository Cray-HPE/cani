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

// sessionStore, when non-nil, holds an in-memory datastore session. While a
// session is active SetDeviceStore returns it instead of constructing a fresh
// disk-backed store, so a batch runner can share one loaded inventory across
// many commands in a single process and defer the disk write until the end.
var sessionStore DeviceStore

// SetDeviceStore selects the datastore implementation for storeType and assigns
// it to the package-level Datastore. Resolving storeType from CLI flags or
// configuration is the command layer's responsibility, which keeps this
// persistence package free of any CLI dependency.
//
// When an in-memory session is active (see BeginSession) the session store is
// used regardless of storeType, so commands re-dispatched inside a batch keep
// operating on the shared inventory rather than reopening the file.
func SetDeviceStore(storeType string) error {
	if sessionStore != nil {
		Datastore = sessionStore
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

// BeginSession starts an in-memory datastore session backed by inv and returns
// the session store. While the session is active every Load returns inv (with
// derived state rebuilt, matching the disk store) and every Save updates it in
// memory without touching disk. The caller persists inv once via a disk store
// after calling EndSession.
func BeginSession(inv *devicetypes.Inventory) *MemStore {
	ms := &MemStore{inv: inv}
	sessionStore = ms
	Datastore = ms
	return ms
}

// EndSession clears the active in-memory session so the next SetDeviceStore
// constructs a disk-backed store again. It is safe to call more than once.
func EndSession() {
	sessionStore = nil
}
