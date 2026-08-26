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

package tools

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCovercheckPolicy verifies coverage floors and test failures are enforced.
//
// Why it matters: malformed policy must not silently disable the CI gate.
// Inputs: floor files and controlled go test output. Outputs: status and message.
// Data choice: boundary values, empty policies and bad rows exercise both false
// acceptance and false rejection without recursively running the unit suite.
func TestCovercheckPolicy(t *testing.T) {
	const packageName = "github.com/Cray-HPE/cani/cmd/add"
	const zeroCoverage = "ok " + packageName + " 0.01s coverage: 0.0% of statements"
	const fullCoverage = "ok " + packageName + " 0.01s coverage: 100.0% of statements"
	cases := []struct {
		name     string
		floors   string
		output   string
		goStatus string
		wantExit int
		wantText string
	}{
		{"empty", "", zeroCoverage, "0", 0, "floors met for 0 package(s)"},
		{"comments only", "# no floors\n \t\n", zeroCoverage, "0", 0, "floors met for 0 package(s)"},
		{"zero", packageName + " 0", zeroCoverage, "0", 0, "floors met for 1 package(s)"},
		{"maximum", packageName + " 100.0 # exact boundary", fullCoverage, "0", 0, "floors met for 1 package(s)"},
		{"decimal", packageName + " 18.5", fullCoverage, "0", 0, "floors met for 1 package(s)"},
		{"drop", packageName + " 18.5", zeroCoverage, "0", 1, "DROPPED"},
		{"missing coverage", packageName + " 18.5", "ok another/package", "0", 1, "MISSING"},
		{"nonnumeric", packageName + " typo", zeroCoverage, "0", 1, "invalid coverage floor"},
		{"missing threshold", packageName, zeroCoverage, "0", 1, "invalid coverage floor"},
		{"negative", packageName + " -1", zeroCoverage, "0", 1, "invalid coverage floor"},
		{"above maximum", packageName + " 100.1", fullCoverage, "0", 1, "invalid coverage floor"},
		{"extra field", packageName + " 0 extra", zeroCoverage, "0", 1, "invalid coverage floor"},
		{"test failure", "", "FAIL unit test", "7", 7, "FAIL unit test"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			output, exitCode := runCovercheck(t, testCase.floors, testCase.output, testCase.goStatus)
			if exitCode != testCase.wantExit {
				t.Errorf("exit = %d, want %d; output:\n%s", exitCode, testCase.wantExit, output)
			}
			if !strings.Contains(output, testCase.wantText) {
				t.Errorf("output = %q, want %q", output, testCase.wantText)
			}
		})
	}
}

func runCovercheck(t *testing.T, floors, output, goStatus string) (string, int) {
	t.Helper()
	tempDir := t.TempDir()
	goStub := "#!/bin/sh\nprintf '%s\\n' \"$COVER_TEST_OUTPUT\"\nexit \"$COVER_TEST_STATUS\"\n"
	for name, content := range map[string]string{"go": goStub, "floors": floors} {
		if err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0700); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("sh", "./covercheck.sh")
	command.Env = append(os.Environ(),
		"PATH="+tempDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"COVERAGE_FLOORS="+filepath.Join(tempDir, "floors"),
		"COVER_TEST_OUTPUT="+output,
		"COVER_TEST_STATUS="+goStatus,
	)
	result, err := command.CombinedOutput()
	if err == nil {
		return string(result), 0
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("running covercheck: %v", err)
	}
	return string(result), exitError.ExitCode()
}
