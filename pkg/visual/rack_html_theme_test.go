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
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// componentStylesheets are the theme-agnostic stylesheets of the page.
var componentStylesheets = map[string]string{
	"rack_html.css":         rackHTMLStyles,
	"rack_html_cabling.css": rackHTMLCablingStyles,
	"rack_html_pane.css":    rackHTMLPaneStyles,
	"rack_html_list.css":    rackHTMLListStyles,
}

var (
	colourLiteral = regexp.MustCompile(`(?i)#[0-9a-f]{3,8}\b|\b(?:rgba?|hsla?|hwb|lab|lch|oklab|oklch|color|light-dark)\(|[:,(\s](?:white|black|red|green|blue|gr[ae]y|orange|yellow|purple)[\s;,)}]`)
	fontName      = regexp.MustCompile(`(?i)graphik|arial|helvetica|menlo|consolas|segoe|roboto|system-ui|ui-monospace|sans-serif|monospace|\bserif\b`)
	fontProperty  = regexp.MustCompile(`\bfont(?:-family)?\s*:\s*([^;}]*)`)
	tokenUse      = regexp.MustCompile(`var\((--(?:c|font)-[a-z0-9-]+)`)
)

func TestComponentCSSHoldsNoThemeLiterals(t *testing.T) {
	for name, css := range componentStylesheets {
		if strings.Contains(css, "--hpe-") {
			t.Errorf("%s: references an HPE token; component CSS must use the semantic --c-* tokens", name)
		}
		if m := colourLiteral.FindString(css); m != "" {
			t.Errorf("%s: contains the colour literal %q; colours belong to the theme", name, m)
		}
		if m := fontName.FindString(css); m != "" {
			t.Errorf("%s: contains the font family name %q; fonts belong to the theme", name, m)
		}
		for _, m := range fontProperty.FindAllStringSubmatch(css, -1) {
			if v := strings.TrimSpace(m[1]); v != "inherit" && !strings.Contains(v, "var(--font-") {
				t.Errorf("%s: font declaration %q must use --font-sans or --font-mono", name, m[0])
			}
		}
	}
}

func TestThemeTokenContractMatchesComponentCSS(t *testing.T) {
	used := map[string]bool{}
	for _, css := range componentStylesheets {
		for _, m := range tokenUse.FindAllStringSubmatch(css, -1) {
			used[m[1]] = true
		}
	}
	for _, tok := range rackThemeTokens {
		if !used[tok] {
			t.Errorf("contract token %s is not used by any component stylesheet, or is listed twice", tok)
		}
		delete(used, tok)
	}
	for tok := range used {
		t.Errorf("component CSS uses %s, which is missing from the theme contract", tok)
	}
}

func TestBuiltInThemesDefineEveryToken(t *testing.T) {
	for _, name := range ValidRackThemes() {
		src, err := loadRackTheme(name)
		if err != nil {
			t.Fatalf("load theme %s: %v", name, err)
		}
		if missing := missingRackThemeTokens(src); len(missing) > 0 {
			t.Errorf("theme %s misses %v", name, missing)
		}
		if !strings.Contains(src, "color-scheme:") {
			t.Errorf("theme %s must declare color-scheme", name)
		}
	}
	neutral, _ := loadRackTheme(RackThemeNeutral)
	if strings.Contains(neutral, "url(") || strings.Contains(neutral, "--hpe-") {
		t.Errorf("the neutral theme must not load external assets or HPE tokens")
	}
	hpe, _ := loadRackTheme(RackThemeHPE)
	if strings.Count(hpe, "https://www.hpe.com/content/dam/hpe/fonts/graphik/") != 4 || !strings.Contains(hpe, "--font-sans: 'HPE Graphik', Arial, sans-serif;") {
		t.Errorf("the hpe theme must keep the four Graphik faces and the HPE font stack")
	}
}

func renderThemed(t *testing.T, theme string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	err := RenderRackHTML(&buf, htmlRackInventory(), CompactRenderOptions{Theme: theme})
	return buf.String(), err
}

func TestRenderRackHTMLDefaultThemeIsHPE(t *testing.T) {
	out, err := renderThemed(t, "")
	if err != nil {
		t.Fatalf("RenderRackHTML: %v", err)
	}
	themeAt, layersAt := strings.Index(out, "<style>@layer theme {\n"), strings.Index(out, "@layer theme, tokens, base,")
	if themeAt < 0 || layersAt < 0 || themeAt > layersAt {
		t.Errorf("the theme block must precede the component stylesheets, which declare the layer order")
	}
	for _, want := range []string{"--c-brand: var(--hpe-color-decorative-brand);", "font-family: 'HPE Graphik'", "color-scheme: light dark;"} {
		if !strings.Contains(out, want) {
			t.Errorf("default theme must be hpe; missing %q", want)
		}
	}
	explicit, err := renderThemed(t, RackThemeHPE)
	if err != nil || explicit[:strings.Index(explicit, "Generated ")] != out[:strings.Index(out, "Generated ")] {
		t.Errorf("an empty theme and %q must render the same page", RackThemeHPE)
	}
}

func TestRenderRackHTMLNeutralTheme(t *testing.T) {
	out, err := renderThemed(t, RackThemeNeutral)
	if err != nil {
		t.Fatalf("RenderRackHTML: %v", err)
	}
	if !strings.Contains(out, "--font-sans: system-ui,") || strings.Contains(out, "hpe.com") || strings.Contains(out, "--hpe-") {
		t.Errorf("the neutral theme must use system fonts and no HPE asset or token")
	}
	if strings.Count(out, "<style>") != 3 {
		t.Errorf("expected theme, component and highlight style blocks")
	}
}

// completeTheme is a minimal theme file declaring every contract token.
func completeTheme(skip ...string) string {
	var b strings.Builder
	b.WriteString(":root { color-scheme: light;\n")
	for _, tok := range rackThemeTokens {
		if !slices.Contains(skip, tok) {
			b.WriteString(tok + ": #123456;\n")
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func writeTheme(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "custom.css")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRenderRackHTMLCustomThemeFile(t *testing.T) {
	src := completeTheme()
	out, err := renderThemed(t, writeTheme(t, src))
	if err != nil {
		t.Fatalf("RenderRackHTML: %v", err)
	}
	if !strings.Contains(out, "<style>@layer theme {\n"+src+"\n}\n</style>") {
		t.Errorf("the custom theme must be injected verbatim inside the theme layer")
	}
	if strings.Contains(out, "--hpe-") || regexp.MustCompile(`(?i)<script`).MatchString(out) || regexp.MustCompile(`(?i)\son[a-z]+=`).MatchString(out) {
		t.Errorf("a custom theme must replace the built-in one and the page must stay script-free")
	}
}

func TestRenderRackHTMLCustomThemeRejectsStyleBreakout(t *testing.T) {
	for name, src := range map[string]string{
		"</style>":   completeTheme() + "</style><script>alert(1)</script>",
		"</STYLE >":  completeTheme() + "</STYLE ><p>x",
		"NUL byte":   completeTheme() + "\x00",
		"nested </ ": "/* </ */" + completeTheme(),
	} {
		out, err := renderThemed(t, writeTheme(t, src))
		if err == nil {
			t.Errorf("%s: a theme file that can leave the <style> element must be rejected", name)
		}
		if out != "" {
			t.Errorf("%s: nothing may be written when the theme is rejected", name)
		}
	}
}

func TestRenderRackHTMLCustomThemeMissingTokens(t *testing.T) {
	_, err := renderThemed(t, writeTheme(t, completeTheme("--c-accent", "--font-mono")))
	if err == nil || !strings.HasSuffix(err.Error(), "does not define required tokens: --c-accent, --font-mono") {
		t.Errorf("expected the missing tokens in contract order, got %v", err)
	}
	if missing := missingRackThemeTokens(":root{--c-text : red}"); slices.Contains(missing, "--c-text") {
		t.Errorf("a declaration with space before the colon counts as defined")
	}
}

func TestRenderRackHTMLThemeErrors(t *testing.T) {
	_, err := renderThemed(t, "dark")
	if err == nil || !strings.Contains(err.Error(), `unknown theme "dark": use hpe, neutral, or the path of a CSS file`) {
		t.Errorf("an unknown theme name must be a clear error, got %v", err)
	}
	_, err = renderThemed(t, filepath.Join(t.TempDir(), "missing.css"))
	if err == nil || !strings.Contains(err.Error(), "read theme file") {
		t.Errorf("an unreadable theme file must be a clear error, got %v", err)
	}
}
