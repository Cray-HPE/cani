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
	"testing"

	"github.com/Cray-HPE/cani/internal/provider/csm/validate/common"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	TestDataDir = "../../../../testdata/fixtures/sls"
)

func loadSchemaForTest(t *testing.T, schemafile string) (schema *jsonschema.Schema, err error) {
	subnetsSchema, err := loadSchema(schemafile)
	if err != nil {
		t.Fatalf("Unexpected error loading schema: %s, error: %s", schemafile, err)
	}
	return subnetsSchema, err
}

func unmarshalToInterfaceForTest(t *testing.T, filename string, content []byte) (RawJson, common.ValidationResult, error) {
	rawJson, result, err := unmarshalToInterface(content)
	if err != nil {
		t.Fatalf("Unexpected error unmarshaling content file: %s, error: %s", filename, err)
	}
	if result.Result != common.Pass {
		t.Fatalf("Unexpected result unmarshaling content file: %s, result: %v", filename, result)
	}
	return rawJson, result, err
}

func loadSchemaAndRawJson(t *testing.T, schemafile string, datafile string) (schema *jsonschema.Schema, rawJson RawJson) {
	schema, _ = loadSchemaForTest(t, schemafile)
	content := loadTestData(t, datafile)
	rawJson, _, _ = unmarshalToInterfaceForTest(t, datafile, content)
	return schema, rawJson
}

func logResults(t *testing.T, results []common.ValidationResult) {
	passCount := 0
	warnCount := 0
	failCount := 0
	for _, r := range results {
		switch r.Result {
		case common.Pass:
			passCount++
		case common.Warning:
			warnCount++
		case common.Fail:
			failCount++
		}
	}
	t.Logf("results: total: %d, pass: %d, warning: %d, fail: %d", len(results), passCount, warnCount, failCount)
	for _, r := range results {
		t.Logf("    %v", r)
	}
}

func TestNetworks(t *testing.T) {
	schemafile := "sls_networks_schema.json"
	datafile := "mug-dumpstate.json"
	networksSchema, rawJson := loadSchemaAndRawJson(t, schemafile, datafile)

	networks, found := common.GetMap(rawJson, "Networks")
	if !found {
		t.Fatalf("Failed to find Networks field in json data file. file: %s", datafile)
	}

	results := validateSchemaNetworks(networksSchema, networks)
	logResults(t, results)
	for _, r := range results {
		if r.Result != common.Pass {
			t.Errorf("Non passing result for datafile: %s, schemafile: %s, result: %v", datafile, schemafile, r)
		}
	}
}

func TestNetworksInvalid(t *testing.T) {
	schemafile := "sls_networks_schema.json"
	datafile := "dumpstate-invalid.json"
	networksSchema, rawJson := loadSchemaAndRawJson(t, schemafile, datafile)

	networks, found := common.GetMap(rawJson, "Networks")
	if !found {
		t.Fatalf("Failed to find Networks field in json data file. file: %s", datafile)
	}

	results := validateSchemaNetworks(networksSchema, networks)
	logResults(t, results)
	err := common.AllError(results)
	if err == nil {
		t.Errorf("Expected failures but there were none datafile: %s, schemafile: %s, results: \n%v", datafile, schemafile, results)
	}
}
