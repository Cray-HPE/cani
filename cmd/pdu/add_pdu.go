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
package pdu

import (
	"os"
	"path/filepath"
	"strings"

	nb "github.com/Cray-HPE/cani/pkg/netbox"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AddPduCmd represents the pdu add command
var AddPduCmd = &cobra.Command{
	Use:     "pdu",
	Short:   "Add pdus to the inventory.",
	Long:    `Add pdus to the inventory.`,
	PreRunE: validHardware, // Hardware can only be valid if defined in the hardware library
	RunE:    addPdu,        // Add a pdu when this sub-command is called
}

// addPdu adds a pdu to the inventory
func addPdu(cmd *cobra.Command, args []string) error {
	// Loading the environment variables from '.env' file.
	log.Info().Msgf("%+v", "Loading config from .env file")
	ie, err := nb.LoadEnvs()
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	nogit := strings.TrimSuffix(ie.RepoUrl, ".git")
	dir2 := filepath.Base(nogit)
	path := filepath.Join(".", dir2, "device-types")

	client, ctx, err := ie.NewClient()
	if err != nil {
		log.Error().Msgf("%+v", err)
		os.Exit(1)
	}

	// create a map of manufacturers to import in case they do not already exist
	manufacturersToImport := make(map[string]string, 0)
	deviceTypesToImport := make(map[string]nb.DeviceType, 0)
	err = filepath.Walk(path, nb.CreateDeviceTypeMap(manufacturersToImport, deviceTypesToImport, []string{}))
	if err != nil {
		return err
	}

	// manufactureres need to exist in order to create device types
	err = nb.CreateManufacturers(client, ctx, manufacturersToImport)
	if err != nil {
		return err
	}

	// create the device types
	err = nb.CreateDeviceTypes(client, ctx, deviceTypesToImport)
	if err != nil {
		return err
	}

	return nil
}
