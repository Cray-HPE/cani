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

// Package store adapts CLI flags to the pkg/datastores backend selection. It
// lives in the command layer so that pkg/datastores stays free of any CLI
// dependency: flag parsing happens here, the persistence package only receives
// a resolved store type.
package store

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// datastoreFlag is the persistent root flag that selects the datastore backend.
const datastoreFlag = "datastore"

type validatingStore struct {
	datastores.DeviceStore
}

func (store validatingStore) Save(inventory *devicetypes.Inventory) error {
	if inventory == nil {
		return fmt.Errorf("relationship validation failed: inventory is nil")
	}
	result := inventory.VerifyParentChildRelationships()
	if err := result.Err(); err != nil {
		return fmt.Errorf("relationship validation failed: %w", err)
	}
	return store.DeviceStore.Save(inventory)
}

// Setup resolves the datastore type from the root command's persistent
// "datastore" flag, selects the matching backend, and wraps command saves with
// relationship validation. Concrete stores remain pure read/write so migration
// code can deliberately persist through the raw implementation.
func Setup(cmd *cli.Command) error {
	storeType := cmd.Root().PersistentFlags().Lookup(datastoreFlag).Value.String()
	if err := datastores.SetDeviceStore(storeType); err != nil {
		return err
	}
	datastores.Datastore = validatingStore{DeviceStore: datastores.Datastore}
	return nil
}
