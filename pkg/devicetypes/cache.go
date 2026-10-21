package devicetypes

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const cacheFileName = ".cani-types-cache.json"

// dirCacheVersion must be bumped whenever YAML loading semantics change, so
// caches written by older binaries are discarded and rebuilt from YAML.
// Version 1: slugs are derived from manufacturer/model when absent.
// Version 2: caches hold every parsed type, including ones shadowed by an
// earlier source; version 1 caches could silently omit those.
const dirCacheVersion = 2

// dirCache holds pre-parsed hardware types for a single directory source.
// It is serialized to JSON so subsequent loads skip YAML parsing.
type dirCache struct {
	Version       int                      `json:"version"`
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
	if c.Version != dirCacheVersion {
		return nil, fmt.Errorf("types cache %s has version %d, want %d", path, c.Version, dirCacheVersion)
	}
	return &c, nil
}

// writeDirCache persists the cache to disk. Errors are logged but
// not fatal — a missing cache just means the next load will re-parse.
func writeDirCache(path string, c *dirCache) {
	c.Version = dirCacheVersion
	data, err := json.Marshal(c)
	if err != nil {
		log.Printf("Warning: failed to marshal types cache: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Printf("Warning: failed to create types cache directory: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("Warning: failed to write types cache %s: %v", path, err)
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

// removeDirCache deletes the cache file for a directory. Called when
// the underlying source has changed (e.g. git pull).
func removeDirCache(root string) {
	path := filepath.Join(root, cacheFileName)
	os.Remove(path)
}
