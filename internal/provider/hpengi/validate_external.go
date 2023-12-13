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
package hpengi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/Cray-HPE/cani/internal/provider/hpcm"
	"github.com/Cray-HPE/cani/pkg/canu"
	sls "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ValidateExternal
func (hpengi *Hpengi) ValidateExternal(cmd *cobra.Command, args []string) error {
	// hpengi may be a brand new system but could also be a migration from a different provider
	// the site survey can run interactively, import from config files, or query a HPCM CMDB
	err := hpengi.siteSurvey(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

// siteSurvey
func (hpengi *Hpengi) siteSurvey(cmd *cobra.Command, args []string) error {
	// [non]interactively get the hpcm cluster config file if one is passed in
	if cmd.Flags().Changed("cm-config") {
		cm, err := hpengi.getCmConfig(cmd, args)
		if err != nil {
			return err
		}
		hpengi.CmConfig = cm
	}

	// validate the customer intent document if one is passed in
	if cmd.Flags().Changed("cid") {
		cid, err := hpengi.getSmartCid(cmd, args)
		if err != nil {
			return err
		}
		hpengi.Cid = cid
	}

	// validate the customer intent document if one is passed in
	if cmd.Flags().Changed("paddle") {
		ccj, err := hpengi.getPaddle(cmd, args)
		if err != nil {
			return err
		}
		hpengi.Paddle = ccj
	}

	// validate the customer intent document if one is passed in
	if cmd.Flags().Changed("sls-dumpstate") {
		dump, err := hpengi.getSlsDumpstate(cmd, args)
		if err != nil {
			return err
		}
		hpengi.SlsInput = dump
	}

	return nil
}

// getSlsDumpstate
func (hpengi *Hpengi) getSlsDumpstate(cmd *cobra.Command, args []string) (dump sls.SLSState, err error) {
	f, _ := cmd.Flags().GetString("sls-dumpstate")
	s, err := os.ReadFile(f)
	if err != nil {
		return dump, err
	}

	err = json.Unmarshal(s, &dump)
	if err != nil {
		return dump, err
	}

	return dump, nil
}

// getPaddle
func (hpengi *Hpengi) getPaddle(cmd *cobra.Command, args []string) (ccj *canu.Paddle, err error) {
	f, _ := cmd.Flags().GetString("paddle")
	p, err := os.ReadFile(f)
	if err != nil {
		return ccj, err
	}

	err = json.Unmarshal(p, &ccj)
	if err != nil {
		return ccj, err
	}

	return ccj, nil
}

// getCmConfig
func (hpengi *Hpengi) getCmConfig(cmd *cobra.Command, args []string) (hpcm.HpcmConfig, error) {
	f, _ := cmd.Flags().GetString("cm-config")
	cm, err := hpcm.LoadCmConfig(f)
	if err != nil {
		return cm, err
	}

	return cm, nil
}

// getSmartCid
func (hpengi *Hpengi) getSmartCid(cmd *cobra.Command, args []string) (Cid, error) {
	cid := Cid{}
	f, _ := cmd.Flags().GetString("cid")

	// read the CID into memory
	body, err := os.ReadFile(f)
	if err != nil {
		return cid, err
	}

	// unmarshal to struct
	err = json.Unmarshal(body, &cid)
	if err != nil {
		return cid, err
	}

	return cid, nil
}

// stringPrompt asks for a string value using the label
func stringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+"")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

// passwordPrompt asks for a string value using the label.
// The entered value will not be displayed on the screen
// while typing.
func passwordPrompt(label string) string {
	var s string
	for {
		fmt.Fprint(os.Stderr, label+" ")
		b, _ := term.ReadPassword(int(syscall.Stdin))
		s = string(b)
		if s != "" {
			break
		}
	}
	fmt.Println()
	return s
}
