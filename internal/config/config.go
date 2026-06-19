/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/Cray-HPE/cani/internal/core"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Providers      map[string]map[string]any `yaml:"providers" head_comment:"A map of settings for each provider"`
	Path           string                    `yaml:"-"`
	Datastore      string                    `yaml:"datastore" line_comment:"Path to the datastore file"`
	Debug          bool                      `yaml:"debug" line_comment:"Enable debug logging"`
	Strict         bool                      `yaml:"strict" line_comment:"Require resolved device type (slug) for all devices"`
	TypesDirs      []string                  `yaml:"types_dirs" head_comment:"Local directories with extra hardware types"`
	TypesRepos     []string                  `yaml:"types_repos" head_comment:"Remote git repos with extra hardware types"`
	TypesRepoClone bool                      `yaml:"types_repo_clone" line_comment:"Clone types repos that are not yet cached locally"`
	TypesRepoPull  bool                      `yaml:"types_repo_pull" line_comment:"Pull latest from types repos on startup"`
	StepMode       bool                      `yaml:"-"`
	NoColor        bool                      `yaml:"-"`
	RootNode       *yaml.Node                `yaml:"-"`
}

// DefaultTypesRepo is the netbox-community device type library, used as the
// default entry in types_repos when creating a new config file.
const DefaultTypesRepo = "https://github.com/netbox-community/devicetype-library.git"

var Cfg *Config // the singleton

// GetNestedValue safely retrieves a nested value from the provider configuration
// Returns the value and a boolean indicating if the value was found
func GetNestedValue(providerName string, keys ...string) (interface{}, bool) {
	if Cfg == nil || Cfg.Providers == nil {
		return nil, false
	}

	provider, ok := Cfg.Providers[providerName]
	if !ok {
		return nil, false
	}

	current := interface{}(provider)

	for _, key := range keys {
		switch v := current.(type) {
		case map[string]any:
			if val, exists := v[key]; exists {
				current = val
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// GetNestedString safely retrieves a nested string value with a default fallback
func GetNestedString(providerName string, defaultValue string, keys ...string) string {
	if value, ok := GetNestedValue(providerName, keys...); ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetNestedInt safely retrieves a nested int value with a default fallback
func GetNestedInt(providerName string, defaultValue int, keys ...string) int {
	if value, ok := GetNestedValue(providerName, keys...); ok {
		switch v := value.(type) {
		case int:
			return v
		case float64: // JSON numbers are often float64
			return int(v)
		}
	}
	return defaultValue
}

// HasTopLevelKey reports whether the loaded config file explicitly contained the
// given top-level key.  It lets callers distinguish a value that came from the
// file from a struct zero value, preserving flag defaults during precedence
// resolution (the role viper's config layer previously played).
func HasTopLevelKey(name string) bool {
	if Cfg == nil || Cfg.RootNode == nil || len(Cfg.RootNode.Content) == 0 {
		return false
	}
	mapping := Cfg.RootNode.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == name {
			return true
		}
	}
	return false
}

// Load reads or creates the config at path.
// It preserves the YAML node tree for comment and ordering preservation on save.
func Load(path string) error {
	Cfg = &Config{Providers: map[string]map[string]any{}}
	Cfg.Path = path
	Cfg.Datastore = core.DsPath

	f, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, create a default config and save it
		if os.IsNotExist(err) {
			log.Printf("Config file not found at %s, creating default config", path)
			// Initialize an empty root node for new files
			Cfg.RootNode = &yaml.Node{
				Kind: yaml.DocumentNode,
				Content: []*yaml.Node{{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				}},
			}
			return Save(path)
		}
		return err
	}
	defer f.Close()

	// Step 1: Decode into a Node tree (preserves comments and ordering)
	var root yaml.Node
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&root); err != nil {
		return err
	}

	// Step 1.5: Migrate legacy config formats (0.1.x–0.5.x) to 0.6.x
	if isLegacyFormat(&root) {
		f.Close() // release before rename
		if err := backupConfig(path); err != nil {
			return fmt.Errorf("backing up legacy config: %w", err)
		}
		newRoot, err := migrateConfig(&root)
		if err != nil {
			return fmt.Errorf("migrating legacy config: %w", err)
		}
		Cfg.RootNode = newRoot
		if err := newRoot.Decode(Cfg); err != nil {
			return err
		}
		log.Printf("Migrated legacy config; backup saved to %s.canisave", path)
		return Save(path)
	}

	Cfg.RootNode = &root

	// Step 2: Also decode into the Config struct for convenient programmatic access
	if err := root.Decode(Cfg); err != nil {
		return err
	}

	if Cfg.Debug {
		log.Printf("Loaded config: %s", Cfg.Path)
	}
	return nil
}
