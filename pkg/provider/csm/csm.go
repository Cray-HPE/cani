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
	"fmt"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

const (
	CsmSlug = "csm"
)

type CSM struct {
	// Clients
	slsClient       *sls_client.APIClient
	hsmClient       *hsm_client.APIClient
	hardwareLibrary *hardwaretypes.Library
	TBV             *validate.ToBeValidated
	Options         *CsmOpts
}

// Need to load from existing if not init
func New(cmd *cobra.Command, args []string, hwlib *hardwaretypes.Library, opts interface{}) (csm *CSM, err error) {
	// create a starting object
	csm = &CSM{
		slsClient:       &sls_client.APIClient{},
		hsmClient:       &hsm_client.APIClient{},
		TBV:             &validate.ToBeValidated{},
		hardwareLibrary: hwlib,
		Options:         &CsmOpts{},
	}

	if cmd.Parent().Name() == "init" {
		useSimulation := cmd.Flags().Changed("use-simulator")
		slsUrl, _ := cmd.Flags().GetString("csm-url-sls")
		hsmUrl, _ := cmd.Flags().GetString("csm-url-hsm")
		insecure := cmd.Flags().Changed("insecure")
		providerHost, _ := cmd.Flags().GetString("csm-api-host")
		tokenUsername, _ := cmd.Flags().GetString("csm-keycloak-username")
		tokenPassword, _ := cmd.Flags().GetString("csm-keycloak-password")
		k8sPodsCidr, _ := cmd.Flags().GetString("csm-k8s-pods-cidr")
		k8sServicesCidr, _ := cmd.Flags().GetString("csm-k8s-services-cidr")
		kubeconfig, _ := cmd.Flags().GetString("csm-kube-config")
		caCertPath, _ := cmd.Flags().GetString("csm-ca-cert")
		secretName, _ := cmd.Flags().GetString("csm-secret-name")
		clientId, _ := cmd.Flags().GetString("csm-client-id")
		clientSecret, _ := cmd.Flags().GetString("csm-client-secret")

		if useSimulation {
			log.Warn().Msg("Using simulation mode")
			csm.Options.UseSimulation = useSimulation
			insecure = true
			if !cmd.Flags().Changed("csm-api-host") {
				providerHost = "localhost:8443"
			}
		}
		if insecure {
			csm.Options.InsecureSkipVerify = true
		}
		if slsUrl != "" {
			csm.Options.BaseUrlSLS = slsUrl
		} else {
			csm.Options.BaseUrlSLS = fmt.Sprintf("https://%s/apis/sls/v1", providerHost)
		}
		if hsmUrl != "" {
			csm.Options.BaseUrlHSM = hsmUrl
		} else {
			csm.Options.BaseUrlHSM = fmt.Sprintf("https://%s/apis/smd/hsm/v2", providerHost)
		}
		csm.Options.InsecureSkipVerify = insecure
		csm.Options.SecretName = secretName
		// todo get these values from bss in the Global bootparameters in
		// the fields: kubernetes-pods-cidr and kubernetes-services-cidr
		csm.Options.K8sPodsCidr = k8sPodsCidr
		csm.Options.K8sServicesCidr = k8sServicesCidr
		csm.Options.KubeConfig = kubeconfig
		csm.Options.CaCertPath = caCertPath
		csm.Options.ClientID = clientId
		csm.Options.ClientSecret = clientSecret
		csm.Options.ProviderHost = strings.TrimRight(providerHost, "/") // Remove trailing slash if present
		csm.Options.TokenUsername = tokenUsername
		csm.Options.TokenPassword = tokenPassword

		err := csm.setupClients()
		if err != nil {
			return csm, err
		}

		// get valid roles and subroles from hsm
		hsmValues, _, err := csm.hsmClient.ServiceInfoApi.DoValuesGet(cmd.Context())
		if err != nil {
			return csm, nil
		}

		// Load system specific config data
		csm.Options.ValidRoles = hsmValues.Role
		csm.Options.ValidSubRoles = hsmValues.SubRole

		csm.TBV.ValidRoles = csm.Options.ValidRoles
		csm.TBV.ValidSubRoles = csm.Options.ValidSubRoles
		csm.TBV.K8sPodsCidr, _ = cmd.Flags().GetString("csm-k8s-pods-cidr")
		csm.TBV.K8sServicesCidr, _ = cmd.Flags().GetString("csm-k8s-services-cidr")
		return csm, nil
	} else {
		// use the existing options if a new session is not being initialized
		// unmarshal to a CsmOpts object
		optsMarshaled, _ := yaml.Marshal(opts)
		csmOpts := CsmOpts{}
		err = yaml.Unmarshal(optsMarshaled, &csmOpts)
		if err != nil {
			return csm, err
		}

		// set the options to the object
		csm.Options = &csmOpts

		// if not init, just setup the clients
		err = csm.setupClients()
		if err != nil {
			return csm, err
		}

		csm.TBV.ValidRoles = csmOpts.ValidRoles
		csm.TBV.ValidSubRoles = csmOpts.ValidSubRoles
		csm.TBV.K8sPodsCidr = csmOpts.K8sPodsCidr
		csm.TBV.K8sServicesCidr = csmOpts.K8sServicesCidr

		return csm, nil
	}

	// return csm, nil
}

func (csm *CSM) Slug() string {
	return CsmSlug
}

func (csm *CSM) setupClients() (err error) {
	// Setup HTTP client and context using csm options
	httpClient, _, err := csm.newClient()
	if err != nil {
		return err
	}

	slsClientConfiguration := &sls_client.Configuration{
		BasePath:   csm.Options.BaseUrlSLS,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	hsmClientConfiguration := &hsm_client.Configuration{
		BasePath:   csm.Options.BaseUrlHSM,
		HTTPClient: httpClient.StandardClient(),
		UserAgent:  taxonomy.App,
		DefaultHeader: map[string]string{
			"Content-Type": "application/json",
		},
	}

	if csm.Options.APIGatewayToken != "" {
		// Set the token for use in the clients
		slsClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", csm.Options.APIGatewayToken)
		hsmClientConfiguration.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", csm.Options.APIGatewayToken)
	}

	// Set the clients
	csm.slsClient = sls_client.NewAPIClient(slsClientConfiguration)
	csm.hsmClient = hsm_client.NewAPIClient(hsmClientConfiguration)

	return nil
}

// listCabinetMetadataColumns returns a slice of strings, which are columns names of CSM-specific metadata to be shown when listing cabinets
func listCabinetMetadataColumns() (columns []string) {
	return []string{"HMN VLAN"}
}

// listCabinetMetadataRow returns a slice of strings, which are values from the hardware that correlate to the columns they will be shown in
func listCabinetMetadataRow(hw inventory.Hardware) (values []string, err error) {
	md, err := DecodeProviderMetadata(hw)
	if err != nil {
		return values, err
	}
	vlan := strconv.Itoa(*md.Cabinet.HMNVlan)
	values = append(values, vlan)
	return values, nil
}
