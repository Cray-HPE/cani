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
	"fmt"
	"strings"
)

// ProviderMetadataField names one key of a provider's metadata that an
// export sends as a custom field. Provider metadata belongs to the provider
// that wrote it and is not exported; an entry of the export option
// provider_metadata_fields is how an operator asks for one key of it.
type ProviderMetadataField struct {
	// Provider is the slug of the provider whose bucket holds the key.
	Provider string
	// Key is the key in that bucket.
	Key string
	// CustomField is the custom-field key the value is sent as.
	CustomField string
}

// ParseProviderMetadataFields reads the entries of provider_metadata_fields.
// An entry is "provider.key", sent as the custom field "key", or
// "provider.key=custom_field", sent under another name. The provider is the
// text before the first dot. An entry that names no provider or no key is an
// error, and so are two different keys sent as the same custom field; an
// entry listed twice counts once.
func ParseProviderMetadataFields(entries []string) ([]ProviderMetadataField, error) {
	fields := make([]ProviderMetadataField, 0, len(entries))
	sent := make(map[string]ProviderMetadataField, len(entries))
	for _, entry := range entries {
		field, err := parseProviderMetadataField(strings.TrimSpace(entry))
		if err != nil {
			return nil, err
		}
		other, taken := sent[field.CustomField]
		if taken && other == field {
			continue
		}
		if taken {
			return nil, fmt.Errorf("provider metadata fields %s.%s and %s.%s are both sent as custom field %q",
				other.Provider, other.Key, field.Provider, field.Key, field.CustomField)
		}
		sent[field.CustomField] = field
		fields = append(fields, field)
	}
	return fields, nil
}

// parseProviderMetadataField reads one entry. Spaces around the provider,
// the key and the custom field are not part of them.
func parseProviderMetadataField(entry string) (ProviderMetadataField, error) {
	source, customField, renamed := strings.Cut(entry, "=")
	provider, key, _ := strings.Cut(source, ".")
	provider, key = strings.TrimSpace(provider), strings.TrimSpace(key)
	customField = strings.TrimSpace(customField)
	if !renamed {
		customField = key
	}
	if provider == "" || key == "" || customField == "" {
		return ProviderMetadataField{}, fmt.Errorf(
			"provider metadata field %q is not provider.key or provider.key=custom_field", entry)
	}
	return ProviderMetadataField{Provider: provider, Key: key, CustomField: customField}, nil
}
