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
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// cmdRoot locates the command layer relative to this package. go test always
// runs with the working directory set to the package under test.
const cmdRoot = "../../../cmd"

// providerRoot locates the provider implementations relative to this package.
const providerRoot = "../../../pkg/provider"

// requiredVerbs must appear among the scanned packages. They are the tripwire:
// if this package moves, or a verb is renamed, the scan silently covering
// nothing would otherwise look like a pass.
var requiredVerbs = []string{"add", "remove", "show", "update"}

// ordinaryWords are provider directory names that are also ordinary English
// words, so their appearance in cmd/ says nothing about provider coupling.
var ordinaryWords = map[string]bool{"example": true}

// providerSlugs reads the concrete provider names from the directory that holds
// them, so a provider added tomorrow is covered the day it appears rather than
// when somebody remembers to extend a hard-coded list here.
func providerSlugs(t *testing.T) map[string]bool {
	t.Helper()

	entries, err := os.ReadDir(providerRoot)
	if err != nil {
		t.Fatalf("reading %s: %v (has this package moved?)", providerRoot, err)
	}

	slugs := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() || ordinaryWords[entry.Name()] {
			continue
		}
		slugs[entry.Name()] = true
	}

	// An empty set would make every scan below pass for the wrong reason.
	if len(slugs) == 0 {
		t.Fatalf("no providers discovered under %s", providerRoot)
	}
	return slugs
}

// TestCmdHasNoProviderNameLiterals verifies no package under cmd/ contains a
// concrete provider's name as a string literal.
//
// Why it matters: provider identity must not control generic CLI behavior.
// Inputs: non-test Go files in cmd/ and all descendants, plus discovered slugs.
// Outputs: failures naming each literal and any required package not scanned.
// Data choice: root and nested packages are included so new command paths get
// the same protection as the CRUD verbs without maintaining a package list.
func TestCmdHasNoProviderNameLiterals(t *testing.T) {
	scanned, literals, err := scanCommandTree(cmdRoot, providerSlugs(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, literal := range literals {
		t.Errorf("%s: provider name %s used as a string literal in the command layer; "+
			"cmd/ must not branch on provider identity", literal.position, literal.value)
	}
	assertScannedDirectories(t, scanned, cmdRoot, append([]string{"."}, requiredVerbs...))
}

type providerLiteral struct {
	position token.Position
	value    string
}

func scanCommandTree(root string, slugs map[string]bool) (map[string]bool, []providerLiteral, error) {
	scanned := make(map[string]bool)
	var literals []providerLiteral
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		found, err := scanProviderFile(path, slugs)
		if err != nil {
			return err
		}
		scanned[filepath.Dir(path)] = true
		literals = append(literals, found...)
		return nil
	})
	return scanned, literals, err
}

func scanProviderFile(path string, slugs map[string]bool) ([]providerLiteral, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	var literals []providerLiteral
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, _ := strconv.Unquote(literal.Value)
		if slugs[strings.ToLower(value)] {
			literals = append(literals, providerLiteral{fset.Position(literal.Pos()), literal.Value})
		}
		return true
	})
	return literals, nil
}

// TestProviderNameScanCoversCommandTree verifies discovery and literal matching.
//
// Why it matters: a guard that skips root or nested commands can pass falsely.
// Inputs: a temporary command tree with literals, comments and ignored files.
// Outputs: exact offending paths and source positions, with no false positives.
// Data choice: immediate, root, nested, raw and escaped literals exercise the
// tree traversal and Go string decoding independently of real command sources.
func TestProviderNameScanCoversCommandTree(t *testing.T) {
	root := t.TempDir()
	fixtures := map[string]string{
		"root.go":                "package fixture\nconst name = \"csm\"\n",
		"add/device.go":          "package fixture\nconst name = `csm`\n",
		"alpha/nested/device.go": "package fixture\nconst name = \"CSM\"\n",
		"alpha/nested/escape.go": "package fixture\nconst name = \"\\x63sm\"\n",
		"update/clean.go":        "package fixture\n// csm\nconst csm = \"ordinary text\"\n",
		"root_test.go":           "ignored invalid Go containing csm",
		"alpha/nested/a_test.go": "ignored invalid Go containing csm",
		"README.md":              "csm",
	}
	for name, source := range fixtures {
		writeCommandSource(t, root, name, source)
	}

	scanned, literals, err := scanCommandTree(root, map[string]bool{"csm": true})
	if err != nil {
		t.Fatal(err)
	}
	var gotPaths []string
	for _, literal := range literals {
		gotPaths = append(gotPaths, literal.position.Filename)
		if literal.position.Line != 2 {
			t.Errorf("position = %v, want line 2", literal.position)
		}
	}
	var wantPaths []string
	for _, name := range []string{"root.go", "add/device.go", "alpha/nested/device.go", "alpha/nested/escape.go"} {
		wantPaths = append(wantPaths, filepath.Join(root, name))
	}
	slices.Sort(gotPaths)
	slices.Sort(wantPaths)
	if !slices.Equal(gotPaths, wantPaths) {
		t.Errorf("offending paths = %v, want %v", gotPaths, wantPaths)
	}
	assertScannedDirectories(t, scanned, root, []string{".", "add", "alpha/nested", "update"})
}

func assertScannedDirectories(t *testing.T, scanned map[string]bool, root string, directories []string) {
	t.Helper()
	for _, directory := range directories {
		if !scanned[filepath.Join(root, directory)] {
			t.Errorf("directory %s was not scanned", directory)
		}
	}
}

// TestProviderNameScanRejectsInvalidSource verifies scan errors cannot pass.
//
// Why it matters: unreadable or malformed command source must not bypass guards.
// Inputs: a missing root and invalid nested Go. Outputs: non-nil scan errors.
// Data choice: both traversal and parsing failures are checked separately.
func TestProviderNameScanRejectsInvalidSource(t *testing.T) {
	root := t.TempDir()
	writeCommandSource(t, root, "nested/invalid.go", "not valid Go")
	for _, directory := range []string{filepath.Join(root, "missing"), root} {
		if _, _, err := scanCommandTree(directory, map[string]bool{"csm": true}); err == nil {
			t.Errorf("scan %s succeeded, want an error", directory)
		}
	}
}

func writeCommandSource(t *testing.T, root, name, source string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}
