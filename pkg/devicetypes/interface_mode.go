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
package devicetypes

import (
	"fmt"
	"strings"
)

// Switchport mode values accepted by Nautobot for an interface.
const (
	InterfaceModeAccess    = "access"
	InterfaceModeTagged    = "tagged"
	InterfaceModeTaggedAll = "tagged-all"
)

// ValidateInterfaceMode normalizes mode to lowercase and verifies it is one of
// the supported switchport modes (access, tagged, tagged-all). An empty (or
// whitespace-only) value returns an empty string with no error so callers may
// treat "unset" as valid (e.g. to clear the mode).
func ValidateInterfaceMode(mode string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(mode))
	switch trimmed {
	case "", InterfaceModeAccess, InterfaceModeTagged, InterfaceModeTaggedAll:
		return trimmed, nil
	default:
		return "", fmt.Errorf("invalid interface mode %q: must be one of access, tagged, tagged-all", mode)
	}
}
