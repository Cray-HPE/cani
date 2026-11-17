package devicetypes

import "sort"

// sortedKeys returns the keys of m in ascending order, so that a scan of
// ProviderMetadata does not depend on Go's randomized map iteration order.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetProviderSubMap returns the provider-specific sub-map for the given
// provider name (e.g. "redfish", "csm", "hpcm").
func (m *ObjectMeta) GetProviderSubMap(provider string) (map[string]any, bool) {
	if m == nil || m.ProviderMetadata == nil {
		return nil, false
	}
	v, ok := m.ProviderMetadata[provider]
	if !ok {
		return nil, false
	}
	sub, ok := v.(map[string]any)
	return sub, ok
}

// SetProviderMeta sets a key inside the named provider sub-map, creating
// the sub-map if it does not exist.
func (m *ObjectMeta) SetProviderMeta(provider, key string, value any) {
	if m == nil {
		return
	}
	if m.ProviderMetadata == nil {
		m.ProviderMetadata = make(map[string]any)
	}
	sub, ok := m.ProviderMetadata[provider].(map[string]any)
	if !ok {
		sub = make(map[string]any)
		m.ProviderMetadata[provider] = sub
	}
	sub[key] = value
}

// SetImportSource sets the "import_source" key inside the named provider
// sub-map, creating the sub-map if it does not exist.
func (m *ObjectMeta) SetImportSource(provider, source string) {
	m.SetProviderMeta(provider, "import_source", source)
}

// GetImportSource returns the "import_source" value from the named provider
// sub-map. Returns an empty string if not set.
func (m *ObjectMeta) GetImportSource(provider string) string {
	sub, ok := m.GetProviderSubMap(provider)
	if !ok {
		return ""
	}
	v, ok := sub["import_source"].(string)
	if !ok {
		return ""
	}
	return v
}
