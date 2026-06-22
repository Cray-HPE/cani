/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2025 Hewlett Packard Enterprise Development LP
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
	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/provider/csm/commands"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/Cray-HPE/cani/pkg/provider/csm/transform"
)

// instance is the singleton provider instance
var instance *Csm

func init() {
	instance = New()
	provider.Register("csm", instance)

	// Register the provider getter with the import package to break import cycle.
	import_.SetProviderGetter(func() interface {
		ClearRawData()
		SetSls(sls *import_.SlsDumpstate)
		SetSmd(smd *import_.SmdComponentList)
	} {
		return instance
	})

	// Register the provider getter with the transform package to break import cycle.
	transform.SetProviderGetter(func() interface {
		GetSls() *import_.SlsDumpstate
		GetSmd() *import_.SmdComponentList
	} {
		return instance
	})
}

// NewProviderCmd returns provider-specific CLI commands.
// This is called for each base command (import, add, show, etc.) to allow
// the provider to customize or extend the command.
func (p *Csm) NewProviderCmd(base *cli.Command) (*cli.Command, error) {
	// Switch on the base command name to provide customizations
	switch base.Name() {
	case "import":
		return commands.NewImportCommand(base)

	case "export":
		return commands.NewExportCommand(base)

	case "show":
		return commands.NewShowCommand(base)

	case "add":
		return commands.NewAddCommand(base)

	case "remove":
		return commands.NewRemoveCommand(base)

	case "update":
		return commands.NewUpdateCommand(base)

	default:
		// No customization for this command
		return base, nil
	}
}
