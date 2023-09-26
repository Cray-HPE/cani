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
	"strings"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/provider/csm/validate"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type CsmOpts struct {
	UseSimulation      bool     `json:"UseSimulation" yaml:"use_simulation"`
	InsecureSkipVerify bool     `json:"InsecureSkipVerify" yaml:"insecure_skip_verify"`
	APIGatewayToken    string   `json:"APIGatewayToken" yaml:"api_gateway_token"`
	BaseUrlSLS         string   `json:"BaseUrlSLS" yaml:"base_url_sls"`
	BaseUrlHSM         string   `json:"BaseUrlHSM" yaml:"base_url_hsm"`
	SecretName         string   `json:"SecretName" yaml:"secret_name"`
	K8sPodsCidr        string   `json:"K8sPodsCidr" yaml:"k8s_pods_cidr"`
	K8sServicesCidr    string   `json:"K8sServicesCidr" yaml:"k8s_services_cidr"`
	KubeConfig         string   `json:"KubeConfig" yaml:"kubeconfig"`
	ClientID           string   `json:"-" yaml:"-"` // omit credentials from cani.yml
	ClientSecret       string   `json:"-" yaml:"-"` // omit credentials from cani.yml
	ProviderHost       string   `json:"ProviderHost" yaml:"provider_host"`
	TokenUsername      string   `json:"-" yaml:"-"` // omit credentials from cani.yml
	TokenPassword      string   `json:"-" yaml:"-"` // omit credentials from cani.yml
	CaCertPath         string   `json:"CaCertPath" yaml:"ca_cert_path"`
	ValidRoles         []string `json:"ValidRoles" yaml:"valid_roles"`
	ValidSubRoles      []string `json:"ValidSubRoles" yaml:"valid_sub_roles"`
}

func (csm *CSM) SetProviderOptions(cmd *cobra.Command, args []string) error {
	if cmd.Name() == "init" {
		useSimulation := cmd.Flags().Changed("csm-simulator")
		slsUrl, _ := cmd.Flags().GetString("csm-url-sls")
		hsmUrl, _ := cmd.Flags().GetString("csm-url-hsm")
		insecure := cmd.Flags().Changed("csm-insecure-https")
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
	}

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

	if cmd.Name() == "init" || cmd.Name() == "apply" {
		// get valid roles and subroles from hsm
		hsmValues, _, err := csm.hsmClient.ServiceInfoApi.DoValuesGet(cmd.Context())
		if err != nil {
			return err
		}
		csm.Options.ValidRoles = hsmValues.Role
		csm.Options.ValidSubRoles = hsmValues.SubRole

		tbv := &validate.ToBeValidated{}
		tbv.K8sPodsCidr = csm.Options.K8sPodsCidr
		tbv.K8sServicesCidr = csm.Options.K8sServicesCidr
		tbv.ValidRoles = csm.Options.ValidRoles
		tbv.ValidSubRoles = csm.Options.ValidSubRoles
		csm.TBV = tbv
	}

	return nil
}

func (csm *CSM) GetProviderOptions() (interface{}, error) {
	if csm.Options == nil {
		return nil, fmt.Errorf("options from CSM are nil")
	}
	return csm.Options, nil
}

func (csm *CSM) SetProviderOptionsInterface(opts interface{}) error {
	// type assert this is the options we need
	csm.Options = opts.(*CsmOpts)
	return nil
}
