package config

// Test coverage for config.go
//
// | Function       | Happy-path test                  | Failure test                       |
// |----------------|----------------------------------|------------------------------------|
// | GetNestedValue | TestGetNestedValueReturnsDeepKey  | TestGetNestedValueMissingProvider   |

import "testing"

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
