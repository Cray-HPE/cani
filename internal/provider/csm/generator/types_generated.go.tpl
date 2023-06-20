// MIT License
//
// (C) Copyright [2023] Hewlett Packard Enterprise Development LP
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

// WARNING GENERATED FILE DO NOT EDIT!

package csm

import (
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

func BuildXname(cHardware inventory.Hardware, locationPath inventory.LocationPath) (xnames.Xname, error) {
	hsmType, err :=  GetHMSType(cHardware, locationPath)
	if err != nil {
		return nil, err
	}

	switch hsmType {
	case xnametypes.HMSTypeInvalid:
		return nil, nil 
{{- range $xnameType := . }}
	case xnametypes.{{$xnameType.Entry.Type}}:
		return xnames.{{$xnameType.Entry.Type}}{
			{{- range $i, $field := $xnameType.Fields }}
			{{ $field.Name }}: locationPath[{{$field.LocationIndex}}].Ordinal,
			{{- end }}
		}, nil
{{- end }}
	}
	return nil, fmt.Errorf("unknown xnametype '%s'", hsmType.String())
}