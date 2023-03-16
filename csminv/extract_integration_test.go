/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package csminv

import (
	"bytes"
	"io"
	"testing"
)

const (
	system_config_valid = "testdata/fixtures/csi/system_config_valid.yaml"
	paddle_valid        = "testdata/fixtures/canu/paddle.json"
	sls_input_valid     = "testdata/fixtures/sls/sls_input_file_valid.json"
)

// input files to test against and the output expected after running the command
var tests = []struct {
	args     []string
	expected string
}{
	{
		args:     []string{system_config_valid},
		expected: "system_config_valid.yaml",
	},
	{
		args:     []string{"extract", paddle_valid},
		expected: "paddle_valid.json",
	},
	{
		args:     []string{"extract", sls_input_valid},
		expected: "sls_input_valid.json",
	},
}

func Test_Extract(t *testing.T) {
	for _, test := range tests {
		// Create a buffer to store the output
		// b := bytes.NewBufferString("sdsd")
		b := new(bytes.Buffer)
		// Set the output to the buffer
		rootCmd.SetOut(b)
		rootCmd.SetErr(b)
		// Set the args for this test
		rootCmd.SetArgs([]string{"extract", "--help"})
		// Execute the command
		err := extractCmd.Execute()
		if err != nil {
			t.Fatal(err)
		}
		// Read the output
		actual, err := io.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		// // Assert the output matches what is expected
		// assert.Equal(t, actual, test.expected)
		t.Log("actia", actual, test.expected)
	}
}
