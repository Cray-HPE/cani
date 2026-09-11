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

import (
	"encoding/json"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// interfacePatch wraps a PatchedWritableInterfaceRequest for the interface
// create/enrich PATCH paths, which always re-send the role FK so that adding
// LAG/VLAN settings never silently drops a role (FORGE-305).
//
// Nautobot 3.2 marks Role as `json:"role,omitempty"`, so a nil Role would be
// omitted rather than cleared. MarshalJSON emits an explicit `"role":null` when
// no role is set, restoring the ability to clear an interface's role on the wire.
type interfacePatch struct {
	nautobotapi.PatchedWritableInterfaceRequest
}

// MarshalJSON serializes the embedded request and forces `"role":null` when the
// role is absent, so an emptied local role clears it in Nautobot.
func (p interfacePatch) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(p.PatchedWritableInterfaceRequest)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	if _, ok := fields["role"]; !ok {
		fields["role"] = json.RawMessage("null")
	}
	return json.Marshal(fields)
}
