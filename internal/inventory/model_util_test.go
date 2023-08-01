/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
package inventory

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
)

func buildLocation(ordinals ...int) LocationPath {
	loc := LocationPath{}
	for _, ordinal := range ordinals {
		loc = append(loc, LocationToken{Ordinal: ordinal})
	}
	return loc
}

func toLocationPath(location string) (LocationPath, error) {
	path := LocationPath{}
	tokens := strings.Split(location, "->")
	for _, t := range tokens {
		parts := strings.SplitN(t, ":", 2)
		if len(parts) != 2 {
			return path, fmt.Errorf("location path string invalid. path: %v, part: %v", location, t)
		}
		tt := hardwaretypes.HardwareType(parts[0])
		ordinal, err := strconv.ParseInt(parts[1], 10, 0)
		if err != nil {
			return path, err
		}
		token := LocationToken{HardwareType: tt, Ordinal: int(ordinal)}

		path = append(path, token)
	}
	return path, nil
}

func TestCompareHardwareByTypeThenLocation(t *testing.T) {
	hw1 := Hardware{Type: hardwaretypes.Cabinet}

	hw2 := Hardware{Type: hardwaretypes.Node}

	v := CompareHardwareByTypeThenLocation(&hw1, &hw1)
	if v {
		t.Fatalf("Expected false when comparing same hardware")
	}

	v = CompareHardwareByTypeThenLocation(&hw1, &hw2)
	if !v {
		t.Fatalf("Expected true when comparing Cabinet to Node, hw1: %v, hw2: %v", hw1, hw2)
	}

	v = CompareHardwareByTypeThenLocation(&hw2, &hw1)
	if v {
		t.Fatalf("Expected false when comparing Node to Cabinet, hw2: %v, hw1: %v", hw2, hw1)
	}

	hw3 := Hardware{Type: hardwaretypes.Node}
	hw2.LocationPath = buildLocation(1, 2, 3)
	hw3.LocationPath = buildLocation(1, 2, 3)
	v = CompareHardwareByTypeThenLocation(&hw2, &hw3)
	if v {
		t.Fatalf("Expected false when comparing same hardware with locations, hw2: %v, hw3: %v", hw2, hw3)
	}

	hw2.LocationPath = buildLocation(1, 2, 3)
	hw3.LocationPath = buildLocation(1, 2, 4)
	v = CompareHardwareByTypeThenLocation(&hw2, &hw3)
	if !v {
		t.Fatalf("Expected true when comparing hardware with locations, hw2: %v, hw3: %v", hw2, hw3)
	}

	v = CompareHardwareByTypeThenLocation(&hw3, &hw2)
	if v {
		t.Fatalf("Expected false when comparing hardware with locations, hw3: %v, hw2: %v", hw3, hw2)
	}
}

func TestCompareLocationPath(t *testing.T) {
	location0, err := toLocationPath("System:0->Cabinet:1003->Chassis:7->NodeBlade:7->NodeCard:1->Node:0")
	if err != nil {
		t.Fatalf("error parsing location0, err: %v", err)
	}
	location1, err := toLocationPath("System:0->Cabinet:1003->Chassis:7->NodeBlade:7->NodeCard:1->Node:1")
	if err != nil {
		t.Fatalf("error parsing location1, err: %v", err)
	}

	v := CompareLocationPath(location0, location1)
	if !v {
		t.Fatalf("Expected true when comparing locations, location0: %v, location1: %v", location0, location1)
	}

	v = CompareLocationPath(location1, location0)
	if v {
		t.Fatalf("Expected true when comparing locations, location1: %v, location0: %v", location1, location0)
	}

	locationChassis, err := toLocationPath("System:0->Cabinet:1003->Chassis:7")
	if err != nil {
		t.Fatalf("error parsing locationChassis, err: %v", err)
	}

	v = CompareLocationPath(location1, locationChassis)
	if !v {
		t.Fatalf("Expected true when comparing locations, location1: %v, locationChassis: %v", location1, locationChassis)
	}

	v = CompareLocationPath(locationChassis, location1)
	if v {
		t.Fatalf("Expected false when comparing locations, locationChassis: %v, location1: %v", locationChassis, location1)
	}
}
