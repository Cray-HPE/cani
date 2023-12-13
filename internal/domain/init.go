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
	"fmt"

	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/cani/internal/provider/hpcm"
	"github.com/Cray-HPE/cani/internal/provider/hpengi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Init() {
	log.Trace().Msgf("%+v", "github.com/Cray-HPE/cani/internal/domain.init")
}

// NewProviderCmd returns the appropriate command to the cmd layer
func NewProviderCmd(caniCmd *cobra.Command, p string) (providerCmd *cobra.Command, err error) {
	providerCmd = &cobra.Command{}
	switch p {
	case "csm":
		providerCmd, err = csm.NewProviderCmd(caniCmd)
	case "hpcm":
		providerCmd, err = hpcm.NewProviderCmd(caniCmd)
	case "hpengi":
		providerCmd, err = hpengi.NewProviderCmd(caniCmd)
	default:
		err = fmt.Errorf("no command matched for provider %s", p)
	}
	if err != nil {
		return providerCmd, nil
	}
	return providerCmd, nil
}
