/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
	"io"
	"os"
	"strings"
)

// ETLOptions controls ETL visual output
type ETLOptions struct {
	NoColor bool
	Writer  io.Writer
}

// StepTally tracks running counts during step-through import
type StepTally struct {
	Racks   int
	Devices int
	Cables  int
}

// Box drawing characters for phase headers
const (
	boxDouble     = "═"
	boxCornerTL   = "╔"
	boxCornerTR   = "╗"
	boxCornerBL   = "╚"
	boxCornerBR   = "╝"
	boxVertDouble = "║"
	boxLight      = "─"
	boxLightVert  = "│"
)

// maxRawDataLen is the maximum length for raw data display before truncation
const maxRawDataLen = 80

// colorFuncs returns closures for coloring text based on NoColor option
func (opts ETLOptions) colorFuncs() (cyan, yellow, green, gray, bold func(string) string) {
	if opts.NoColor {
		identity := func(s string) string { return s }
		return identity, identity, identity, identity, identity
	}
	cyan = func(s string) string { return ColorCyan + s + ColorReset }
	yellow = func(s string) string { return ColorYellow + s + ColorReset }
	green = func(s string) string { return ColorGreen + s + ColorReset }
	gray = func(s string) string { return ColorGray + s + ColorReset }
	bold = func(s string) string { return ColorBold + s + ColorReset }
	return
}

// getWriter returns the configured writer or os.Stdout
func (opts ETLOptions) getWriter() io.Writer {
	if opts.Writer != nil {
		return opts.Writer
	}
	return os.Stdout
}

// PrintPhaseHeader prints a boxed header for an ETL phase
func PrintPhaseHeader(phase string, opts ETLOptions) {
	w := opts.getWriter()
	cyan, _, _, _, bold := opts.colorFuncs()

	phase = strings.ToUpper(phase)
	width := len(phase) + 4

	topBorder := boxCornerTL + strings.Repeat(boxDouble, width) + boxCornerTR
	bottomBorder := boxCornerBL + strings.Repeat(boxDouble, width) + boxCornerBR
	content := boxVertDouble + "  " + bold(phase) + "  " + boxVertDouble

	fmt.Fprintln(w)
	fmt.Fprintln(w, cyan(topBorder))
	fmt.Fprintln(w, cyan(content))
	fmt.Fprintln(w, cyan(bottomBorder))
	fmt.Fprintln(w)
}

// PrintPhaseComplete prints a completion marker for an ETL phase
func PrintPhaseComplete(phase string, opts ETLOptions) {
	w := opts.getWriter()
	_, _, green, _, _ := opts.colorFuncs()

	phase = strings.ToUpper(phase)
	fmt.Fprintf(w, "%s %s phase complete\n\n", green("✓"), phase)
}

// PrintCaniOperation prints a cani framework operation message (cyan)
func PrintCaniOperation(msg string, opts ETLOptions) {
	w := opts.getWriter()
	cyan, _, _, _, _ := opts.colorFuncs()

	fmt.Fprintf(w, "%s %s\n", cyan("│"), msg)
}

// PrintProviderOperation prints a provider-specific operation message (yellow)
func PrintProviderOperation(msg string, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, _, _ := opts.colorFuncs()

	fmt.Fprintf(w, "%s %s\n", yellow("│"), msg)
}

// PrintStepItem prints details about an item being processed (for step-through)
func PrintStepItem(itemDesc string, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, gray, _ := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", yellow("→"), itemDesc)
	fmt.Fprintf(w, "  %s\n", gray("Press Enter to continue..."))
}

// PromptStep prints item details and waits for user to press Enter
func PromptStep(itemDesc string, opts ETLOptions) error {
	PrintStepItem(itemDesc, opts)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}

// truncateRawData truncates a string to maxLen and adds ellipsis if needed
func truncateRawData(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PrintCSVRowStep prints enhanced step info for a CSV row with raw data and tally
func PrintCSVRowStep(rowNum, totalRows int, rawData string, parsed string, tally StepTally, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, gray, bold := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", yellow("─────────────────────────────────────────────────────────────"), "")

	// Row header
	fmt.Fprintf(w, "%s CSV Row %d of %d\n", yellow("→"), rowNum, totalRows)

	// Raw data (truncated if needed)
	truncatedRaw := truncateRawData(rawData, maxRawDataLen)
	fmt.Fprintf(w, "  %s %s\n", gray("raw:"), truncatedRaw)

	// Parsed interpretation
	fmt.Fprintf(w, "  %s %s\n", gray("parsed:"), parsed)

	// Running tally
	fmt.Fprintf(w, "  %s racks: %s  devices: %s  cables: %s\n",
		gray("tally:"),
		bold(fmt.Sprintf("%d", tally.Racks)),
		bold(fmt.Sprintf("%d", tally.Devices)),
		bold(fmt.Sprintf("%d", tally.Cables)))

	fmt.Fprintf(w, "\n  %s\n", gray("Press Enter to continue..."))
}

// PrintCSVRowStepRaw prints step info for a raw CSV row without tally (for extract phase)
func PrintCSVRowStepRaw(rowNum, totalRows int, rawData string, parsed string, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, gray, _ := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", yellow("─────────────────────────────────────────────────────────────"), "")

	// Row header
	fmt.Fprintf(w, "%s CSV Row %d of %d\n", yellow("→"), rowNum, totalRows)

	// Raw data (truncated if needed)
	truncatedRaw := truncateRawData(rawData, maxRawDataLen)
	fmt.Fprintf(w, "  %s %s\n", gray("raw:"), truncatedRaw)

	// Parsed interpretation
	fmt.Fprintf(w, "  %s %s\n", gray("parsed:"), parsed)

	fmt.Fprintf(w, "\n  %s\n", gray("Press Enter to continue..."))
}

// PrintRecordStepRaw prints step info for a record without tally (for extract phase).
// identifier is an optional line (e.g. hostname or serial) shown in gray below the header.
func PrintRecordStepRaw(rowNum, totalRows int, rawData string, parsed string, identifier string, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, gray, _ := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", yellow("─────────────────────────────────────────────────────────────"), "")

	// Record header
	fmt.Fprintf(w, "%s Record %d of %d\n", yellow("→"), rowNum, totalRows)

	// Optional unique identifier in gray
	if identifier != "" {
		fmt.Fprintf(w, "  %s\n", gray(identifier))
	}

	// Raw data (truncated if needed)
	truncatedRaw := truncateRawData(rawData, maxRawDataLen)
	fmt.Fprintf(w, "  %s %s\n", gray("raw:"), truncatedRaw)

	// Parsed interpretation
	fmt.Fprintf(w, "  %s %s\n", gray("parsed:"), parsed)

	fmt.Fprintf(w, "\n  %s\n", gray("Press Enter to continue..."))
}

// PromptRecordStepRaw prints record step info and waits for Enter (no tally).
// identifier is an optional line (e.g. hostname or serial) shown in gray below the header.
func PromptRecordStepRaw(rowNum, totalRows int, rawData string, parsed string, identifier string, opts ETLOptions) error {
	PrintRecordStepRaw(rowNum, totalRows, rawData, parsed, identifier, opts)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}

// PromptCSVRowStepRaw prints raw CSV row step info and waits for Enter (no tally)
func PromptCSVRowStepRaw(rowNum, totalRows int, rawData string, parsed string, opts ETLOptions) error {
	PrintCSVRowStepRaw(rowNum, totalRows, rawData, parsed, opts)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}

// PromptCSVRowStep prints CSV row step info and waits for Enter
func PromptCSVRowStep(rowNum, totalRows int, rawData string, parsed string, tally StepTally, opts ETLOptions) error {
	PrintCSVRowStep(rowNum, totalRows, rawData, parsed, tally, opts)

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}

// PrintFieldMappings prints a formatted table of CSV-to-target field mappings.
// Shows how each CSV field maps to the resulting inventory object field.
// Direct copies show "→", derived/computed fields show "→→".
func PrintFieldMappings(mappings []FieldMapping, opts ETLOptions) {
	w := opts.getWriter()
	cyan, _, green, gray, _ := opts.colorFuncs()

	if len(mappings) == 0 {
		return
	}

	// Calculate column widths for alignment
	maxSourceField := 0
	maxSourceValue := 0
	maxTargetField := 0
	for _, m := range mappings {
		if len(m.SourceField) > maxSourceField {
			maxSourceField = len(m.SourceField)
		}
		// Account for quotes around value
		quotedLen := len(m.SourceValue) + 2
		if quotedLen > maxSourceValue {
			maxSourceValue = quotedLen
		}
		targetField := m.TargetType + "." + m.TargetField
		if len(targetField) > maxTargetField {
			maxTargetField = len(targetField)
		}
	}

	// Cap max widths for readability
	if maxSourceValue > 32 {
		maxSourceValue = 32
	}

	fmt.Fprintf(w, "  %s\n", cyan("Field Mappings:"))
	for _, m := range mappings {
		sourceVal := m.SourceValue
		if len(sourceVal) > 30 {
			sourceVal = sourceVal[:27] + "..."
		}
		quotedSource := fmt.Sprintf("%q", sourceVal)

		arrow := gray(" → ")
		if m.IsDerived {
			arrow = green("→→ ")
		}

		targetField := m.TargetType + "." + m.TargetField
		fmt.Fprintf(w, "    %-*s %-*s %s%-*s = %q\n",
			maxSourceField, m.SourceField,
			maxSourceValue, quotedSource,
			arrow,
			maxTargetField, targetField,
			m.TargetValue)
	}
}

// PromptFieldMappingStep prints field mappings with tally and waits for Enter.
// Used for step-through mode during transform phase.
func PromptFieldMappingStep(rowNum, totalRows int, hwType string, mappings []FieldMapping, tally StepTally, opts ETLOptions) error {
	info := TransformStepInfo{
		Quantity: 1,
		HwType:   hwType,
		Mappings: mappings,
	}
	return PromptTransformStep(rowNum, totalRows, info, tally, opts)
}

// PromptTransformStep prints detailed transform step info with quantity and created items.
func PromptTransformStep(rowNum, totalRows int, info TransformStepInfo, tally StepTally, opts ETLOptions) error {
	w := opts.getWriter()
	cyan, yellow, green, gray, bold := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s %s\n", yellow("─────────────────────────────────────────────────────────────"), "")

	// Row header with hardware type and quantity
	if info.Quantity > 1 {
		fmt.Fprintf(w, "%s CSV Row %d of %d → [%s] × %s\n",
			yellow("→"), rowNum, totalRows, bold(info.HwType), cyan(fmt.Sprintf("%d", info.Quantity)))
	} else {
		fmt.Fprintf(w, "%s CSV Row %d of %d → [%s]\n",
			yellow("→"), rowNum, totalRows, bold(info.HwType))
	}

	// Field mappings table (template from first item)
	PrintFieldMappings(info.Mappings, opts)

	// Show all created items
	if len(info.CreatedItems) > 0 {
		fmt.Fprintf(w, "\n  %s\n", cyan("Created Items:"))
		for i, item := range info.CreatedItems {
			fmt.Fprintf(w, "    %s %s %s\n",
				green(fmt.Sprintf("[%d]", i+1)),
				gray(item.ID),
				item.Name)
		}
	}

	// Running tally
	fmt.Fprintf(w, "\n  %s racks: %s  devices: %s  cables: %s\n",
		gray("tally:"),
		bold(fmt.Sprintf("%d", tally.Racks)),
		bold(fmt.Sprintf("%d", tally.Devices)),
		bold(fmt.Sprintf("%d", tally.Cables)))

	fmt.Fprintf(w, "\n  %s\n", gray("Press Enter to continue..."))

	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	return err
}

// ImportSummary holds summary data for post-import display
type ImportSummary struct {
	RackNames     []string
	DevicesByRack map[string][]string // rack name -> device names
	Cables        []CableSummary
}

// CableSummary describes a cable connection for display
type CableSummary struct {
	SourceDevice string
	SourcePort   string
	DestDevice   string
	DestPort     string
	CableType    string
}

// PrintImportSummary prints a summary of imported items and their relationships
func PrintImportSummary(summary ImportSummary, opts ETLOptions, stepMode bool) {
	w := opts.getWriter()
	cyan, _, green, gray, bold := opts.colorFuncs()

	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", cyan("════════════════════════════════════════════════════════════"))
	fmt.Fprintf(w, "%s %s\n", cyan("║"), bold("IMPORT SUMMARY"))
	fmt.Fprintf(w, "%s\n", cyan("════════════════════════════════════════════════════════════"))
	fmt.Fprintln(w)

	// Racks and their devices
	if len(summary.RackNames) > 0 {
		fmt.Fprintf(w, "%s %s (%d)\n", green("■"), bold("Racks"), len(summary.RackNames))
		for _, rackName := range summary.RackNames {
			fmt.Fprintf(w, "  %s %s\n", cyan("├──"), rackName)
			if devices, ok := summary.DevicesByRack[rackName]; ok && len(devices) > 0 {
				for i, devName := range devices {
					prefix := "│   ├──"
					if i == len(devices)-1 {
						prefix = "│   └──"
					}
					fmt.Fprintf(w, "  %s %s\n", gray(prefix), devName)
				}
			}
		}
		fmt.Fprintln(w)
	}

	// Standalone devices (not in racks)
	if standaloneDevs, ok := summary.DevicesByRack[""]; ok && len(standaloneDevs) > 0 {
		fmt.Fprintf(w, "%s %s (%d)\n", green("■"), bold("Standalone Devices"), len(standaloneDevs))
		for _, devName := range standaloneDevs {
			fmt.Fprintf(w, "  %s %s\n", gray("└──"), devName)
		}
		fmt.Fprintln(w)
	}

	// Cables
	if len(summary.Cables) > 0 {
		fmt.Fprintf(w, "%s %s (%d)\n", green("■"), bold("Cables"), len(summary.Cables))
		for _, cable := range summary.Cables {
			cableInfo := fmt.Sprintf("%s:%s ←→ %s:%s",
				cable.SourceDevice, cable.SourcePort,
				cable.DestDevice, cable.DestPort)
			if cable.CableType != "" {
				cableInfo += fmt.Sprintf(" [%s]", cable.CableType)
			}
			fmt.Fprintf(w, "  %s %s\n", gray("└──"), cableInfo)
		}
		fmt.Fprintln(w)
	}

	// In step mode, wait for user to continue
	if stepMode {
		fmt.Fprintf(w, "  %s\n", gray("Press Enter to continue..."))
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}
}

// PrintWarning prints a warning message
func PrintWarning(msg string, opts ETLOptions) {
	w := opts.getWriter()
	_, yellow, _, _, _ := opts.colorFuncs()

	fmt.Fprintf(w, "%s %s\n", yellow("⚠"), msg)
}

// PrintError prints an error message
func PrintError(msg string, opts ETLOptions) {
	w := opts.getWriter()
	// Use yellow since we don't have red defined
	_, yellow, _, _, _ := opts.colorFuncs()

	fmt.Fprintf(w, "%s %s\n", yellow("✗"), msg)
}
