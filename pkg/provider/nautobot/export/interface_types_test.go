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
package export

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

// generatedInterfaceTypes parses the pinned Nautobot client and returns the
// string values of every InterfaceTypeValue constant.
func generatedInterfaceTypes(t *testing.T) map[string]struct{} {
	t.Helper()
	source := filepath.Join("..", "..", "..", "nautobot", "nautobot_api.go")
	file, err := parser.ParseFile(token.NewFileSet(), source, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse generated client: %v", err)
	}
	values := make(map[string]struct{})
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		collectInterfaceTypeValues(gen, values)
	}
	return values
}

func collectInterfaceTypeValues(gen *ast.GenDecl, values map[string]struct{}) {
	for _, spec := range gen.Specs {
		if value, ok := interfaceTypeConstValue(spec); ok {
			values[value] = struct{}{}
		}
	}
}

// interfaceTypeConstValue returns the string literal of a `X InterfaceTypeValue = "..."` spec.
func interfaceTypeConstValue(spec ast.Spec) (string, bool) {
	value, ok := spec.(*ast.ValueSpec)
	if !ok || len(value.Values) != 1 {
		return "", false
	}
	if ident, ok := value.Type.(*ast.Ident); !ok || ident.Name != "InterfaceTypeValue" {
		return "", false
	}
	lit, ok := value.Values[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	unquoted, err := strconv.Unquote(lit.Value)
	return unquoted, err == nil
}

// TestNautobotInterfaceTypesMatchGeneratedEnum verifies the hand-kept
// nautobotInterfaceTypes set equals the InterfaceTypeValue enum in the
// generated client, in both directions.
//
// Why it matters: the exporter decides which library types pass through
// unchanged from this set; a stale entry would send a value Nautobot rejects,
// a missing one would rewrite or skip a port Nautobot accepts.
// Inputs: the parsed constants of pkg/nautobot/nautobot_api.go and the set.
// Outputs: no value present in only one of the two.
// Data choice: parsing the generated source is the only enumerable form of the
// enum, since Go constants cannot be listed at run time.
func TestNautobotInterfaceTypesMatchGeneratedEnum(t *testing.T) {
	generated := generatedInterfaceTypes(t)
	if len(generated) == 0 {
		t.Fatal("no InterfaceTypeValue constants found in the generated client")
	}
	for value := range generated {
		if !isValidNautobotInterfaceType(value) {
			t.Errorf("generated enum value %q missing from nautobotInterfaceTypes", value)
		}
	}
	for value := range nautobotInterfaceTypes {
		if _, ok := generated[value]; !ok {
			t.Errorf("nautobotInterfaceTypes has %q, which the generated client lacks", value)
		}
	}
}

// TestMapInterfaceTypePassesThroughEveryNautobotValue verifies mapInterfaceType
// returns every Nautobot enum value unchanged, including case-folded input.
//
// Why it matters: the previous substring matcher folded distinct form factors
// (QSFP-DD onto OSFP, XFP onto SFP+) that Nautobot distinguishes.
// Inputs: every value in nautobotInterfaceTypes, plus "10GBASE-X-XFP".
// Outputs: the same value, lower-cased.
// Data choice: iterating the whole set guards every entry, and the XFP case
// was a silent casualty of the old "10gbase-x" substring rule.
func TestMapInterfaceTypePassesThroughEveryNautobotValue(t *testing.T) {
	for value := range nautobotInterfaceTypes {
		if got := mapInterfaceType(value); got != value {
			t.Errorf("mapInterfaceType(%q) = %q, want unchanged", value, got)
		}
	}
	if got := mapInterfaceType("10GBASE-X-XFP"); got != "10gbase-x-xfp" {
		t.Errorf("mapInterfaceType(10GBASE-X-XFP) = %q, want 10gbase-x-xfp", got)
	}
}

// TestMapInterfaceTypeAliasesAndUnknown verifies the alias table rewrites only
// spellings Nautobot lacks and unknown types are returned verbatim.
//
// Why it matters: callers validate the result with isValidNautobotInterfaceType,
// so an unknown type must survive untouched for them to report it.
// Inputs: "1gbase-t", " nvlink " and "". Outputs: "1000base-t", " nvlink " and
// "1000base-t".
// Data choice: the alias is the one remaining non-identity mapping; nvlink is a
// real library type Nautobot lacks; the empty default is legacy behaviour.
func TestMapInterfaceTypeAliasesAndUnknown(t *testing.T) {
	if got := mapInterfaceType("1gbase-t"); got != "1000base-t" {
		t.Errorf("alias 1gbase-t = %q, want 1000base-t", got)
	}
	if got := mapInterfaceType(" nvlink "); got != " nvlink " {
		t.Errorf("unknown type = %q, want verbatim input", got)
	}
	if got := mapInterfaceType(""); got != "1000base-t" {
		t.Errorf("empty type = %q, want 1000base-t", got)
	}
}

// TestSupportedInterfaceSpecsDropsAndCountsUnsupported verifies
// supportedInterfaceSpecs keeps specs whose type Nautobot stores, in order,
// and counts each dropped spec in IfacesUnsupported.
//
// Why it matters: both the device and the module export paths run this
// filter before dry-run reporting and before any create, so it alone decides
// which ports reach Nautobot and how many omissions the summary reports.
// Inputs: eth0 (1000base-t), nv0 (nvlink) and hsn0 (400gbase-x-qsfpdd).
// Outputs: [eth0, hsn0] and IfacesUnsupported == 1.
// Data choice: the unsupported spec sits between two supported ones to prove
// the filter preserves order rather than truncating at the first miss.
func TestSupportedInterfaceSpecsDropsAndCountsUnsupported(t *testing.T) {
	specs := []interfaceSpec{
		{Name: "eth0", Type: "1000base-t"},
		{Name: "nv0", Type: "nvlink"},
		{Name: "hsn0", Type: "400gbase-x-qsfpdd"},
	}
	result := &LoadResult{}

	kept := supportedInterfaceSpecs(specs, "gpu-node", result)

	if len(kept) != 2 || kept[0].Name != "eth0" || kept[1].Name != "hsn0" {
		t.Errorf("kept = %+v, want eth0 then hsn0", kept)
	}
	if result.IfacesUnsupported != 1 {
		t.Errorf("IfacesUnsupported = %d, want 1", result.IfacesUnsupported)
	}
}
