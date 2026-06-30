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

// Package schema generates a machine-readable JSON Schema (Draft 2020-12)
// describing the on-disk JSON form of the cani inventory datastore. The schema
// is produced by reflecting over the devicetypes.Inventory struct so it can
// never drift from the Go source of truth, and is committed to the repository
// so consumers can validate datastores without building or running cani.
package schema

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

const (
	// dialect is the JSON Schema dialect the generated artifact conforms to.
	dialect = "https://json-schema.org/draft/2020-12/schema"
	// schemaID is the canonical $id consumers use to reference the schema.
	schemaID = "https://raw.githubusercontent.com/Cray-HPE/cani/refs/heads/main/pkg/devicetypes/schema/inventory.schema.json"
	// title is the human-readable schema title.
	title = "cani inventory datastore"
	// description documents the artifact's provenance for human readers.
	description = "Machine-readable schema for the cani inventory datastore. " +
		"Generated from the Go structs in pkg/devicetypes by tools/genschema; do not edit by hand."
)

// Generate builds the JSON Schema describing devicetypes.Inventory and returns
// it as indented JSON bytes with a trailing newline. It reflects over the live
// struct definitions, so the result always matches the current Go types. The
// output is deterministic: object keys are sorted and required lists ordered.
func Generate() ([]byte, error) {
	g := newGenerator()
	root := g.structSchema(reflect.TypeOf(devicetypes.Inventory{}))
	root["$schema"] = dialect
	root["$id"] = schemaID
	root["title"] = title
	root["description"] = description
	root["x-cani-schema-version"] = devicetypes.SchemaVersionV1Alpha3
	if len(g.defs) > 0 {
		root["$defs"] = g.defs
	}
	return marshal(root)
}

// marshal renders a schema node as indented JSON with HTML escaping disabled
// and deterministic (sorted) object keys.
func marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
