package devicetypes

import (
	"embed"
	"io/fs"
	"log"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Embedded filesystem variables containing built-in hardware type YAML files.
// These are compiled into the binary so built-in types are always available.
var (
	//go:embed device-types
	embeddedDeviceTypes embed.FS

	//go:embed module-types
	embeddedModuleTypes embed.FS

	//go:embed cable-types
	embeddedCableTypes embed.FS

	//go:embed rack-types
	embeddedRackTypes embed.FS
)

const sourceBuiltin = "builtin"

var embeddedOnce sync.Once

func init() {
	// Load embedded types at package init so they are available
	// before Cobra's PersistentPreRunE (e.g. Args validators).
	embeddedOnce.Do(func() {
		if err := loadAllEmbedded(); err != nil {
			log.Fatalf("failed to load embedded types: %v", err)
		}
	})
}

// LoadEmbedded is a no-op if embedded types were already loaded at init.
// It exists so callers can still reference it explicitly without harm.
func LoadEmbedded() error {
	var err error
	embeddedOnce.Do(func() {
		err = loadAllEmbedded()
	})
	return err
}

func loadAllEmbedded() error {
	if err := loadEmbeddedDeviceTypes(); err != nil {
		return err
	}
	if err := loadEmbeddedModuleTypes(); err != nil {
		return err
	}
	if err := loadEmbeddedCableTypes(); err != nil {
		return err
	}
	if err := loadEmbeddedRackTypes(); err != nil {
		return err
	}
	return nil
}

func loadEmbeddedDeviceTypes() error {
	return fs.WalkDir(embeddedDeviceTypes, "device-types", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := embeddedDeviceTypes.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read embedded device type %s: %v", path, err)
			return nil
		}
		var dt CaniDeviceType
		if err := yaml.Unmarshal(data, &dt); err != nil {
			log.Printf("Warning: failed to parse embedded device type %s: %v", path, err)
			return nil
		}
		if dt.Slug == "" {
			return nil
		}
		dt.Source = sourceBuiltin
		RegisterDeviceType(dt)
		return nil
	})
}

func loadEmbeddedModuleTypes() error {
	return fs.WalkDir(embeddedModuleTypes, "module-types", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := embeddedModuleTypes.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read embedded module type %s: %v", path, err)
			return nil
		}
		var mt CaniModuleType
		if err := yaml.Unmarshal(data, &mt); err != nil {
			log.Printf("Warning: failed to parse embedded module type %s: %v", path, err)
			return nil
		}
		if mt.Slug == "" {
			return nil
		}
		mt.Source = sourceBuiltin
		RegisterModuleType(mt)
		return nil
	})
}

func loadEmbeddedCableTypes() error {
	return fs.WalkDir(embeddedCableTypes, "cable-types", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := embeddedCableTypes.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read embedded cable type %s: %v", path, err)
			return nil
		}
		var ct CaniCableType
		if err := yaml.Unmarshal(data, &ct); err != nil {
			log.Printf("Warning: failed to parse embedded cable type %s: %v", path, err)
			return nil
		}
		if ct.Slug == "" {
			return nil
		}
		ct.Source = sourceBuiltin
		RegisterCableType(ct)
		return nil
	})
}

func loadEmbeddedRackTypes() error {
	return fs.WalkDir(embeddedRackTypes, "rack-types", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := embeddedRackTypes.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read embedded rack type %s: %v", path, err)
			return nil
		}
		var rt CaniRackType
		if err := yaml.Unmarshal(data, &rt); err != nil {
			log.Printf("Warning: failed to parse embedded rack type %s: %v", path, err)
			return nil
		}
		if rt.Slug == "" {
			return nil
		}
		rt.Source = sourceBuiltin
		RegisterRackType(rt)
		return nil
	})
}

// isYAML returns true if the file path has a .yaml or .yml extension.
func isYAML(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}
