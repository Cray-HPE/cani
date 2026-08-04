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
	"slices"
	"strings"
	"testing"
)

// TestCompareClassifiesCompatibility verifies semantic changes are grouped by
// their effect on readers and migration requirements.
//
// Why it matters: release notes must distinguish optional additions from data
// loss risks instead of presenting consumers with an undifferentiated JSON diff.
// Inputs: compact v1alpha3 and v1alpha4 schemas with added, removed, required,
// and type-changed properties. Outputs: deterministic breaking and compatible lists.
// Data choice: one property per classification makes a missing or inverted rule
// visible without relying on the much larger generated inventory schema.
func TestCompareClassifiesCompatibility(t *testing.T) {
	oldSchema := []byte(`{
		"x-cani-schema-version":"v1alpha3",
		"properties":{
			"changed":{"type":"string"},
			"removed":{"type":"boolean"},
			"stable":{"type":"string","description":"old text"}
		}
	}`)
	newSchema := []byte(`{
		"x-cani-schema-version":"v1alpha4",
		"x-cani-schema-digest":"sha256:abc",
		"required":["requiredNew"],
		"properties":{
			"added":{"type":"number"},
			"changed":{"type":"integer"},
			"requiredNew":{"type":"string"},
			"stable":{"type":"string","description":"new text"}
		}
	}`)

	report, err := Compare(oldSchema, newSchema)
	if err != nil {
		t.Fatalf("Compare() returned unexpected error: %v", err)
	}

	wantBreaking := []string{
		"added required property `/requiredNew`",
		"changed constraints for `/changed` from `{\"type\":\"string\"}` to `{\"type\":\"integer\"}`",
		"removed property `/removed`",
		"schema generation changed from `v1alpha3` to `v1alpha4`",
	}
	if !slices.Equal(report.Breaking, wantBreaking) {
		t.Errorf("Breaking = %#v, want %#v", report.Breaking, wantBreaking)
	}
	wantCompatible := []string{"added optional property `/added`"}
	if !slices.Equal(report.Compatible, wantCompatible) {
		t.Errorf("Compatible = %#v, want %#v", report.Compatible, wantCompatible)
	}
	if report.ToDigest != "sha256:abc" {
		t.Errorf("ToDigest = %q, want sha256:abc", report.ToDigest)
	}
	if !report.WireChanged {
		t.Error("WireChanged = false, want true")
	}
}

// TestValidateEvolutionRequiresGenerationBump verifies a serialized shape
// change cannot retain the same generation.
//
// Why it matters: regenerating the golden schema alone must not permit another
// release to silently change the writer contract under an existing label.
// Inputs: two v1alpha4 schemas where the latter adds one optional property.
// Outputs: an error directing the contributor to bump the generation.
// Data choice: an optional addition is accepted by older readers but dropped by
// older writers, making it the least obvious change that still needs a bump.
func TestValidateEvolutionRequiresGenerationBump(t *testing.T) {
	oldSchema := []byte(`{"x-cani-schema-version":"v1alpha4","properties":{"stable":{"type":"string"}}}`)
	newSchema := []byte(`{"x-cani-schema-version":"v1alpha4","properties":{"stable":{"type":"string"},"added":{"type":"string"}}}`)

	err := ValidateEvolution(oldSchema, newSchema)

	if err == nil || !strings.Contains(err.Error(), "without a generation bump") {
		t.Fatalf("ValidateEvolution() error = %v, want generation-bump context", err)
	}
}

// TestValidateEvolutionAllowsGenerationAndAnnotationChanges verifies valid
// evolution and documentation-only edits pass the gate.
//
// Why it matters: enforcement must prevent contract drift without blocking a
// deliberate generation change or harmless schema documentation maintenance.
// Inputs: a generation change with an added field and a same-generation title
// change. Outputs: nil for both comparisons.
// Data choice: the two cases distinguish wire contract identity from generated
// artifact byte identity.
func TestValidateEvolutionAllowsGenerationAndAnnotationChanges(t *testing.T) {
	cases := []struct {
		name string
		old  []byte
		new  []byte
	}{
		{
			name: "generation bump",
			old:  []byte(`{"x-cani-schema-version":"v1alpha3","properties":{"stable":{"type":"string"}}}`),
			new:  []byte(`{"x-cani-schema-version":"v1alpha4","properties":{"stable":{"type":"string"},"added":{"type":"string"}}}`),
		},
		{
			name: "annotation only",
			old:  []byte(`{"x-cani-schema-version":"v1alpha4","title":"old","properties":{"stable":{"type":"string"}}}`),
			new:  []byte(`{"x-cani-schema-version":"v1alpha4","title":"new","properties":{"stable":{"type":"string"}}}`),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateEvolution(tc.old, tc.new); err != nil {
				t.Fatalf("ValidateEvolution() error = %v", err)
			}
		})
	}
}

// TestSemanticDiffMarkdownDocumentsWriterRisk verifies release-note rendering
// includes the compatibility warning and both change categories.
//
// Why it matters: an optional field is not forward-compatible when an older
// writer drops unknown data, so the generated notes must not overstate safety.
// Inputs: a report with one breaking and one compatible change. Outputs: Markdown
// containing the generation, digest, warning, headings, and both entries.
// Data choice: short sentinel descriptions keep the assertion focused on report
// structure rather than the property comparison algorithm covered separately.
func TestSemanticDiffMarkdownDocumentsWriterRisk(t *testing.T) {
	report := SemanticDiff{
		FromVersion: "v1alpha3",
		ToVersion:   "v1alpha4",
		ToDigest:    "sha256:abc",
		Breaking:    []string{"breaking sentinel"},
		Compatible:  []string{"compatible sentinel"},
	}

	markdown := report.Markdown()
	for _, want := range []string{
		"Generation: `v1alpha3` -> `v1alpha4`",
		"Schema digest: `sha256:abc`",
		"older writers must not rewrite newer generations",
		"### Migration-required changes",
		"- breaking sentinel.",
		"### Reader-compatible additions",
		"- compatible sentinel.",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("Markdown() missing %q:\n%s", want, markdown)
		}
	}
}

// TestCompareRejectsInvalidSchema verifies malformed input is reported with
// which side of the comparison failed.
//
// Why it matters: release automation must fail visibly instead of publishing an
// empty or misleading compatibility report when a schema artifact is corrupt.
// Inputs: malformed old JSON and a valid empty new schema. Outputs: an error with
// old-schema decoding context.
// Data choice: the malformed opening brace reaches JSON decoding immediately and
// isolates error attribution from semantic comparison.
func TestCompareRejectsInvalidSchema(t *testing.T) {
	_, err := Compare([]byte(`{"broken"`), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "decoding old schema") {
		t.Fatalf("Compare() error = %v, want old schema decoding context", err)
	}
}
