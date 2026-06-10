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
	"path/filepath"
	"sort"

	"github.com/Cray-HPE/cani/internal/core"
	"github.com/Cray-HPE/cani/internal/provider"
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

// findNodeByKey finds a key's value node in a MappingNode
// Returns the value node and the index of the key node, or nil and -1 if not found
func findNodeByKey(mapNode *yaml.Node, key string) (*yaml.Node, int) {
	if mapNode == nil || mapNode.Kind != yaml.MappingNode {
		return nil, -1
	}
	// Content is [key, value, key, value, ...]
	for i := 0; i+1 < len(mapNode.Content); i += 2 {
		if mapNode.Content[i].Value == key {
			return mapNode.Content[i+1], i
		}
	}
	return nil, -1
}

// setOrAddKey updates an existing key or adds a new one to a MappingNode
func setOrAddKey(mapNode *yaml.Node, key string, valueNode *yaml.Node) {
	if mapNode == nil || mapNode.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(mapNode.Content); i += 2 {
		if mapNode.Content[i].Value == key {
			mapNode.Content[i+1] = valueNode
			return
		}
	}
	// Key doesn't exist, add it
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	mapNode.Content = append(mapNode.Content, keyNode, valueNode)
}

// valueToNode converts a Go value to a yaml.Node
func valueToNode(value interface{}) *yaml.Node {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		// Fallback to string representation
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprint(value)}
	}
	return &node
}

// mergeIntoNode recursively merges defaults into an existing MappingNode
// Only adds keys that don't already exist (preserves user values)
// New keys are added in sorted order for deterministic output
func mergeIntoNode(target *yaml.Node, defaults map[string]any) error {
	if target == nil {
		return fmt.Errorf("target node is nil")
	}
	if target.Kind != yaml.MappingNode {
		return fmt.Errorf("target must be a mapping node, got kind %d", target.Kind)
	}

	// Build set of existing keys
	existingKeys := make(map[string]bool)
	for i := 0; i+1 < len(target.Content); i += 2 {
		existingKeys[target.Content[i].Value] = true
	}

	// Collect missing keys and sort them for deterministic order
	var missingKeys []string
	for key := range defaults {
		if !existingKeys[key] {
			missingKeys = append(missingKeys, key)
		}
	}
	sort.Strings(missingKeys)

	// Add missing keys in sorted order
	for _, key := range missingKeys {
		value := defaults[key]
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
		valueNode := valueToNode(value)
		target.Content = append(target.Content, keyNode, valueNode)
	}

	// Recursively merge nested maps
	for key, value := range defaults {
		if nestedDefaults, ok := value.(map[string]any); ok {
			if existingNode, _ := findNodeByKey(target, key); existingNode != nil {
				if existingNode.Kind == yaml.MappingNode {
					if err := mergeIntoNode(existingNode, nestedDefaults); err != nil {
						return fmt.Errorf("failed to merge nested key %q: %w", key, err)
					}
				}
			}
		}
	}

	return nil
}

// ensureNodePath ensures a path of keys exists in a node tree, creating missing nodes as needed
// Returns the node at the end of the path
func ensureNodePath(root *yaml.Node, keys ...string) *yaml.Node {
	if root == nil || len(keys) == 0 {
		return root
	}

	current := root
	for _, key := range keys {
		if current.Kind != yaml.MappingNode {
			return nil
		}
		child, _ := findNodeByKey(current, key)
		if child == nil {
			// Create the missing node
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
			child = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			current.Content = append(current.Content, keyNode, child)
		}
		current = child
	}
	return current
}

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

// Save writes Cfg back to path, preserving existing YAML structure, comments, and ordering.
// Only missing keys are added with default values.
func Save(path string) error {
	// Ensure the directory exists. Use 0700 because the config may hold
	// provider API tokens; it must not be readable by other local users.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Ensure we have a root node to work with
	if Cfg.RootNode == nil {
		Cfg.RootNode = &yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			}},
		}
	}

	// Get the document content (the main mapping node)
	docContent := Cfg.RootNode.Content[0]

	// Ensure "providers" key exists
	providersNode, _ := findNodeByKey(docContent, "providers")
	if providersNode == nil {
		providersNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "providers",
		}
		docContent.Content = append([]*yaml.Node{keyNode, providersNode}, docContent.Content...)
	}

	// Ensure "datastore" key exists
	datastoreNode, _ := findNodeByKey(docContent, "datastore")
	if datastoreNode == nil {
		setOrAddKey(docContent, "datastore", &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: Cfg.Datastore,
		})
	}

	// Ensure "debug" key exists (default: false)
	debugNode, _ := findNodeByKey(docContent, "debug")
	if debugNode == nil {
		setOrAddKey(docContent, "debug", &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: "false",
		})
	}

	// Ensure "types_dirs" key exists (default: empty list)
	typesDirsNode, _ := findNodeByKey(docContent, "types_dirs")
	if typesDirsNode == nil {
		setOrAddKey(docContent, "types_dirs", &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		})
	}

	// Ensure "types_repos" key exists (default: list with DefaultTypesRepo)
	typesReposNode, _ := findNodeByKey(docContent, "types_repos")
	if typesReposNode == nil {
		setOrAddKey(docContent, "types_repos", &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
			Content: []*yaml.Node{{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: DefaultTypesRepo,
			}},
		})
	}

	// Ensure "types_repo_clone" key exists (default: false)
	typesRepoCloneNode, _ := findNodeByKey(docContent, "types_repo_clone")
	if typesRepoCloneNode == nil {
		setOrAddKey(docContent, "types_repo_clone", &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: "false",
		})
	}

	// Ensure "types_repo_pull" key exists (default: false)
	typesRepoPullNode, _ := findNodeByKey(docContent, "types_repo_pull")
	if typesRepoPullNode == nil {
		setOrAddKey(docContent, "types_repo_pull", &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: "false",
		})
	}

	// Process each registered provider
	for name, p := range provider.GetProviders() {
		// Ensure provider section exists in node tree
		providerNode, _ := findNodeByKey(providersNode, name)
		if providerNode == nil {
			providerNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: name}
			providersNode.Content = append(providersNode.Content, keyNode, providerNode)
		}

		// Ensure Cfg.Providers map is in sync
		if _, ok := Cfg.Providers[name]; !ok {
			Cfg.Providers[name] = map[string]any{}
		}

		// For providers with HasOptions: merge top-level defaults (only missing keys)
		if hasOptions, ok := p.(provider.HasOptions); ok {
			defaults := hasOptions.GetDefaultOptions()
			// Filter out import/export from top-level defaults (handled separately below)
			topLevelDefaults := make(map[string]any)
			for k, v := range defaults {
				if k != "import" && k != "export" {
					topLevelDefaults[k] = v
				}
			}
			if err := mergeIntoNode(providerNode, topLevelDefaults); err != nil {
				log.Printf("Warning: failed to merge defaults for provider %s: %v", name, err)
			}
		}

		// Auto-migrate: ensure import section exists with defaults
		if hasImport, ok := p.(provider.HasImportOptions); ok {
			importNode := ensureNodePath(providerNode, "import")
			if importNode != nil {
				defaults := hasImport.GetImportDefaults()
				if err := mergeIntoNode(importNode, defaults); err != nil {
					log.Printf("Warning: failed to merge import defaults for provider %s: %v", name, err)
				}
			}
		}

		// Auto-migrate: ensure export section exists with defaults
		if hasExport, ok := p.(provider.HasExportOptions); ok {
			exportNode := ensureNodePath(providerNode, "export")
			if exportNode != nil {
				defaults := hasExport.GetExportDefaults()
				if err := mergeIntoNode(exportNode, defaults); err != nil {
					log.Printf("Warning: failed to merge export defaults for provider %s: %v", name, err)
				}
			}
		}
	}

	// Apply comments from struct tags to top-level config keys
	applyComments(docContent, extractComments(Config{}))

	// Apply comments from provider option structs
	for name, p := range provider.GetProviders() {
		providerNode, _ := findNodeByKey(providersNode, name)
		if providerNode == nil {
			continue
		}
		if hasOptions, ok := p.(provider.HasOptions); ok {
			applyComments(providerNode, extractComments(hasOptions.GetOptionsStruct()))
		}
		if hasImport, ok := p.(provider.HasImportOptions); ok {
			if importNode, _ := findNodeByKey(providerNode, "import"); importNode != nil {
				applyComments(importNode, extractComments(hasImport.GetImportOptionsStruct()))
			}
		}
		if hasExport, ok := p.(provider.HasExportOptions); ok {
			if exportNode, _ := findNodeByKey(providerNode, "export"); exportNode != nil {
				applyComments(exportNode, extractComments(hasExport.GetExportOptionsStruct()))
			}
		}
	}

	// Write the node tree to file. Use 0600 because the config may contain
	// provider API tokens; it must not be readable by other local users.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Tighten permissions on a pre-existing file that may have been created
	// with a looser umask before this safeguard existed.
	if err := f.Chmod(0600); err != nil {
		return err
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(Cfg.RootNode); err != nil {
		return err
	}

	return nil
}
