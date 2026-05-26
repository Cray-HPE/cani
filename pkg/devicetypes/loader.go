package devicetypes

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadAll loads hardware types from all configured sources in priority order:
//  1. Built-in (embedded) types — always loaded first.
//  2. Local directory types — loaded from each entry in typesDirs.
//  3. Remote git repo types — cloned/pulled then loaded from each entry in typesRepos.
//
// When typesRepoPull is false each repo is cloned on first use but never pulled.
// Types registered first win; duplicates by slug are skipped.
func LoadAll(typesDirs, typesRepos []string, typesRepoClone, typesRepoPull bool) error {
	// 1. Built-in types are always available
	if err := LoadEmbedded(); err != nil {
		return fmt.Errorf("loading embedded types: %w", err)
	}

	// 2. Optional local directories
	for _, dir := range typesDirs {
		if dir == "" {
			continue
		}
		if err := LoadFromDir(dir, "local:"+dir); err != nil {
			return fmt.Errorf("loading types from dir %s: %w", dir, err)
		}
	}

	// 3. Optional remote git repositories
	for _, repo := range typesRepos {
		if repo == "" {
			continue
		}
		if err := LoadFromGitRepo(repo, typesRepoClone, typesRepoPull); err != nil {
			return fmt.Errorf("loading types from git %s: %w", repo, err)
		}
	}

	return nil
}

// LoadFromDir scans a root directory for hardware type subdirectories
// (device-types, module-types, cable-types, rack-types, inventory-types)
// and registers any YAML types found. Existing slugs are not overwritten.
func LoadFromDir(root, source string) error {
	// Fast path: load from a pre-built JSON cache when available.
	cachePath := filepath.Join(root, cacheFileName)
	if cache, err := readDirCache(cachePath); err == nil {
		registerCachedTypes(cache, source)
		return nil
	}

	// Slow path: walk YAML files and register types.
	subDirs := []struct {
		name string
		load func(dir, source string) error
	}{
		{"device-types", loadDeviceTypesFromDir},
		{"module-types", loadModuleTypesFromDir},
		{"cable-types", loadCableTypesFromDir},
		{"rack-types", loadRackTypesFromDir},
		{"inventory-types", loadFruTypesFromDir},
		{"location-types", loadLocationTypesFromDir},
	}
	for _, sd := range subDirs {
		dir := filepath.Join(root, sd.name)
		if !dirExists(dir) {
			continue
		}
		if err := sd.load(dir, source); err != nil {
			return fmt.Errorf("loading %s from %s: %w", sd.name, root, err)
		}
	}

	// Persist cache for next invocation.
	writeDirCache(cachePath, collectTypesFromSource(source))
	return nil
}

func loadDeviceTypesFromDir(dir, source string) error {
	return walkYAMLFiles(dir, func(data []byte, path string) {
		forEachYAMLDoc(data, func(doc []byte) {
			var dt CaniDeviceType
			if err := yaml.Unmarshal(doc, &dt); err != nil {
				log.Printf("Warning: failed to parse device type %s: %v", path, err)
				return
			}
			if dt.Slug == "" {
				return
			}
			if _, exists := allDeviceTypes[dt.Slug]; exists {
				return
			}
			dt.Source = source
			RegisterDeviceType(dt)
			if Debug {
				log.Printf("Loaded device type: %s [%s]", dt.Slug, source)
			}
		})
	})
}

func loadModuleTypesFromDir(dir, source string) error {
	return walkYAMLFiles(dir, func(data []byte, path string) {
		var mt CaniModuleType
		if err := yaml.Unmarshal(data, &mt); err != nil {
			log.Printf("Warning: failed to parse module type %s: %v", path, err)
			return
		}
		if mt.Slug == "" {
			return
		}
		if _, exists := allModuleTypes[mt.Slug]; exists {
			return
		}
		mt.Source = source
		RegisterModuleType(mt)
		if Debug {
			log.Printf("Loaded module type: %s [%s]", mt.Slug, source)
		}
	})
}

func loadCableTypesFromDir(dir, source string) error {
	return walkYAMLFiles(dir, func(data []byte, path string) {
		var ct CaniCableType
		if err := yaml.Unmarshal(data, &ct); err != nil {
			log.Printf("Warning: failed to parse cable type %s: %v", path, err)
			return
		}
		if ct.Slug == "" {
			return
		}
		if _, exists := allCableTypes[ct.Slug]; exists {
			return
		}
		ct.Source = source
		RegisterCableType(ct)
		if Debug {
			log.Printf("Loaded cable type: %s [%s]", ct.Slug, source)
		}
	})
}

func loadRackTypesFromDir(dir, source string) error {
	return walkYAMLFiles(dir, func(data []byte, path string) {
		var rt CaniRackType
		if err := yaml.Unmarshal(data, &rt); err != nil {
			log.Printf("Warning: failed to parse rack type %s: %v", path, err)
			return
		}
		if rt.Slug == "" {
			return
		}
		if _, exists := allRackTypes[rt.Slug]; exists {
			return
		}
		rt.Source = source
		RegisterRackType(rt)
		if Debug {
			log.Printf("Loaded rack type: %s [%s]", rt.Slug, source)
		}
	})
}

func loadFruTypesFromDir(dir, source string) error {
	return walkYAMLFiles(dir, func(data []byte, path string) {
		var ft CaniFruType
		if err := yaml.Unmarshal(data, &ft); err != nil {
			log.Printf("Warning: failed to parse FRU type %s: %v", path, err)
			return
		}
		if ft.Slug == "" {
			return
		}
		if _, exists := allFruTypes[ft.Slug]; exists {
			return
		}
		ft.Source = source
		RegisterFruType(ft)
		if Debug {
			log.Printf("Loaded FRU type: %s [%s]", ft.Slug, source)
		}
	})
}

func loadLocationTypesFromDir(dir, source string) error {
	return walkYAMLFiles(dir, func(data []byte, path string) {
		var lt LocationTypeDefinition
		if err := yaml.Unmarshal(data, &lt); err != nil {
			log.Printf("Warning: failed to parse location type %s: %v", path, err)
			return
		}
		if lt.Slug == "" {
			return
		}
		if _, exists := allLocationTypes[lt.Slug]; exists {
			return
		}
		lt.Source = source
		RegisterLocationType(lt)
		if Debug {
			log.Printf("Loaded location type: %s [%s]", lt.Slug, source)
		}
	})
}

// walkYAMLFiles walks a directory tree and calls fn for every YAML file found.
func walkYAMLFiles(dir string, fn func(data []byte, path string)) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !isYAML(path) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", path, err)
			return nil
		}
		fn(data, path)
		return nil
	})
}

// dirExists returns true if path is an existing directory.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// sanitizeRepoName converts a git URL into a safe directory name.
func sanitizeRepoName(url string) string {
	name := url
	// Strip common prefixes
	for _, prefix := range []string{"https://", "http://", "git@", "ssh://"} {
		name = strings.TrimPrefix(name, prefix)
	}
	// Replace path separators and special chars with dashes
	r := strings.NewReplacer("/", "-", ":", "-", ".git", "")
	return r.Replace(name)
}

// forEachYAMLDoc splits multi-document YAML data (separated by "---")
// and calls fn for each document. Single-document files work as well.
func forEachYAMLDoc(data []byte, fn func(doc []byte)) {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	for {
		var node yaml.Node
		if err := dec.Decode(&node); err != nil {
			break // io.EOF or parse error — stop iteration
		}
		doc, err := yaml.Marshal(&node)
		if err != nil {
			continue
		}
		fn(doc)
	}
}
