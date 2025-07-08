package core

import (
	"regexp"
	"strings"
)

// Slugify converts a string to a slug with lowercase letters and dashes.
func Slugify(s string) string {
	// Convert to lowercase.
	s = strings.ToLower(s)
	// Replace any sequence of non-alphanumeric characters with a single dash.
	re := regexp.MustCompile("[^a-z0-9]+")
	s = re.ReplaceAllString(s, "-")
	// Trim dashes from the beginning and end.
	s = strings.Trim(s, "-")
	return s
}
