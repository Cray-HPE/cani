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
package session

import (
	"fmt"
	"os"
	"text/tabwriter"

	root "github.com/Cray-HPE/cani/cmd"
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// SessionSummaryCmd represents the session stop command
var SessionSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show the summary of a stopped session",
	Long:  `Show the summary of a stopped session`,
	RunE:  showSummary,
}

func showSummary(cmd *cobra.Command, args []string) error {
	// resetup the domain to get fresh info
	err := root.D.SetupDomain(cmd, args, root.Conf.Session.Domains)
	if err != nil {
		return err
	}

	// Get the entire inventory
	inv, err := root.D.List()
	if err != nil {
		return err
	}

	// print a header
	fmt.Println("Summary:")
	fmt.Println("--------")
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tTYPE\tSTATUS")

	staged := inv.FilterHardwareByStatus(inventory.HardwareStatusStaged)

	// Create the colors you want to use
	black := color.New(color.FgBlack).FprintfFunc()
	red := color.New(color.FgRed).FprintfFunc()
	green := color.New(color.FgGreen).FprintfFunc()
	yellow := color.New(color.FgYellow).FprintfFunc()
	blue := color.New(color.FgBlue).FprintfFunc()
	// magenta := color.New(color.FgMagenta).FprintfFunc()
	// cyan := color.New(color.FgCyan).FprintfFunc()

	const format = "%s\t%s\t(%s)\n"
	// for each new hardware, print some details
	for i, hw := range staged {
		// Only show staged hardware
		// TODO: Better logic as staged hardware could have been added in a different session
		if hw.Status == inventory.HardwareStatusStaged {
			switch hw.Type {
			case hardwaretypes.Cabinet:
				red(tw, format, i.String(), hw.Type, hw.Status)
			case hardwaretypes.Chassis:
				yellow(tw, format, i.String(), hw.Type, hw.Status)
			case hardwaretypes.NodeBlade:
				green(tw, format, i.String(), hw.Type, hw.Status)
			case hardwaretypes.Node:
				blue(tw, format, i.String(), hw.Type, hw.Status)
			default:
				black(tw, format, i.String(), hw.Type, hw.Status)
			}
			staged[i] = hw
		}
	}

	tw.Flush()

	fmt.Printf("\n%d new hardware item(s) are in the inventory:\n", len(staged))

	return nil
}
