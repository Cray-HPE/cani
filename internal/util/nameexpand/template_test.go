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
package nameexpand

import "testing"

func TestIsTemplate(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"gh-%{RACK}u%{U}", true},
		{"%{PARENT}", true},
		{"plain-name", false},
		{"x370{1..2}", false}, // braces but no %
		{"%{SEQ}", true},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsTemplate(tt.input); got != tt.want {
			t.Errorf("IsTemplate(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestExpandTemplateRack(t *testing.T) {
	got, err := ExpandTemplate("device-%{RACK}", map[string]string{"RACK": "x3701"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "device-x3701" {
		t.Errorf("got %q, want %q", got, "device-x3701")
	}
}

func TestExpandTemplateParentAlias(t *testing.T) {
	got, err := ExpandTemplate("gh-%{PARENT}", map[string]string{"RACK": "x3701"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "gh-x3701" {
		t.Errorf("got %q, want %q", got, "gh-x3701")
	}
}

func TestExpandTemplateU(t *testing.T) {
	got, err := ExpandTemplate("u%{U}", map[string]string{"U": "44"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "u44" {
		t.Errorf("got %q, want %q", got, "u44")
	}
}

func TestExpandTemplateSeq(t *testing.T) {
	got, err := ExpandTemplate("node-%{SEQ}", map[string]string{"SEQ": "3"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "node-3" {
		t.Errorf("got %q, want %q", got, "node-3")
	}
}

func TestExpandTemplateFace(t *testing.T) {
	got, err := ExpandTemplate("%{FACE}-dev", map[string]string{"FACE": "front"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "front-dev" {
		t.Errorf("got %q, want %q", got, "front-dev")
	}
}

func TestExpandTemplateCombined(t *testing.T) {
	vars := map[string]string{
		"RACK": "x3701",
		"U":    "44",
		"SEQ":  "1",
		"FACE": "front",
	}
	got, err := ExpandTemplate("gh-%{RACK}u%{U}", vars)
	if err != nil {
		t.Fatal(err)
	}
	if got != "gh-x3701u44" {
		t.Errorf("got %q, want %q", got, "gh-x3701u44")
	}
}

func TestExpandTemplateUnknownToken(t *testing.T) {
	_, err := ExpandTemplate("%{BOGUS}", map[string]string{})
	if err == nil {
		t.Fatal("expected error for unknown token")
	}
}

func TestExpandTemplateUnclosedToken(t *testing.T) {
	_, err := ExpandTemplate("%{RACK", map[string]string{"RACK": "r1"})
	if err == nil {
		t.Fatal("expected error for unclosed token")
	}
}

func TestExpandTemplateMissingValue(t *testing.T) {
	_, err := ExpandTemplate("%{U}", map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
}

func TestExpandTemplateLiteralOnly(t *testing.T) {
	got, err := ExpandTemplate("no-tokens-here", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "no-tokens-here" {
		t.Errorf("got %q", got)
	}
}

func TestExpandTemplateDeviceToken(t *testing.T) {
	got, err := ExpandTemplate("gpu-%{DEVICE}-%{BAY}", map[string]string{
		"DEVICE": "gh-x3701u1",
		"BAY":    "GPU 0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "gpu-gh-x3701u1-GPU 0" {
		t.Errorf("expected 'gpu-gh-x3701u1-GPU 0', got %q", got)
	}
}

func TestExpandTemplateDeviceBaySeq(t *testing.T) {
	got, err := ExpandTemplate("%{DEVICE}-bay%{BAY}-seq%{SEQ}", map[string]string{
		"DEVICE": "server-a",
		"BAY":    "3",
		"SEQ":    "7",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "server-a-bay3-seq7" {
		t.Errorf("expected 'server-a-bay3-seq7', got %q", got)
	}
}

func TestIsTemplateDeviceToken(t *testing.T) {
	if !IsTemplate("%{DEVICE}") {
		t.Error("expected %{DEVICE} to be recognized as template")
	}
	if !IsTemplate("gpu-%{BAY}") {
		t.Error("expected gpu-%{BAY} to be recognized as template")
	}
}
