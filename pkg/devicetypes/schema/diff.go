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
	"fmt"
	"maps"
	"sort"
	"strings"
)

// SemanticDiff summarizes compatibility-relevant changes between two
// generated inventory schemas.
type SemanticDiff struct {
	FromVersion string
	ToVersion   string
	ToDigest    string
	WireChanged bool
	Breaking    []string
	Compatible  []string
}

type propertyContract struct {
	required    bool
	constraints string
}

// Compare reports property and constraint changes without being distracted by
// JSON object ordering or descriptive metadata.
func Compare(oldData, newData []byte) (SemanticDiff, error) {
	oldDoc, err := decodeSchema(oldData)
	if err != nil {
		return SemanticDiff{}, fmt.Errorf("decoding old schema: %w", err)
	}
	newDoc, err := decodeSchema(newData)
	if err != nil {
		return SemanticDiff{}, fmt.Errorf("decoding new schema: %w", err)
	}

	report := SemanticDiff{
		FromVersion: schemaGeneration(oldDoc),
		ToVersion:   schemaGeneration(newDoc),
		ToDigest:    stringValue(newDoc["x-cani-schema-digest"]),
	}
	if report.FromVersion != report.ToVersion {
		report.Breaking = append(report.Breaking, fmt.Sprintf(
			"schema generation changed from `%s` to `%s`", report.FromVersion, report.ToVersion))
	}

	oldProperties := map[string]propertyContract{}
	newProperties := map[string]propertyContract{}
	collectProperties(oldDoc, "", oldProperties)
	collectProperties(newDoc, "", newProperties)
	report.WireChanged = !maps.Equal(wireProperties(oldProperties), wireProperties(newProperties))
	compareProperties(&report, oldProperties, newProperties)
	sort.Strings(report.Breaking)
	sort.Strings(report.Compatible)
	return report, nil
}

// ValidateEvolution rejects wire-shape changes that retain the same schema
// generation. Generation-only corrections and annotation changes are allowed.
func ValidateEvolution(oldData, newData []byte) error {
	report, err := Compare(oldData, newData)
	if err != nil {
		return err
	}
	if report.WireChanged && report.FromVersion == report.ToVersion {
		return fmt.Errorf("inventory wire schema changed without a generation bump from %s", report.FromVersion)
	}
	return nil
}

func wireProperties(properties map[string]propertyContract) map[string]propertyContract {
	wire := make(map[string]propertyContract, len(properties))
	for path, contract := range properties {
		if path != "/schemaVersion" {
			wire[path] = contract
		}
	}
	return wire
}

// Markdown renders a release-note section from a semantic diff.
func (d SemanticDiff) Markdown() string {
	var out strings.Builder
	out.WriteString("## Inventory schema contract\n\n")
	fmt.Fprintf(&out, "Generation: `%s` -> `%s`\n\n", d.FromVersion, d.ToVersion)
	if d.ToDigest != "" {
		fmt.Fprintf(&out, "Schema digest: `%s`\n\n", d.ToDigest)
	}
	out.WriteString("Cani does not preserve unknown JSON fields. Reader-compatible additions are safe for older readers only; older writers must not rewrite newer generations.\n\n")
	writeChanges(&out, "Migration-required changes", d.Breaking)
	writeChanges(&out, "Reader-compatible additions", d.Compatible)
	return out.String()
}

func decodeSchema(data []byte) (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func schemaGeneration(doc map[string]any) string {
	if version := stringValue(doc["x-cani-schema-version"]); version != "" {
		return version
	}
	properties, _ := doc["properties"].(map[string]any)
	versionSchema, _ := properties["schemaVersion"].(map[string]any)
	if version := stringValue(versionSchema["const"]); version != "" {
		return version
	}
	return "unknown"
}

func collectProperties(node map[string]any, path string, out map[string]propertyContract) {
	required := requiredNames(node["required"])
	if properties, ok := node["properties"].(map[string]any); ok {
		for name, value := range properties {
			property, ok := value.(map[string]any)
			if !ok {
				continue
			}
			propertyPath := joinPointer(path, name)
			out[propertyPath] = propertyContract{
				required:    required[name],
				constraints: semanticJSON(property),
			}
			collectProperties(property, propertyPath, out)
		}
	}
	collectDefinitions(node, path, out)
	collectNestedSchema(node, "items", path, out)
	collectNestedSchema(node, "additionalProperties", path, out)
}

func collectDefinitions(node map[string]any, path string, out map[string]propertyContract) {
	definitions, ok := node["$defs"].(map[string]any)
	if !ok {
		return
	}
	defsPath := joinPointer(path, "$defs")
	for name, value := range definitions {
		definition, ok := value.(map[string]any)
		if ok {
			collectProperties(definition, joinPointer(defsPath, name), out)
		}
	}
}

func collectNestedSchema(node map[string]any, key, path string, out map[string]propertyContract) {
	nested, ok := node[key].(map[string]any)
	if ok {
		collectProperties(nested, joinPointer(path, key), out)
	}
}

func compareProperties(report *SemanticDiff, oldProperties, newProperties map[string]propertyContract) {
	paths := map[string]bool{}
	for path := range oldProperties {
		paths[path] = true
	}
	for path := range newProperties {
		paths[path] = true
	}
	for path := range paths {
		compareProperty(report, path, oldProperties, newProperties)
	}
}

func compareProperty(report *SemanticDiff, path string, oldProperties, newProperties map[string]propertyContract) {
	oldProperty, hadOld := oldProperties[path]
	newProperty, hasNew := newProperties[path]
	switch {
	case !hasNew:
		report.Breaking = append(report.Breaking, fmt.Sprintf("removed property `%s`", path))
	case !hadOld && newProperty.required:
		report.Breaking = append(report.Breaking, fmt.Sprintf("added required property `%s`", path))
	case !hadOld:
		report.Compatible = append(report.Compatible, fmt.Sprintf("added optional property `%s`", path))
	default:
		compareExistingProperty(report, path, oldProperty, newProperty)
	}
}

func compareExistingProperty(report *SemanticDiff, path string, oldProperty, newProperty propertyContract) {
	if oldProperty.required != newProperty.required {
		if newProperty.required {
			report.Breaking = append(report.Breaking, fmt.Sprintf("made property `%s` required", path))
		} else {
			report.Compatible = append(report.Compatible, fmt.Sprintf("made property `%s` optional", path))
		}
	}
	if oldProperty.constraints != newProperty.constraints {
		report.Breaking = append(report.Breaking, fmt.Sprintf(
			"changed constraints for `%s` from `%s` to `%s`",
			path, oldProperty.constraints, newProperty.constraints))
	}
}

func semanticJSON(value any) string {
	normalized := semanticValue(value)
	data, _ := json.Marshal(normalized)
	return string(data)
}

func semanticValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for key, child := range typed {
			if !ignoredSemanticKey(key) {
				out[key] = semanticValue(child)
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			out[i] = semanticValue(child)
		}
		return out
	default:
		return value
	}
}

func ignoredSemanticKey(key string) bool {
	switch key {
	case "$defs", "$id", "$comment", "description", "title", "default",
		"examples", "properties", "required", "x-cani-schema-digest",
		"x-cani-schema-version":
		return true
	default:
		return false
	}
}

func requiredNames(value any) map[string]bool {
	out := map[string]bool{}
	items, _ := value.([]any)
	for _, item := range items {
		if name := stringValue(item); name != "" {
			out[name] = true
		}
	}
	return out
}

func joinPointer(path, token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	token = strings.ReplaceAll(token, "/", "~1")
	return path + "/" + token
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func writeChanges(out *strings.Builder, heading string, changes []string) {
	fmt.Fprintf(out, "### %s\n\n", heading)
	if len(changes) == 0 {
		out.WriteString("- None.\n\n")
		return
	}
	for _, change := range changes {
		fmt.Fprintf(out, "- %s.\n", change)
	}
	out.WriteString("\n")
}
