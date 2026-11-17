/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package devicetypes

import "sync"

var (
	bucketsMu    sync.RWMutex
	knownBuckets = map[string]bool{}
)

// RegisterProviderBucket records that a provider of the given slug is built
// into this binary, so that validation can tell its bucket in provider
// metadata from one no provider owns. provider.Register calls it with the
// registry key, which is the provider's slug and the name of its bucket.
func RegisterProviderBucket(slug string) {
	bucketsMu.Lock()
	defer bucketsMu.Unlock()
	knownBuckets[slug] = true
}

// knownBucket reports whether a provider of the slug is built in, and
// whether any provider is. With none registered, as when the model is used
// on its own, no bucket can be told from another and the check does not run.
func knownBucket(slug string) (known, anyRegistered bool) {
	bucketsMu.RLock()
	defer bucketsMu.RUnlock()
	return knownBuckets[slug], len(knownBuckets) > 0
}
