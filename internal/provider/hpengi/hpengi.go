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
	"github.com/Cray-HPE/cani/internal/provider/hpcm"
	"github.com/Cray-HPE/cani/pkg/canu"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	sls "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/cobra"
)

type Hpengi struct {
	hardwareLibrary *hardwaretypes.Library
	Options         *HpengiOpts
	Cid             Cid             `json:"CID"`
	CmConfig        hpcm.HpcmConfig `json:"CmConfig"`
	HpcmCmdb        hpcm.Cmdb       `json:"HpcmCmdb"`
	Paddle          *canu.Paddle    `json:"Paddle"`
	SlsInput        sls.SLSState    `json:"SlsInputFile"`
}

type HpengiOpts struct {
}

// New returns a new Hpengi object, which represents the Hpengi inventory
// When initializing a new sesssion with 'init,' this function is called.
// ValidateInternal is then called to ensure CANI's inventory structure is ok
// ValidateExternal is then called to ensure the provider inventory is ok
// Import is the last to be called and if successful, a session is activated
func New(cmd *cobra.Command, args []string, hardwareLibrary *hardwaretypes.Library, opts interface{}) (hpengi *Hpengi, err error) {
	hpengi = &Hpengi{
		hardwareLibrary: hardwareLibrary,
		Options:         &HpengiOpts{},
	}

	return hpengi, nil
}

func (f Hpengi) Slug() string {
	return "hpengi"
}
