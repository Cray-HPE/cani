// MIT License
//
// (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package configs

import "sort"

// The Key is xname as that is the only thing that can link the CCJ and SLS.
type ApplicationNodeMetadataMap map[string]ApplicationNodeMetadata

type ApplicationNodeMetadata struct {
	CANUCommonName string   `yaml:"canu_common_name,omitempty"`
	SubRole        string   `yaml:"subrole"`
	Aliases        []string `yaml:"aliases"`
}

func (m ApplicationNodeMetadataMap) AllAliases() map[string][]string {
	allAliases := map[string][]string{}
	for xname, metadata := range m {
		for _, alias := range metadata.Aliases {
			allAliases[alias] = append(allAliases[alias], xname)
		}
	}

	for _, nodes := range allAliases {
		sort.Strings(nodes)
	}

	return allAliases

}
