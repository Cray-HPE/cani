/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package blade

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"internal/hsm"
	"internal/sls"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	X "github.com/Cray-HPE/cani/pkg/xname"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	HSM        *hsm_client.APIClient
	SLS        *sls_client.APIClient
	simulation bool
	token      = os.ExpandEnv("${TOKEN}")
	apiGw      = "https://api-gw-service-nmn.local/apis"
	tr         = &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client = &http.Client{Transport: tr}
)

func EnableSimulation() {
	HSM = hsm.EnableSimulation()
	SLS = sls.EnableSimulation()
}

func DisableSimulation() {
	HSM = hsm.DisableSimulation()
	SLS = sls.DisableSimulation()
}

func EnableDebug() {
	debug = true
}

// validateArgsAreValidBladeXnames validates that the arguments passed to the command are valid blade/slot/ComputeModule xnames.
func validateArgsAreValidBladeXnames(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("At least one blade xname must be specified, received %d", len(args))
	}

	// Check that each arg is a valid blade xname
	for _, arg := range args {
		xn := X.Xname(arg)
		if xn.IsValid() {
			blade := xn.Type()
			if blade.Type != "ComputeModule" {
				return fmt.Errorf("Invalid blade xname %s: %s", blade.Letters, arg)
			}
		} else {
			return fmt.Errorf("Not an xname: %s", arg)
		}
	}

	return nil
}
