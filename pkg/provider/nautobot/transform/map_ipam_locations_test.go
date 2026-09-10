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
package transform

import (
	"testing"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

func TestBuildVLANLocationMapUsesFirstValidAssignment(t *testing.T) {
	vlanID := uuid.New()
	firstLocationID := uuid.New()
	secondLocationID := uuid.New()
	assignments := make([]nautobotapi.VLANLocationAssignment, 3)
	setNBRef(&assignments[0].Location, uuid.New())
	setNBRef(&assignments[1].Vlan, vlanID)
	setNBRef(&assignments[1].Location, firstLocationID)
	setNBRef(&assignments[2].Vlan, vlanID)
	setNBRef(&assignments[2].Location, secondLocationID)

	got := BuildVLANLocationMap(assignments)
	if got[vlanID] != firstLocationID {
		t.Fatalf("VLAN location = %s, want first assignment %s", got[vlanID], firstLocationID)
	}
}

func TestBuildPrefixLocationMapUsesFirstValidAssignment(t *testing.T) {
	prefixID := uuid.New()
	firstLocationID := uuid.New()
	secondLocationID := uuid.New()
	assignments := make([]nautobotapi.PrefixLocationAssignment, 3)
	setNBRef(&assignments[0].Prefix, prefixID)
	setNBRef(&assignments[1].Prefix, prefixID)
	setNBRef(&assignments[1].Location, firstLocationID)
	setNBRef(&assignments[2].Prefix, prefixID)
	setNBRef(&assignments[2].Location, secondLocationID)

	got := BuildPrefixLocationMap(assignments)
	if got[prefixID] != firstLocationID {
		t.Fatalf("prefix location = %s, want first assignment %s", got[prefixID], firstLocationID)
	}
}
