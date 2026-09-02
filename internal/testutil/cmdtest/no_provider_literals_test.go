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

package cmdtest_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
	"testing"
)

// crudPackages are the verbs required to work on the portable model alone,
// relative to this package's directory.
var crudPackages = []string{
	"../../../cmd/add",
	"../../../cmd/remove",
	"../../../cmd/update",
	"../../../cmd/show",
}

// providerSlugs are the concrete providers whose names must never steer CRUD.
var providerSlugs = map[string]bool{
	"csm":      true,
	"hpcm":     true,
	"nautobot": true,
	"ochami":   true,
	"redfish":  true,
}

// TestCRUDHasNoProviderNameLiterals verifies no CRUD source file contains a
// concrete provider's name as a string literal.
//
// Why it matters: RequireNoProviders only notices a provider that registers
// itself, which requires an import. Branching on a name — `if p == "csm"` —
// needs no import and would slip past it entirely, while doing exactly the
// thing the portable model forbids: making a verb behave differently for one
// provider. This closes that blind spot.
// Inputs: every non-test .go file under cmd/add, cmd/remove, cmd/update and
// cmd/show, parsed rather than grepped so comments and identifiers are ignored.
// Outputs: a failure naming the file, position and offending literal.
// Data choice: the five concrete provider slugs are checked; "example" is
// omitted because it is an ordinary English word that would false-positive.
func TestCRUDHasNoProviderNameLiterals(t *testing.T) {
	scanned := 0

	for _, pkgDir := range crudPackages {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, pkgDir, func(info fs.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", pkgDir, err)
		}

		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				scanned++
				ast.Inspect(file, func(node ast.Node) bool {
					lit, ok := node.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						return true
					}
					value := strings.ToLower(strings.Trim(lit.Value, "`\""))
					if providerSlugs[value] {
						t.Errorf("%s: provider name %q used as a string literal in a CRUD command; "+
							"CRUD must not branch on provider identity",
							fset.Position(lit.Pos()), value)
					}
					return true
				})
			}
		}
	}

	if scanned == 0 {
		t.Fatal("no CRUD source files were scanned; the package paths are probably wrong")
	}
}
