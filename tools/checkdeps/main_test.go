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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunChecksDirectDependencies verifies the guard follows Go module syntax.
//
// Why it matters: comments must neither hide direct dependencies nor reject
// valid declarations, and malformed manifests must fail closed.
// Inputs: manifests and an allowlist. Outputs: errors or sanctioned counts.
// Data choice: the reported comment cases, quoted paths and genuine indirect
// annotations distinguish Go parsing from the previous text approximation.
func TestRunChecksDirectDependencies(t *testing.T) {
	cases := []struct {
		name         string
		requirements string
		wantError    string
		wantCount    string
	}{
		{"empty", "", "", "0"},
		{"sanctioned", "require github.com/google/uuid v1.6.0", "", "1"},
		{"commented block", "require ( // sanctioned modules\n github.com/google/uuid v1.6.0\n)", "", "1"},
		{"quoted path", "require \"github.com/google/uuid\" v1.6.0", "", "1"},
		{"unapproved", "require example.com/unapproved v1.0.0", "unsanctioned direct dependency", ""},
		{"similar comment", "require example.com/unapproved v1.0.0 // indirectly required", "unsanctioned direct dependency", ""},
		{"similar block comment", "require (\n example.com/unapproved v1.0.0 // indirectly required\n)", "unsanctioned direct dependency", ""},
		{"indirect", "require example.com/unapproved v1.0.0 // indirect", "", "0"},
		{"indirect explanation", "require example.com/unapproved v1.0.0 // indirect; explanation", "", "0"},
		{"mixed", "require (\n github.com/google/uuid v1.6.0\n example.com/unapproved v1.0.0 // indirect\n)", "", "1"},
		{"invalid syntax", "require (", "reading go.mod", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			tempDir := t.TempDir()
			manifest := "module example.com/guard\n\n" + testCase.requirements + "\n"
			modulePath := writeFixture(t, tempDir, "go.mod", manifest)
			allowPath := writeFixture(t, tempDir, "allowed", "# allowed modules\n github.com/google/uuid # UUIDs\n")
			var output bytes.Buffer

			err := run([]string{modulePath, allowPath}, &output)

			assertResult(t, err, output.String(), testCase.wantError, testCase.wantCount)
			unchanged, err := os.ReadFile(modulePath)
			if err != nil || string(unchanged) != manifest {
				t.Errorf("manifest was modified: content = %q, error = %v", unchanged, err)
			}
		})
	}
}

func assertResult(t *testing.T, err error, output, wantError, wantCount string) {
	t.Helper()
	if wantError != "" {
		if err == nil || !strings.Contains(err.Error(), wantError) {
			t.Errorf("error = %v, want %q", err, wantError)
		}
		if output != "" {
			t.Errorf("failed guard emitted success output: %q", output)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	want := "go.mod direct dependencies OK (" + wantCount + " sanctioned module(s) in use)\n"
	if output != want {
		t.Errorf("output = %q, want %q", output, want)
	}
}

// TestRunRejectsMissingInputs verifies absent files and empty policies fail closed.
//
// Why it matters: a missing policy must not silently permit third-party imports.
// Inputs: missing paths or an empty allowlist. Outputs: actionable errors.
// Data choice: one direct dependency makes an empty policy distinguishable from
// the valid no-dependency manifest in TestRunChecksDirectDependencies.
func TestRunRejectsMissingInputs(t *testing.T) {
	tempDir := t.TempDir()
	modulePath := writeFixture(t, tempDir, "go.mod", "module example.com/guard\nrequire example.com/unapproved v1.0.0\n")
	allowPath := writeFixture(t, tempDir, "allowed", "")
	badAllowPath := writeFixture(t, tempDir, "bad-allowed", "example.com/one example.com/two\n")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"arguments", nil, "usage:"},
		{"missing module", []string{filepath.Join(tempDir, "missing"), allowPath}, "reading go.mod"},
		{"missing allowlist", []string{modulePath, filepath.Join(tempDir, "missing")}, "reading allowlist"},
		{"empty allowlist", []string{modulePath, allowPath}, "example.com/unapproved"},
		{"bad allowlist", []string{modulePath, badAllowPath}, "invalid allowlist entry"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var output bytes.Buffer
			err := run(testCase.args, &output)
			assertResult(t, err, output.String(), testCase.want, "")
		})
	}
}

func writeFixture(t *testing.T, directory, name, content string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
