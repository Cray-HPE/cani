package config

// Test coverage for config.go
//
// | Function       | Happy-path test                  | Failure test                       |
// |----------------|----------------------------------|------------------------------------|
// | GetNestedValue | TestGetNestedValueReturnsDeepKey  | TestGetNestedValueMissingProvider   |

import (
	"os"
	"testing"
)

// ---------- GetNestedValue ----------

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
