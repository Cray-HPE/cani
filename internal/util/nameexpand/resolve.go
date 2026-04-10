/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package nameexpand

import (
	"fmt"
	"strings"
)

// ResolveNames determines the list of names to assign to items based on
// the combination of CLI flags provided. It supports three modes:
//
//  1. Brace expansion: --name "x370{1..4}" expands to [x3701, x3702, x3703, x3704].
//  2. Prefix+increment: --prefix "x370" --start 1 generates sequential numeric names.
//  3. Plain name: --name "myRack" with qty=1 returns ["myRack"].
//
// Returns an error when:
//   - Both --name (with braces) and --prefix are provided (mutually exclusive).
//   - The expanded name count does not match qty.
//   - A plain --name is used with qty > 1 (ambiguous intent).
func ResolveNames(nameFlag, prefix string, start, padWidth, qty int) ([]string, error) {
	hasTemplate := IsTemplate(nameFlag)
	hasBraces := !hasTemplate && strings.Contains(nameFlag, "{") && strings.Contains(nameFlag, "}")
	hasPrefix := prefix != ""
	if hasTemplate && hasPrefix {
		return nil, fmt.Errorf("template tokens (%%{...}) and --prefix are mutually exclusive")
	}

	// Template mode — return nil to signal deferred resolution.
	if hasTemplate {
		return nil, nil
	}

	if hasBraces && hasPrefix {
		return nil, fmt.Errorf("--name with brace expansion and --prefix are mutually exclusive")
	}

	if hasBraces {
		return resolveBraceExpansion(nameFlag, qty)
	}

	if hasPrefix {
		return Sequence(prefix, start, qty, padWidth), nil
	}

	if nameFlag != "" {
		if qty > 1 {
			return nil, fmt.Errorf(
				"plain --name %q with --qty %d is ambiguous; use brace expansion "+
					"(e.g., --name \"%s{1..%d}\") or --prefix",
				nameFlag, qty, nameFlag, qty,
			)
		}
		return []string{nameFlag}, nil
	}

	// No name flags provided — return nil so the caller can fall back to defaults.
	return nil, nil
}

// resolveBraceExpansion expands the pattern and validates the result count.
func resolveBraceExpansion(pattern string, qty int) ([]string, error) {
	names, err := Expand(pattern)
	if err != nil {
		return nil, fmt.Errorf("name expansion failed: %w", err)
	}

	if len(names) != qty {
		return nil, fmt.Errorf(
			"name expansion produced %d names but --qty is %d; they must match",
			len(names), qty,
		)
	}
	return names, nil
}
