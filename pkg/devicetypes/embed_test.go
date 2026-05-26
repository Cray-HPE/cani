package devicetypes

// Test coverage for embed.go
//
// | Function               | Happy-path test                                | Failure test                                     |
// |------------------------|------------------------------------------------|--------------------------------------------------|
// | LoadEmbedded           | TestLoadEmbeddedIdempotent                     | TestLoadEmbeddedRegistriesPopulated              |
// | loadAllEmbedded        | TestLoadAllEmbeddedPopulatesAll                | TestLoadAllEmbeddedNoEmptySlugs                  |
// | loadEmbeddedDeviceTypes| TestLoadEmbeddedDeviceTypesPopulates           | TestLoadEmbeddedDeviceTypesSourceBuiltin         |
// | loadEmbeddedModuleTypes| TestLoadEmbeddedModuleTypesPopulates           | TestLoadEmbeddedModuleTypesSourceBuiltin         |
// | loadEmbeddedCableTypes | TestLoadEmbeddedCableTypesPopulates            | TestLoadEmbeddedCableTypesSourceBuiltin          |
// | loadEmbeddedRackTypes  | TestLoadEmbeddedRackTypesPopulates             | TestLoadEmbeddedRackTypesSourceBuiltin           |
// | isYAML                 | TestIsYAMLTrueForYamlExtensions                | TestIsYAMLFalseForNonYamlExtensions              |

import "testing"

// ---------- isYAML ----------

func TestIsYAMLTrueForYamlExtensions(t *testing.T) {
	cases := []string{
		"device-types/HPE/hpe-node.yaml",
		"rack-types/HPE/rack.yml",
		"simple.yaml",
		"deep/path/file.yml",
	}
	for _, path := range cases {
		if !isYAML(path) {
			t.Errorf("expected isYAML(%q) = true", path)
		}
	}
}

func TestIsYAMLFalseForNonYamlExtensions(t *testing.T) {
	cases := []string{
		"README.md",
		"config.json",
		"image.png",
		"data.txt",
		"no-extension",
		"tricky.yamll",
		"tricky.yamlx",
	}
	for _, path := range cases {
		if isYAML(path) {
			t.Errorf("expected isYAML(%q) = false", path)
		}
	}
}

// ---------- LoadEmbedded ----------

func TestLoadEmbeddedIdempotent(t *testing.T) {
	// LoadEmbedded is a no-op after init() already loaded; must return nil.
	err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded returned error on idempotent call: %v", err)
	}
}

func TestLoadEmbeddedRegistriesPopulated(t *testing.T) {
	// Verify that loading embedded types populates at least one device type.
	// Reset first because earlier tests (all_test.go) may have cleared registries
	// and LoadEmbedded is a no-op after init()'s sync.Once.
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadAllEmbedded(); err != nil {
		t.Fatalf("loadAllEmbedded returned error: %v", err)
	}
	if len(All()) == 0 {
		t.Fatal("expected at least one device type after LoadEmbedded, got 0")
	}
}

// ---------- loadAllEmbedded ----------

func TestLoadAllEmbeddedPopulatesAll(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }() // restore state for other tests

	if err := loadAllEmbedded(); err != nil {
		t.Fatalf("loadAllEmbedded returned error: %v", err)
	}
	if len(All()) == 0 {
		t.Error("expected device types to be populated")
	}
	if len(AllModules()) == 0 {
		t.Error("expected module types to be populated")
	}
	if len(AllCables()) == 0 {
		t.Error("expected cable types to be populated")
	}
	if len(AllRackTypes()) == 0 {
		t.Error("expected rack types to be populated")
	}
}

func TestLoadAllEmbeddedNoEmptySlugs(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadAllEmbedded(); err != nil {
		t.Fatalf("loadAllEmbedded returned error: %v", err)
	}
	for slug := range All() {
		if slug == "" {
			t.Fatal("found device type with empty slug in registry")
		}
	}
	for slug := range AllCables() {
		if slug == "" {
			t.Fatal("found cable type with empty slug in registry")
		}
	}
	for slug := range AllRackTypes() {
		if slug == "" {
			t.Fatal("found rack type with empty slug in registry")
		}
	}
}

// ---------- loadEmbeddedDeviceTypes ----------

func TestLoadEmbeddedDeviceTypesPopulates(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedDeviceTypes(); err != nil {
		t.Fatalf("loadEmbeddedDeviceTypes returned error: %v", err)
	}
	if len(All()) == 0 {
		t.Fatal("expected at least one device type after loading embedded device types")
	}
}

func TestLoadEmbeddedDeviceTypesSourceBuiltin(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedDeviceTypes(); err != nil {
		t.Fatalf("loadEmbeddedDeviceTypes returned error: %v", err)
	}
	for slug, dt := range All() {
		if dt.Source != sourceBuiltin {
			t.Errorf("device type %q has Source=%q, want %q", slug, dt.Source, sourceBuiltin)
		}
	}
}

// ---------- loadEmbeddedModuleTypes ----------

func TestLoadEmbeddedModuleTypesPopulates(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedModuleTypes(); err != nil {
		t.Fatalf("loadEmbeddedModuleTypes returned error: %v", err)
	}
	if len(AllModules()) == 0 {
		t.Fatal("expected at least one module type after loading embedded module types")
	}
}

func TestLoadEmbeddedModuleTypesSourceBuiltin(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedModuleTypes(); err != nil {
		t.Fatalf("loadEmbeddedModuleTypes returned error: %v", err)
	}
	for slug, mt := range AllModules() {
		if mt.Source != sourceBuiltin {
			t.Errorf("module type %q has Source=%q, want %q", slug, mt.Source, sourceBuiltin)
		}
	}
}

// ---------- loadEmbeddedCableTypes ----------

func TestLoadEmbeddedCableTypesPopulates(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedCableTypes(); err != nil {
		t.Fatalf("loadEmbeddedCableTypes returned error: %v", err)
	}
	if len(AllCables()) == 0 {
		t.Fatal("expected at least one cable type after loading embedded cable types")
	}
}

func TestLoadEmbeddedCableTypesSourceBuiltin(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedCableTypes(); err != nil {
		t.Fatalf("loadEmbeddedCableTypes returned error: %v", err)
	}
	for slug, ct := range AllCables() {
		if ct.Source != sourceBuiltin {
			t.Errorf("cable type %q has Source=%q, want %q", slug, ct.Source, sourceBuiltin)
		}
	}
}

// ---------- loadEmbeddedRackTypes ----------

func TestLoadEmbeddedRackTypesPopulates(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedRackTypes(); err != nil {
		t.Fatalf("loadEmbeddedRackTypes returned error: %v", err)
	}
	if len(AllRackTypes()) == 0 {
		t.Fatal("expected at least one rack type after loading embedded rack types")
	}
}

func TestLoadEmbeddedRackTypesSourceBuiltin(t *testing.T) {
	resetRegistries()
	defer func() { _ = loadAllEmbedded() }()

	if err := loadEmbeddedRackTypes(); err != nil {
		t.Fatalf("loadEmbeddedRackTypes returned error: %v", err)
	}
	for slug, rt := range AllRackTypes() {
		if rt.Source != sourceBuiltin {
			t.Errorf("rack type %q has Source=%q, want %q", slug, rt.Source, sourceBuiltin)
		}
	}
}
