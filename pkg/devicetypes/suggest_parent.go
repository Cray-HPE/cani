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

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// ParentSuggestion is a scored candidate parent for an orphan.
type ParentSuggestion struct {
	ID     uuid.UUID
	Name   string
	Kind   string // "rack", "device", "location"
	Score  int    // 0–100
	Reason string // human-readable explanation
	Detail string // extra context: location, model, device count, etc.
}

// SuggestParents returns ranked candidate parents for an orphan.
// The hierarchy is strictly: location → rack → device → module → FRU.
// Orphan devices suggest racks only; orphan racks suggest locations only.
// identityKeys are the keys the providers declare as identifying a device
// (see TransformResult.IdentityKeys); a rack whose bucket holds one with the
// same prefix as the orphan's scores higher.
func SuggestParents(inv *Inventory, orphan OrphanItem, maxResults int, identityKeys []string) []ParentSuggestion {
	if inv == nil {
		return nil
	}

	var candidates []ParentSuggestion
	switch orphan.Kind {
	case "device":
		candidates = suggestDeviceParents(inv, orphan, identityKeys)
	case "rack":
		candidates = suggestRackParents(inv, orphan)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	if maxResults > 0 && len(candidates) > maxResults {
		candidates = candidates[:maxResults]
	}
	return candidates
}

// suggestDeviceParents scores racks as potential parents for a device.
// Following the strict hierarchy: location → rack → device.
func suggestDeviceParents(inv *Inventory, orphan OrphanItem, identityKeys []string) []ParentSuggestion {
	var out []ParentSuggestion

	for id, rack := range inv.Racks {
		if rack == nil {
			continue
		}
		score, reasons := scoreDeviceToRack(orphan, rack, identityKeys)
		if score < 1 {
			score = 1 // minimum so every rack appears
		}
		out = append(out, ParentSuggestion{
			ID:     id,
			Name:   rack.Name,
			Kind:   "rack",
			Score:  clampScore(score),
			Reason: joinReasons(reasons),
			Detail: rackDetail(inv, rack),
		})
	}

	return out
}

// suggestRackParents scores locations as potential parents for a rack.
func suggestRackParents(inv *Inventory, orphan OrphanItem) []ParentSuggestion {
	var out []ParentSuggestion
	for id, loc := range inv.Locations {
		if loc == nil {
			continue
		}
		score, reasons := scoreRackToLocation(orphan, loc)
		if score < 1 {
			score = 1
		}
		out = append(out, ParentSuggestion{
			ID:     id,
			Name:   loc.Name,
			Kind:   "location",
			Score:  clampScore(score),
			Reason: joinReasons(reasons),
			Detail: locationDetail(loc),
		})
	}
	return out
}

// --- detail helpers ---

// rackDetail builds a short summary string for a rack candidate.
func rackDetail(inv *Inventory, rack *CaniRackType) string {
	var parts []string
	if rack.Model != "" {
		parts = append(parts, rack.Model)
	} else if rack.Type != "" {
		parts = append(parts, string(rack.Type))
	}
	if rack.UHeight > 0 {
		used := len(rack.Devices)
		parts = append(parts, fmt.Sprintf("%dU, %d/%d occupied", rack.UHeight, used, rack.UHeight))
	}
	if rack.Location != uuid.Nil && inv != nil {
		if loc, ok := inv.Locations[rack.Location]; ok && loc != nil {
			parts = append(parts, "location: "+loc.Name)
		}
	}
	return strings.Join(parts, ", ")
}

// locationDetail builds a short summary string for a location candidate.
func locationDetail(loc *CaniLocationType) string {
	var parts []string
	if loc.LocationType != "" {
		parts = append(parts, loc.LocationType)
	}
	if len(loc.Racks) > 0 {
		parts = append(parts, fmt.Sprintf("%d racks", len(loc.Racks)))
	}
	if loc.Facility != "" {
		parts = append(parts, loc.Facility)
	}
	if loc.Description != "" {
		parts = append(parts, loc.Description)
	}
	return strings.Join(parts, ", ")
}

// --- scoring helpers ---

func scoreDeviceToRack(orphan OrphanItem, rack *CaniRackType, identityKeys []string) (int, []string) {
	score := 0
	var reasons []string

	// Name similarity
	if ns := nameSimilarity(orphan.Name, rack.Name); ns > 0 {
		score += ns
		reasons = append(reasons, "name similarity")
	}

	// Provider metadata: identity key prefix match
	if ps := providerPrefixScore(orphan.ProviderMetadata, rack.ProviderMetadata, identityKeys); ps > 0 {
		score += ps
		reasons = append(reasons, "provider key prefix")
	}

	// Proximity: rack has available space (minor bonus)
	if rack.UHeight > 0 {
		used := len(rack.Devices)
		if used < rack.UHeight {
			score += 10
			reasons = append(reasons, "rack has capacity")
		}
	}

	return score, reasons
}

func scoreRackToLocation(orphan OrphanItem, loc *CaniLocationType) (int, []string) {
	score := 0
	var reasons []string

	if ns := nameSimilarity(orphan.Name, loc.Name); ns > 0 {
		score += ns
		reasons = append(reasons, "name similarity")
	}

	// If the location already has racks, it's a proven parent.
	if len(loc.Racks) > 0 {
		score += 15
		reasons = append(reasons, "location has existing racks")
	}

	return score, reasons
}

// nameSimilarity returns up to 20 points based on the longest common prefix
// between two names (case-insensitive).
func nameSimilarity(a, b string) int {
	la := strings.ToLower(a)
	lb := strings.ToLower(b)
	prefixLen := longestCommonPrefix(la, lb)
	if prefixLen < 2 {
		return 0
	}
	maxLen := len(la)
	if len(lb) > maxLen {
		maxLen = len(lb)
	}
	if maxLen == 0 {
		return 0
	}
	ratio := float64(prefixLen) / float64(maxLen)
	return int(ratio * 20)
}

func longestCommonPrefix(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// providerPrefixScore reports whether an identity key the orphan and the
// candidate both record in one provider's bucket shares a prefix of at least
// three characters, as the xname of a node does with its cabinet's. Buckets
// are compared by name, so a key is only read where the same provider wrote
// both sides. Returns 30 points for a match and 0 otherwise.
func providerPrefixScore(orphanMeta, candidateMeta map[string]any, identityKeys []string) int {
	for _, provider := range sortedKeys(orphanMeta) {
		orphanSub, _ := orphanMeta[provider].(map[string]any)
		candidateSub, _ := candidateMeta[provider].(map[string]any)
		if orphanSub == nil || candidateSub == nil {
			continue
		}
		if sharesIdentityPrefix(orphanSub, candidateSub, identityKeys) {
			return 30
		}
	}
	return 0
}

// minIdentityPrefix is how many leading characters two identity values must
// share to count as related.
const minIdentityPrefix = 3

// sharesIdentityPrefix reports whether, for some identity key both buckets
// hold, the values share at least minIdentityPrefix characters, case aside.
func sharesIdentityPrefix(orphanSub, candidateSub map[string]any, identityKeys []string) bool {
	for _, key := range identityKeys {
		ov, cv := metaString(orphanSub, key), metaString(candidateSub, key)
		if ov == "" || cv == "" {
			continue
		}
		if longestCommonPrefix(strings.ToLower(ov), strings.ToLower(cv)) >= minIdentityPrefix {
			return true
		}
	}
	return false
}

func metaString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func clampScore(s int) int {
	if s > 100 {
		return 100
	}
	if s < 0 {
		return 0
	}
	return s
}

func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	return strings.Join(reasons, ", ")
}

// SearchParentCandidates searches the next level up in the hierarchy by name
// substring. Devices → racks, racks → locations.
func SearchParentCandidates(inv *Inventory, query string, orphanKind string, maxResults int) []ParentSuggestion {
	lower := strings.ToLower(query)
	var out []ParentSuggestion

	switch orphanKind {
	case "device":
		for id, rack := range inv.Racks {
			if rack != nil && strings.Contains(strings.ToLower(rack.Name), lower) {
				out = append(out, ParentSuggestion{ID: id, Name: rack.Name, Kind: "rack", Detail: rackDetail(inv, rack)})
			}
		}
	case "rack":
		for id, loc := range inv.Locations {
			if loc != nil && strings.Contains(strings.ToLower(loc.Name), lower) {
				out = append(out, ParentSuggestion{ID: id, Name: loc.Name, Kind: "location", Detail: locationDetail(loc)})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	if maxResults > 0 && len(out) > maxResults {
		out = out[:maxResults]
	}
	return out
}
