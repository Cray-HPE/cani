/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
package netbox

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// stripProtocolsAndSpecialChars removes protocols and special characters from a string
func stripProtocolsAndSpecialChars(url string) string {
	// Remove protocols at the start of the string
	re := regexp.MustCompile(`^(?:\w+://)?(.*)$`)
	url = re.ReplaceAllString(url, "$1")

	// Remove trailing special characters
	re = regexp.MustCompile(`[^a-zA-Z0-9]+$`)
	url = re.ReplaceAllString(url, "")

	return url
}

// stringToSlug converts a string to a slug
func stringToSlug(s string) string {
	encoder := json.Encoder{}
	encoder.SetEscapeHTML(true)
	// make string lowercase
	one := strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	// remove special characters
	re := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	two := re.ReplaceAllString(one, "")
	// replace unicode characters make url safe
	three := strings.ReplaceAll(two, " ", "-")

	// remove unicode characters
	urlSafe, _ := strconv.Unquote(`"` + three + `"`)

	return urlSafe
}
