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
package provider

import (
	"fmt"
	"sync"
)

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

// Register makes a plugin available under name.
// This should be called in the plugin's init() function.
//
// Register panics if p is nil or if name was already registered: both indicate
// a wiring bug that should fail loudly at startup rather than silently
// overwrite an existing provider.
func Register(name string, p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if p == nil {
		panic("provider: Register provider is nil")
	}
	if _, dup := providers[name]; dup {
		panic(fmt.Sprintf("provider: Register called twice for provider %q", name))
	}
	providers[name] = p
}

// GetProvider returns a registered plugin or nil.
func GetProvider(name string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[name]
}

// GetProviders returns a snapshot of every registered plugin. The returned map
// is a copy, so callers may range over it safely while providers register and
// cannot mutate the registry through it.
func GetProviders() map[string]Provider {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]Provider, len(providers))
	for name, p := range providers {
		out[name] = p
	}
	return out
}
