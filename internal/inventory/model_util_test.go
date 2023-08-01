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
		t.Fatalf("Expected false when comparing Node to Cabinet, hw1: %v, hw2: %v", hw2, hw1)
	}

	hw3 := Hardware{Type: hardwaretypes.Node}
	hw2.LocationPath = buildLocation(1, 2, 3)
	hw3.LocationPath = buildLocation(1, 2, 3)
	v = CompareHardwareByTypeThenLocation(&hw2, &hw3)
	if v {
		t.Fatalf("Expected false when comparing same hardware with locations, hw1: %v, hw2: %v", hw2, hw3)
	}

	hw2.LocationPath = buildLocation(1, 2, 3)
	hw3.LocationPath = buildLocation(1, 2, 4)
	v = CompareHardwareByTypeThenLocation(&hw2, &hw3)
	if !v {
		t.Fatalf("Expected true when comparing hardware with locations, hw1: %v, hw2: %v", hw2, hw3)
	}

	v = CompareHardwareByTypeThenLocation(&hw3, &hw2)
	if v {
		t.Fatalf("Expected false when comparing hardware with locations, hw1: %v, hw2: %v", hw3, hw2)
	}
}
