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
package devicetypes

// Test coverage for cache.go
//
// | Function             | Happy-path test                 | Failure / edge test            |
// |----------------------|---------------------------------|--------------------------------|
// | writeDirCache        | TestDirCacheRoundTrip           | -                              |
// | readDirCache         | TestDirCacheRoundTrip           | TestReadDirCacheErrors         |
// | registerCachedTypes  | TestRegisterCachedTypes         | TestRegisterCachedTypes        |
// | collectTypesFromSource | TestCollectTypesFromSource    | -                              |
// | removeDirCache       | TestRemoveDirCache              | TestRemoveDirCache             |

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDirCacheRoundTrip verifies writeDirCache serializes a cache to disk and
// readDirCache reads it back with all type slices intact.
//
// Why it matters: the on-disk cache lets subsequent loads skip YAML parsing, so
// a write/read mismatch would silently drop hardware types from the library.
// Inputs: a dirCache holding one device type, written to a temp path and read
// back. Outputs: a cache whose DeviceTypes slice contains the original slug.
// Data choice: a single device type is enough to prove JSON marshalling and the
// directory-creation step both succeed without coupling to other kinds.
func TestDirCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", cacheFileName)
	in := &dirCache{DeviceTypes: []CaniDeviceType{{Slug: "cache-rt-dev", Model: "M"}}}

	writeDirCache(path, in)

	out, err := readDirCache(path)
	if err != nil {
		t.Fatalf("readDirCache error: %v", err)
	}
	if len(out.DeviceTypes) != 1 || out.DeviceTypes[0].Slug != "cache-rt-dev" {
		t.Errorf("round-trip device types = %+v, want one slug cache-rt-dev", out.DeviceTypes)
	}
}

// TestReadDirCacheErrors verifies readDirCache returns an error for a missing
// file and for malformed JSON.
//
// Why it matters: a corrupt or absent cache must surface an error so the loader
// falls back to parsing YAML rather than registering empty data.
// Inputs: a path that does not exist, and a file containing invalid JSON.
// Outputs: a non-nil error in both cases. Data choice: these are the only two
// failure modes readDirCache can encounter (ReadFile error and Unmarshal
// error), so one input targets each.
func TestReadDirCacheErrors(t *testing.T) {
	if _, err := readDirCache(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("expected error for missing cache file, got nil")
	}

	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}
	if _, err := readDirCache(bad); err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

// TestRegisterCachedTypes verifies registerCachedTypes registers every kind from
// a cache, stamps the source, skips empty slugs, and skips already-registered
// slugs on a second pass.
//
// Why it matters: cached registration is the fast path on startup, so it must
// populate all six registries exactly once and never clobber a slug already
// loaded from a higher-priority source.
// Inputs: a dirCache with one device, module, cable, rack, FRU, and location
// type plus an empty-slug device entry, registered twice with source "cache".
// Outputs: each registry contains the slug with Source="cache", and the second
// pass leaves counts unchanged. Data choice: one entry per kind covers every
// registration branch, the empty-slug entry covers the skip guard, and the
// repeat pass covers the already-exists guard.
func TestRegisterCachedTypes(t *testing.T) {
	t.Cleanup(func() {
		delete(allDeviceTypes, "cache-dev")
		delete(allModuleTypes, "cache-mod")
		delete(allCableTypes, "cache-cable")
		delete(allRackTypes, "cache-rack")
		delete(allFruTypes, "cache-fru")
		delete(allLocationTypes, "cache-loc")
	})

	c := &dirCache{
		DeviceTypes:   []CaniDeviceType{{Slug: "cache-dev"}, {Slug: ""}},
		ModuleTypes:   []CaniModuleType{{Slug: "cache-mod"}},
		CableTypes:    []CaniCableType{{Slug: "cache-cable"}},
		RackTypes:     []CaniRackType{{Slug: "cache-rack"}},
		FruTypes:      []CaniFruType{{Slug: "cache-fru"}},
		LocationTypes: []LocationTypeDefinition{{Slug: "cache-loc"}},
	}

	registerCachedTypes(c, "cache")

	if dt, ok := allDeviceTypes["cache-dev"]; !ok || dt.Source != "cache" {
		t.Errorf("device type not registered with source: %+v ok=%v", dt, ok)
	}
	if _, ok := allModuleTypes["cache-mod"]; !ok {
		t.Error("module type not registered")
	}
	if _, ok := allCableTypes["cache-cable"]; !ok {
		t.Error("cable type not registered")
	}
	if _, ok := allRackTypes["cache-rack"]; !ok {
		t.Error("rack type not registered")
	}
	if _, ok := allFruTypes["cache-fru"]; !ok {
		t.Error("fru type not registered")
	}
	if _, ok := allLocationTypes["cache-loc"]; !ok {
		t.Error("location type not registered")
	}
	if _, ok := allDeviceTypes[""]; ok {
		t.Error("empty-slug device type should not be registered")
	}

	before := len(allDeviceTypes)
	registerCachedTypes(c, "second") // already-exists path
	if len(allDeviceTypes) != before {
		t.Errorf("second registration changed device count: %d -> %d", before, len(allDeviceTypes))
	}
	if allDeviceTypes["cache-dev"].Source != "cache" {
		t.Error("second pass overwrote existing source")
	}
}

// TestCollectTypesFromSource verifies collectTypesFromSource gathers only the
// registered types whose Source matches the requested value.
//
// Why it matters: writing a directory cache must capture exactly the types that
// came from that directory, so source filtering governs cache correctness.
// Inputs: two device types registered with different sources ("src-a",
// "src-b"), collected for "src-a". Outputs: a cache containing only the src-a
// device type. Data choice: two differing sources prove the filter both
// includes a match and excludes a non-match.
func TestCollectTypesFromSource(t *testing.T) {
	t.Cleanup(func() {
		delete(allDeviceTypes, "src-a-dev")
		delete(allDeviceTypes, "src-b-dev")
	})

	RegisterDeviceType(CaniDeviceType{Slug: "src-a-dev", Source: "src-a"})
	RegisterDeviceType(CaniDeviceType{Slug: "src-b-dev", Source: "src-b"})

	got := collectTypesFromSource("src-a")
	if len(got.DeviceTypes) != 1 || got.DeviceTypes[0].Slug != "src-a-dev" {
		t.Errorf("collectTypesFromSource(src-a) = %+v, want only src-a-dev", got.DeviceTypes)
	}
}

// TestRemoveDirCache verifies removeDirCache deletes an existing cache file and
// is a no-op when the file is absent.
//
// Why it matters: after a source changes (e.g. a git pull) the stale cache must
// be removed so the next load re-parses, and removal must not panic when there
// is nothing to delete.
// Inputs: a directory with a cache file present, then the same call against a
// directory with no cache file. Outputs: the file no longer exists after the
// first call and the second call returns without error or panic. Data choice:
// exercising both the present and absent cases covers the os.Remove success and
// ignored-error branches.
func TestRemoveDirCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cacheFileName)
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	removeDirCache(dir)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("cache file still present after removeDirCache, stat err = %v", err)
	}

	// Absent file: must not panic or error.
	removeDirCache(t.TempDir())
}
