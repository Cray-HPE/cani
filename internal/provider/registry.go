package provider

var providers = map[string]Provider{}

// Register makes a plugin available under name.
// This should be called in the plugin's init() function.
func Register(name string, p Provider) {
	providers[name] = p
}

// GetProvider returns a registered plugin or nil.
func GetProvider(name string) Provider {
	return providers[name]
}

// All returns every registered plugin
func GetProviders() map[string]Provider {
	return providers
}
