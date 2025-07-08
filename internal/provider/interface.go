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
package provider

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var ErrDataValidationFailure = fmt.Errorf("data validation failure")

var ActiveProvider Provider

type Provider interface {
	// Extract runs first and lets the provider ensure their external inventory source is valid
	// This might be simply validating an API is reachable and has expected data
	// It could also parse a file and set up a queue of devices to be created during Import()
	// This is the "Extract" step in ETL
	// A common pattern is to extract data from an external source, such as a REST API or a file,
	// then add it to the provider structure for later processing in the Transform method.
	Extract(cmd *cobra.Command, args []string) error

	// Transform runs after Extract
	// It typically does the "Transform" step in ETL to convert the external source data into CANI's format
	// CANI will load the existing inventory from the datastore and pass it to Transform in case the provider needs to
	// check for existing devices or racks.
	// It should return a map of UUIDs to DeviceType pointers, which will be added to the datastore by CANI
	Transform(existing devicetypes.Inventory) (map[uuid.UUID]*devicetypes.CaniDeviceType, error)

	// NewProviderCmd returns a new cobra.Command for the provider
	// This is used to add provider-specific info the the CLI
	// This is usually a switch statement that returns a command for each command the provider supports
	NewProviderCmd(base *cobra.Command) (*cobra.Command, error)

	// Add is called when the user runs `cani add <device> <device-type-slug> <args>`
	// The provider should handle the logic and give back a CaniDeviceType with any additional fields set
	Add(cmd *cobra.Command, args []string, deviceType devicetypes.DeviceType) (devicesToAdd map[uuid.UUID]*devicetypes.CaniDeviceType, err error)

	// Show is called when the user runs `cani list`
	Show(cmd *cobra.Command, args []string, devices []*devicetypes.CaniDeviceType) (err error)

	// Remove is called when the user runs `cani remove <device> <device-type-slug> <args>`
	// The provider should handle the logic and return a slice of UUIDs to be removed from the datastore
	Remove(cmd *cobra.Command, args []string) (ids []uuid.UUID, err error)

	// Slug simply returns the provider's slug, which is used to identify it in the system
	Slug() string
}

var providers = map[string]Provider{}

// Register makes a plugin available under name.
// This should be called in the plugin's init() function.
func Register(name string, p Provider) {
	providers[name] = p
}

// GetActiveProvider returns a registered plugin or nil.
func GetActiveProvider(cmd *cobra.Command, args []string) (err error) {
	if config.Cfg == nil {
		return nil
	}
	if config.Cfg.ActiveProvider == "" {
		return nil
	}
	ActiveProvider = GetProvider(config.Cfg.ActiveProvider)
	if ActiveProvider == nil {
		return fmt.Errorf("no active provider found. Run 'cani session init <provider>' to initialize a session")
	}
	return nil
}

// GetProvider returns a registered plugin or nil.
func GetProvider(name string) Provider {
	return providers[name]
}

// All returns every registered plugin
func GetProviders() map[string]Provider {
	return providers
}
