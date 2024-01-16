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
	"github.com/spf13/cobra"
)

// ValidateExternal validates the CMDB, a config, or runs a [non]interactive survey
func (hpcm *Hpcm) ValidateExternal(cmd *cobra.Command, args []string) (err error) {
	err = hpcm.siteSurvey(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

// siteSurvey checks the flags passed in during 'session init' and validates
// the appropriate data source.  It adds the validated data to the *Hpcm object
func (hpcm *Hpcm) siteSurvey(cmd *cobra.Command, args []string) error {
	// [non]interactively get the hpcm cluster config file if one is passed in
	if cmd.Flags().Changed("cm-config") {
		f, _ := cmd.Flags().GetString("cm-config")
		cm, err := LoadCmConfig(f)
		if err != nil {
			return err
		}
		hpcm.CmConfig = cm
		// by default, import from the CMDB if no flags were passed
	} else {
		err := hpcm.LoadCmdb(cmd, args)
		if err != nil {
			return err
		}
	}
	return nil
}
