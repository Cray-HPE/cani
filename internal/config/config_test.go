/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
package config

// Test coverage for config.go
//
// | Function       | Happy-path test                  | Failure test                       |
// |----------------|----------------------------------|------------------------------------|
// | GetNestedValue | TestGetNestedValueReturnsDeepKey  | TestGetNestedValueMissingProvider   |

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// ---------- GetNestedValue ----------

// TestGetNestedValueReturnsDeepKey verifies nested provider map values are found.
//
// Why it matters: provider code reads config through this helper instead of raw maps.
// Inputs: a provider with top-level and nested import settings. Outputs: found values with ok=true.
// Data choice: string values make exact key traversal failures easy to diagnose.
func TestGetNestedValueReturnsDeepKey(t *testing.T) {
	// Save and restore global singleton.
	orig := Cfg
	defer func() { Cfg = orig }()

	Cfg = &Config{
		Providers: map[string]map[string]any{
			"csm": {
				"provider_host": "api.example.com",
				"import": map[string]any{
					"source": "sls",
				},
			},
		},
	}

	// Top-level key.
	val, ok := GetNestedValue("csm", "provider_host")
	if !ok {
		t.Fatal("expected ok for existing top-level key")
	}
	if val != "api.example.com" {
		t.Errorf("expected api.example.com, got %v", val)
	}

	// Nested key.
	val, ok = GetNestedValue("csm", "import", "source")
	if !ok {
		t.Fatal("expected ok for nested key")
	}
	if val != "sls" {
		t.Errorf("expected sls, got %v", val)
	}
}

// TestGetNestedValueMissingProvider verifies missing providers, keys, and globals return false.
//
// Why it matters: callers rely on absent config paths falling back without panics.
// Inputs: empty, partial, and nil global configs. Outputs: ok=false for every missing lookup.
// Data choice: a single known provider isolates missing-key behavior from provider lookup behavior.
func TestGetNestedValueMissingProvider(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	Cfg = &Config{
		Providers: map[string]map[string]any{},
	}

	// Unknown provider.
	_, ok := GetNestedValue("nonexistent", "key")
	if ok {
		t.Error("expected false for missing provider")
	}

	// Known provider, missing key.
	Cfg.Providers["csm"] = map[string]any{"host": "x"}
	_, ok = GetNestedValue("csm", "missing_key")
	if ok {
		t.Error("expected false for missing key")
	}

	// Nil Cfg.
	Cfg = nil
	_, ok = GetNestedValue("csm", "anything")
	if ok {
		t.Error("expected false when Cfg is nil")
	}
}

// ---------- GetNestedString ----------

// TestGetNestedStringReturnsValue verifies string lookups return configured values.
//
// Why it matters: provider defaults should only be used when a string setting is absent or invalid.
// Inputs: top-level and nested string settings. Outputs: the exact configured strings.
// Data choice: Nautobot-shaped keys mirror real provider configuration paths.
func TestGetNestedStringReturnsValue(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	Cfg = &Config{
		Providers: map[string]map[string]any{
			"nautobot": {
				"provider_host": "nautobot.example.com",
				"export": map[string]any{
					"default_status": "Active",
				},
			},
		},
	}

	// Top-level string key.
	got := GetNestedString("nautobot", "fallback", "provider_host")
	if got != "nautobot.example.com" {
		t.Errorf("expected nautobot.example.com, got %s", got)
	}

	// Nested string key.
	got = GetNestedString("nautobot", "fallback", "export", "default_status")
	if got != "Active" {
		t.Errorf("expected Active, got %s", got)
	}
}

// TestGetNestedStringReturnsFallback verifies missing or non-string values use the fallback.
//
// Why it matters: config consumers need stable defaults when YAML contains unexpected types.
// Inputs: a missing key, an integer value, and an unknown provider. Outputs: the supplied fallback.
// Data choice: an integer exercises the type guard without relying on YAML decoding.
func TestGetNestedStringReturnsFallback(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	Cfg = &Config{
		Providers: map[string]map[string]any{
			"nautobot": {
				"port": 8080, // not a string
			},
		},
	}

	// Missing key returns default.
	got := GetNestedString("nautobot", "default-val", "missing")
	if got != "default-val" {
		t.Errorf("expected default-val for missing key, got %s", got)
	}

	// Non-string value returns default.
	got = GetNestedString("nautobot", "default-val", "port")
	if got != "default-val" {
		t.Errorf("expected default-val for non-string value, got %s", got)
	}

	// Missing provider returns default.
	got = GetNestedString("unknown", "fallback", "key")
	if got != "fallback" {
		t.Errorf("expected fallback for unknown provider, got %s", got)
	}
}

// ---------- GetNestedInt ----------

// TestGetNestedIntReturnsValue verifies integer lookups accept int and JSON-style float64 values.
//
// Why it matters: config values may come from multiple decoders before callers request integers.
// Inputs: int and float64 provider settings. Outputs: integer values returned without fallback.
// Data choice: port and timeout names represent common numeric config fields.
func TestGetNestedIntReturnsValue(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	Cfg = &Config{
		Providers: map[string]map[string]any{
			"nautobot": {
				"port":    8443,
				"timeout": float64(30), // simulates JSON decode
			},
		},
	}

	got := GetNestedInt("nautobot", 0, "port")
	if got != 8443 {
		t.Errorf("expected 8443, got %d", got)
	}

	// float64 coercion path.
	got = GetNestedInt("nautobot", 0, "timeout")
	if got != 30 {
		t.Errorf("expected 30, got %d", got)
	}
}

// TestGetNestedIntReturnsFallback verifies missing or non-numeric values use the fallback.
//
// Why it matters: callers should not treat malformed numeric config as a real value.
// Inputs: a string value and a missing key. Outputs: the supplied integer fallback.
// Data choice: a host string is a realistic wrong type for an integer accessor.
func TestGetNestedIntReturnsFallback(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	Cfg = &Config{
		Providers: map[string]map[string]any{
			"nautobot": {
				"host": "localhost", // string, not int
			},
		},
	}

	// Wrong type returns default.
	got := GetNestedInt("nautobot", 42, "host")
	if got != 42 {
		t.Errorf("expected 42 for string value, got %d", got)
	}

	// Missing key returns default.
	got = GetNestedInt("nautobot", 99, "missing_key")
	if got != 99 {
		t.Errorf("expected 99 for missing key, got %d", got)
	}
}

// ---------- Load ----------

// TestLoadInvalidYAML verifies invalid YAML returns an error from Load.
//
// Why it matters: malformed config files must fail clearly before mutating global config state further.
// Inputs: a temporary file containing invalid YAML. Outputs: a non-nil Load error.
// Data choice: mixed punctuation and indentation creates a parser-level failure.
func TestLoadInvalidYAML(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	// Write invalid YAML to a temp file.
	tmp := t.TempDir()
	path := tmp + "/bad.yaml"
	if err := os.WriteFile(path, []byte(":::not valid yaml\n\t{["), 0644); err != nil {
		t.Fatal(err)
	}

	err := Load(path)
	if err == nil {
		t.Fatal("expected Load to return an error for invalid YAML")
	}
}

// TestLoadCreatesDefaultConfig verifies missing config files are created with defaults.
//
// Why it matters: first-run CLI behavior depends on Load being able to bootstrap configuration.
// Inputs: a path that does not exist in a temporary directory. Outputs: a file and populated Cfg.
// Data choice: an empty temp directory isolates creation from any user config on disk.
func TestLoadCreatesDefaultConfig(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	tmp := t.TempDir()
	path := tmp + "/new-config.yaml"

	// File does not exist yet – Load should create it.
	err := Load(path)
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}

	// Verify the file was created.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected config file to be created")
	}

	// Verify Cfg is populated.
	if Cfg == nil {
		t.Fatal("expected Cfg to be non-nil")
	}
	if Cfg.Providers == nil {
		t.Fatal("expected Cfg.Providers to be non-nil")
	}
}

// ---------- Comments ----------

type commentTagFixture struct {
	Scalar  string `yaml:"scalar,omitempty" head_comment:"scalar head" line_comment:"scalar line" foot_comment:"scalar foot"`
	Ignored string `yaml:"-" head_comment:"ignored head"`
	Plain   string `yaml:"plain"`
	Inline  string `yaml:",inline" line_comment:"inline line"`
}

// TestExtractCommentsReadsYAMLCommentTags verifies comment tags are extracted
// only for YAML fields that have a concrete key and at least one comment.
//
// Why it matters: Save relies on this metadata to restore comments while preserving user YAML.
// Inputs: a struct with commented, ignored, plain, and inline-style YAML fields. Outputs: one FieldComment entry.
// Data choice: the fields cover the tag forms accepted or skipped by config structs.
func TestExtractCommentsReadsYAMLCommentTags(t *testing.T) {
	comments := extractComments(&commentTagFixture{})

	got, ok := comments["scalar"]
	if !ok {
		t.Fatal("expected scalar comments to be extracted")
	}
	if got.HeadComment != "scalar head" {
		t.Errorf("HeadComment = %q, want %q", got.HeadComment, "scalar head")
	}
	if got.LineComment != "scalar line" {
		t.Errorf("LineComment = %q, want %q", got.LineComment, "scalar line")
	}
	if got.FootComment != "scalar foot" {
		t.Errorf("FootComment = %q, want %q", got.FootComment, "scalar foot")
	}

	for _, key := range []string{"-", "plain", ""} {
		if _, ok := comments[key]; ok {
			t.Errorf("did not expect comments for skipped key %q", key)
		}
	}
	if len(comments) != 1 {
		t.Fatalf("extractComments returned %d entries, want 1: %#v", len(comments), comments)
	}
}

// TestApplyCommentsKeepsLineCommentsOffNestedNodes verifies line comments are
// not attached to populated mapping or sequence nodes.
//
// Why it matters: yaml.v3 can misplace nested-node line comments and produce confusing or malformed config output.
// Inputs: scalar, empty mapping, populated mapping, and sequence YAML nodes with comment metadata. Outputs: safe node comments and parseable YAML.
// Data choice: the node shapes match config scalars and provider import/export sections.
func TestApplyCommentsKeepsLineCommentsOffNestedNodes(t *testing.T) {
	mapNode := &yaml.Node{Kind: yaml.MappingNode, Tag: tagMap, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: tagStr, Value: "scalar"},
		{Kind: yaml.ScalarNode, Tag: tagStr, Value: "value"},
		{Kind: yaml.ScalarNode, Tag: tagStr, Value: "empty_map"},
		{Kind: yaml.MappingNode, Tag: tagMap},
		{Kind: yaml.ScalarNode, Tag: tagStr, Value: "nested_map"},
		{Kind: yaml.MappingNode, Tag: tagMap, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: tagStr, Value: "child"},
			{Kind: yaml.ScalarNode, Tag: tagStr, Value: "value"},
		}},
		{Kind: yaml.ScalarNode, Tag: tagStr, Value: "sequence"},
		{Kind: yaml.SequenceNode, Tag: tagSeq, Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: tagStr, Value: "item"},
		}},
	}}
	comments := map[string]FieldComment{
		"scalar":     {HeadComment: "scalar head", LineComment: "scalar line", FootComment: "scalar foot"},
		"empty_map":  {LineComment: "empty line"},
		"nested_map": {LineComment: "nested line", FootComment: "nested foot"},
		"sequence":   {LineComment: "sequence line", FootComment: "sequence foot"},
	}

	applyComments(mapNode, comments)

	if got := yamlMappingKey(t, mapNode, "scalar").HeadComment; got != "scalar head" {
		t.Errorf("scalar head comment = %q, want %q", got, "scalar head")
	}
	if got := yamlMappingValue(t, mapNode, "scalar").LineComment; got != "scalar line" {
		t.Errorf("scalar line comment = %q, want %q", got, "scalar line")
	}
	if got := yamlMappingValue(t, mapNode, "scalar").FootComment; got != "scalar foot" {
		t.Errorf("scalar foot comment = %q, want %q", got, "scalar foot")
	}
	if got := yamlMappingValue(t, mapNode, "empty_map").LineComment; got != "empty line" {
		t.Errorf("empty mapping line comment = %q, want %q", got, "empty line")
	}
	if got := yamlMappingValue(t, mapNode, "nested_map").LineComment; got != "" {
		t.Errorf("nested mapping line comment = %q, want empty", got)
	}
	if got := yamlMappingValue(t, mapNode, "sequence").LineComment; got != "" {
		t.Errorf("sequence line comment = %q, want empty", got)
	}
	if got := yamlMappingValue(t, mapNode, "nested_map").FootComment; got != "nested foot" {
		t.Errorf("nested mapping foot comment = %q, want %q", got, "nested foot")
	}

	out := renderYAML(t, &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{mapNode}})
	assertNotContains(t, out, "nested line")
	assertNotContains(t, out, "sequence line")
	var reparsed yaml.Node
	if err := yaml.Unmarshal([]byte(out), &reparsed); err != nil {
		t.Fatalf("rendered YAML should parse after comment application: %v\n%s", err, out)
	}
}

func yamlMappingKey(t *testing.T, mapNode *yaml.Node, key string) *yaml.Node {
	t.Helper()
	for i := 0; i+1 < len(mapNode.Content); i += 2 {
		if mapNode.Content[i].Value == key {
			return mapNode.Content[i]
		}
	}
	t.Fatalf("missing YAML key %q", key)
	return nil
}

func yamlMappingValue(t *testing.T, mapNode *yaml.Node, key string) *yaml.Node {
	t.Helper()
	node, _ := findNodeByKey(mapNode, key)
	if node == nil {
		t.Fatalf("missing YAML value for key %q", key)
	}
	return node
}
