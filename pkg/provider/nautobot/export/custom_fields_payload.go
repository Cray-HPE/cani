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
	"fmt"
	"maps"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// customFieldSet is the custom-field projection of one inventory object: the
// values a write sends, and the keys whose value comes from a provider's
// metadata in place of a different one the object's own custom fields hold.
type customFieldSet struct {
	values     map[string]any
	overridden []ProviderMetadataField
}

// Object kinds that label a device and a rack kept as a device in log lines.
const (
	kindDevice     = "device"
	kindRackDevice = "rack device"
)

// customFieldProjection is the one projection from an object's metadata to
// the custom_fields Nautobot receives, shared by every create, merge and
// compare path. It holds the provider-metadata keys the operator asked to
// export; with none, an object's custom fields are all that is sent.
type customFieldProjection []ProviderMetadataField

// of projects one object. Its custom fields are sent as they are: false,
// zero and empty values are kept, because Nautobot fills an omitted key from
// the definition's default, and an explicit null is kept, because it is how
// a field is cleared. Provider metadata is not sent, except the keys the
// projection lists, each under its custom-field key. Such a key belongs to
// its provider, so its value replaces a custom field of the same key. That
// custom field is usually the copy an import read back; were it to win, the
// field would stay at the first value exported. A custom field that is null
// holds no value to replace: an import reads every unset field back as null.
func (p customFieldProjection) of(meta *devicetypes.ObjectMeta) customFieldSet {
	if meta == nil {
		return customFieldSet{}
	}
	set := customFieldSet{values: maps.Clone(meta.CustomFields)}
	for _, field := range p {
		bucket, _ := meta.ProviderMetadata[field.Provider].(map[string]any)
		value, held := bucket[field.Key]
		if !held {
			continue
		}
		if own := set.values[field.CustomField]; own != nil && fmt.Sprint(own) != fmt.Sprint(value) {
			set.overridden = append(set.overridden, field)
		}
		if set.values == nil {
			set.values = make(map[string]any)
		}
		set.values[field.CustomField] = value
	}
	return set
}

// payload is the set in the request shape, nil when there is nothing to send.
func (s customFieldSet) payload() *map[string]*interface{} {
	if len(s.values) == 0 {
		return nil
	}
	return toNautobotCustomFields(s.values)
}

// noted projects one object and notes each custom field a listed provider
// key replaced with a different value, naming the object, the key and the
// provider but never a value. A path that compares before it writes notes
// here, so the note is made once whether or not a write follows; a custom
// field that had no effect would otherwise go unmentioned.
func (p customFieldProjection) noted(kind, name string, meta *devicetypes.ObjectMeta) customFieldSet {
	set := p.of(meta)
	for _, field := range set.overridden {
		clog.Detail("%s %s: custom field %q comes from %s provider metadata; the %s's own value differs and is not sent",
			kind, name, field.CustomField, field.Provider, kind)
	}
	return set
}

// payload projects one object's custom fields for a create and notes what a
// listed provider key replaced.
func (p customFieldProjection) payload(kind, name string, meta *devicetypes.ObjectMeta) *map[string]*interface{} {
	return p.noted(kind, name, meta).payload()
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
