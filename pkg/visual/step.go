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
package visual

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// PrintRackDetails displays rack information during step mode
func PrintRackDetails(index, total int, rack *devicetypes.CaniRackType, opts ETLOptions) {
	fmt.Printf("\n[Rack %d/%d]\n", index, total)
	fmt.Printf("  Name:     %s\n", rack.Name)
	if rack.Location != uuid.Nil {
		fmt.Printf("  Location: %s\n", rack.Location.String()[:8])
	}
	if rack.UHeight > 0 {
		fmt.Printf("  Height:   %d U\n", rack.UHeight)
	}
}

// PrintDeviceDetails displays device information during step mode
func PrintDeviceDetails(index, total int, device *devicetypes.CaniDeviceType, opts ETLOptions) {
	fmt.Printf("\n[Device %d/%d]\n", index, total)
	fmt.Printf("  Name:     %s\n", device.Name)
	fmt.Printf("  Type:     %s\n", device.Slug)
	if device.RackPosition > 0 {
		fmt.Printf("  Position: U%d\n", device.RackPosition)
	}
	if string(device.Type) != "" {
		fmt.Printf("  HW Type:  %s\n", device.Type)
	}
}

// PrintCableDetails displays cable information during step mode
func PrintCableDetails(index, total int, cable *devicetypes.CaniCableType, opts ETLOptions) {
	fmt.Printf("\n[Cable %d/%d]\n", index, total)
	fmt.Printf("  Label:    %s\n", cable.Label)
	fmt.Printf("  Type:     %s\n", cable.Slug)
	if cable.TerminationAPort != "" && cable.TerminationBPort != "" {
		fmt.Printf("  From:     %s\n", cable.TerminationAPort)
		fmt.Printf("  To:       %s\n", cable.TerminationBPort)
	}
	if cable.Length != nil && *cable.Length > 0 {
		fmt.Printf("  Length:   %.1f %s\n", *cable.Length, cable.LengthUnit)
	}
}

// NodeStepInfo holds all info needed for an HPCM node transform step display.
type NodeStepInfo struct {
	NodeNum         int
	Total           int
	RawName         string
	RawType         string
	RawUUID         string
	RawRack         string
	FruCount        int
	FruNames        []string // unique group IDs (e.g. "dimm.432890BE", "disk.disk0")
	Mappings        []FieldMapping
	LibMatch        string // slug of matched library entry, or ""
	MatchQuery      string // query string that produced the match
	MatchScore      int    // confidence score (0-100)
	LibModel        string // model field of matched library entry
	LibManufacturer string // manufacturer field of matched library entry

}

// PrintNodeTransformStep shows raw HPCM node data and how fields map to CANI
// types. Includes per-node field mappings and a running tally.
func PrintNodeTransformStep(info NodeStepInfo, tally StepTally, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, gray, bold, _ := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", yellow("─────────────────────────────────────────────────────────────"), "")

	// Node header
	fmt.Fprintf(w, "%s Node %d of %d: %s\n",
		yellow("→"), info.NodeNum, info.Total, bold(info.RawName))

	// Raw data
	fmt.Fprintf(w, "  %s\n", gray("Raw HPCM Data:"))
	fmt.Fprintf(w, "    name: %s\n", info.RawName)
	fmt.Fprintf(w, "    type: %s\n", info.RawType)
	fmt.Fprintf(w, "    uuid: %s\n", info.RawUUID)
	if info.RawRack != "" {
		fmt.Fprintf(w, "    rack: %s\n", info.RawRack)
	}

	// Library match
	if info.LibMatch != "" {
		fmt.Fprintf(w, "  %s %s\n", gray("Library match:"), info.LibMatch)
		if info.MatchQuery != "" {
			fmt.Fprintf(w, "  %s\n", gray(fmt.Sprintf(
				`    winning query: %q  (score: %d, %s)`,
				info.MatchQuery, info.MatchScore,
				devicetypes.ScoreTierLabel(info.MatchScore))))
			fmt.Fprintf(w, "  %s\n", gray(fmt.Sprintf(
				"    scored against:  slug=%q  model=%q  mfr=%q",
				info.LibMatch, info.LibModel, info.LibManufacturer)))
		}
	} else {
		fmt.Fprintf(w, "  %s %s\n", gray("Library match:"), gray("(none)"))
	}

	// Field mappings table
	PrintFieldMappings(info.Mappings, opts)

	// FRU summary
	fmt.Fprintf(w, "\n  %s %s\n", gray("FRUs:"), bold(fmt.Sprintf("%d", info.FruCount)))
	for _, name := range info.FruNames {
		fmt.Fprintf(w, "    %s\n", name)
	}

	// Running tally
	fmt.Fprintf(w, "  %s racks: %s  devices: %s  FRUs: %s\n",
		gray("tally:"),
		bold(fmt.Sprintf("%d", tally.Racks)),
		bold(fmt.Sprintf("%d", tally.Devices)),
		bold(fmt.Sprintf("%d", tally.Cables))) // reuse Cables as FRU count

	fmt.Fprintf(w, "\n  %s\n", gray("Press Enter to continue..."))
}

// PromptNodeTransformStep prints node transform step info and waits for Enter.
func PromptNodeTransformStep(info NodeStepInfo, tally StepTally, opts ETLOptions) error {
	PrintNodeTransformStep(info, tally, opts)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}
