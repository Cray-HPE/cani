/*
MIT License

(C) Copyright 2023 Hewlett Packard Enterprise Development LP

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

package validate

import (
	"os"
	"testing"
)

func loadTestData(t *testing.T, name string) []byte {
	content, err := os.ReadFile(TestDataDir + "/" + name)
	if err != nil {
		t.Fatalf("Failed to load file %s. error: %v", name, err)
	}
	return content
}

func TestUnmarshalToString(t *testing.T) {
	datafile := "mug-dumpstate.json"
	content := loadTestData(t, datafile)

	raw, result, err := unmarshalToInterface(content)
	if err != nil {
		t.Fatalf("Failed to unmarshal %s. error: %s", datafile, err)
	}

	if result.Result != Pass {
		t.Fatalf("Failed to unmarshal %s. result: %v, error: %s", datafile, result, err)
	}

	if raw == nil {
		t.Fatalf("Failed to unmarshal %s. the returned interface{} is nil", datafile)
	}
}

func TestUnmarshalToSlsState(t *testing.T) {
	datafile := "mug-dumpstate.json"
	content := loadTestData(t, datafile)
	slsState, result, err := unmarshalToSlsState(content)

	if err != nil {
		t.Fatalf("failed to unmarshal %s. error: %s", datafile, err)
	}

	if result.Result != Pass {
		t.Fatalf("Failed to unmarshal %s. result: %v, error: %s", datafile, result, err)
	}

	if slsState == nil {
		t.Fatalf("Failed to unmarshal %s. the returned slsState is nil", datafile)
	}

	if len(slsState.Hardware) == 0 {
		t.Errorf("Failed to unmarshal %s. Found zero hardware", datafile)
	}

	if len(slsState.Networks) == 0 {
		t.Errorf("Failed to unmarshal %s. Found zero networks", datafile)
	}
}
