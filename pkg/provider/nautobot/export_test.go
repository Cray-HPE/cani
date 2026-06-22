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

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
)

// newExportFlagCmd returns a cli command registering the flags that
// applyFlagOverrides inspects, all left in their unchanged default state.
func newExportFlagCmd() *cli.Command {
	cmd := &cli.Command{}
	cmd.Flags().String("default-location", "", "")
	cmd.Flags().String("default-role", "", "")
	cmd.Flags().String("default-status", "", "")
	cmd.Flags().Bool("merge", false, "")
	cmd.Flags().Bool("dry-run", false, "")
	return cmd
}

func TestApplyFlagOverrides_ChangedFlagsOverrideOptions(t *testing.T) {
	cmd := newExportFlagCmd()
	mustSet(t, cmd, "default-location", "CLI-Loc")
	mustSet(t, cmd, "default-role", "CLI-Role")
	mustSet(t, cmd, "default-status", "CLI-Status")
	mustSet(t, cmd, "merge", "true")
	mustSet(t, cmd, "dry-run", "true")

	p := New()
	p.applyFlagOverrides(cmd)

	if p.Options.DefaultLocation != "CLI-Loc" {
		t.Errorf("DefaultLocation = %q, want %q", p.Options.DefaultLocation, "CLI-Loc")
	}
	if p.Options.DefaultRole != "CLI-Role" {
		t.Errorf("DefaultRole = %q, want %q", p.Options.DefaultRole, "CLI-Role")
	}
	if p.Options.DefaultStatus != "CLI-Status" {
		t.Errorf("DefaultStatus = %q, want %q", p.Options.DefaultStatus, "CLI-Status")
	}
	if !p.Options.Export.Merge {
		t.Error("Export.Merge = false, want true")
	}
	if !p.Options.Export.DryRun {
		t.Error("Export.DryRun = false, want true")
	}
}

func TestApplyFlagOverrides_UnchangedFlagsLeaveOptionsIntact(t *testing.T) {
	cmd := newExportFlagCmd()

	p := New()
	p.Options.DefaultLocation = "preset-loc"
	p.Options.Export.Merge = true

	p.applyFlagOverrides(cmd)

	if p.Options.DefaultLocation != "preset-loc" {
		t.Errorf("DefaultLocation = %q, want it unchanged (%q)", p.Options.DefaultLocation, "preset-loc")
	}
	if !p.Options.Export.Merge {
		t.Error("Export.Merge should remain true when --merge is not changed")
	}
}

func TestApplyFlagOverrides_InitializesNilExportOptions(t *testing.T) {
	cmd := newExportFlagCmd()

	p := New()
	p.Options.Export = nil

	p.applyFlagOverrides(cmd)

	if p.Options.Export == nil {
		t.Fatal("Export options should be initialized when nil")
	}
}

// mustSet sets a flag value, marking it as changed, and fails the test on error.
func mustSet(t *testing.T, cmd *cli.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("failed to set flag %q=%q: %v", name, value, err)
	}
}
