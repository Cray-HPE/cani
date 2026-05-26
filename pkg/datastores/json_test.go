package datastores

// | Function     | Happy-path test          | Failure test              |
// |--------------|--------------------------|---------------------------|
// | NewJSONStore | TestNewJSONStoreHappyPath | TestNewJSONStoreNilConfig |
// | Load         | TestLoadHappyPath        | TestLoadInvalidJSON       |
// | Save         | TestSaveHappyPath        | TestSaveInvalidPath       |

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func TestNewJSONStoreHappyPath(t *testing.T) {
	original := config.Cfg
	config.Cfg = &config.Config{
		Path:      "/tmp/cani-test/config.yaml",
		Datastore: "inventory.json",
	}
	defer func() { config.Cfg = original }()

	store := NewJSONStore()

	expected := filepath.Join("/tmp/cani-test", "inventory.json")
	if store.Path != expected {
		t.Errorf("expected path %q, got %q", expected, store.Path)
	}
}

func TestNewJSONStoreAbsolutePath(t *testing.T) {
	original := config.Cfg
	config.Cfg = &config.Config{
		Path:      "/tmp/cani-test/config.yaml",
		Datastore: "/tmp/override/test.json",
	}
	defer func() { config.Cfg = original }()

	store := NewJSONStore()

	expected := "/tmp/override/test.json"
	if store.Path != expected {
		t.Errorf("expected path %q, got %q", expected, store.Path)
	}
}

func TestNewJSONStoreNilConfig(t *testing.T) {
	original := config.Cfg
	config.Cfg = nil
	defer func() { config.Cfg = original }()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when config.Cfg is nil, but did not panic")
		}
	}()

	NewJSONStore()
}

func TestLoadHappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.json")

	inv := devicetypes.NewInventory()
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		t.Fatalf("marshaling test inventory: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	store := &JSONStore{Path: path}
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load() returned nil inventory")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.json")

	if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	store := &JSONStore{Path: path}
	_, err := store.Load()
	if err == nil {
		t.Error("Load() expected error for invalid JSON, got nil")
	}
}

func TestSaveHappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "inventory.json")

	store := &JSONStore{Path: path}
	inv := devicetypes.NewInventory()

	if err := store.Save(inv); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create the file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}

	var loaded devicetypes.Inventory
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Errorf("saved file contains invalid JSON: %v", err)
	}
}

func TestSaveInvalidPath(t *testing.T) {
	dir := t.TempDir()

	// Create a file where a directory is expected
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatalf("creating blocker file: %v", err)
	}

	// Path requires "blocker" to be a directory, but it is a file
	path := filepath.Join(blocker, "sub", "inventory.json")

	store := &JSONStore{Path: path}
	inv := devicetypes.NewInventory()

	if err := store.Save(inv); err == nil {
		t.Error("Save() expected error for invalid path, got nil")
	}
}
