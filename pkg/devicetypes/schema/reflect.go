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
	"reflect"
	"strings"
)

// uuidPkgPath is the import path of the UUID library used for IDs across the
// inventory model; UUID values serialize as strings.
const uuidPkgPath = "github.com/google/uuid"

// generator accumulates reusable type definitions while walking the struct
// tree so shared types are emitted once under $defs and referenced via $ref.
type generator struct {
	defs map[string]any  // type name -> schema node, emitted under $defs
	seen map[string]bool // guards against infinite recursion on cyclic types
}

func newGenerator() *generator {
	return &generator{defs: map[string]any{}, seen: map[string]bool{}}
}

// schemaFor returns the JSON Schema node for an arbitrary Go type.
func (g *generator) schemaFor(t reflect.Type) any {
	t = deref(t)
	if s, ok := specialScalar(t); ok {
		return s
	}
	switch t.Kind() {
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": g.schemaFor(t.Elem())}
	case reflect.Map:
		return g.mapSchema(t)
	case reflect.Struct:
		return g.structRef(t)
	default:
		return map[string]any{} // interface{} and anything else: accept any value
	}
}

// mapSchema models a Go map as a JSON object with typed values. UUID-keyed
// maps advertise their key format via propertyNames.
func (g *generator) mapSchema(t reflect.Type) map[string]any {
	out := map[string]any{
		"type":                 "object",
		"additionalProperties": g.schemaFor(t.Elem()),
	}
	if isUUID(t.Key()) {
		out["propertyNames"] = map[string]any{"format": "uuid"}
	}
	return out
}

// structRef registers a named struct under $defs (once) and returns a $ref to
// it. Anonymous structs are inlined.
func (g *generator) structRef(t reflect.Type) map[string]any {
	name := t.Name()
	if name == "" {
		return g.structSchema(t)
	}
	if !g.seen[name] {
		g.seen[name] = true // mark before recursing so cyclic types terminate
		g.defs[name] = g.structSchema(t)
	}
	return map[string]any{"$ref": "#/$defs/" + name}
}

// structSchema builds the object schema for a struct type, flattening embedded
// metadata structs exactly as encoding/json promotes them.
//
// The schema intentionally omits a "required" list. Go's encoding/json accepts
// datastores with any field absent (missing fields decode to zero values), and
// a struct tag's omitempty governs only write output, not what makes a valid
// document. Deriving "required" from omitempty would reject real, loadable
// datastores (e.g. hand-authored or migrated files that omit partNumber), so
// the artifact validates field types and structure only.
func (g *generator) structSchema(t reflect.Type) map[string]any {
	props := map[string]any{}
	g.collectFields(t, props)
	return map[string]any{"type": "object", "properties": props}
}

// collectFields walks a struct's fields, promoting embedded structs and
// honoring json tags, to populate the properties map.
func (g *generator) collectFields(t reflect.Type, props map[string]any) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Tag.Get("json") == "" && isStruct(f.Type) {
			g.collectFields(deref(f.Type), props)
			continue
		}
		if f.PkgPath != "" {
			continue // unexported: encoding/json ignores it
		}
		name, skip := jsonPropertyName(f)
		if skip {
			continue
		}
		props[name] = g.schemaFor(f.Type)
	}
}

// jsonPropertyName resolves a field's JSON property name and whether it is
// skipped entirely (json:"-").
func jsonPropertyName(f reflect.StructField) (name string, skip bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", true
	}
	if i := strings.IndexByte(tag, ','); i >= 0 {
		tag = tag[:i]
	}
	if tag == "" {
		return f.Name, false
	}
	return tag, false
}

// specialScalar maps well-known library types to their JSON string forms,
// taking priority over their underlying reflect.Kind.
func specialScalar(t reflect.Type) (map[string]any, bool) {
	switch t.PkgPath() + "." + t.Name() {
	case uuidPkgPath + ".UUID":
		return map[string]any{"type": "string", "format": "uuid"}, true
	case "time.Time":
		return map[string]any{"type": "string", "format": "date-time"}, true
	}
	return nil, false
}

// isUUID reports whether t is the uuid.UUID type.
func isUUID(t reflect.Type) bool {
	return t.PkgPath() == uuidPkgPath && t.Name() == "UUID"
}

// isStruct reports whether t (after dereferencing pointers) is a struct.
func isStruct(t reflect.Type) bool {
	return deref(t).Kind() == reflect.Struct
}

// deref unwraps pointer indirection to reach the underlying element type.
func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
