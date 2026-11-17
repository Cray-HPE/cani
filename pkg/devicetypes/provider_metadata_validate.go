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

import (
	"fmt"
	"maps"
	"slices"
)

// bucketIssue counts the objects that hold a bucket no provider owns, and
// keeps the first of them by label as the example to report.
type bucketIssue struct {
	count   int
	example string
}

// validateProviderMetadata reports provider metadata that no provider owns.
// Provider metadata is one map per provider, under the provider's slug, and
// belongs to that provider. A value stored beside the buckets has no owner:
// no provider, command or export reads it, so it is reported for the
// operator to move or remove rather than kept in silence. A bucket named for
// a provider that is not built in is reported too, once per name with a
// count and one example, since nothing reads it either; the providers that
// are built in are known from their registration, so when none is, as when
// the model is used on its own, buckets are not checked.
func (inv *Inventory) validateProviderMetadata() *RelationshipResult {
	result := &RelationshipResult{}
	unknown := make(map[string]*bucketIssue)
	inv.eachObjectMeta(func(owner metaOwner) {
		for _, key := range sortedKeys(owner.provider) {
			if _, isBucket := owner.provider[key].(map[string]any); isBucket {
				countUnknownBucket(unknown, key, owner.label)
				continue
			}
			result.Unresolved = append(result.Unresolved, fmt.Sprintf(
				"%s: provider metadata %q is outside every provider's bucket, so nothing reads it; move it into a bucket or into customFields",
				owner.label, key))
		}
	})
	for _, name := range slices.Sorted(maps.Keys(unknown)) {
		issue := unknown[name]
		result.Unresolved = append(result.Unresolved, fmt.Sprintf(
			"provider metadata bucket %q is held by %d object(s), e.g. %s, but no provider %q is built in, so nothing reads it",
			name, issue.count, issue.example, name))
	}
	return result
}

// countUnknownBucket records one more object holding the bucket name, when
// providers are registered and none of them owns it.
func countUnknownBucket(unknown map[string]*bucketIssue, name, label string) {
	if known, anyRegistered := knownBucket(name); known || !anyRegistered {
		return
	}
	issue := unknown[name]
	if issue == nil {
		issue = &bucketIssue{}
		unknown[name] = issue
	}
	issue.count++
	if issue.example == "" || label < issue.example {
		issue.example = label
	}
}
