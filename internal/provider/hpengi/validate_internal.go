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
package hpengi

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ValidateInternal validates the representation of the CANI inventory data into
// the provider's inventory system is consistent. The default set of checks will
// verify all currently provided data is valid. If enableRequiredDataChecks is
// set to true, additional checks focusing on missing data will be ran.
func (hpengi *Hpengi) ValidateInternal(cmd *cobra.Command, args []string, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]provider.HardwareValidationResult, error) {
	log.Warn().Msgf("ValidateInternal partially implemented")
	// Get all hardware
	inv, err := datastore.List()
	if err != nil {
		return nil, err
	}

	// Build up the validation results map
	results := map[uuid.UUID]provider.HardwareValidationResult{}
	for _, hw := range inv.Hardware {
		results[hw.ID] = provider.HardwareValidationResult{
			Hardware: hw,
		}
	}

	// Do some sort of checks

	// otherwise, return an empty object because nothing is wrong
	return map[uuid.UUID]provider.HardwareValidationResult{}, nil
}
