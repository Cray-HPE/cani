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
package csm

import (
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/spf13/cobra"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type CSM struct {
	// Clients
	slsClient       *sls_client.APIClient
	hsmClient       *hsm_client.APIClient
	hardwareLibrary *hardwaretypes.Library
	TBV             *validate.ToBeValidated
	Options         *CsmOpts
}

func New(cmd *cobra.Command, args []string, hardwareLibrary *hardwaretypes.Library) (csm *CSM, err error) {
	csm = &CSM{
		hardwareLibrary: hardwareLibrary,
		// slsClient:       &sls_client.APIClient{},
		// hsmClient:       &hsm_client.APIClient{},
		// TBV:             &validate.ToBeValidated{},
		Options: &CsmOpts{},
	}
	return csm, nil
}
