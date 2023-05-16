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
	"errors"
	"net/http"
	"os"

	"internal/hsm"
	"internal/sls"

	"github.com/Cray-HPE/cani/internal/cani/domain"
	hardware_type_library "github.com/Cray-HPE/cani/pkg/hardware-type-library"
	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
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

// validHardware checks that the hardware type is valid by comparing it against the list of hardware types
func validHardware(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("list-supported-types") {
		domain.Data.ListSupportedTypes(hardware_type_library.HardwareTypeNodeBlade)
		os.Exit(0)
	}

	if len(args) == 0 {
		return errors.New("No hardware type provided")
	}

	library, err := hardware_type_library.NewEmbeddedLibrary()
	if err != nil {
		return err
	}

	// Get the list of hardware types that are blades
	deviceTypes := library.GetDeviceTypesByHardwareType(hardware_type_library.HardwareTypeNodeBlade)

	// Check that each arg is a valid blade xname
	for _, arg := range args {
		matchFound := false
		for _, device := range deviceTypes {
			if arg == device.Slug {
				matchFound = true
				break
			}
		}
		if !matchFound {
			return errors.New("Invalid hardware type: " + arg)
		}
	}

	return nil
}
