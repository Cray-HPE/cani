/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const fixtureDir = "../../testdata/fixtures/cani/configs"

func loadFixture(t *testing.T, name string) *yaml.Node {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatalf("parsing fixture %s: %v", name, err)
	}
	return &root
}

// TestIsLegacyFormat verifies legacy config detection across supported fixture versions.
//
// Why it matters: Load depends on this gate to decide whether to migrate or decode in place.
// Inputs: 0.1.x through 0.6.x config fixtures. Outputs: true only for pre-0.6.x configs.
// Data choice: versioned fixtures mirror the integration migration coverage and include both legacy families.
func TestIsLegacyFormat(t *testing.T) {
	tests := []struct {
		fixture string
		want    bool
	}{
		{"cani_0.1.x.yml", true},
		{"cani_0.2.x.yml", true},
		{"cani_0.3.x.yml", true},
		{"cani_0.4.x.yml", true},
		{"cani_0.5.x.yml", true},
		{"cani_0.6.x.yml", false},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			root := loadFixture(t, tt.fixture)
			if got := isLegacyFormat(root); got != tt.want {
				t.Errorf("isLegacyFormat(%s) = %v, want %v", tt.fixture, got, tt.want)
			}
		})
	}
}

// TestBackupConfig verifies backupConfig renames a config file to the .canisave path.
//
// Why it matters: migration must preserve the user's original config before writing the new format.
// Inputs: a temporary cani.yml file. Outputs: missing source file and backup file with identical content.
// Data choice: a small YAML-like payload makes content preservation obvious without unrelated config parsing.
func TestBackupConfig(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "cani.yml")
	content := []byte("test: true\n")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := backupConfig(src); err != nil {
		t.Fatalf("backupConfig: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("original file should not exist after backup")
	}

	backup := src + ".canisave"
	got, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("backup content = %q, want %q", got, content)
	}
}

// TestMigrateConfig_01x verifies a 0.1.x family-A config migrates to the current schema.
//
// Why it matters: early CSM configs lacked newer fields but still need a usable provider and datastore section.
// Inputs: the cani_0.1.x fixture. Outputs: current-format YAML without session and without nonexistent k8s fields.
// Data choice: this fixture represents the oldest supported migration shape.
func TestMigrateConfig_01x(t *testing.T) {
	root := loadFixture(t, "cani_0.1.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)
	decoded := decodeRenderedConfig(t, out)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "datastore: /tmp/.cani/canidb.json")
	assertMapValue(t, decoded, "datastore", "/tmp/.cani/canidb.json")
	assertContains(t, out, "types_repos:")
	assertContains(t, out, "types_repo_pull: false")
	assertNotContains(t, out, "session:")
	assertMapMissing(t, decoded, "session")

	assertContains(t, out, "use_simulation:")
	assertContains(t, out, "insecure:")
	assertContains(t, out, "provider_host:")
	assertContains(t, out, "ca_cert:")

	// No k8s fields in 0.1.x
	assertNotContains(t, out, "k8s_pods_cidr:")
	assertNotContains(t, out, "k8s_services_cidr:")

	assertContains(t, out, "types_dirs: []")
}

// TestMigrateConfig_02x verifies a 0.2.x family-A config migrates k8s CSM fields.
//
// Why it matters: these CIDR fields are consumed by the CSM provider after migration.
// Inputs: the cani_0.2.x fixture. Outputs: current-format YAML with mapped CSM CIDR keys.
// Data choice: this fixture is the first family-A version containing k8s CIDR options.
func TestMigrateConfig_02x(t *testing.T) {
	root := loadFixture(t, "cani_0.2.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)
	decoded := decodeRenderedConfig(t, out)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "datastore: /Users/jsalmela/.cani/canidb.json")
	assertMapValue(t, decoded, "datastore", "/Users/jsalmela/.cani/canidb.json")
	assertNotContains(t, out, "session:")
	assertMapMissing(t, decoded, "session")

	assertContains(t, out, "k8s_pods_cidr:")
	assertContains(t, out, "k8s_services_cidr:")
}

// TestMigrateConfig_03x verifies a 0.3.x family-A config migrates custom hardware types.
//
// Why it matters: legacy custom_hardware_types_dir must become types_dirs without preserving the old key.
// Inputs: the cani_0.3.x fixture. Outputs: current-format YAML with the hardware types path in types_dirs.
// Data choice: this fixture is the family-A case that carries a custom hardware types directory.
func TestMigrateConfig_03x(t *testing.T) {
	root := loadFixture(t, "cani_0.3.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)
	decoded := decodeRenderedConfig(t, out)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertNotContains(t, out, "session:")
	assertMapMissing(t, decoded, "session")

	assertContains(t, out, "/tmp/.cani/hardware-types")
	assertNotContains(t, out, "custom_hardware_types_dir")
}

// TestMigrateConfig_04x verifies a 0.4.x family-B config migrates mapped and legacy CSM options.
//
// Why it matters: family-B configs introduced domains and unmapped options that must remain available under _legacy.
// Inputs: the cani_0.4.x fixture. Outputs: current-format YAML with mapped CSM keys and non-empty _legacy data.
// Data choice: this fixture includes scalar, sequence, empty, and renamed legacy option values.
func TestMigrateConfig_04x(t *testing.T) {
	root := loadFixture(t, "cani_0.4.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)
	decoded := decodeRenderedConfig(t, out)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertNotContains(t, out, "session:")
	assertNotContains(t, out, "domains:")
	assertMapMissing(t, decoded, "session")

	assertContains(t, out, "use_simulation: \"true\"")
	assertContains(t, out, "insecure: \"true\"")
	assertContains(t, out, "provider_host: localhost:8443")
	assertContains(t, out, "secret_name: admin-client-auth")
	assertContains(t, out, "k8s_pods_cidr: 10.32.0.0/12")
	assertContains(t, out, "k8s_services_cidr: 10.16.0.0/12")

	assertContains(t, out, "_legacy:")
	assertContains(t, out, "apigatewaytoken: migrated")
	assertContains(t, out, "baseurlsls:")
	assertContains(t, out, "baseurlhsm:")
	assertContains(t, out, "validroles:")
	assertContains(t, out, "validsubroles:")

	assertContains(t, out, "/tmp/.cani/hardware-types")
}

// TestMigrateConfig_05x verifies a 0.5.x family-B config migrates multiple providers.
//
// Why it matters: migration must preserve each domain as a provider section, even when options are nil.
// Inputs: the cani_0.5.x fixture. Outputs: current-format YAML with csm and ngsm providers and no domains key.
// Data choice: this fixture has one populated provider and one provider with null options.
func TestMigrateConfig_05x(t *testing.T) {
	root := loadFixture(t, "cani_0.5.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)
	decoded := decodeRenderedConfig(t, out)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "ngsm:")
	assertNotContains(t, out, "session:")
	assertNotContains(t, out, "domains:")
	assertMapMissing(t, decoded, "session")

	assertContains(t, out, "use_simulation:")
	assertContains(t, out, "provider_host: localhost:8443")
}

// TestLoadMigratesLegacyConfigBacksUpAndWritesParseableYAML verifies Load runs
// the migration path end to end for a legacy config file.
//
// Why it matters: integration tests exercise the CLI, but unit tests should catch backup, save, and YAML parse regressions near config code.
// Inputs: a copied 0.4.x fixture at a temporary config path. Outputs: backup file, current-format YAML, and populated Cfg.
// Data choice: the 0.4.x fixture includes mapped fields, legacy fields, and custom hardware types.
func TestLoadMigratesLegacyConfigBacksUpAndWritesParseableYAML(t *testing.T) {
	orig := Cfg
	defer func() { Cfg = orig }()

	fixturePath := filepath.Join(fixtureDir, "cani_0.4.x.yml")
	fixtureData, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	path := filepath.Join(t.TempDir(), "cani.yml")
	if err := os.WriteFile(path, fixtureData, 0600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	if err := Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	backupData, err := os.ReadFile(path + ".canisave")
	if err != nil {
		t.Fatalf("reading backup config: %v", err)
	}
	if string(backupData) != string(fixtureData) {
		t.Fatalf("backup content changed during migration")
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading migrated config: %v", err)
	}
	out := string(written)
	decoded := decodeRenderedConfig(t, out)
	assertMapMissing(t, decoded, "session")
	assertMapMissing(t, decoded, "domains")
	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "_legacy:")
	assertContains(t, out, "datastore:")

	if Cfg == nil {
		t.Fatal("expected Cfg to be populated")
	}
	csm, ok := Cfg.Providers["csm"]
	if !ok {
		t.Fatal("expected migrated csm provider in Cfg")
	}
	if got := csm["provider_host"]; got != "localhost:8443" {
		t.Errorf("provider_host = %v, want localhost:8443", got)
	}
}

// TestMigrateConfigSkipsUnsupportedLegacyAliasValues verifies unsupported YAML
// node kinds in legacy provider options do not render malformed _legacy YAML.
//
// Why it matters: legacy configs may contain YAML constructs that are not represented by the known scalar/map/sequence paths.
// Inputs: a family-B config with an anchor, alias, and nested legacy map option. Outputs: parseable YAML with supported legacy values preserved.
// Data choice: aliases exercise an unsupported node kind while nested maps and sequences prove supported complex values still survive.
func TestMigrateConfigSkipsUnsupportedLegacyAliasValues(t *testing.T) {
	legacy := []byte(`session:
  domains:
    csm:
      datastore_path: /tmp/cani.json
      custom_hardware_types_dir: /tmp/hardware-types
      options:
        providerhost: localhost:8443
        roles: &roles
          - Management
          - Compute
        copied_roles: *roles
        nested:
          list:
            - System
          enabled: true
`)
	var root yaml.Node
	if err := yaml.Unmarshal(legacy, &root); err != nil {
		t.Fatalf("parsing legacy YAML: %v", err)
	}

	newRoot, err := migrateConfig(&root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)
	decoded := decodeRenderedConfig(t, out)

	assertMapValue(t, decoded, "datastore", "/tmp/cani.json")
	assertContains(t, out, "provider_host: localhost:8443")
	assertContains(t, out, "_legacy:")
	assertContains(t, out, "roles:")
	assertContains(t, out, "nested:")
	assertNotContains(t, out, "copied_roles:")
}

func renderYAML(t *testing.T, root *yaml.Node) string {
	t.Helper()
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		t.Fatalf("encoding yaml: %v", err)
	}
	return buf.String()
}

func decodeRenderedConfig(t *testing.T, out string) map[string]any {
	t.Helper()
	decoded := map[string]any{}
	if err := yaml.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("rendered YAML should parse: %v\n--- output ---\n%s", err, out)
	}
	return decoded
}

func assertMapValue(t *testing.T, decoded map[string]any, key string, want any) {
	t.Helper()
	got, ok := decoded[key]
	if !ok {
		t.Fatalf("decoded YAML missing key %q", key)
	}
	if got != want {
		t.Fatalf("decoded YAML key %q = %v, want %v", key, got, want)
	}
}

func assertMapMissing(t *testing.T, decoded map[string]any, key string) {
	t.Helper()
	if _, ok := decoded[key]; ok {
		t.Fatalf("decoded YAML should not contain key %q", key)
	}
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("output should contain %q\n--- output ---\n%s", needle, haystack)
	}
}

func assertNotContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("output should NOT contain %q\n--- output ---\n%s", needle, haystack)
	}
}
