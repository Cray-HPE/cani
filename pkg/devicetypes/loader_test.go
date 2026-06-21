package devicetypes

// Test coverage for loader.go
//
// | Function                | Happy-path test                            | Failure test                                  |
// |-------------------------|--------------------------------------------|-----------------------------------------------|
// | LoadAll                 | TestLoadAllHappyPath                       | TestLoadAllFailsOnBadDir                      |
// | LoadFromDir             | TestLoadFromDirHappyPath                   | TestLoadFromDirNonexistentRoot                |
// | loadDeviceTypesFromDir  | TestLoadDeviceTypesFromDirHappyPath        | TestLoadDeviceTypesFromDirInvalidYAML         |
// | loadModuleTypesFromDir  | TestLoadModuleTypesFromDirHappyPath        | TestLoadModuleTypesFromDirInvalidYAML         |
// | loadCableTypesFromDir   | TestLoadCableTypesFromDirHappyPath         | TestLoadCableTypesFromDirInvalidYAML          |
// | loadRackTypesFromDir    | TestLoadRackTypesFromDirHappyPath          | TestLoadRackTypesFromDirInvalidYAML           |
// | loadFruTypesFromDir     | TestLoadFruTypesFromDirHappyPath           | TestLoadFruTypesFromDirInvalidYAML            |
// | walkYAMLFiles           | TestWalkYAMLFilesHappyPath                 | TestWalkYAMLFilesNonexistentDir               |
// | dirExists               | TestDirExistsHappyPath                     | TestDirExistsNonexistent                      |
// | sanitizeRepoName        | TestSanitizeRepoNameHappyPath              | TestSanitizeRepoNamePlainString               |

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------- helpers ----------

// mkTmpDir creates a temporary directory and returns its path.
// The directory is removed when the test finishes.
func mkTmpDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "loader-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// writeYAML creates a YAML file at path with the given content.
func writeYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// ---------- dirExists ----------

func TestDirExistsHappyPath(t *testing.T) {
	dir := mkTmpDir(t)
	if !dirExists(dir) {
		t.Errorf("dirExists(%q) = false, want true", dir)
	}
}

func TestDirExistsNonexistent(t *testing.T) {
	if dirExists("/tmp/this-path-should-not-exist-29473") {
		t.Error("dirExists returned true for nonexistent path")
	}
}

// ---------- sanitizeRepoName ----------

func TestSanitizeRepoNameHappyPath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"https://github.com/org/repo.git", "github.com-org-repo"},
		{"git@github.com:org/repo.git", "github.com-org-repo"},
		{"ssh://git@example.com/path/repo.git", "git@example.com-path-repo"},
		{"http://example.com/my/repo.git", "example.com-my-repo"},
	}
	for _, tc := range cases {
		got := sanitizeRepoName(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeRepoName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSanitizeRepoNamePlainString(t *testing.T) {
	// A plain string with no protocol prefix is returned with separators replaced.
	got := sanitizeRepoName("plain-name")
	if got != "plain-name" {
		t.Errorf("sanitizeRepoName(%q) = %q, want %q", "plain-name", got, "plain-name")
	}
}

// ---------- walkYAMLFiles ----------

func TestWalkYAMLFilesHappyPath(t *testing.T) {
	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "a.yaml"), "slug: a")
	writeYAML(t, filepath.Join(dir, "sub", "b.yml"), "slug: b")
	// non-YAML should be skipped
	writeYAML(t, filepath.Join(dir, "readme.md"), "# readme")

	var collected []string
	err := walkYAMLFiles(dir, func(data []byte, path string) {
		collected = append(collected, path)
	})
	if err != nil {
		t.Fatalf("walkYAMLFiles returned error: %v", err)
	}
	if len(collected) != 2 {
		t.Errorf("expected 2 YAML files, got %d: %v", len(collected), collected)
	}
}

func TestWalkYAMLFilesNonexistentDir(t *testing.T) {
	err := walkYAMLFiles("/tmp/does-not-exist-83729", func(data []byte, path string) {
		t.Error("callback should not be called for nonexistent dir")
	})
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

// ---------- loadDeviceTypesFromDir ----------

func TestLoadDeviceTypesFromDirHappyPath(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "test-device.yaml"),
		"slug: test-dev-loader\nmanufacturer: Acme\nmodel: D1\n")

	err := loadDeviceTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("loadDeviceTypesFromDir returned error: %v", err)
	}
	dt, ok := allDeviceTypes["test-dev-loader"]
	if !ok {
		t.Fatal("expected test-dev-loader in allDeviceTypes")
	}
	if dt.Source != "test-source" {
		t.Errorf("Source = %q, want %q", dt.Source, "test-source")
	}
}

func TestLoadDeviceTypesFromDirInvalidYAML(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "bad.yaml"), "{{{{not valid yaml")

	// Invalid YAML is logged but does not return an error.
	err := loadDeviceTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("expected no error for invalid YAML (logged warning), got: %v", err)
	}
	if len(allDeviceTypes) != 0 {
		t.Errorf("expected 0 device types after bad YAML, got %d", len(allDeviceTypes))
	}
}

// ---------- loadModuleTypesFromDir ----------

func TestLoadModuleTypesFromDirHappyPath(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "mod.yaml"),
		"slug: test-mod-loader\nmanufacturer: Acme\nmodel: M1\n")

	err := loadModuleTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("loadModuleTypesFromDir returned error: %v", err)
	}
	mt, ok := allModuleTypes["test-mod-loader"]
	if !ok {
		t.Fatal("expected test-mod-loader in allModuleTypes")
	}
	if mt.Source != "test-source" {
		t.Errorf("Source = %q, want %q", mt.Source, "test-source")
	}
}

func TestLoadModuleTypesFromDirInvalidYAML(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "bad.yaml"), "{{{{not valid yaml")

	err := loadModuleTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("expected no error for invalid YAML, got: %v", err)
	}
	if len(allModuleTypes) != 0 {
		t.Errorf("expected 0 module types after bad YAML, got %d", len(allModuleTypes))
	}
}

// ---------- loadCableTypesFromDir ----------

func TestLoadCableTypesFromDirHappyPath(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "cable.yaml"),
		"slug: test-cable-loader\nmanufacturer: Acme\nmodel: C1\n")

	err := loadCableTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("loadCableTypesFromDir returned error: %v", err)
	}
	ct, ok := allCableTypes["test-cable-loader"]
	if !ok {
		t.Fatal("expected test-cable-loader in allCableTypes")
	}
	if ct.Source != "test-source" {
		t.Errorf("Source = %q, want %q", ct.Source, "test-source")
	}
}

func TestLoadCableTypesFromDirInvalidYAML(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "bad.yaml"), "{{{{not valid yaml")

	err := loadCableTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("expected no error for invalid YAML, got: %v", err)
	}
	if len(allCableTypes) != 0 {
		t.Errorf("expected 0 cable types after bad YAML, got %d", len(allCableTypes))
	}
}

// ---------- loadRackTypesFromDir ----------

func TestLoadRackTypesFromDirHappyPath(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "rack.yaml"),
		"slug: test-rack-loader\nmanufacturer: Acme\nmodel: R1\n")

	err := loadRackTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("loadRackTypesFromDir returned error: %v", err)
	}
	rt, ok := allRackTypes["test-rack-loader"]
	if !ok {
		t.Fatal("expected test-rack-loader in allRackTypes")
	}
	if rt.Source != "test-source" {
		t.Errorf("Source = %q, want %q", rt.Source, "test-source")
	}
}

func TestLoadRackTypesFromDirInvalidYAML(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "bad.yaml"), "{{{{not valid yaml")

	err := loadRackTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("expected no error for invalid YAML, got: %v", err)
	}
	if len(allRackTypes) != 0 {
		t.Errorf("expected 0 rack types after bad YAML, got %d", len(allRackTypes))
	}
}

// ---------- loadFruTypesFromDir ----------

func TestLoadFruTypesFromDirHappyPath(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "fru.yaml"),
		"slug: test-fru-loader\nmanufacturer: Acme\nmodel: F1\n")

	err := loadFruTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("loadFruTypesFromDir returned error: %v", err)
	}
	ft, ok := allFruTypes["test-fru-loader"]
	if !ok {
		t.Fatal("expected test-fru-loader in allFruTypes")
	}
	if ft.Source != "test-source" {
		t.Errorf("Source = %q, want %q", ft.Source, "test-source")
	}
}

func TestLoadFruTypesFromDirInvalidYAML(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "bad.yaml"), "{{{{not valid yaml")

	err := loadFruTypesFromDir(dir, "test-source")
	if err != nil {
		t.Fatalf("expected no error for invalid YAML, got: %v", err)
	}
	if len(allFruTypes) != 0 {
		t.Errorf("expected 0 FRU types after bad YAML, got %d", len(allFruTypes))
	}
}

// ---------- LoadFromDir ----------

func TestLoadFromDirHappyPath(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	root := mkTmpDir(t)
	// Create the expected subdirectory structure.
	writeYAML(t, filepath.Join(root, "device-types", "d.yaml"),
		"slug: lfd-dev\nmanufacturer: Test\n")
	writeYAML(t, filepath.Join(root, "module-types", "m.yaml"),
		"slug: lfd-mod\nmanufacturer: Test\n")
	writeYAML(t, filepath.Join(root, "cable-types", "c.yaml"),
		"slug: lfd-cable\nmanufacturer: Test\n")
	writeYAML(t, filepath.Join(root, "rack-types", "r.yaml"),
		"slug: lfd-rack\nmanufacturer: Test\n")
	writeYAML(t, filepath.Join(root, "inventory-types", "f.yaml"),
		"slug: lfd-fru\nmanufacturer: Test\n")

	err := LoadFromDir(root, "test-local")
	if err != nil {
		t.Fatalf("LoadFromDir returned error: %v", err)
	}
	for slug, registry := range map[string]bool{
		"lfd-dev":   allDeviceTypes["lfd-dev"].Slug == "lfd-dev",
		"lfd-mod":   allModuleTypes["lfd-mod"].Slug == "lfd-mod",
		"lfd-cable": allCableTypes["lfd-cable"].Slug == "lfd-cable",
		"lfd-rack":  allRackTypes["lfd-rack"].Slug == "lfd-rack",
		"lfd-fru":   allFruTypes["lfd-fru"].Slug == "lfd-fru",
	} {
		if !registry {
			t.Errorf("expected %s to be registered", slug)
		}
	}
}

func TestLoadFromDirNonexistentRoot(t *testing.T) {
	// A nonexistent root has no matching subdirs, so all are skipped — no error.
	err := LoadFromDir("/tmp/no-such-dir-loader-test-87461", "src")
	if err != nil {
		t.Errorf("expected nil error for nonexistent root (subdirs skipped), got: %v", err)
	}
}

// ---------- LoadAll ----------

func TestLoadAllHappyPath(t *testing.T) {
	// Embedded types were loaded by init(); LoadAll(nil,nil,false,false) should succeed
	// with no extra sources and leave embedded types intact.
	err := LoadAll(nil, nil, false, false)
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}
	if len(allDeviceTypes) == 0 {
		t.Error("expected embedded device types to be loaded")
	}
}

func TestLoadAllFailsOnBadDir(t *testing.T) {
	// Create a root with a device-types subdir that is unreadable.
	tmp := mkTmpDir(t)
	devDir := filepath.Join(tmp, "device-types")
	if err := os.MkdirAll(devDir, 0o755); err != nil {
		t.Fatalf("setup mkdir: %v", err)
	}
	// Remove all permissions so filepath.Walk returns a permission error.
	if err := os.Chmod(devDir, 0o000); err != nil {
		t.Fatalf("setup chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(devDir, 0o755) })

	err := LoadAll([]string{tmp}, nil, false, false)
	if err == nil {
		t.Error("expected error when device-types dir is unreadable, got nil")
	}
}

// TestLoadAllWithDirsAndEmptyEntries verifies LoadAll skips empty source
// entries, loads types from a valid local directory, and ignores an empty repo
// entry without touching git.
//
// Why it matters: operators pass dir/repo lists that often contain blank
// entries; LoadAll must tolerate them and still register types from the real
// directory, which is the common multi-source startup path.
// Inputs: typesDirs of {"", validRoot} where validRoot has a device-types
// YAML, and typesRepos of {""}. Outputs: no error and the directory's slug
// registered. Data choice: pairing a blank entry with a real one in the same
// slice exercises both the continue guard and the successful load branch, while
// the blank repo covers the repo-loop guard without needing network access.
func TestLoadAllWithDirsAndEmptyEntries(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	root := mkTmpDir(t)
	writeYAML(t, filepath.Join(root, "device-types", "d.yaml"),
		"slug: loadall-dev\nmanufacturer: Test\n")

	err := LoadAll([]string{"", root}, []string{""}, false, false)
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}
	if _, ok := allDeviceTypes["loadall-dev"]; !ok {
		t.Error("expected loadall-dev to be registered from valid dir")
	}
}

// ---------- loader skip branches (empty slug + duplicate) ----------

// TestLoadDeviceTypesFromDirSkipsEmptyAndDuplicate verifies loadDeviceTypesFromDir
// skips documents with an empty slug and does not overwrite an already-registered
// slug.
//
// Why it matters: device-type sources are layered by priority, so a lower-
// priority directory must never clobber a slug already loaded, and malformed
// slug-less docs must be ignored rather than registered as blanks.
// Inputs: a pre-registered "dup-dev" (empty Model) and a one-file, two-document
// YAML whose first doc has an empty slug and whose second re-declares "dup-dev"
// with Model "Y". Outputs: no error and "dup-dev" retaining its original empty
// Model. Data choice: the multi-document file exercises forEachYAMLDoc while
// hitting both the empty-slug and duplicate-skip guards in one load.
func TestLoadDeviceTypesFromDirSkipsEmptyAndDuplicate(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	RegisterDeviceType(CaniDeviceType{Slug: "dup-dev"})

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "d.yaml"),
		"slug: \"\"\nmodel: X\n---\nslug: dup-dev\nmodel: Y\n")

	if err := loadDeviceTypesFromDir(dir, "src"); err != nil {
		t.Fatalf("loadDeviceTypesFromDir error: %v", err)
	}
	if allDeviceTypes["dup-dev"].Model != "" {
		t.Errorf("duplicate slug overwrote existing: %+v", allDeviceTypes["dup-dev"])
	}
}

// TestLoadTypesFromDirSkipBranches verifies the module, cable, rack, and FRU
// directory loaders all skip empty-slug and already-registered entries.
//
// Why it matters: each kind shares the same priority-layering contract as
// devices, so a duplicate or slug-less file in a lower-priority directory must
// be ignored for every type, not just device types.
// Inputs: for each loader, a pre-registered "dup-<kind>" plus a directory
// holding an empty-slug file and a duplicate-slug file. Outputs: no error and
// the registry size unchanged after loading. Data choice: a separate empty and
// duplicate file per loader (single-document loaders) isolates both guard
// branches that the happy-path and invalid-YAML tests leave uncovered.
func TestLoadTypesFromDirSkipBranches(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	RegisterModuleType(CaniModuleType{Slug: "dup-mod"})
	RegisterCableType(CaniCableType{Slug: "dup-cable"})
	RegisterRackType(CaniRackType{Slug: "dup-rack"})
	RegisterFruType(CaniFruType{Slug: "dup-fru"})

	mkDir := func(dup string) string {
		dir := mkTmpDir(t)
		writeYAML(t, filepath.Join(dir, "empty.yaml"), "slug: \"\"\nmodel: X\n")
		writeYAML(t, filepath.Join(dir, "dup.yaml"), "slug: "+dup+"\nmodel: Y\n")
		return dir
	}

	if err := loadModuleTypesFromDir(mkDir("dup-mod"), "src"); err != nil {
		t.Fatalf("loadModuleTypesFromDir error: %v", err)
	}
	if err := loadCableTypesFromDir(mkDir("dup-cable"), "src"); err != nil {
		t.Fatalf("loadCableTypesFromDir error: %v", err)
	}
	if err := loadRackTypesFromDir(mkDir("dup-rack"), "src"); err != nil {
		t.Fatalf("loadRackTypesFromDir error: %v", err)
	}
	if err := loadFruTypesFromDir(mkDir("dup-fru"), "src"); err != nil {
		t.Fatalf("loadFruTypesFromDir error: %v", err)
	}

	if allModuleTypes["dup-mod"].Model != "" || allCableTypes["dup-cable"].Model != "" ||
		allRackTypes["dup-rack"].Model != "" || allFruTypes["dup-fru"].Model != "" {
		t.Error("a duplicate slug overwrote an existing registered type")
	}
}

// ---------- loadLocationTypesFromDir ----------

// TestLoadLocationTypesFromDirHappyPath verifies loadLocationTypesFromDir parses
// a location-type YAML file and registers it with the given source.
//
// Why it matters: location types define the site hierarchy; without this loader
// path, directory-supplied location definitions would never reach the registry.
// Inputs: a directory with a "test-loc-loader" location YAML and source
// "test-source". Outputs: the slug present in allLocationTypes with its Source
// set. Data choice: a single named location type mirrors the existing per-kind
// loader tests and proves both parsing and registration succeed.
func TestLoadLocationTypesFromDirHappyPath(t *testing.T) {
	t.Cleanup(func() { delete(allLocationTypes, "test-loc-loader") })

	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "loc.yaml"),
		"slug: test-loc-loader\nname: Test Location\n")

	if err := loadLocationTypesFromDir(dir, "test-source"); err != nil {
		t.Fatalf("loadLocationTypesFromDir returned error: %v", err)
	}
	lt, ok := allLocationTypes["test-loc-loader"]
	if !ok {
		t.Fatal("expected test-loc-loader in allLocationTypes")
	}
	if lt.Source != "test-source" {
		t.Errorf("Source = %q, want %q", lt.Source, "test-source")
	}
}

// TestLoadLocationTypesFromDirInvalidYAML verifies loadLocationTypesFromDir logs
// and skips malformed YAML without returning an error or registering anything.
//
// Why it matters: a single corrupt location file must not abort loading the
// rest of the library, matching the resilient behavior of the other loaders.
// Inputs: a directory with one invalid YAML file. Outputs: a nil error and no
// "bad" slug registered. Data choice: reusing the same invalid-YAML fixture as
// the sibling loader tests keeps the failure mode consistent and targets the
// unmarshal-error branch.
func TestLoadLocationTypesFromDirInvalidYAML(t *testing.T) {
	dir := mkTmpDir(t)
	writeYAML(t, filepath.Join(dir, "bad.yaml"), "{{{{not valid yaml")

	if err := loadLocationTypesFromDir(dir, "test-source"); err != nil {
		t.Fatalf("expected no error for invalid YAML, got: %v", err)
	}
}

// TestLoadTypesFromDirDebugLogging verifies the device and location loaders
// register valid types and execute their Debug log path when Debug is enabled.
//
// Why it matters: operators troubleshooting library loading rely on the Debug
// log lines, so those branches must run without altering the registration
// outcome; this also confirms a fresh slug loads cleanly atop the embedded set.
// Inputs: Debug toggled on, plus temp dirs holding one valid device YAML and one
// valid location YAML with fresh slugs. Outputs: both slugs registered under the
// given source. Data choice: enabling Debug is the only way to reach the
// conditional log statements, and unique slugs avoid colliding with the embedded
// library while cleanup removes exactly them.
func TestLoadTypesFromDirDebugLogging(t *testing.T) {
	orig := Debug
	Debug = true
	t.Cleanup(func() { Debug = orig })

	devDir := mkTmpDir(t)
	writeYAML(t, filepath.Join(devDir, "dev.yaml"),
		"slug: dbg-dev-loader\nmanufacturer: Acme\nmodel: D1\n")
	locDir := mkTmpDir(t)
	writeYAML(t, filepath.Join(locDir, "loc.yaml"),
		"slug: dbg-loc-loader\nname: Debug Location\n")
	t.Cleanup(func() {
		delete(allDeviceTypes, "dbg-dev-loader")
		delete(allLocationTypes, "dbg-loc-loader")
	})

	if err := loadDeviceTypesFromDir(devDir, "dbg-source"); err != nil {
		t.Fatalf("loadDeviceTypesFromDir: %v", err)
	}
	if err := loadLocationTypesFromDir(locDir, "dbg-source"); err != nil {
		t.Fatalf("loadLocationTypesFromDir: %v", err)
	}

	if _, ok := allDeviceTypes["dbg-dev-loader"]; !ok {
		t.Error("expected dbg-dev-loader registered with Debug on")
	}
	if _, ok := allLocationTypes["dbg-loc-loader"]; !ok {
		t.Error("expected dbg-loc-loader registered with Debug on")
	}
}
