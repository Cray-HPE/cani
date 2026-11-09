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
package export

import (
	"encoding/json"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// customFieldSet is the custom-field projection of one inventory object: the
// values a write sends, and the keys that both the explicit CustomFields and
// the flattened ProviderMetadata supplied.
type customFieldSet struct {
	values     map[string]any
	collisions []string
}

// customFieldsOf is the one projection from an object's metadata to the
// custom_fields Nautobot receives, shared by every create, merge and compare
// path. Explicit CustomFields are merged with the flattened ProviderMetadata;
// on a shared key the provider value wins, which is the documented
// precedence, and the key is reported rather than resolved differently.
// Explicit false, zero and empty values are kept: Nautobot fills an omitted
// key from the definition's default, so omission and false are different
// writes. Keys named in exclude never reach the payload; the legacy rack
// mapper uses this for geometry it maps to dedicated fields.
func customFieldsOf(meta *devicetypes.ObjectMeta, exclude ...string) customFieldSet {
	if meta == nil {
		return customFieldSet{}
	}
	excluded := stringSet(exclude)
	flat := meta.FlattenProviderMetadata()
	set := customFieldSet{values: mergedCustomFields(meta.CustomFields, flat)}
	for k := range flat {
		if _, dup := meta.CustomFields[k]; dup && !excluded[k] {
			set.collisions = append(set.collisions, k)
		}
	}
	for k := range excluded {
		delete(set.values, k)
	}
	sort.Strings(set.collisions)
	return set
}

// payload is the set in the request shape, nil when there is nothing to send.
func (s customFieldSet) payload() *map[string]*interface{} {
	if len(s.values) == 0 {
		return nil
	}
	return toNautobotCustomFields(s.values)
}

// customFieldsPayload projects one object's custom fields for a write and
// warns once per colliding key, naming the object and the key but never the
// value.
func customFieldsPayload(kind, name string, meta *devicetypes.ObjectMeta, exclude ...string) *map[string]*interface{} {
	set := customFieldsOf(meta, exclude...)
	for _, key := range set.collisions {
		clog.Warn("%s %s: custom field %q is set both explicitly and by provider metadata; the provider value is sent", kind, name, key)
	}
	return set.payload()
}

// recordDroppedCustomFields counts the custom-field keys a write sent that
// the response does not carry. Nautobot answers 2xx for a key with no
// definition for the object's content type: its clean() logs a server-side
// warning and keeps the value in the object's data, but the API never
// exposes it, so the response body is the only evidence that a value did
// not arrive. A body without a custom_fields object says nothing and is not
// counted. Each missing key is named with its object, never with its value.
func recordDroppedCustomFields(kind, name string, sent *map[string]*interface{}, body []byte, result *LoadResult) {
	if sent == nil || len(*sent) == 0 || len(body) == 0 {
		return
	}
	var resp struct {
		CustomFields *map[string]json.RawMessage `json:"custom_fields"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || resp.CustomFields == nil {
		return
	}
	for _, key := range sortedCustomFieldKeys(*sent) {
		if _, arrived := (*resp.CustomFields)[key]; arrived {
			continue
		}
		result.CustomFieldsDropped++
		clog.Warn("%s %s: custom field %q did not arrive; Nautobot has no definition of it for this object type", kind, name, key)
	}
}

// sortedCustomFieldKeys returns the keys of a request custom_fields map in
// sorted order so reports are stable.
func sortedCustomFieldKeys(cf map[string]*interface{}) []string {
	keys := make([]string, 0, len(cf))
	for k := range cf {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
