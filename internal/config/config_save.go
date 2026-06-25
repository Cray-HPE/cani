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
	"bytes"
	"log"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/internal/provider"
	"gopkg.in/yaml.v3"
)

const falseValue = "false"

const (
	keyImport = "import"
	keyExport = "export"
)

// Save writes Cfg back to path, preserving existing YAML structure, comments, and ordering.
// Only missing keys are added with default values.
func Save(path string) error {
	// Ensure the directory exists. Use 0700 because the config may hold
	// provider API tokens; it must not be readable by other local users.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	docContent := ensureRootDocument()
	providersNode := ensureProvidersNode(docContent)
	ensureTopLevelKeys(docContent)
	mergeProviderDefaults(providersNode)
	applyAllComments(docContent, providersNode)

	return writeConfigFile(path, Cfg.RootNode)
}

// ensureRootDocument ensures Cfg.RootNode holds a document with a mapping node
// and returns that mapping node.
func ensureRootDocument() *yaml.Node {
	if Cfg.RootNode == nil {
		Cfg.RootNode = &yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{{
				Kind: yaml.MappingNode,
				Tag:  tagMap,
			}},
		}
	}
	return Cfg.RootNode.Content[0]
}

// ensureProvidersNode ensures the "providers" mapping exists at the front of
// docContent and returns it.
func ensureProvidersNode(docContent *yaml.Node) *yaml.Node {
	providersNode, _ := findNodeByKey(docContent, "providers")
	if providersNode == nil {
		providersNode = &yaml.Node{Kind: yaml.MappingNode, Tag: tagMap}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tagStr, Value: "providers"}
		docContent.Content = append([]*yaml.Node{keyNode, providersNode}, docContent.Content...)
	}
	return providersNode
}

// ensureScalarKey adds a scalar key with the given tag and value if it is missing.
func ensureScalarKey(docContent *yaml.Node, key, tag, value string) {
	if node, _ := findNodeByKey(docContent, key); node == nil {
		setOrAddKey(docContent, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value})
	}
}

// ensureTopLevelKeys adds the default top-level config keys if they are missing.
func ensureTopLevelKeys(docContent *yaml.Node) {
	ensureScalarKey(docContent, "datastore", tagStr, Cfg.Datastore)
	ensureScalarKey(docContent, "debug", tagBool, falseValue)

	if node, _ := findNodeByKey(docContent, "types_dirs"); node == nil {
		setOrAddKey(docContent, "types_dirs", &yaml.Node{Kind: yaml.SequenceNode, Tag: tagSeq})
	}
	if node, _ := findNodeByKey(docContent, "types_repos"); node == nil {
		setOrAddKey(docContent, "types_repos", &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  tagSeq,
			Content: []*yaml.Node{{
				Kind:  yaml.ScalarNode,
				Tag:   tagStr,
				Value: DefaultTypesRepo,
			}},
		})
	}

	ensureScalarKey(docContent, "types_repo_clone", tagBool, falseValue)
	ensureScalarKey(docContent, "types_repo_pull", tagBool, falseValue)
}

// mergeProviderDefaults ensures each registered provider has a section in the
// node tree and in Cfg.Providers, then merges its default options.
func mergeProviderDefaults(providersNode *yaml.Node) {
	for name, p := range provider.GetProviders() {
		providerNode, _ := findNodeByKey(providersNode, name)
		if providerNode == nil {
			providerNode = &yaml.Node{Kind: yaml.MappingNode, Tag: tagMap}
			keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: tagStr, Value: name}
			providersNode.Content = append(providersNode.Content, keyNode, providerNode)
		}

		if _, ok := Cfg.Providers[name]; !ok {
			Cfg.Providers[name] = map[string]any{}
		}

		mergeProviderOptionDefaults(providerNode, p, name)
	}
}

// mergeProviderOptionDefaults merges top-level, import, and export option
// defaults for a single provider into providerNode.
func mergeProviderOptionDefaults(providerNode *yaml.Node, p provider.Provider, name string) {
	mergeTopLevelOptionDefaults(providerNode, p, name)
	mergeSectionOptionDefaults(providerNode, p, name)
}

// mergeTopLevelOptionDefaults merges a provider's top-level option defaults,
// excluding the import/export sections that are handled separately.
func mergeTopLevelOptionDefaults(providerNode *yaml.Node, p provider.Provider, name string) {
	hasOptions, ok := p.(provider.HasOptions)
	if !ok {
		return
	}
	topLevelDefaults := make(map[string]any)
	for k, v := range hasOptions.GetDefaultOptions() {
		if k != keyImport && k != keyExport {
			topLevelDefaults[k] = v
		}
	}
	if err := mergeIntoNode(providerNode, topLevelDefaults); err != nil {
		log.Printf("Warning: failed to merge defaults for provider %s: %v", name, err)
	}
}

// mergeSectionOptionDefaults merges a provider's import and export option
// defaults into their respective sub-sections.
func mergeSectionOptionDefaults(providerNode *yaml.Node, p provider.Provider, name string) {
	if hasImport, ok := p.(provider.HasImportOptions); ok {
		mergeSubsectionDefaults(providerNode, keyImport, hasImport.GetImportDefaults(), name)
	}
	if hasExport, ok := p.(provider.HasExportOptions); ok {
		mergeSubsectionDefaults(providerNode, keyExport, hasExport.GetExportDefaults(), name)
	}
}

// mergeSubsectionDefaults ensures a named sub-section exists under providerNode
// and merges the provided defaults into it.
func mergeSubsectionDefaults(providerNode *yaml.Node, section string, defaults map[string]any, name string) {
	node := ensureNodePath(providerNode, section)
	if node == nil {
		return
	}
	if err := mergeIntoNode(node, defaults); err != nil {
		log.Printf("Warning: failed to merge %s defaults for provider %s: %v", section, name, err)
	}
}

// applyAllComments applies struct-tag comments to the top-level config keys and
// to each registered provider's option, import, and export sections.
func applyAllComments(docContent, providersNode *yaml.Node) {
	applyComments(docContent, extractComments(Config{}))

	for name, p := range provider.GetProviders() {
		providerNode, _ := findNodeByKey(providersNode, name)
		if providerNode == nil {
			continue
		}
		if hasOptions, ok := p.(provider.HasOptions); ok {
			applyComments(providerNode, extractComments(hasOptions.GetOptionsStruct()))
		}
		if hasImport, ok := p.(provider.HasImportOptions); ok {
			if importNode, _ := findNodeByKey(providerNode, keyImport); importNode != nil {
				applyComments(importNode, extractComments(hasImport.GetImportOptionsStruct()))
			}
		}
		if hasExport, ok := p.(provider.HasExportOptions); ok {
			if exportNode, _ := findNodeByKey(providerNode, keyExport); exportNode != nil {
				applyComments(exportNode, extractComments(hasExport.GetExportOptionsStruct()))
			}
		}
	}
}

// writeConfigFile writes the node tree to path with 0600 permissions. The config
// may contain provider API tokens, so it must not be readable by other users.
//
// The encoded output is compared against the current file first; when it is
// identical the write is skipped entirely, avoiding needless disk I/O in large
// scripts of sequential commands where setupDomain runs Save on each process
// start.
//
// When a write is needed it is atomic: the encoded YAML is written to a
// temporary file in the same directory and then renamed over path. os.Rename is
// atomic on POSIX, so a concurrent reader (e.g. another cani process loading the
// same config in a shell pipeline) always observes either the complete old file
// or the complete new file — never a half-written, unparseable one.
func writeConfigFile(path string, root *yaml.Node) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}

	// Skip the write when the on-disk content already matches.
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, buf.Bytes()) {
		return nil
	}

	dir := filepath.Dir(path)
	// os.CreateTemp creates the file with 0600 permissions, matching the
	// token-bearing config's confidentiality requirement.
	tmp, err := os.CreateTemp(dir, ".cani-config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, writeErr := tmp.Write(buf.Bytes()); writeErr != nil {
		tmp.Close()
		os.Remove(tmpName)
		return writeErr
	}
	if closeErr := tmp.Close(); closeErr != nil {
		os.Remove(tmpName)
		return closeErr
	}

	// Atomically replace the destination. On failure, clean up the temp file.
	if renameErr := os.Rename(tmpName, path); renameErr != nil {
		os.Remove(tmpName)
		return renameErr
	}
	return nil
}
