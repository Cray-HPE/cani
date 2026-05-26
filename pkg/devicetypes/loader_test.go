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
