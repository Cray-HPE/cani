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
package hpcm

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (hpcm *Hpcm) ImportInit(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	// Clone the datastore during import
	tempDatastore, err := datastore.Clone()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to clone datastore"), err)
	}

	// Get the parent system
	sys, err := datastore.GetSystemZero()
	if err != nil {
		return err
	}

	// Get the provider name
	p, err := datastore.InventoryProvider()
	if err != nil {
		return err
	}

	// Set top-level meta to the "system"
	sysMeta := inventory.ProviderMetadataRaw{}
	sys.ProviderMetadata = make(map[inventory.Provider]inventory.ProviderMetadataRaw)
	sys.ProviderMetadata[p] = sysMeta

	// Update the datastore
	err = tempDatastore.Update(&sys)
	if err != nil {
		return err
	}

	// merge in changes
	err = datastore.Merge(tempDatastore)
	if err != nil {
		return err
	}

	for _, node := range hpcm.Nodes {
		log.Info().Msgf("Importing %+v", node.Name)
		hw := &inventory.Hardware{}
		hw.Name = node.Name
		hw.Model = node.Type_
		hw.Type = hardwaretypes.Node
		hw.Architecture = node.Platform.Architecture
		err = tempDatastore.Add(hw)
		if err != nil {
			return err
		}
	}

	err = datastore.Merge(tempDatastore)
	if err != nil {
		return err
	}

	return datastore.Flush()
}

func (hpcm *Hpcm) Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	log.Warn().Msgf("Import not yet implemented")
	return nil
}
