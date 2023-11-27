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
package domain

import (
	"errors"
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (d *Domain) UpdateNode(cmd *cobra.Command, args []string, cabinet, chassis, slot, bmc, node int) (AddHardwareResult, error) {
	// Get the node object from the datastore
	locationPath := inventory.LocationPath{
		{HardwareType: hardwaretypes.System, Ordinal: 0},
		{HardwareType: hardwaretypes.Cabinet, Ordinal: cabinet},
		{HardwareType: hardwaretypes.Chassis, Ordinal: chassis},
		{HardwareType: hardwaretypes.NodeBlade, Ordinal: slot},
		{HardwareType: hardwaretypes.NodeCard, Ordinal: bmc}, // Yes I mean put the BMC location for the node card location
		{HardwareType: hardwaretypes.Node, Ordinal: node},
	}
	hw, err := d.datastore.GetAtLocation(locationPath)
	if err != nil {
		return AddHardwareResult{}, err
	}

	log.Debug().Msgf("Found node at: %s with ID (%s)", locationPath, hw.ID)

	// Ask the inventory provider to craft a metadata object for this information
	err = d.externalInventoryProvider.NewHardwareMetadata(&hw, cmd, args)
	if err != nil {
		return AddHardwareResult{}, err
	}

	// Push it back into the data store
	if err := d.datastore.Update(&hw); err != nil {
		return AddHardwareResult{}, err
	}

	// Validate the current state of CANI's inventory data against the provider plugin
	// for provider specific data.
	var result AddHardwareResult
	if failedValidations, err := d.externalInventoryProvider.ValidateInternal(cmd.Context(), d.datastore, false); len(failedValidations) > 0 {
		result.ProviderValidationErrors = failedValidations
		return result, provider.ErrDataValidationFailure
	} else if err != nil {
		return AddHardwareResult{}, errors.Join(
			fmt.Errorf("failed to validate inventory against inventory provider plugin"),
			err,
		)
	}

	return result, d.datastore.Flush()
}
