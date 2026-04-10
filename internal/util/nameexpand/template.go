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

// knownTokens lists all supported template variable names.
var knownTokens = map[string]bool{
	"RACK":   true,
	"PARENT": true, // alias for RACK
	"U":      true,
	"SEQ":    true,
	"FACE":   true,
	"ZONE":   true,
	"DEVICE": true,
	"BAY":    true,
}

// IsTemplate returns true if the string contains %{...} template tokens.
func IsTemplate(s string) bool {
	return strings.Contains(s, "%{") && strings.Contains(s, "}")
}

// ExpandTemplate replaces %{KEY} tokens in pattern with values from vars.
// PARENT is an alias for RACK — if PARENT is requested and not in vars,
// the RACK value is used. Returns an error for unknown tokens.
func ExpandTemplate(pattern string, vars map[string]string) (string, error) {
	var result strings.Builder
	i := 0
	for i < len(pattern) {
		// Look for next %{ token.
		start := strings.Index(pattern[i:], "%{")
		if start == -1 {
			result.WriteString(pattern[i:])
			break
		}
		// Write literal text before the token.
		result.WriteString(pattern[i : i+start])

		// Find closing brace.
		tokenStart := i + start + 2 // skip "%{"
		end := strings.Index(pattern[tokenStart:], "}")
		if end == -1 {
			return "", fmt.Errorf("unclosed template token at position %d", i+start)
		}

		key := strings.ToUpper(pattern[tokenStart : tokenStart+end])
		if !knownTokens[key] {
			return "", fmt.Errorf("unknown template token %%{%s}", key)
		}

		val, ok := vars[key]
		if !ok && key == "PARENT" {
			val, ok = vars["RACK"]
		}
		if !ok {
			return "", fmt.Errorf("template token %%{%s} has no value", key)
		}
		result.WriteString(val)

		i = tokenStart + end + 1 // advance past "}"
	}
	return result.String(), nil
}
