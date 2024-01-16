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
package hpcm

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (hpcm *Hpcm) BuildHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string, recommendations provider.HardwareRecommendations) error {
	log.Warn().Msgf("BuildHardwareMetadata not yet implemented")
	return nil
}

func (hpcm *Hpcm) NewHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string) error {
	log.Warn().Msgf("NewHardwareMetadata not yet implemented")
	return nil
}

func (hpcm *Hpcm) GetFieldMetadata() ([]provider.FieldMetadata, error) {
	log.Warn().Msgf("GetFieldMetadata not yet implemented")
	return []provider.FieldMetadata{}, nil
}

func (hpcm *Hpcm) GetFields(hw *inventory.Hardware, fieldNames []string) (values []string, err error) {
	log.Warn().Msgf("GetFields not yet implemented")
	return []string{}, nil
}
