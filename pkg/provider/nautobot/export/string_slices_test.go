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
package export

import "testing"

// -----------------------------------------------------------------------------
// stringSlicesEqual — pure logic
// -----------------------------------------------------------------------------

func TestStringSlicesEqual_SameElements(t *testing.T) {
	if !stringSlicesEqual([]string{"a", "b", "c"}, []string{"a", "b", "c"}) {
		t.Error("identical slices should be equal")
	}
}

func TestStringSlicesEqual_DifferentOrder(t *testing.T) {
	if !stringSlicesEqual([]string{"c", "a", "b"}, []string{"a", "b", "c"}) {
		t.Error("same elements in different order should be equal")
	}
}

func TestStringSlicesEqual_DifferentLengths(t *testing.T) {
	if stringSlicesEqual([]string{"a", "b"}, []string{"a", "b", "c"}) {
		t.Error("different lengths should not be equal")
	}
}

func TestStringSlicesEqual_DifferentElements(t *testing.T) {
	if stringSlicesEqual([]string{"a", "b", "c"}, []string{"a", "b", "d"}) {
		t.Error("different elements should not be equal")
	}
}

func TestStringSlicesEqual_Duplicates(t *testing.T) {
	if stringSlicesEqual([]string{"a", "a", "b"}, []string{"a", "b", "b"}) {
		t.Error("different duplicate counts should not be equal")
	}
}

func TestStringSlicesEqual_BothEmpty(t *testing.T) {
	if !stringSlicesEqual([]string{}, []string{}) {
		t.Error("two empty slices should be equal")
	}
}

func TestStringSlicesEqual_BothNil(t *testing.T) {
	if !stringSlicesEqual(nil, nil) {
		t.Error("two nil slices should be equal")
	}
}
