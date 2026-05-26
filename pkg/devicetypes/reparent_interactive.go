/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package devicetypes

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// PromptForParent displays an interactive menu that lets the user pick a
// parent for an orphan item. Returns the selected parent UUID or uuid.Nil
// if the user chose to skip.
//
// The inv parameter is used for search; it may be nil if search is not needed.
func PromptForParent(inv *Inventory, orphan OrphanItem, suggestions []ParentSuggestion, opts ClassifyOptions) (uuid.UUID, error) {
	w := opts.writer()
	r := bufio.NewReader(opts.reader())

	cyan, yellow, green, gray, bold := colorFuncs(opts.NoColor)

	// Display orphan summary
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\n", bold("─── Orphan "+orphan.Kind+" ───"))
	fmt.Fprintf(w, "  Name:          %s\n", cyan(orphan.Name))
	fmt.Fprintf(w, "  ID:            %s\n", gray(orphan.ID.String()))
	if orphan.DeviceType != "" {
		fmt.Fprintf(w, "  Type:          %s\n", orphan.DeviceType)
	}
	if orphan.Model != "" {
		fmt.Fprintf(w, "  Model:         %s\n", orphan.Model)
	}
	if orphan.Manufacturer != "" {
		fmt.Fprintf(w, "  Manufacturer:  %s\n", orphan.Manufacturer)
	}
	printProviderMetadata(w, orphan.ProviderMetadata, gray)

	// Show suggestions
	if len(suggestions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%s\n", yellow("Candidate parents:"))
		for i, s := range suggestions {
			label := scoreTier(s.Score)
			reason := ""
			if s.Reason != "" {
				reason = " — " + s.Reason
			}
			detail := ""
			if s.Detail != "" {
				detail = "\n          " + s.Detail
			}
			fmt.Fprintf(w, "  %s  %s %s%s\n",
				bold(fmt.Sprintf("[%d]", i+1)),
				s.Name,
				gray(fmt.Sprintf("(%s, %d%% %s%s)", s.Kind, s.Score, label, reason)),
				gray(detail),
			)
		}
	} else {
		fmt.Fprintf(w, "\n%s\n", gray("  No candidate parents found"))
	}

	// Show options
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s select candidate, %s to search, %s to skip\n",
		bold("[1-N]"), bold("[s]"), bold("[k]"))
	fmt.Fprintf(w, "  > ")

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return uuid.Nil, fmt.Errorf("reading input: %w", err)
		}
		input := strings.TrimSpace(line)

		switch strings.ToLower(input) {
		case "k", "skip":
			return uuid.Nil, nil
		case "s", "search":
			id, err := promptParentSearch(inv, orphan.Kind, w, r, opts.NoColor)
			if err != nil {
				return uuid.Nil, err
			}
			if id != uuid.Nil {
				fmt.Fprintf(w, "  %s selected\n", green(id.String()))
			}
			return id, nil
		default:
			n, err := strconv.Atoi(input)
			if err == nil && n >= 1 && n <= len(suggestions) {
				sel := suggestions[n-1]
				fmt.Fprintf(w, "  %s %s (%s)\n", green("✓"), sel.Name, sel.Kind)
				return sel.ID, nil
			}
			// Try as UUID
			if parsed, perr := uuid.Parse(input); perr == nil {
				fmt.Fprintf(w, "  %s %s\n", green("✓"), parsed)
				return parsed, nil
			}
			fmt.Fprintf(w, "  Invalid input. Enter a number, 's' to search, or 'k' to skip: ")
		}
	}
}

// promptParentSearch presents a free-text search prompt for parent candidates.
func promptParentSearch(inv *Inventory, orphanKind string, w io.Writer, r *bufio.Reader, noColor bool) (uuid.UUID, error) {
	_, yellow, _, gray, bold := colorFuncs(noColor)

	fmt.Fprintf(w, "  %s ", yellow("Search query:"))
	line, err := r.ReadString('\n')
	if err != nil {
		return uuid.Nil, fmt.Errorf("reading search input: %w", err)
	}
	query := strings.TrimSpace(line)
	if query == "" {
		return uuid.Nil, nil
	}

	results := SearchParentCandidates(inv, query, orphanKind, 10)
	if len(results) == 0 {
		fmt.Fprintf(w, "  %s\n", gray("No matches found"))
		return uuid.Nil, nil
	}

	fmt.Fprintf(w, "  %s\n", yellow("Search results:"))
	for i, s := range results {
		detail := ""
		if s.Detail != "" {
			detail = " — " + s.Detail
		}
		fmt.Fprintf(w, "    %s  %s %s\n",
			bold(fmt.Sprintf("[%d]", i+1)),
			s.Name,
			gray(fmt.Sprintf("(%s%s)", s.Kind, detail)),
		)
	}
	fmt.Fprintf(w, "  Select or 'k' to skip: ")

	line, err = r.ReadString('\n')
	if err != nil {
		return uuid.Nil, fmt.Errorf("reading selection: %w", err)
	}
	input := strings.TrimSpace(line)
	if strings.ToLower(input) == "k" {
		return uuid.Nil, nil
	}
	n, err := strconv.Atoi(input)
	if err == nil && n >= 1 && n <= len(results) {
		return results[n-1].ID, nil
	}
	return uuid.Nil, nil
}

// scoreTier returns a human label for a score value.
func scoreTier(score int) string {
	switch {
	case score >= 70:
		return "strong"
	case score >= 40:
		return "moderate"
	default:
		return "weak"
	}
}
