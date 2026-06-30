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

package schema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// schemaFile is the committed artifact, resolved relative to this package's
// directory (the working directory used by `go test`).
const schemaFile = "inventory.schema.json"

// TestGeneratedSchemaMatchesCommittedFile verifies that the committed
// inventory.schema.json is byte-for-byte identical to the output of Generate().
//
// Why it matters: the committed file is the machine-readable contract other
// tools consume straight from the repository. If a contributor changes an
// inventory struct without regenerating, the published schema would silently
// drift from the code; this test turns that drift into a failing build and the
// schema diff becomes the human-readable record of the change.
//
// Inputs/outputs: reads the on-disk schema and compares it to freshly generated
// bytes; no inputs beyond the live struct definitions. Fails with instructions
// to run `make schema` when they differ.
//
// Data choice: the real committed artifact and the real Inventory type are used
// so the test exercises the exact contract shipped to consumers.
func TestGeneratedSchemaMatchesCommittedFile(t *testing.T) {
	want, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("reading committed schema %q: %v (run `make schema` to create it)", schemaFile, err)
	}

	got, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned an unexpected error: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("generated schema differs from %s: an inventory struct changed.\n"+
			"Run `make schema` to regenerate and commit the updated artifact.", schemaFile)
	}
}

// TestGeneratedSchemaStructure verifies the structural invariants of the
// generated schema so that a blind regeneration cannot bless a broken contract.
//
// Why it matters: the golden test alone would pass if both the generator and
// the committed file were wrong together. These assertions independently pin
// the dialect, the version stamp, the UUID-keyed collection shape, the
// flattening of embedded ObjectMeta, and the presence of every core type def,
// which are the properties consumers and downstream validators depend on.
//
// Inputs/outputs: parses Generate() output into a generic map and asserts on
// specific keys; produces only pass/fail.
//
// Data choice: assertions target representative, stable parts of the model (the
// devices collection, CaniDeviceType, the promoted status field) so they remain
// meaningful without being brittle to unrelated field additions.
func TestGeneratedSchemaStructure(t *testing.T) {
	var doc map[string]any
	if err := json.Unmarshal(mustGenerate(t), &doc); err != nil {
		t.Fatalf("generated schema is not valid JSON: %v", err)
	}

	if got := doc["$schema"]; got != dialect {
		t.Errorf("$schema = %v, want %s", got, dialect)
	}
	if got := doc["x-cani-schema-version"]; got != devicetypes.SchemaVersionV1Alpha3 {
		t.Errorf("x-cani-schema-version = %v, want %s", got, devicetypes.SchemaVersionV1Alpha3)
	}

	defs := childMap(t, doc, "$defs")
	for _, name := range []string{
		"CaniDeviceType", "CaniRackType", "CaniLocationType", "CaniModuleType",
		"CaniCableType", "CaniFruType", "CaniInterface", "CaniPrefix",
		"CaniIPAddress", "CaniVLAN", "InventoryMetadata",
	} {
		if _, ok := defs[name]; !ok {
			t.Errorf("$defs missing expected type %q", name)
		}
	}
	if _, ok := defs["ObjectMeta"]; ok {
		t.Error("$defs must not contain ObjectMeta; its fields should be flattened into each type")
	}

	assertUUIDKeyedRef(t, childMap(t, doc, "properties"), "devices", "#/$defs/CaniDeviceType")

	// The embedded ObjectMeta.Status field must be promoted onto each type.
	device := childMap(t, defs, "CaniDeviceType")
	deviceProps := childMap(t, device, "properties")
	if _, ok := deviceProps["status"]; !ok {
		t.Error("CaniDeviceType is missing the promoted 'status' property from ObjectMeta")
	}
}

// mustGenerate returns the generated schema bytes or fails the test.
func mustGenerate(t *testing.T) []byte {
	t.Helper()
	data, err := Generate()
	if err != nil {
		t.Fatalf("Generate() returned an unexpected error: %v", err)
	}
	return data
}

// childMap fetches a nested object by key, failing if it is absent or not an object.
func childMap(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()
	child, ok := parent[key].(map[string]any)
	if !ok {
		t.Fatalf("expected %q to be a JSON object, got %T", key, parent[key])
	}
	return child
}

// assertUUIDKeyedRef checks that a collection property is a UUID-keyed object
// whose values reference the expected definition.
func assertUUIDKeyedRef(t *testing.T, props map[string]any, field, ref string) {
	t.Helper()
	coll := childMap(t, props, field)
	if coll["type"] != "object" {
		t.Errorf("%s.type = %v, want object", field, coll["type"])
	}
	names := childMap(t, coll, "propertyNames")
	if names["format"] != "uuid" {
		t.Errorf("%s.propertyNames.format = %v, want uuid", field, names["format"])
	}
	values := childMap(t, coll, "additionalProperties")
	if values["$ref"] != ref {
		t.Errorf("%s.additionalProperties.$ref = %v, want %s", field, values["$ref"], ref)
	}
}
