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
package domain

import (
	"github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/spf13/cobra"
)

func NewSessionInitCommand(p string) (providerCmd *cobra.Command, err error) {
	switch p {
	case "csm":
		providerCmd, err = csm.NewSessionInitCommand()
	}
	if err != nil {
		return providerCmd, err
	}
	return providerCmd, nil
}

// NewProviderCmd returns the appropriate command to the cmd layer
func NewProviderCmd(bootstrapCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{}
	for _, p := range GetProviders() {
		switch p.Slug() {
		case taxonomy.CSM:
			providerCmd, err = csm.NewProviderCmd(bootstrapCmd)
		}
		if err != nil {
			return providerCmd, nil
		}
	}

	return providerCmd, nil
}

// UpdateProviderCmd allows the provider to make updates to the command after cani defines its settings
func UpdateProviderCmd(bootstrapCmd *cobra.Command) (err error) {
	for _, p := range GetProviders() {
		switch p.Slug() {
		case taxonomy.CSM:
			err = csm.UpdateProviderCmd(bootstrapCmd)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
