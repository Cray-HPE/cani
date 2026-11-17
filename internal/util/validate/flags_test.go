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
package validate

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
)

// TestRejectFlags_NamesTheFlagAndTheCommand verifies that a set flag is
// refused with its name and the full command path, and that the same flag
// left unset, or a name the command never registered, is not an error.
//
// Why it matters: a parent command hands its persistent flags to every
// subcommand, including ones with nothing to store them in; the refusal has
// to say which flag and which command so the user can tell it from a typo.
// Inputs: root "cani" -> "add" -> "metadata" with an inherited --tag that
// is set and an inherited --status that is not.
// Outputs: an error reading `--tag does not apply to "cani add metadata"`
// for --tag; nil for --status and for an unregistered name.
// Data choice: a three-level tree proves the path is walked to the root,
// and the unset sibling proves only flags the user typed are refused.
func TestRejectFlags_NamesTheFlagAndTheCommand(t *testing.T) {
	// Arrange.
	root := &cli.Command{Use: "cani"}
	add := &cli.Command{Use: "add"}
	metadata := &cli.Command{Use: "metadata <kind>"}
	root.AddCommand(add)
	add.AddCommand(metadata)
	metadata.Flags().StringArray("tag", nil, "tags")
	metadata.Flags().String("status", "", "status")
	if err := metadata.Flags().Set("tag", "audit"); err != nil {
		t.Fatalf("set --tag: %v", err)
	}

	// Act.
	refused := RejectFlags(metadata, "status", "tag")
	unset := RejectFlags(metadata, "status")
	unknown := RejectFlags(metadata, "no-such-flag")

	// Assert.
	if want := `--tag does not apply to "cani add metadata"`; refused == nil || refused.Error() != want {
		t.Errorf("RejectFlags(status, tag) = %v, want %s", refused, want)
	}
	if unset != nil || unknown != nil {
		t.Errorf("unset flag error = %v, unregistered flag error = %v, want nil for both", unset, unknown)
	}
}
