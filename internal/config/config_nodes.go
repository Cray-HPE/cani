/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
	"sort"

	"gopkg.in/yaml.v3"
)

// YAML node tags used when building config nodes.
const (
	tagStr  = "!!str"
	tagBool = "!!bool"
	tagMap  = "!!map"
	tagSeq  = "!!seq"
)

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
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tagStr, Value: key}
	mapNode.Content = append(mapNode.Content, keyNode, valueNode)
}

// valueToNode converts a Go value to a yaml.Node
func valueToNode(value interface{}) *yaml.Node {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		// Fallback to string representation
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tagStr, Value: fmt.Sprint(value)}
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

	existingKeys := collectExistingKeys(target)
	addMissingKeys(target, defaults, existingKeys)
	return mergeNestedMaps(target, defaults)
}

// collectExistingKeys builds a set of the keys already present in a MappingNode.
func collectExistingKeys(target *yaml.Node) map[string]bool {
	existingKeys := make(map[string]bool)
	for i := 0; i+1 < len(target.Content); i += 2 {
		existingKeys[target.Content[i].Value] = true
	}
	return existingKeys
}

// addMissingKeys appends defaults not already present, in sorted order for
// deterministic output.
func addMissingKeys(target *yaml.Node, defaults map[string]any, existingKeys map[string]bool) {
	var missingKeys []string
	for key := range defaults {
		if !existingKeys[key] {
			missingKeys = append(missingKeys, key)
		}
	}
	sort.Strings(missingKeys)

	for _, key := range missingKeys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tagStr, Value: key}
		valueNode := valueToNode(defaults[key])
		target.Content = append(target.Content, keyNode, valueNode)
	}
}

// mergeNestedMaps recurses into nested map defaults whose key already exists as
// a mapping node in target.
func mergeNestedMaps(target *yaml.Node, defaults map[string]any) error {
	for key, value := range defaults {
		nestedDefaults, ok := value.(map[string]any)
		if !ok {
			continue
		}
		existingNode, _ := findNodeByKey(target, key)
		if existingNode == nil || existingNode.Kind != yaml.MappingNode {
			continue
		}
		if err := mergeIntoNode(existingNode, nestedDefaults); err != nil {
			return fmt.Errorf("failed to merge nested key %q: %w", key, err)
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
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tagStr, Value: key}
			child = &yaml.Node{Kind: yaml.MappingNode, Tag: tagMap}
			current.Content = append(current.Content, keyNode, child)
		}
		current = child
	}
	return current
}
