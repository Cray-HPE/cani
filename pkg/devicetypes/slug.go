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
	"log"
	"regexp"
	"strings"
)

// nonWordRun matches runs of characters outside Python's Unicode \w class.
var nonWordRun = regexp.MustCompile(`[^\p{L}\p{N}_]+`)

// deriveSlug builds a slug from manufacturer and model using the same rule as
// the Nautobot devicetype-library importer: lowercase "<manufacturer> <model>"
// with each run of non-word characters replaced by a single dash. Leading and
// trailing dashes are kept so derived slugs match Nautobot exactly.
func deriveSlug(manufacturer, model string) string {
	if manufacturer == "" || model == "" {
		return ""
	}
	return nonWordRun.ReplaceAllString(strings.ToLower(manufacturer+" "+model), "-")
}

// resolveSlug returns slug when set, otherwise one derived from manufacturer
// and model. It warns when a file that names a manufacturer or model still
// cannot be given a slug, since that type would otherwise vanish silently.
func resolveSlug(slug, manufacturer, model, path string) string {
	if slug != "" {
		return slug
	}
	if derived := deriveSlug(manufacturer, model); derived != "" {
		return derived
	}
	if manufacturer != "" || model != "" {
		log.Printf("Warning: skipping %s: no slug, and manufacturer and model are both required to derive one", path)
	}
	return ""
}
