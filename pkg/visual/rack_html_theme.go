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
package visual

import (
	"cmp"
	"embed"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"
)

// Built-in themes of the rack HTML output.
const (
	RackThemeHPE     = "hpe"
	RackThemeNeutral = "neutral"
)

//go:embed themes/hpe.css themes/neutral.css
var rackHTMLThemes embed.FS

// ValidRackThemes returns the built-in theme names; any other value of the
// theme option is read as the path of a CSS file.
func ValidRackThemes() []string {
	return []string{RackThemeHPE, RackThemeNeutral}
}

// rackThemeTokens is the contract between the component stylesheets and a
// theme: every custom property the stylesheets read for colour and type. A
// theme must declare each of them on :root (themes/hpe.css documents their
// meaning); a test keeps this list equal to the set the stylesheets use.
var rackThemeTokens = []string{
	"--c-bg-page", "--c-bg-card", "--c-bg-tint", "--c-bg-warn",
	"--c-text", "--c-text-strong", "--c-text-weak", "--c-text-link", "--c-text-on-strong",
	"--c-border", "--c-border-weak", "--c-border-strong",
	"--c-accent", "--c-accent-fill", "--c-brand",
	"--c-ok", "--c-warn", "--c-crit", "--c-info", "--c-unknown",
	"--c-kind-server", "--c-kind-server-weak",
	"--c-kind-chassis", "--c-kind-chassis-weak",
	"--c-kind-switch", "--c-kind-switch-weak",
	"--c-kind-hsn", "--c-kind-hsn-weak",
	"--c-kind-other", "--c-kind-other-weak",
	"--c-shadow",
	"--font-sans", "--font-mono",
}

// cssDeclaration matches the name of a custom property declaration.
var cssDeclaration = regexp.MustCompile(`(--[a-z0-9-]+)\s*:`)

// rackThemeCSS resolves the theme option ("" or a built-in name, else a file
// path) to the stylesheet that fills the theme layer. A file is accepted only
// when it cannot leave the <style> element and defines every token.
func rackThemeCSS(theme string) (template.CSS, error) {
	name := cmp.Or(theme, RackThemeHPE)
	src, err := loadRackTheme(name)
	if err != nil {
		return "", err
	}
	if err := checkRackTheme(src); err != nil {
		return "", fmt.Errorf("theme %s: %w", name, err)
	}
	return template.CSS("@layer theme {\n" + src + "\n}\n"), nil
}

// loadRackTheme returns the raw CSS of a built-in theme or of a theme file.
// A value with a path separator or a .css suffix is a file; any other
// unknown name is an error so that a typo never reads a file by accident.
func loadRackTheme(theme string) (string, error) {
	switch theme {
	case RackThemeHPE, RackThemeNeutral:
		data, err := rackHTMLThemes.ReadFile("themes/" + theme + ".css")
		if err != nil {
			return "", fmt.Errorf("embedded theme %s: %w", theme, err)
		}
		return string(data), nil
	}
	if !strings.ContainsAny(theme, `/\`) && !strings.HasSuffix(strings.ToLower(theme), ".css") {
		return "", fmt.Errorf("unknown theme %q: use %s, or the path of a CSS file", theme, strings.Join(ValidRackThemes(), ", "))
	}
	data, err := os.ReadFile(theme)
	if err != nil {
		return "", fmt.Errorf("read theme file: %w", err)
	}
	return string(data), nil
}

// checkRackTheme rejects CSS that could end the <style> element (the theme is
// written verbatim as template.CSS, and "</" has no case, so this covers
// every spelling of </style>) and CSS that leaves tokens undefined.
func checkRackTheme(src string) error {
	if strings.Contains(src, "</") {
		return fmt.Errorf("contains \"</\", which could end the <style> element")
	}
	if strings.ContainsRune(src, 0) {
		return fmt.Errorf("contains a NUL byte")
	}
	if missing := missingRackThemeTokens(src); len(missing) > 0 {
		return fmt.Errorf("does not define required tokens: %s", strings.Join(missing, ", "))
	}
	return nil
}

// missingRackThemeTokens lists the contract tokens src never declares, in
// contract order.
func missingRackThemeTokens(src string) []string {
	declared := map[string]bool{}
	for _, m := range cssDeclaration.FindAllStringSubmatch(src, -1) {
		declared[m[1]] = true
	}
	var missing []string
	for _, tok := range rackThemeTokens {
		if !declared[tok] {
			missing = append(missing, tok)
		}
	}
	return missing
}
