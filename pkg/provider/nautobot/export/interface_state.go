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
	"sort"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// interfaceState is the set of fields updateInterface writes, in a comparable
// form, so --merge can tell an interface that already matches from one that
// needs a PATCH.
type interfaceState struct {
	Type        string
	MgmtOnly    bool
	Mac         string // upper-cased, as Nautobot stores it
	Label       string
	Description string
	StatusID    uuid.UUID
	RoleID      uuid.UUID
	Tags        string // sorted tag UUIDs joined, keeping the struct comparable
}

// remoteInterfaceState captures the comparable fields of an interface as
// Nautobot returned it.
func remoteInterfaceState(iface nautobotapi.Interface) *interfaceState {
	state := &interfaceState{
		Mac:         strings.ToUpper(ptrStr(iface.MacAddress)),
		Label:       ptrStr(iface.Label),
		Description: ptrStr(iface.Description),
		StatusID:    refID(iface.Status.Id),
		Tags:        joinIDs(refIDs(iface.Tags)),
	}
	if iface.Type.Value != nil {
		state.Type = string(*iface.Type.Value)
	}
	if iface.MgmtOnly != nil {
		state.MgmtOnly = *iface.MgmtOnly
	}
	if iface.Role != nil {
		state.RoleID = refID(iface.Role.Id)
	}
	return state
}

// interfaceCurrent reports whether an existing interface already holds what
// updateInterface would write for spec. Unknown remote state counts as
// stale, which keeps the previous always-PATCH behaviour for it.
func (e *Exporter) interfaceCurrent(existing *CachedItem, spec interfaceSpec) bool {
	if existing == nil || existing.Iface == nil {
		return false
	}
	desired, err := e.desiredInterfaceState(spec, existing.Iface.Mac)
	if err != nil {
		return false
	}
	return desired == *existing.Iface
}

// desiredInterfaceState mirrors updateInterface's request: status is always
// Active, an empty MAC leaves the remote one alone, label and description
// are written as they are, and an empty role clears the remote role.
func (e *Exporter) desiredInterfaceState(spec interfaceSpec, remoteMac string) (interfaceState, error) {
	statusID, err := e.statusRef(string(devicetypes.StatusActive))
	if err != nil {
		return interfaceState{}, err
	}
	roleID, err := e.roleRef(spec.Role)
	if err != nil {
		return interfaceState{}, err
	}
	mac := strings.ToUpper(spec.Mac)
	if mac == "" {
		mac = remoteMac
	}
	return interfaceState{
		Type:        spec.Type,
		MgmtOnly:    spec.MgmtOnly,
		Mac:         mac,
		Label:       spec.Label,
		Description: spec.Description,
		StatusID:    statusID,
		RoleID:      roleID,
		Tags:        joinIDs(e.Cache.resolveTagRefs(spec.Tags)),
	}, nil
}

// refIDs reads the UUIDs out of a generated `*[]struct{Id ...}` reference
// slice through its JSON form, which every such shape shares.
func refIDs(refs any) []uuid.UUID {
	raw, err := json.Marshal(refs)
	if err != nil || string(raw) == "null" {
		return nil
	}
	var items []struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func joinIDs(ids []uuid.UUID) string {
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, id.String())
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}
