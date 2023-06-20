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
package blade

import (
	root "github.com/Cray-HPE/cani/cmd"
)

var (
	auto      bool
	cabinet   int
	chassis   int
	blade     int
	recursion bool
)

func init() {
	// Add blade variants to root commands
	root.AddCmd.AddCommand(AddBladeCmd)
	root.ListCmd.AddCommand(ListBladeCmd)
	root.RemoveCmd.AddCommand(RemoveBladeCmd)

	// Add a flag to show supported types
	AddBladeCmd.Flags().BoolP("list-supported-types", "L", false, "List supported hardware types.")

	// Blades have several parents, so we need to add flags for each
	AddBladeCmd.Flags().IntVar(&cabinet, "cabinet", 1001, "Parent cabinet")
	AddBladeCmd.MarkFlagRequired("cabinet")
	AddBladeCmd.Flags().IntVar(&chassis, "chassis", 7, "Parent chassis")
	AddBladeCmd.MarkFlagRequired("chassis")
	AddBladeCmd.Flags().IntVar(&blade, "blade", 1, "Blade")
	AddBladeCmd.MarkFlagRequired("blade")
	AddBladeCmd.MarkFlagsRequiredTogether("cabinet", "chassis", "blade")

	AddBladeCmd.Flags().BoolVar(&auto, "auto", false, "Automatically recommend values for parent hardware")
	AddBladeCmd.MarkFlagsRequiredTogether("list-supported-types")

	RemoveBladeCmd.Flags().BoolVarP(&recursion, "recursive", "R", false, "Recursively delete child hardware")

}
