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
	"os"

	"gopkg.in/yaml.v3"
)

// csmFieldMap maps legacy CSM option keys to their new config key names.
// Keys present here are migrated into the top-level provider section.
var csmFieldMap = map[string]string{
	"usesimulation":      "use_simulation",
	"insecureskipverify": "insecure",
	"providerhost":       "provider_host",
	"cacertpath":         "ca_cert",
	"secretname":         "secret_name",
	"k8spodscidr":        "k8s_pods_cidr",
	"k8sservicescidr":    "k8s_services_cidr",
}

// isLegacyFormat returns true if the decoded YAML root contains a "session"
// top-level key, which indicates a pre-0.6.x config format.
func isLegacyFormat(root *yaml.Node) bool {
	if root == nil {
		return false
	}
	doc := root
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		doc = doc.Content[0]
	}
	if doc.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value == "session" {
			return true
		}
	}
	return false
}

// backupConfig renames path to path.canisave (like .rpmsave).
func backupConfig(path string) error {
	dst := path + ".canisave"
	return os.Rename(path, dst)
}

// migrateConfig detects the legacy format family and dispatches to the
// appropriate migration function. It returns a new 0.6.x root node.
func migrateConfig(root *yaml.Node) (*yaml.Node, error) {
	doc := root
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		doc = doc.Content[0]
	}

	sessionNode, _ := findNodeByKey(doc, "session")
	if sessionNode == nil {
		return nil, fmt.Errorf("legacy config missing 'session' key")
	}

	// Family B: session.domains (0.4.x, 0.5.x)
	if domainsNode, _ := findNodeByKey(sessionNode, "domains"); domainsNode != nil {
		return migrateFamilyB(domainsNode)
	}

	// Family A: session.domain_options (0.1.x, 0.2.x, 0.3.x)
	if domainOpts, _ := findNodeByKey(sessionNode, "domain_options"); domainOpts != nil {
		return migrateFamilyA(domainOpts)
	}

	return nil, fmt.Errorf("legacy config has unrecognised session structure")
}

// migrateFamilyA handles 0.1.x–0.3.x configs structured as:
//
//	session.domain_options.provider
//	session.domain_options.datastore_path
//	session.domain_options.csm_options.*
//	session.domain_options.custom_hardware_types_dir
func migrateFamilyA(domainOpts *yaml.Node) (*yaml.Node, error) {
	provider := nodeScalar(domainOpts, "provider")
	datastore := nodeScalar(domainOpts, "datastore_path")
	typesDir := nodeScalar(domainOpts, "custom_hardware_types_dir")

	// Extract provider-specific options (e.g. csm_options)
	optKey := provider + "_options"
	optsNode, _ := findNodeByKey(domainOpts, optKey)

	mapped, legacy := splitCSMOptions(optsNode)

	return buildNewRoot(datastore, typesDir, map[string]providerMigration{
		provider: {mapped: mapped, legacy: legacy},
	})
}

// migrateFamilyB handles 0.4.x–0.5.x configs structured as:
//
//	session.domains.<name>.provider
//	session.domains.<name>.datastore_path
//	session.domains.<name>.options.*
//	session.domains.<name>.custom_hardware_types_dir
func migrateFamilyB(domainsNode *yaml.Node) (*yaml.Node, error) {
	if domainsNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("domains node is not a mapping")
	}

	providers := map[string]providerMigration{}
	var datastore, typesDir string

	for i := 0; i+1 < len(domainsNode.Content); i += 2 {
		name := domainsNode.Content[i].Value
		domainNode := domainsNode.Content[i+1]
		if domainNode.Kind != yaml.MappingNode {
			continue
		}

		// Use the first non-empty datastore_path we find
		if ds := nodeScalar(domainNode, "datastore_path"); ds != "" && datastore == "" {
			datastore = ds
		}
		if td := nodeScalar(domainNode, "custom_hardware_types_dir"); td != "" && typesDir == "" {
			typesDir = td
		}

		optsNode, _ := findNodeByKey(domainNode, "options")
		mapped, legacy := splitCSMOptions(optsNode)
		providers[name] = providerMigration{mapped: mapped, legacy: legacy}
	}

	return buildNewRoot(datastore, typesDir, providers)
}

// providerMigration holds the split results for one provider.
type providerMigration struct {
	mapped map[string]string // renamed key → value
	legacy map[string]any    // unmapped key → value
}

// splitCSMOptions iterates over a legacy options MappingNode and separates
// fields into mapped (renamed) and legacy (unmapped, non-empty) buckets.
func splitCSMOptions(optsNode *yaml.Node) (mapped map[string]string, legacy map[string]any) {
	mapped = map[string]string{}
	legacy = map[string]any{}
	if optsNode == nil || optsNode.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i+1 < len(optsNode.Content); i += 2 {
		key := optsNode.Content[i].Value
		valNode := optsNode.Content[i+1]

		if newKey, ok := csmFieldMap[key]; ok {
			mapped[newKey] = valNode.Value
			continue
		}

		// Collect non-empty unmapped fields into legacy
		val := decodeNodeValue(valNode)
		if !isZeroValue(val) {
			legacy[key] = val
		}
	}
	return
}

// buildNewRoot constructs a fresh 0.6.x-format document node.
func buildNewRoot(datastore, typesDir string, providers map[string]providerMigration) (*yaml.Node, error) {
	docMap := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	// providers:
	providersNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for name, pm := range providers {
		provNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

		// Add mapped fields in deterministic order
		for _, pair := range sortedPairs(pm.mapped) {
			addScalar(provNode, pair.key, pair.val)
		}

		// Add _legacy sub-map if there are unmapped fields
		if len(pm.legacy) > 0 {
			legacyNode := valueToNode(pm.legacy)
			addKeyVal(provNode, "_legacy", legacyNode)
		}

		// Add import/export stubs
		addKeyVal(provNode, "import", &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"})
		addKeyVal(provNode, "export", &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"})

		addKeyVal(providersNode, name, provNode)
	}
	addKeyVal(docMap, "providers", providersNode)

	// datastore:
	if datastore == "" {
		datastore = "/tmp/.cani/canidb.json"
	}
	addScalar(docMap, "datastore", datastore)

	// debug:
	addKeyVal(docMap, "debug", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"})

	// types_dirs:
	typesDirsNode := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	if typesDir != "" {
		typesDirsNode.Content = append(typesDirsNode.Content, &yaml.Node{
			Kind: yaml.ScalarNode, Tag: "!!str", Value: typesDir,
		})
	}
	addKeyVal(docMap, "types_dirs", typesDirsNode)

	// types_repos:
	typesReposNode := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: DefaultTypesRepo},
	}}
	addKeyVal(docMap, "types_repos", typesReposNode)

	// types_repo_pull:
	addKeyVal(docMap, "types_repo_pull", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"})

	return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{docMap}}, nil
}

// --- small helpers ---------------------------------------------------------

// nodeScalar returns the scalar value for key inside a MappingNode, or "".
func nodeScalar(mapNode *yaml.Node, key string) string {
	n, _ := findNodeByKey(mapNode, key)
	if n == nil {
		return ""
	}
	return n.Value
}

// addScalar appends a string key/value pair to a MappingNode.
func addScalar(m *yaml.Node, key, val string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val},
	)
}

// addKeyVal appends a generic key → value-node pair to a MappingNode.
func addKeyVal(m *yaml.Node, key string, val *yaml.Node) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		val,
	)
}

// sortedPair is a key-value pair for deterministic iteration.
type sortedPair struct {
	key, val string
}

// sortedPairs returns mapped entries sorted by key.
func sortedPairs(m map[string]string) []sortedPair {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// simple insertion sort (small maps, no imports needed)
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	out := make([]sortedPair, len(keys))
	for i, k := range keys {
		out[i] = sortedPair{k, m[k]}
	}
	return out
}

// decodeNodeValue converts a yaml.Node into a Go value for the legacy bucket.
func decodeNodeValue(n *yaml.Node) any {
	switch n.Kind {
	case yaml.ScalarNode:
		return n.Value
	case yaml.SequenceNode:
		items := make([]any, 0, len(n.Content))
		for _, c := range n.Content {
			items = append(items, decodeNodeValue(c))
		}
		return items
	case yaml.MappingNode:
		m := map[string]any{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			m[n.Content[i].Value] = decodeNodeValue(n.Content[i+1])
		}
		return m
	}
	return nil
}

// isZeroValue returns true for empty strings, nil, and empty slices/maps.
func isZeroValue(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	}
	return false
}
