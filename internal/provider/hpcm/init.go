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
package hpcm

import "github.com/spf13/cobra"

func NewSessionInitCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}
	cmd.Short = `HPCM`
	cmd.Long = `HPCM`

	// Session init flags
	cmd.Flags().BoolP("use-simulator", "S", false, "Use simulation environtment settings")
	cmd.Flags().String("cmdb-url", "cmu/v1", "Base URL for the CMDB")
	cmd.Flags().BoolP("insecure", "k", false, "Allow insecure connections when using HTTPS")
	cmd.Flags().String("host", "localhost:8080", "Host or FQDN for APIs")
	cmd.Flags().String("cacert", "", "Path to the CA certificate file")
	cmd.Flags().String("token", "", "API token")
	return cmd, nil
}

func NewAddCabinetCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	return cmd, nil
}

func UpdateAddCabinetCommand(caniCmd *cobra.Command) error {
	return nil
}

func NewAddNodeCommand() (cmd *cobra.Command, err error) {
	// cmd represents for cani alpha add node
	cmd = &cobra.Command{}

	return cmd, nil
}

func NewUpdateNodeCommand() (cmd *cobra.Command, err error) {
	// cmd represents for cani alpha update node
	cmd = &cobra.Command{}

	return cmd, nil
}

// UpdateUpdateNodeCommand
func UpdateUpdateNodeCommand(caniCmd *cobra.Command) error {

	return nil
}

func NewExportCommand() (cmd *cobra.Command, err error) {
	// cmd represents cani alpha export
	cmd = &cobra.Command{}
	cmd.Flags().Bool("hpcm", false, "Export inventory to HPCM format.")

	return cmd, nil
}

func NewImportCommand() (cmd *cobra.Command, err error) {
	// cmd represents cani alpha import
	cmd = &cobra.Command{}

	return cmd, nil
}
