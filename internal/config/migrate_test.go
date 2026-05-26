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

func TestMigrateConfig_01x(t *testing.T) {
	root := loadFixture(t, "cani_0.1.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "datastore: /tmp/.cani/canidb.json")
	assertContains(t, out, "types_repos:")
	assertContains(t, out, "types_repo_pull: false")
	assertNotContains(t, out, "session:")

	assertContains(t, out, "use_simulation:")
	assertContains(t, out, "insecure:")
	assertContains(t, out, "provider_host:")
	assertContains(t, out, "ca_cert:")

	// No k8s fields in 0.1.x
	assertNotContains(t, out, "k8s_pods_cidr:")
	assertNotContains(t, out, "k8s_services_cidr:")

	assertContains(t, out, "types_dirs: []")
}

func TestMigrateConfig_02x(t *testing.T) {
	root := loadFixture(t, "cani_0.2.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "datastore: /Users/jsalmela/.cani/canidb.json")
	assertNotContains(t, out, "session:")

	assertContains(t, out, "k8s_pods_cidr:")
	assertContains(t, out, "k8s_services_cidr:")
}

func TestMigrateConfig_03x(t *testing.T) {
	root := loadFixture(t, "cani_0.3.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertNotContains(t, out, "session:")

	assertContains(t, out, "/tmp/.cani/hardware-types")
	assertNotContains(t, out, "custom_hardware_types_dir")
}

func TestMigrateConfig_04x(t *testing.T) {
	root := loadFixture(t, "cani_0.4.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertNotContains(t, out, "session:")
	assertNotContains(t, out, "domains:")

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

func TestMigrateConfig_05x(t *testing.T) {
	root := loadFixture(t, "cani_0.5.x.yml")
	newRoot, err := migrateConfig(root)
	if err != nil {
		t.Fatalf("migrateConfig: %v", err)
	}
	out := renderYAML(t, newRoot)

	assertContains(t, out, "providers:")
	assertContains(t, out, "csm:")
	assertContains(t, out, "ngsm:")
	assertNotContains(t, out, "session:")
	assertNotContains(t, out, "domains:")

	assertContains(t, out, "use_simulation:")
	assertContains(t, out, "provider_host: localhost:8443")
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
