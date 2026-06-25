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
package batch

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// readLines reads path and returns its lines (without trailing newlines).
func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading batch file %s: %w", path, err)
	}
	return strings.Split(string(data), "\n"), nil
}

// commandTokens parses one batch-file line into command arguments. It returns
// ok=false for blank lines, comment lines, and lines that are not cani
// invocations (for example rm or make helpers in an existing setup script). A
// leading program token ("cani", "bin/cani", "./cani", or any path ending in
// "/cani") is stripped, leaving the arguments to dispatch.
func commandTokens(line string) ([]string, bool) {
	tokens := tokenize(line)
	if len(tokens) == 0 {
		return nil, false
	}
	if !isCaniProgram(tokens[0]) {
		return nil, false
	}
	args := tokens[1:]
	return args, len(args) > 0
}

// isCaniProgram reports whether tok names the cani program, with or without a
// leading path (e.g. "cani", "bin/cani", "./cani", "/usr/local/bin/cani").
func isCaniProgram(tok string) bool {
	return tok == "cani" || strings.HasSuffix(tok, "/cani")
}

// tokenize splits a line into shell-style tokens. It supports single and double
// quotes (which group spaces) and treats an unquoted '#' as the start of a
// comment to end of line. It intentionally does not implement backslash
// escaping or variable expansion, which the batch file format does not require.
func tokenize(line string) []string {
	var tokens []string
	var cur strings.Builder
	inToken := false
	var quote rune // 0 when outside quotes, else the open quote rune

	flush := func() {
		if inToken {
			tokens = append(tokens, cur.String())
			cur.Reset()
			inToken = false
		}
	}

	for _, r := range line {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			inToken = true
		case r == '\'' || r == '"':
			quote = r
			inToken = true
		case r == '#':
			flush()
			return tokens
		case unicode.IsSpace(r):
			flush()
		default:
			cur.WriteRune(r)
			inToken = true
		}
	}
	flush()
	return tokens
}
