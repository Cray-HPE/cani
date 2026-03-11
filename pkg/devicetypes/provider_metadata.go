package devicetypes

// GetProviderMeta searches all provider sub-maps in ProviderMetadata for the
// given key and returns the first match. Top-level keys are checked first,
// then each provider sub-map is scanned.
func (c *CaniDeviceType) GetProviderMeta(key string) (any, bool) {
	if c == nil || c.ProviderMetadata == nil {
		return nil, false
	}
	// Check top-level keys first.
	if v, ok := c.ProviderMetadata[key]; ok {
		// Skip provider sub-maps when searching for a plain key.
		if _, isMap := v.(map[string]any); !isMap {
			return v, true
		}
	}
	// Search inside each provider sub-map.
	for _, v := range c.ProviderMetadata {
		sub, ok := v.(map[string]any)
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
func (c *CaniDeviceType) GetProviderSubMap(provider string) (map[string]any, bool) {
	if c == nil || c.ProviderMetadata == nil {
		return nil, false
	}
	v, ok := c.ProviderMetadata[provider]
	if !ok {
		return nil, false
	}
	sub, ok := v.(map[string]any)
	return sub, ok
}

// SetProviderMeta sets a key inside the named provider sub-map, creating
// the sub-map if it does not exist.
func (c *CaniDeviceType) SetProviderMeta(provider, key string, value any) {
	if c == nil {
		return
	}
	if c.ProviderMetadata == nil {
		c.ProviderMetadata = make(map[string]any)
	}
	sub, ok := c.ProviderMetadata[provider].(map[string]any)
	if !ok {
		sub = make(map[string]any)
		c.ProviderMetadata[provider] = sub
	}
	sub[key] = value
}

// SetImportSource sets the "import_source" key inside the named provider
// sub-map, creating the sub-map if it does not exist.
func (c *CaniDeviceType) SetImportSource(provider, source string) {
	c.SetProviderMeta(provider, "import_source", source)
}

// GetImportSource returns the "import_source" value from the named provider
// sub-map. Returns an empty string if not set.
func (c *CaniDeviceType) GetImportSource(provider string) string {
	sub, ok := c.GetProviderSubMap(provider)
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
// last one wins (map iteration order).
func (c *CaniDeviceType) FlattenProviderMetadata() map[string]any {
	if c == nil || c.ProviderMetadata == nil {
		return nil
	}
	flat := make(map[string]any)
	for k, v := range c.ProviderMetadata {
		sub, ok := v.(map[string]any)
		if !ok {
			flat[k] = v
			continue
		}
		for sk, sv := range sub {
			flat[sk] = sv
		}
	}
	return flat
}
