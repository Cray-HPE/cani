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
package cabinet

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	cabinetNumber int
	vlanId        int
	auto          bool
	accept        bool
	format        string
)

func init() {
	// Add variants to root commands
	root.AddCmd.AddCommand(AddCabinetCmd)
	root.ListCmd.AddCommand(ListCabinetCmd)
	root.RemoveCmd.AddCommand(RemoveCabinetCmd)

	// Add a flag to show supported types
	AddCabinetCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	// Cabinets
	AddCabinetCmd.Flags().IntVar(&cabinetNumber, "cabinet", 1001, "Cabinet number.")
	// AddCabinetCmd.MarkFlagRequired("cabinet")
	AddCabinetCmd.Flags().IntVar(&vlanId, "vlan-id", -1, "Vlan ID for the cabinet.")
	// AddCabinetCmd.MarkFlagRequired("vlan-id")
	AddCabinetCmd.MarkFlagsRequiredTogether("cabinet", "vlan-id")
	AddCabinetCmd.Flags().BoolVar(&auto, "auto", false, "Automatically recommend and assign required flags.")
	AddCabinetCmd.MarkFlagsMutuallyExclusive("auto")
	AddCabinetCmd.Flags().BoolVarP(&accept, "accept", "y", false, "Automatically accept recommended values.")
	ListCabinetCmd.Flags().StringVarP(&format, "format", "f", "pretty", "Format out output")
}
