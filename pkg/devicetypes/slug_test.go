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
package devicetypes

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestDeriveSlugMatchesNautobotImporter verifies deriveSlug reproduces the
// Nautobot devicetype-library importer rule re.sub(r'\W+', '-', name.lower()).
//
// Why it matters: libraries such as forge-devicetype-library omit slugs and
// rely on Nautobot to generate them, so cani must derive identical slugs or
// the same hardware would be keyed differently in cani and Nautobot.
// Inputs: manufacturer/model pairs. Outputs: the slug Python produces for the
// same input. Data choice: expected values were generated with the importer's
// Python expression and cover dashes, punctuation runs, a trailing non-word
// character, underscores (kept by \w), and non-ASCII letters.
func TestDeriveSlugMatchesNautobotImporter(t *testing.T) {
	cases := []struct{ manufacturer, model, want string }{
		{"HPE", "Aruba 10000-48Y6C Back-to-Front", "hpe-aruba-10000-48y6c-back-to-front"},
		{"Juniper", "EX4100-24T", "juniper-ex4100-24t"},
		{"HPE", "ProLiant DL380 Gen10 (8SFF)", "hpe-proliant-dl380-gen10-8sff-"},
		{"HPE", "HPE InfiniBand NDR/Ethernet 2x400Gb OSFP", "hpe-hpe-infiniband-ndr-ethernet-2x400gb-osfp"},
		{"Acme", "Model_X  v2", "acme-model_x-v2"},
		{"Ünïcode", "Gerät 9", "ünïcode-gerät-9"},
	}
	for _, c := range cases {
		if got := deriveSlug(c.manufacturer, c.model); got != c.want {
			t.Errorf("deriveSlug(%q, %q) = %q, want %q", c.manufacturer, c.model, got, c.want)
		}
	}
}

// TestDeriveSlugRequiresManufacturerAndModel verifies deriveSlug returns ""
// when either input is missing.
//
// Why it matters: a slug built from only one field would be ambiguous and
// could collide across vendors. Inputs: one or both fields empty. Outputs: "".
func TestDeriveSlugRequiresManufacturerAndModel(t *testing.T) {
	for _, in := range [][2]string{{"", "M1"}, {"Acme", ""}, {"", ""}} {
		if got := deriveSlug(in[0], in[1]); got != "" {
			t.Errorf("deriveSlug(%q, %q) = %q, want empty", in[0], in[1], got)
		}
	}
}

// TestResolveSlug verifies resolveSlug prefers an explicit slug, falls back to
// a derived one, and warns only when a partially described type is skipped.
//
// Why it matters: explicit slugs in the embedded library must never change,
// and a file that cannot be loaded should be visible rather than dropped
// silently. Inputs: explicit, derivable, partial, and empty field sets.
// Outputs: the resolved slug and whether a warning was logged.
func TestResolveSlug(t *testing.T) {
	cases := []struct {
		name, slug, manufacturer, model, want string
		warn                                  bool
	}{
		{"explicit wins", "custom", "Acme", "M1", "custom", false},
		{"derived", "", "Acme", "M1", "acme-m1", false},
		{"partial warns", "", "", "M1", "", true},
		{"empty doc silent", "", "", "", "", false},
	}
	out := captureLog(t)
	for _, c := range cases {
		out.Reset()
		got := resolveSlug(c.slug, c.manufacturer, c.model, "f.yaml")
		if got != c.want {
			t.Errorf("%s: slug = %q, want %q", c.name, got, c.want)
		}
		if warned := strings.Contains(out.String(), "f.yaml"); warned != c.warn {
			t.Errorf("%s: warned = %v, want %v (log %q)", c.name, warned, c.warn, out.String())
		}
	}
}

// TestLoadFromDirDerivesMissingSlugs verifies a Nautobot-format library with
// no slug fields loads its device and module types under derived slugs.
//
// Why it matters: this is the forge-devicetype-library layout; before slug
// derivation every file was skipped. Inputs: a root with device-types/ and
// module-types/ files lacking slugs. Outputs: both types registered under the
// Nautobot-derived slug.
func TestLoadFromDirDerivesMissingSlugs(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	root := mkTmpDir(t)
	writeYAML(t, filepath.Join(root, "device-types", "HPE", "a.yaml"),
		"---\nmanufacturer: HPE\nmodel: Aruba 10000-48Y6C Back-to-Front\nu_height: 1\n")
	writeYAML(t, filepath.Join(root, "module-types", "HPE", "m.yaml"),
		"manufacturer: HPE\nmodel: J9820A\n")

	if err := LoadFromDir(root, "local:"+root); err != nil {
		t.Fatalf("LoadFromDir: %v", err)
	}
	if _, ok := allDeviceTypes["hpe-aruba-10000-48y6c-back-to-front"]; !ok {
		t.Error("expected derived device slug hpe-aruba-10000-48y6c-back-to-front")
	}
	if _, ok := allModuleTypes["hpe-j9820a"]; !ok {
		t.Error("expected derived module slug hpe-j9820a")
	}
}

// TestLoadFromDirCacheIncludesShadowedTypes verifies a directory's cache keeps
// types that lost registration to an earlier source.
//
// Why it matters: a cache built while another library supplied the same slug
// used to omit that slug, so a later run without the other library could not
// find it at all. Inputs: two directories defining the same derived slug,
// loaded in order, then a fresh registry loading only the second directory
// from its cache. Outputs: the slug resolves to the second directory's type.
func TestLoadFromDirCacheIncludesShadowedTypes(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	first, second := mkTmpDir(t), mkTmpDir(t)
	for _, root := range []string{first, second} {
		writeYAML(t, filepath.Join(root, "device-types", "HPE", "dl325.yaml"),
			"manufacturer: HPE\nmodel: ProLiant DL325 Gen10\n")
	}
	if err := LoadFromDir(first, "first"); err != nil {
		t.Fatalf("LoadFromDir(first): %v", err)
	}
	if err := LoadFromDir(second, "second"); err != nil {
		t.Fatalf("LoadFromDir(second): %v", err)
	}

	resetRegistries()
	if err := LoadFromDir(second, "second"); err != nil {
		t.Fatalf("LoadFromDir(second) from cache: %v", err)
	}
	if dt, ok := allDeviceTypes["hpe-proliant-dl325-gen10"]; !ok || dt.Source != "second" {
		t.Errorf("cached slug = %+v (ok=%v), want hpe-proliant-dl325-gen10 from second", dt, ok)
	}
}
