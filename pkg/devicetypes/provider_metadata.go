package devicetypes

import "sort"

// sortedKeys returns the keys of m in deterministic, ascending order so that
// lookups and flattening over ProviderMetadata do not depend on Go's
// randomized map iteration order.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetProviderMeta searches all provider sub-maps in ProviderMetadata for the
// given key and returns the first match. Top-level keys are checked first,
// then each provider sub-map is scanned.
func (m *ObjectMeta) GetProviderMeta(key string) (any, bool) {
	if m == nil || m.ProviderMetadata == nil {
		return nil, false
	}
	// Check top-level keys first.
	if v, ok := m.ProviderMetadata[key]; ok {
		// Skip provider sub-maps when searching for a plain key.
		if _, isMap := v.(map[string]any); !isMap {
			return v, true
		}
	}
	// Search inside each provider sub-map, in deterministic provider order.
	for _, name := range sortedKeys(m.ProviderMetadata) {
		sub, ok := m.ProviderMetadata[name].(map[string]any)
		if !ok {
			continue
		}
		if val, found := sub[key]; found {
			return val, true
		}
	}
	return nil, false
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

// FlattenProviderMetadata returns a flat map of all provider metadata,
// combining all provider sub-maps and top-level keys. Provider sub-map
// keys are NOT prefixed. If multiple providers define the same key, the
// value from the highest-sorted provider name wins (and within a provider,
// the highest-sorted sub-key), so the result is deterministic and does not
// depend on Go's randomized map iteration order.
func (m *ObjectMeta) FlattenProviderMetadata() map[string]any {
	if m == nil || m.ProviderMetadata == nil {
		return nil
	}
	flat := make(map[string]any)
	for _, k := range sortedKeys(m.ProviderMetadata) {
		sub, ok := m.ProviderMetadata[k].(map[string]any)
		if !ok {
			flat[k] = m.ProviderMetadata[k]
			continue
		}
		for _, sk := range sortedKeys(sub) {
			flat[sk] = sub[sk]
		}
	}
	return flat
}
