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
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	translated = make(map[uuid.UUID]*inventory.Hardware, 0)
)

// ValidateExternal validates the external inventory
// for NGSM, this could be an HBOM, a customer intent document, an existing
// CMDB instance, an hpcm.config file, an SLS dumpstate, or a paddle file.
func (ngsm *Ngsm) ValidateExternal(cmd *cobra.Command, args []string) error {
	err := ngsm.siteSurvey(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

// siteSurvey will run interactively or non-interactively to gather the
// necessary information. ngsm may be a brand new system but could also be a
// migration from a different provider
func (ngsm *Ngsm) siteSurvey(cmd *cobra.Command, args []string) error {
	// validate the hbom if one or more were provided
	if cmd.Flags().Changed("bom") {
		err := ngsm.loadBoms(cmd, args)
		if err != nil {
			return err
		}
	}

	return nil
}
