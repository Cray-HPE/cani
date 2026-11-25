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
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

//go:embed rack_html.tmpl
var rackHTMLSource string

//go:embed rack_html_detail.tmpl
var rackHTMLDetailSource string

//go:embed rack_html_list.tmpl
var rackHTMLListSource string

//go:embed rack_html.css
var rackHTMLStyles string

//go:embed rack_html_cabling.css
var rackHTMLCablingStyles string

//go:embed rack_html_pane.css
var rackHTMLPaneStyles string

//go:embed rack_html_list.css
var rackHTMLListStyles string

// rackHTMLTemplate parses the embedded elevation, detail-card and rack
// contents templates once; the error is surfaced to every caller instead of
// panicking at package init.
var rackHTMLTemplate = sync.OnceValues(func() (*template.Template, error) {
	funcs := template.FuncMap{
		"plural":    pluralize,
		"sub":       func(a, b int) int { return a - b },
		"kindLabel": kindLabel,
	}
	tmpl, err := template.New("elevation").Funcs(funcs).Parse(rackHTMLSource)
	if err != nil {
		return nil, err
	}
	for _, src := range []string{rackHTMLDetailSource, rackHTMLListSource} {
		if _, err := tmpl.Parse(src); err != nil {
			return nil, err
		}
	}
	return tmpl, nil
})

// pluralize renders "1 rack", "3 racks" or "2 switches".
func pluralize(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	for _, suffix := range []string{"s", "x", "ch", "sh"} {
		if strings.HasSuffix(word, suffix) {
			return fmt.Sprintf("%d %ses", n, word)
		}
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// kindLabel is the human name of a tile kind shown in the card eyebrow.
func kindLabel(kind string) string {
	switch kind {
	case "server":
		return "Server"
	case "chassis":
		return "Chassis"
	case "switch":
		return "Switch"
	case "hsn":
		return "HSN switch"
	default:
		return "Device"
	}
}

// RenderRackHTML writes one self-contained HTML document showing the front
// elevation of every rack matching opts.RackFilter to w. The page uses no
// JavaScript: the label switch, filters, detail drawers and cabling tables
// are native HTML+CSS. opts.Theme picks the colour and type tokens: a
// built-in theme (hpe by default, neutral) or the path of a CSS file that
// defines the token set.
func RenderRackHTML(w io.Writer, inv *devicetypes.Inventory, opts CompactRenderOptions) error {
	if inv == nil {
		return fmt.Errorf("inventory is nil")
	}
	tmpl, err := rackHTMLTemplate()
	if err != nil {
		return fmt.Errorf("parse rack html template: %w", err)
	}
	theme, err := rackThemeCSS(opts.Theme)
	if err != nil {
		return err
	}
	page := buildHTMLPage(inv, opts.RackFilter, time.Now())
	page.ThemeCSS = theme
	page.CSS = template.CSS(strings.Join([]string{rackHTMLStyles, rackHTMLCablingStyles, rackHTMLPaneStyles, rackHTMLListStyles}, "\n"))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, page); err != nil {
		return fmt.Errorf("render rack html: %w", err)
	}
	if _, err := buf.WriteTo(w); err != nil {
		return fmt.Errorf("write rack html: %w", err)
	}
	return nil
}
