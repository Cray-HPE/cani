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
package nautobot

import "testing"

func TestApplyMetadata(t *testing.T) {
	p := New()

	var pm map[string]any
	p.ApplyMetadata(&pm, map[string]string{"rack": "r1", "site": "dc1"})

	sub, ok := pm[p.Slug()].(map[string]any)
	if !ok {
		t.Fatalf("expected %q sub-map, got %v", p.Slug(), pm)
	}
	if sub["rack"] != "r1" || sub["site"] != "dc1" {
		t.Errorf("sub-map = %v, want rack=r1 site=dc1", sub)
	}
}

func TestApplyMetadataMergesExisting(t *testing.T) {
	p := New()

	pm := map[string]any{
		p.Slug(): map[string]any{"existing": "v"},
	}
	p.ApplyMetadata(&pm, map[string]string{"new": "w"})

	sub := pm[p.Slug()].(map[string]any)
	if sub["existing"] != "v" || sub["new"] != "w" {
		t.Errorf("sub-map = %v, want existing=v new=w", sub)
	}
}

func TestApplyMetadataEmptyNoop(t *testing.T) {
	p := New()

	var pm map[string]any
	p.ApplyMetadata(&pm, nil)
	if pm != nil {
		t.Errorf("expected nil map for empty meta, got %v", pm)
	}
}
