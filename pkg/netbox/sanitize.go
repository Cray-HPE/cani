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
