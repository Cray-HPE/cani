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
package placement

import "strings"

// Strategy describes how devices are distributed across racks.
type Strategy string

const (
	// StrategyFill packs one rack before moving to the next.
	StrategyFill Strategy = "FILL"
	// StrategySpread distributes devices round-robin across racks.
	StrategySpread Strategy = "SPREAD"
)

// ParseStrategy detects a placement strategy token (%{FILL} or %{SPREAD})
// in the rack flag value. Returns the strategy and true if found, or empty
// and false if the value is a literal rack reference.
func ParseStrategy(rackFlag string) (Strategy, bool) {
	upper := strings.ToUpper(rackFlag)
	if strings.Contains(upper, "%{FILL}") {
		return StrategyFill, true
	}
	if strings.Contains(upper, "%{SPREAD}") {
		return StrategySpread, true
	}
	return "", false
}
