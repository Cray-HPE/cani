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
package devicetypes

import "github.com/google/uuid"

// ObjectMeta contains metadata fields shared across all inventory types.
// Embed this struct in any inventory type to gain consistent support for
// status, role, tags, tenant, custom fields, external IDs, and provider
// metadata. Go promotes the embedded fields so callers can access them
// directly (e.g. device.Status, rack.Tags).
type ObjectMeta struct {
	Status           string               `json:"status"                      yaml:"status,omitempty"`
	Role             string               `json:"role,omitempty"              yaml:"role,omitempty"`
	Tags             []string             `json:"tags,omitempty"              yaml:"tags,omitempty"`
	Tenant           string               `json:"tenant,omitempty"            yaml:"tenant,omitempty"`
	CustomFields     map[string]any       `json:"customFields,omitempty"      yaml:"custom_fields,omitempty"`
	ExternalIDs      map[string]uuid.UUID `json:"externalIDs,omitempty"       yaml:"external_ids,omitempty"`
	ProviderMetadata map[string]any       `json:"providerMetadata,omitempty"  yaml:"provider_metadata,omitempty"`
}

// GetRole returns the effective role: the explicit Role field if set,
// otherwise the first "role" value found in ProviderMetadata.
func (m ObjectMeta) GetRole() string {
	if m.Role != "" {
		return m.Role
	}
	for _, v := range m.ProviderMetadata {
		bucket, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if r, ok := bucket["role"].(string); ok && r != "" {
			return r
		}
	}
	return ""
}
