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

// Package logcolor provides ANSI-colored log output for the Nautobot export
// workflow. It categorises messages into semantic levels so the user can
// quickly distinguish primary actions from supplementary information.
//
// Color scheme (Ansible-inspired):
//
//	White/Bold  – primary actions (created, connected, etc.)
//	Gray/Dim    – supplementary / thinking messages
//	Green       – created items (new)
//	DarkGreen   – no-change / idempotent items (already up-to-date)
//	Cyan        – skipped items
//	Yellow      – changed / updated items
//	Red         – errors or would-change diffs
package logcolor

import (
	"fmt"
	"log"
	"os"
)

// ANSI escape sequences.
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	white     = "\033[97m"
	red       = "\033[31m"
	green     = "\033[32m"
	darkGreen = "\033[38;5;34m"
	yellow    = "\033[33m"
	cyan      = "\033[36m"
	gray      = "\033[90m"
)

// Logger wraps a standard log.Logger and adds coloured helpers.
type Logger struct {
	std     *log.Logger
	noColor bool
}

// New creates a Logger that writes to stderr with the given prefix.
func New(prefix string, noColor bool) *Logger {
	return &Logger{
		std:     log.New(os.Stderr, prefix, log.LstdFlags),
		noColor: noColor,
	}
}

// wrap applies an ANSI colour to s when colour is enabled.
func (l *Logger) wrap(code, s string) string {
	if l.noColor {
		return s
	}
	return code + s + reset
}

// --- Semantic log levels ---------------------------------------------------

// Info logs a primary action message in bold white.
// Use for main flow events: "Successfully connected", "Created device", etc.
func (l *Logger) Info(format string, args ...any) {
	l.std.Println(l.wrap(bold+white, fmt.Sprintf(format, args...)))
}

// Detail logs a supplementary/thinking message in dim gray.
// Use for config notes: "Auto-creation enabled", "No location specified", etc.
func (l *Logger) Detail(format string, args ...any) {
	l.std.Println(l.wrap(gray, fmt.Sprintf(format, args...)))
}

// Created logs a successfully-created item in green.
func (l *Logger) Created(format string, args ...any) {
	l.std.Println(l.wrap(green, fmt.Sprintf(format, args...)))
}

// Ok logs an idempotent / no-change item in dark green.
func (l *Logger) Ok(format string, args ...any) {
	l.std.Println(l.wrap(darkGreen, fmt.Sprintf(format, args...)))
}

// Skipped logs a skipped / already-exists item in cyan.
func (l *Logger) Skipped(format string, args ...any) {
	l.std.Println(l.wrap(cyan, fmt.Sprintf(format, args...)))
}

// Changed logs an updated / changed item in yellow.
func (l *Logger) Changed(format string, args ...any) {
	l.std.Println(l.wrap(yellow, fmt.Sprintf(format, args...)))
}

// Diff logs a would-change detail line with red (old) → green (new).
func (l *Logger) Diff(field, oldVal, newVal string) {
	l.std.Println(fmt.Sprintf("      %s: %s --> %s",
		field,
		l.wrap(red, oldVal),
		l.wrap(green, newVal),
	))
}

// Warn logs a warning in yellow.
func (l *Logger) Warn(format string, args ...any) {
	l.std.Println(l.wrap(yellow, fmt.Sprintf(format, args...)))
}

// Error logs an error message in red.
func (l *Logger) Error(format string, args ...any) {
	l.std.Println(l.wrap(red, fmt.Sprintf(format, args...)))
}

// DryRun logs a dry-run message in cyan with a [DRY-RUN] prefix.
func (l *Logger) DryRun(format string, args ...any) {
	l.std.Println(l.wrap(cyan, "[DRY-RUN] "+fmt.Sprintf(format, args...)))
}

// Header logs a section header in bold white.
func (l *Logger) Header(format string, args ...any) {
	l.std.Println(l.wrap(bold+white, fmt.Sprintf(format, args...)))
}

// SummaryCreated logs a summary line for created items (green with + prefix).
func (l *Logger) SummaryCreated(format string, args ...any) {
	l.std.Println(l.wrap(green, "  + "+fmt.Sprintf(format, args...)))
}

// SummaryChanged logs a summary line for changed items (yellow with ~ prefix).
func (l *Logger) SummaryChanged(format string, args ...any) {
	l.std.Println(l.wrap(yellow, "  ~ "+fmt.Sprintf(format, args...)))
}

// SummarySkipped logs a summary line for skipped items (cyan with - prefix).
func (l *Logger) SummarySkipped(format string, args ...any) {
	l.std.Println(l.wrap(cyan, "  - "+fmt.Sprintf(format, args...)))
}

// SummaryError logs a summary line for errors (red with ! prefix).
func (l *Logger) SummaryError(format string, args ...any) {
	l.std.Println(l.wrap(red, "  ! "+fmt.Sprintf(format, args...)))
}

// Plain logs uncolored text (pass-through), useful for blank lines.
func (l *Logger) Plain(format string, args ...any) {
	l.std.Println(fmt.Sprintf(format, args...))
}
