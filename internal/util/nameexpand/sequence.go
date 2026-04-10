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
	"strconv"
)

// Sequence generates count names by concatenating prefix with a zero-padded
// incrementing number starting at start. If padWidth is 0, the width is
// auto-detected from start (i.e., the number of digits in start).
//
// Example: Sequence("x370", 1, 4, 0) → ["x3701", "x3702", "x3703", "x3704"]
// Example: Sequence("node-", 1, 3, 3) → ["node-001", "node-002", "node-003"]
func Sequence(prefix string, start, count, padWidth int) []string {
	if padWidth <= 0 {
		padWidth = len(strconv.Itoa(start + count - 1))
		startWidth := len(strconv.Itoa(start))
		if startWidth > padWidth {
			padWidth = startWidth
		}
	}

	result := make([]string, count)
	for i := range count {
		result[i] = prefix + padNumber(start+i, padWidth)
	}
	return result
}

// SequenceAlpha generates count names by concatenating prefix with an
// incrementing letter sequence starting at startLetter. Letters use
// spreadsheet-style column naming: a-z, then aa, ab, ..., az, ba, ...
//
// Case is preserved from startLetter.
//
// Example: SequenceAlpha("rack-", "a", 4) → ["rack-a", "rack-b", "rack-c", "rack-d"]
// Example: SequenceAlpha("shelf-", "y", 4) → ["shelf-y", "shelf-z", "shelf-aa", "shelf-ab"]
func SequenceAlpha(prefix, startLetter string, count int) []string {
	upper := len(startLetter) > 0 && startLetter[0] >= 'A' && startLetter[0] <= 'Z'
	startIdx := alphaToIndex(startLetter)

	result := make([]string, count)
	for i := range count {
		result[i] = prefix + indexToAlpha(startIdx+i, upper)
	}
	return result
}
