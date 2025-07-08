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
	"fmt"

	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/ngsm/remove"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

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

var ErrIsDeprecated = fmt.Errorf("the HPCM provider is no longer supported.  Use v0.4.x if you need to use HPCM.")

func init() {
	provider.Register("HPCM", &HPCM{})
}

type HPCM struct{}

func (p *HPCM) Slug() string { return "HPCM" }

// NewProviderCmd is called by your CLI init loop.
// base will be each of your core verbs (“add”, “remove”, …).
// Return a new cobra.Command (or nil, nil if you don’t support that verb).
func (p *HPCM) NewProviderCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	switch caniCmd.Name() {
	case "init": // session init
		providerCmd, _ = p.newInitCmd(caniCmd)

	default:
		// this provider doesn’t customize that verb
		return nil, nil
	}

	return providerCmd, nil
}

func (p *HPCM) Transform(existing devicetypes.Inventory) (map[uuid.UUID]*devicetypes.CaniDeviceType, error) {
	// This is where you would implement the import logic for HPCM.
	// For now, we return an empty map.
	return make(map[uuid.UUID]*devicetypes.CaniDeviceType), nil
}

func (p *HPCM) Extract(cmd *cobra.Command, args []string) error {
	// This is where you would implement any validation logic for HPCM.
	// For now, we just return nil to indicate no errors.
	return nil
}

func (p *HPCM) Add(cmd *cobra.Command, args []string, deviceType devicetypes.DeviceType) (devicesToAdd map[uuid.UUID]*devicetypes.CaniDeviceType, err error) {
	return devicesToAdd, nil
}

// Remove is called when the user runs `cani remove <device> <device-type-slug> <args>`
func (p *HPCM) Remove(cmd *cobra.Command, args []string) (idsToRemove []uuid.UUID, err error) {
	return remove.Remove(cmd, args)
}

// Show is called when the user runs `cani list`
func (p *HPCM) Show(cmd *cobra.Command, args []string, devices []*devicetypes.CaniDeviceType) (err error) {
	return nil
}

func (p *HPCM) newInitCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{
		RunE: initializeHPCM,
	}

	return providerCmd, nil
}

func initializeHPCM(cmd *cobra.Command, args []string) error {
	return ErrIsDeprecated
}
