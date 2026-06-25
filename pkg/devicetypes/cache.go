package devicetypes

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const cacheFileName = ".cani-types-cache.json"

// dirCache holds pre-parsed hardware types for a single directory source.
// It is serialized to JSON so subsequent loads skip YAML parsing.
type dirCache struct {
	// Fingerprint identifies the source content a cache was built from. It is
	// only used for the embedded built-in types, where it is a hash of the
	// embedded files; a mismatch forces a rebuild. Directory and git sources
	// leave it empty and are validated by other means.
	Fingerprint   string                   `json:"fingerprint,omitempty"`
	DeviceTypes   []CaniDeviceType         `json:"device_types,omitempty"`
	ModuleTypes   []CaniModuleType         `json:"module_types,omitempty"`
	CableTypes    []CaniCableType          `json:"cable_types,omitempty"`
	RackTypes     []CaniRackType           `json:"rack_types,omitempty"`
	FruTypes      []CaniFruType            `json:"fru_types,omitempty"`
	LocationTypes []LocationTypeDefinition `json:"location_types,omitempty"`
}

// readDirCache loads a previously written cache file.
func readDirCache(path string) (*dirCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c dirCache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// writeDirCache persists the cache to disk. Errors are logged but
// not fatal — a missing cache just means the next load will re-parse.
// The write is atomic (temp file + rename) so concurrent writers, such as
// parallel `go test` processes priming the built-in cache, cannot observe a
// half-written file.
func writeDirCache(path string, c *dirCache) {
	data, err := json.Marshal(c)
	if err != nil {
		log.Printf("Warning: failed to marshal types cache: %v", err)
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Warning: failed to create types cache directory: %v", err)
		return
	}
	tmp, err := os.CreateTemp(dir, ".cani-types-cache-*.tmp")
	if err != nil {
		log.Printf("Warning: failed to create types cache temp file: %v", err)
		return
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		log.Printf("Warning: failed to write types cache %s: %v", path, err)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		log.Printf("Warning: failed to close types cache %s: %v", path, err)
		return
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		log.Printf("Warning: failed to replace types cache %s: %v", path, err)
	}
}

// registerCachedTypes registers all types from a cache, setting Source
// on each. Types whose slug is already registered are skipped.
func registerCachedTypes(c *dirCache, source string) {
	for _, dt := range c.DeviceTypes {
		if dt.Slug == "" {
			continue
		}
		if _, exists := allDeviceTypes[dt.Slug]; exists {
			continue
		}
		dt.Source = source
		RegisterDeviceType(dt)
	}
	for _, mt := range c.ModuleTypes {
		if mt.Slug == "" {
			continue
		}
		if _, exists := allModuleTypes[mt.Slug]; exists {
			continue
		}
		mt.Source = source
		RegisterModuleType(mt)
	}
	for _, ct := range c.CableTypes {
		if ct.Slug == "" {
			continue
		}
		if _, exists := allCableTypes[ct.Slug]; exists {
			continue
		}
		ct.Source = source
		RegisterCableType(ct)
	}
	for _, rt := range c.RackTypes {
		if rt.Slug == "" {
			continue
		}
		if _, exists := allRackTypes[rt.Slug]; exists {
			continue
		}
		rt.Source = source
		RegisterRackType(rt)
	}
	for _, ft := range c.FruTypes {
		if ft.Slug == "" {
			continue
		}
		if _, exists := allFruTypes[ft.Slug]; exists {
			continue
		}
		ft.Source = source
		RegisterFruType(ft)
	}
	for _, lt := range c.LocationTypes {
		if lt.Slug == "" {
			continue
		}
		if _, exists := allLocationTypes[lt.Slug]; exists {
			continue
		}
		lt.Source = source
		RegisterLocationType(lt)
	}
}

// collectTypesFromSource gathers all registered types that match
// the given source string into a dirCache for serialization.
func collectTypesFromSource(source string) *dirCache {
	c := &dirCache{}
	for _, dt := range allDeviceTypes {
		if dt.Source == source {
			c.DeviceTypes = append(c.DeviceTypes, dt)
		}
	}
	for _, mt := range allModuleTypes {
		if mt.Source == source {
			c.ModuleTypes = append(c.ModuleTypes, mt)
		}
	}
	for _, ct := range allCableTypes {
		if ct.Source == source {
			c.CableTypes = append(c.CableTypes, ct)
		}
	}
	for _, rt := range allRackTypes {
		if rt.Source == source {
			c.RackTypes = append(c.RackTypes, rt)
		}
	}
	for _, ft := range allFruTypes {
		if ft.Source == source {
			c.FruTypes = append(c.FruTypes, ft)
		}
	}
	for _, lt := range allLocationTypes {
		if lt.Source == source {
			c.LocationTypes = append(c.LocationTypes, lt)
		}
	}
	return c
}

// removeDirCache deletes the cache file for a directory. Called when
// the underlying source has changed (e.g. git pull).
func removeDirCache(root string) {
	path := filepath.Join(root, cacheFileName)
	os.Remove(path)
}
