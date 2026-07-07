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
package export

import (
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// resolveTagRefs converts a list of tag names into Nautobot tag references,
// creating any tags that do not yet exist. It returns nil when the input is
// empty or no tags resolve, so callers can assign the result directly to an
// optional request field (leaving it unset rather than an empty list).
func (c *LookupCache) resolveTagRefs(names []string) *[]nautobotapi.BulkWritableCableRequestStatus {
	if len(names) == 0 {
		return nil
	}
	refs := make([]nautobotapi.BulkWritableCableRequestStatus, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		tag, err := c.GetOrCreateTag(name)
		if err == nil && tag != nil {
			refs = append(refs, makeStatusRef(tag.ID))
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return &refs
}
