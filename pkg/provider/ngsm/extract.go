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
package ngsm

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// ImportInit imports the external inventory data into CANI's inventory format
func (ngsm *Ngsm) ImportInit(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	var translated map[uuid.UUID]*inventory.Hardware

	// copy the datastore and add set provider metadata
	ds, err := setupTempDatastore(datastore)
	if err != nil {
		return err
	}

	// import from an expert bom
	if cmd.Flags().Changed("bom") {
		translated, err = ngsm.importBoms()
		if err != nil {
			return err
		}
	}

	// add the translated hardware to the datastore
	for _, hw := range translated {
		err := ds.Add(hw)
		if err != nil {
			return err
		}
	}

	// merge the temp datastore with the existing one
	err = datastore.Merge(ds)
	if err != nil {
		return err
	}

	return datastore.Flush()
}

// setupTempDatastore
func setupTempDatastore(datastore inventory.Datastore) (temp inventory.Datastore, err error) {
	temp, err = datastore.Clone()
	if err != nil {
		return temp, errors.Join(fmt.Errorf("failed to clone datastore"), err)
	}

	// Get the parent system
	sys, err := datastore.GetSystemZero()
	if err != nil {
		return temp, err
	}
	// Set additional metadata
	p, err := datastore.InventoryProvider()
	if err != nil {
		return temp, err
	}
	// Set top-level meta to the "system"
	sysMeta := inventory.ProviderMetadataRaw{}
	sys.ProviderMetadata = make(map[inventory.Provider]inventory.ProviderMetadataRaw)
	sys.ProviderMetadata[p] = sysMeta

	// Add it to the datastore
	err = temp.Update(&sys)
	if err != nil {
		return temp, err
	}
	return temp, nil
}
