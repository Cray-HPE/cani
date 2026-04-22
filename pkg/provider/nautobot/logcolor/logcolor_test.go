/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package logcolor

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// newTestLogger creates a Logger that writes to a buffer for capture.
func newTestLogger(noColor bool) (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return &Logger{
		std:     log.New(&buf, "", 0),
		noColor: noColor,
	}, &buf
}

func TestNew(t *testing.T) {
	l := New("[test] ", false)
	if l == nil {
		t.Fatal("expected non-nil Logger")
	}
	if l.noColor {
		t.Error("expected noColor=false")
	}

	l2 := New("[test] ", true)
	if !l2.noColor {
		t.Error("expected noColor=true")
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name      string
		noColor   bool
		code      string
		input     string
		wantAnsi  bool
		wantExact string
	}{
		{
			name:     "color enabled wraps with ANSI",
			noColor:  false,
			code:     green,
			input:    "hello",
			wantAnsi: true,
		},
		{
			name:      "color disabled returns raw text",
			noColor:   true,
			code:      green,
			input:     "hello",
			wantAnsi:  false,
			wantExact: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := newTestLogger(tt.noColor)
			got := l.wrap(tt.code, tt.input)

			if tt.wantAnsi {
				if !strings.Contains(got, "\033[") {
					t.Errorf("expected ANSI codes in %q", got)
				}
				if !strings.Contains(got, "hello") {
					t.Errorf("expected %q in output", "hello")
				}
				if !strings.HasSuffix(got, reset) {
					t.Errorf("expected reset suffix in %q", got)
				}
			} else {
				if got != tt.wantExact {
					t.Errorf("wrap() = %q, want %q", got, tt.wantExact)
				}
			}
		})
	}
}

// logMethodTest defines a test case for a log method.
type logMethodTest struct {
	name           string
	call           func(l *Logger)
	wantSubstring  string
	wantAnsiColor  bool // when noColor=false, should contain ANSI?
	wantNoColorStr string // substring expected when noColor=true
}

func testLogMethods(t *testing.T, tests []logMethodTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name+"/color", func(t *testing.T) {
			l, buf := newTestLogger(false)
			tt.call(l)
			out := buf.String()
			if !strings.Contains(out, tt.wantSubstring) {
				t.Errorf("output %q missing substring %q", out, tt.wantSubstring)
			}
			if tt.wantAnsiColor && !strings.Contains(out, "\033[") {
				t.Errorf("expected ANSI codes in colored output %q", out)
			}
		})

		t.Run(tt.name+"/nocolor", func(t *testing.T) {
			l, buf := newTestLogger(true)
			tt.call(l)
			out := buf.String()
			if strings.Contains(out, "\033[") {
				t.Errorf("unexpected ANSI codes in noColor output %q", out)
			}
			want := tt.wantNoColorStr
			if want == "" {
				want = tt.wantSubstring
			}
			if !strings.Contains(out, want) {
				t.Errorf("output %q missing substring %q", out, want)
			}
		})
	}
}

func TestSemanticLogMethods(t *testing.T) {
	testLogMethods(t, []logMethodTest{
		{
			name:          "Info",
			call:          func(l *Logger) { l.Info("connected to %s", "nautobot") },
			wantSubstring: "connected to nautobot",
			wantAnsiColor: true,
		},
		{
			name:          "Detail",
			call:          func(l *Logger) { l.Detail("loaded %d items", 42) },
			wantSubstring: "loaded 42 items",
			wantAnsiColor: true,
		},
		{
			name:          "Created",
			call:          func(l *Logger) { l.Created("device %s", "server-1") },
			wantSubstring: "device server-1",
			wantAnsiColor: true,
		},
		{
			name:          "Ok",
			call:          func(l *Logger) { l.Ok("already exists") },
			wantSubstring: "already exists",
			wantAnsiColor: true,
		},
		{
			name:          "Skipped",
			call:          func(l *Logger) { l.Skipped("duplicate %s", "eth0") },
			wantSubstring: "duplicate eth0",
			wantAnsiColor: true,
		},
		{
			name:          "Changed",
			call:          func(l *Logger) { l.Changed("updated position") },
			wantSubstring: "updated position",
			wantAnsiColor: true,
		},
		{
			name:          "Warn",
			call:          func(l *Logger) { l.Warn("retrying %d", 3) },
			wantSubstring: "retrying 3",
			wantAnsiColor: true,
		},
		{
			name:          "Error",
			call:          func(l *Logger) { l.Error("failed: %s", "timeout") },
			wantSubstring: "failed: timeout",
			wantAnsiColor: true,
		},
		{
			name:          "DryRun",
			call:          func(l *Logger) { l.DryRun("would create %s", "device") },
			wantSubstring: "[DRY-RUN] would create device",
			wantAnsiColor: true,
		},
		{
			name:          "Header",
			call:          func(l *Logger) { l.Header("Phase %d", 1) },
			wantSubstring: "Phase 1",
			wantAnsiColor: true,
		},
		{
			name:          "SummaryCreated",
			call:          func(l *Logger) { l.SummaryCreated("%d devices", 5) },
			wantSubstring: "  + 5 devices",
			wantAnsiColor: true,
		},
		{
			name:          "SummaryChanged",
			call:          func(l *Logger) { l.SummaryChanged("%d updated", 2) },
			wantSubstring: "  ~ 2 updated",
			wantAnsiColor: true,
		},
		{
			name:          "SummarySkipped",
			call:          func(l *Logger) { l.SummarySkipped("%d skipped", 1) },
			wantSubstring: "  - 1 skipped",
			wantAnsiColor: true,
		},
		{
			name:          "SummaryError",
			call:          func(l *Logger) { l.SummaryError("%d errors", 3) },
			wantSubstring: "  ! 3 errors",
			wantAnsiColor: true,
		},
		{
			name:          "Plain",
			call:          func(l *Logger) { l.Plain("raw %s", "text") },
			wantSubstring: "raw text",
			wantAnsiColor: false,
		},
	})
}

func TestDiff(t *testing.T) {
	t.Run("color mode includes both ANSI codes and field values", func(t *testing.T) {
		l, buf := newTestLogger(false)
		l.Diff("position", "U10", "U20")
		out := buf.String()

		if !strings.Contains(out, "position") {
			t.Errorf("output missing field name: %q", out)
		}
		if !strings.Contains(out, "U10") {
			t.Errorf("output missing old value: %q", out)
		}
		if !strings.Contains(out, "U20") {
			t.Errorf("output missing new value: %q", out)
		}
		if !strings.Contains(out, "-->") {
			t.Errorf("output missing arrow separator: %q", out)
		}
		if !strings.Contains(out, "\033[") {
			t.Errorf("expected ANSI codes in colored diff: %q", out)
		}
	})

	t.Run("noColor mode has no ANSI codes", func(t *testing.T) {
		l, buf := newTestLogger(true)
		l.Diff("status", "Active", "Planned")
		out := buf.String()

		if strings.Contains(out, "\033[") {
			t.Errorf("unexpected ANSI codes in noColor diff: %q", out)
		}
		if !strings.Contains(out, "Active") || !strings.Contains(out, "Planned") {
			t.Errorf("output missing values: %q", out)
		}
	})
}
