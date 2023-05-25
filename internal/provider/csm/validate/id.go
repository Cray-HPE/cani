/*
MIT License

(C) Copyright 2023 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package validate

import "fmt"

type ID struct {
	Pairs []Pair
}

type Pair struct {
	KeyName string
	Name    string
}

func NewID(keyName string, name string) *ID {
	pairs := make([]Pair, 1)
	pairs[0] = Pair{keyName, name}
	// return &ID{keyName, name}
	return &ID{pairs}
}

func (id *ID) append(pair Pair) *ID {
	pairs := append(id.Pairs, pair)
	return &ID{pairs}
}

func (id *ID) str() string {
	s := ""
	for _, pair := range id.Pairs {
		if s != "" {
			s = s + "."
		}
		s = s + pair.Name
	}
	return s
}

func (id *ID) strYaml() string {
	s := "["
	for _, pair := range id.Pairs {
		if s != "[" {
			s += ","
		}
		s = fmt.Sprintf("%s%s:%s", s, pair.KeyName, pair.Name)
	}
	s += "]"
	return s
}

func (id *ID) strPair() string {
	s := ""
	for _, pair := range id.Pairs {
		if s != "" {
			s = s + ","
		}
		s = fmt.Sprintf("%s%s:%s", s, pair.KeyName, pair.Name)
	}
	return s
}
