package devicetypes

import (
	"embed"
	"hash"
	"hash/fnv"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

	//go:embed location-types
	embeddedLocationTypes embed.FS
)

const sourceBuiltin = "builtin"

var embeddedOnce sync.Once

func init() {
	// Load embedded types at package init so they are available
	// before PersistentPreRunE (e.g. Args validators).
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

// loadAllEmbedded registers the built-in hardware types. Parsing ~100 embedded
// YAML files on every process start is the dominant CLI startup cost, so the
// parsed result is cached to a JSON file keyed by a fingerprint of the embedded
// content. On a cache hit the YAML parse is skipped entirely; the cache rebuilds
// automatically whenever the embedded library changes (e.g. a new binary),
// because the fingerprint no longer matches.
func loadAllEmbedded() error {
	fingerprint := embeddedFingerprint()
	cachePath := builtinCachePath()

	if cachePath != "" {
		if cache, err := readDirCache(cachePath); err == nil && cache.Fingerprint == fingerprint {
			registerCachedTypes(cache, sourceBuiltin)
			return nil
		}
	}

	if err := loadEmbeddedTypes(); err != nil {
		return err
	}

	if cachePath != "" {
		cache := collectTypesFromSource(sourceBuiltin)
		cache.Fingerprint = fingerprint
		writeDirCache(cachePath, cache)
	}
	return nil
}

// loadEmbeddedTypes parses every embedded YAML file and registers the types.
func loadEmbeddedTypes() error {
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
	if err := loadEmbeddedLocationTypes(); err != nil {
		return err
	}
	return nil
}

// builtinCachePath returns the on-disk cache location for the built-in types,
// or "" when the home directory cannot be determined (in which case the cache
// is skipped and the YAML is parsed as before).
func builtinCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cani", "types-cache", "builtin", cacheFileName)
}

// embeddedFingerprint returns a fast non-cryptographic hash over the contents
// of every embedded type file. Reading the embedded bytes is an in-memory copy
// (the files are compiled into the binary), so this is far cheaper than parsing
// them as YAML.
func embeddedFingerprint() string {
	h := fnv.New64a()
	for _, efs := range []embed.FS{
		embeddedDeviceTypes,
		embeddedModuleTypes,
		embeddedCableTypes,
		embeddedRackTypes,
		embeddedLocationTypes,
	} {
		hashEmbeddedFS(h, efs)
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

// hashEmbeddedFS folds every YAML file's path and bytes in efs into h.
func hashEmbeddedFS(h hash.Hash64, efs embed.FS) {
	_ = fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := efs.ReadFile(path)
		if err != nil {
			return nil
		}
		_, _ = h.Write([]byte(path))
		_, _ = h.Write(data)
		return nil
	})
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
		forEachYAMLDoc(data, func(doc []byte) {
			var dt CaniDeviceType
			if err := yaml.Unmarshal(doc, &dt); err != nil {
				log.Printf("Warning: failed to parse embedded device type %s: %v", path, err)
				return
			}
			if dt.Slug == "" {
				return
			}
			dt.Source = sourceBuiltin
			RegisterDeviceType(dt)
		})
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

func loadEmbeddedLocationTypes() error {
	return fs.WalkDir(embeddedLocationTypes, "location-types", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := embeddedLocationTypes.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read embedded location type %s: %v", path, err)
			return nil
		}
		var lt LocationTypeDefinition
		if err := yaml.Unmarshal(data, &lt); err != nil {
			log.Printf("Warning: failed to parse embedded location type %s: %v", path, err)
			return nil
		}
		if lt.Slug == "" {
			return nil
		}
		lt.Source = sourceBuiltin
		RegisterLocationType(lt)
		return nil
	})
}

// isYAML returns true if the file path has a .yaml or .yml extension.
func isYAML(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}
