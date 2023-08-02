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

// CompareHardwareByTypeThenLocation returns true if hw1 should sort before h2
// otherwise it returns false
func CompareHardwareByTypeThenLocation(hw1 *Hardware, hw2 *Hardware) bool {
	if hw1.Type == hw2.Type {
		return CompareLocationPath(hw1.LocationPath, hw2.LocationPath)
	}
	return hw1.Type < hw2.Type
}

// CompareLocationPath returns true if location1 should sort before location2
// otherwise it returns false
func CompareLocationPath(location1 LocationPath, location2 LocationPath) bool {
	len2 := len(location2)
	for i, loc1 := range location1 {
		if i >= len2 {
			return true
		}

		loc2 := location2[i]
		if loc1.HardwareType == loc2.HardwareType {
			if loc1.Ordinal == loc2.Ordinal {
				continue // go to the next location
			}
			return loc1.Ordinal < loc2.Ordinal
		} else {
			return loc1.HardwareType < loc2.HardwareType
		}
	}
	// This case means that the two locations are the same
	return false
}
