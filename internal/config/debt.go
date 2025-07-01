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
package config

import (
	"os"

	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"gopkg.in/yaml.v3"
)

// DeprecatedConfig is the single-provider Config object and is left here for migration purposes
type DeprecatedConfig struct {
	Session *DeprecatedSession `yaml:"session"`
}

// DeprecatedSession is the single-provider Session object and is left here for migration purposes
type DeprecatedSession struct {
	DomainOptions *DeprecatedDomainOpts `yaml:"domain_options"`
	Domain        *DeprecatedDomain     `yaml:"domain"`
	Active        bool                  `yaml:"active"`
}

// DeprecatedDomain is the single-provider Domain object and is left here for migration purposes
type DeprecatedDomain struct {
	hardwareTypeLibrary       *hardwaretypes.Library
	datastore                 inventory.Datastore
	externalInventoryProvider provider.InventoryProvider
	configOptions             DeprecatedConfigOptions
}

// DeprecatedDomainOpts is the single-provider Domain object and is left here for migration purposes
type DeprecatedDomainOpts struct {
	DatastorePath          string                 `yaml:"datastore_path"`
	LogFilePath            string                 `yaml:"log_file_path"`
	Provider               string                 `yaml:"provider"`
	CsmOptions             DeprecatedProviderOpts `yaml:"csm_options"`
	CustomHardwareTypesDir string                 `yaml:"custom_hardware_types_dir"`
}

// DeprecatedProviderOpts are the single-provider options and is now handled within each provider's package
type DeprecatedProviderOpts struct {
	UseSimulation      bool
	InsecureSkipVerify bool
	APIGatewayToken    string
	BaseUrlSLS         string
	BaseUrlHSM         string
	SecretName         string
	K8sPodsCidr        string
	K8sServicesCidr    string
	KubeConfig         string
	ClientID           string `json:"-" yaml:"-"` // omit credentials from cani.yml
	ClientSecret       string `json:"-" yaml:"-"` // omit credentials from cani.yml
	ProviderHost       string
	TokenUsername      string `json:"-" yaml:"-"` // omit credentials from cani.yml
	TokenPassword      string `json:"-" yaml:"-"` // omit credentials from cani.yml
	CaCertPath         string
	ValidRoles         []string
	ValidSubRoles      []string
}

// DeprecatedConfigOptions is the single-provider ConfigOptions and is now handled within each provider's package
type DeprecatedConfigOptions struct {
	ValidRoles      []string
	ValidSubRoles   []string
	K8sPodsCidr     string
	K8sServicesCidr string
}

// migrateConfig converts an older single-provider config into a multi-provider formatted one
func migrateConfig(path string, old []byte) (migrated *Config, err error) {
	// unmarshal the old config style to a struct
	dc := &DeprecatedConfig{}
	err = yaml.Unmarshal(old, &dc)
	if err != nil {
		return migrated, err
	}

	// get just the options
	csmOpts, err := yaml.Marshal(dc.Session.DomainOptions.CsmOptions)
	if err != nil {
		return migrated, err
	}

	// rename the old config
	err = os.Rename(path, path+".single")
	if err != nil {
		return migrated, err
	}

	// Generate a new one in its place
	err = InitConfig(path)
	if err != nil {
		return migrated, err
	}

	// load the new config
	new, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// unmarshal the new config to a struct
	err = yaml.Unmarshal(new, &migrated)
	if err != nil {
		return nil, err
	}

	providerOpts := map[string]interface{}{}

	// convert deprecated csm_options to new generic options
	err = yaml.Unmarshal(csmOpts, &providerOpts)
	if err != nil {
		return nil, err
	}

	// the multi-provider structure is slightly different, so map appropriately
	migrated.Session.Domains[taxonomy.CSM].CustomHardwareTypesDir = dc.Session.DomainOptions.CustomHardwareTypesDir
	migrated.Session.Domains[taxonomy.CSM].DatastorePath = dc.Session.DomainOptions.DatastorePath
	migrated.Session.Domains[taxonomy.CSM].LogFilePath = dc.Session.DomainOptions.LogFilePath
	migrated.Session.Domains[taxonomy.CSM].Active = dc.Session.Active
	migrated.Session.Domains[taxonomy.CSM].Provider = dc.Session.DomainOptions.Provider
	// also set the csm_options to the now-generic options
	migrated.Session.Domains[taxonomy.CSM].Options = providerOpts

	err = WriteConfig(path, migrated)
	if err != nil {
		return migrated, err
	}

	return migrated, nil
}
