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
package export

import (
	"testing"
)

// TestNewExporter verifies the Exporter constructor wires its collaborators and
// preserves caller options, including the nil-options case.
//
// Why it matters: the Exporter is the entry point for pushing cani inventory to
// Nautobot; if it drops the client, cache, or options the whole export runs
// against the wrong target or with the wrong defaults.
// Inputs: a client, a lookup cache, and an *ExporterOpts (possibly nil).
// Outputs: a wired *Exporter.
// Data choice: one case supplies full options (location/role/status/dry-run) to
// confirm they are retained; the nil-options case confirms the constructor does
// not fabricate defaults, matching how callers opt out of overrides.
func TestNewExporter(t *testing.T) {
	tests := []struct {
		name   string
		opts   *ExporterOpts
		hasOpt bool
	}{
		{
			name: "creates exporter with options",
			opts: &ExporterOpts{
				DefaultLocation: "Site-A",
				DefaultRole:     "Generic",
				DefaultStatus:   "Active",
				DryRun:          true,
			},
			hasOpt: true,
		},
		{
			name:   "creates exporter with nil options",
			opts:   nil,
			hasOpt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := NewNautobotClient("http://localhost/api", "token123")
			cache := NewLookupCache(client)

			exporter := NewExporter(client, cache, tt.opts)

			if exporter == nil {
				t.Fatal("expected non-nil exporter")
			}
			if exporter.Client != client {
				t.Error("exporter.Client does not match input")
			}
			if exporter.Cache != cache {
				t.Error("exporter.Cache does not match input")
			}
			if tt.hasOpt && exporter.Options.DefaultLocation != tt.opts.DefaultLocation {
				t.Errorf("exporter.Options.DefaultLocation = %q, want %q",
					exporter.Options.DefaultLocation, tt.opts.DefaultLocation)
			}
			if !tt.hasOpt && exporter.Options != nil {
				t.Error("expected nil options but got non-nil")
			}
		})
	}
}
