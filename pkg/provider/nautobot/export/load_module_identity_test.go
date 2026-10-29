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
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// identityModule is Nautobot's view of one module for the identity fake.
type identityModule struct {
	id       uuid.UUID
	serial   string
	assetTag string
	typeID   uuid.UUID
}

// identityListServer answers module list queries filtered by asset_tag,
// serial and module_type from a fixed set, answers the module-type lookup
// with nicTypeID when it is set and with no result otherwise, and records
// every request path and query.
func identityListServer(nicTypeID uuid.UUID, modules []identityModule, calls *[]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		*calls = append(*calls, r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case strings.Contains(r.URL.Path, "dcim/module-types"):
			if nicTypeID == uuid.Nil {
				_, _ = io.WriteString(w, emptyListJSON)
				return
			}
			fmt.Fprintf(w, `{"count":1,"results":[{"id":%q,"model":"NIC","display":"NIC"}]}`, nicTypeID)
		case strings.HasPrefix(r.URL.Path, "/dcim/modules/"):
			query := r.URL.Query()
			var results []string
			for _, m := range modules {
				if (query.Get("asset_tag") == "" && query.Get("serial") == "") ||
					(query.Get("asset_tag") != "" && m.assetTag != query.Get("asset_tag")) ||
					(query.Get("serial") != "" && m.serial != query.Get("serial")) ||
					(query.Get("module_type") != "" && m.typeID.String() != query.Get("module_type")) {
					continue
				}
				results = append(results, fmt.Sprintf(`{"id":%q,"display":"m","serial":%q,"asset_tag":%q}`, m.id, m.serial, m.assetTag))
			}
			fmt.Fprintf(w, `{"count":%d,"results":[%s]}`, len(results), strings.Join(results, ","))
		default:
			_, _ = io.WriteString(w, emptyListJSON)
		}
	}
}

func moduleListCalls(calls []string) int {
	n := 0
	for _, call := range calls {
		if strings.HasPrefix(call, "/dcim/modules/") {
			n++
		}
	}
	return n
}

// TestFindSpareByIdentityMatchesAssetTagThenSerialOnType verifies a spare
// without a recorded ID is found by its asset tag first, otherwise by its
// serial among modules of its own type, and that a spare with neither
// identity, or whose type Nautobot lacks, is new without a module lookup.
//
// Why it matters: an asset tag is unique across Nautobot while a serial is
// unique only per module type, so a serial-only lookup adopted a module of
// another type that shared the serial, and a spare authored with only an
// asset tag was never found and was re-created into a uniqueness error.
// Inputs: Nautobot holding a NIC and an HBA with serial SN-1 and a module
// with asset tag AT-9; spares identified by AT-9, by SN-1 on model NIC, by
// SN-1 on a model Nautobot lacks, and by nothing. Outputs: the AT-9 module,
// the NIC, nil with no module list call, and nil with no call at all.
// Data choice: the shared serial across two types is the case a serial-only
// lookup gets wrong.
func TestFindSpareByIdentityMatchesAssetTagThenSerialOnType(t *testing.T) {
	nicType, hbaType := uuid.New(), uuid.New()
	nic, hba, tagged := identityModule{uuid.New(), "SN-1", "", nicType}, identityModule{uuid.New(), "SN-1", "", hbaType}, identityModule{uuid.New(), "", "AT-9", nicType}
	cases := []struct {
		name    string
		module  *devicetypes.CaniModuleType
		nicType uuid.UUID
		wantID  uuid.UUID
		calls   int
	}{
		{"asset tag", &devicetypes.CaniModuleType{AssetTag: "AT-9", Serial: "other"}, nicType, tagged.id, 1},
		{"serial on type", &devicetypes.CaniModuleType{Model: "NIC", Serial: "SN-1"}, nicType, nic.id, 1},
		{"type Nautobot lacks", &devicetypes.CaniModuleType{Model: "NIC", Serial: "SN-1"}, uuid.Nil, uuid.Nil, 0},
		{"no identity", &devicetypes.CaniModuleType{Model: "NIC"}, nicType, uuid.Nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			e, cleanup := newExporterWithServer(t, identityListServer(tc.nicType, []identityModule{nic, hba, tagged}, &calls))
			defer cleanup()

			remote, err := e.findSpareByIdentity(context.Background(), tc.module)
			if err != nil {
				t.Fatalf("findSpareByIdentity: %v", err)
			}
			gotID := uuid.Nil
			if remote != nil {
				gotID = toUUID(remote.Id)
			}
			if gotID != tc.wantID || moduleListCalls(calls) != tc.calls {
				t.Errorf("remote = %v, module list calls = %d (%v); want %v and %d", gotID, moduleListCalls(calls), calls, tc.wantID, tc.calls)
			}
		})
	}
}

// TestFindSpareByIdentityRefusesSeveralMatches verifies a serial shared by
// several Nautobot modules of unknown type, or an asset tag Nautobot holds
// twice, is an error naming the identity rather than the first result.
//
// Why it matters: adopting found[0] moved whichever module Nautobot listed
// first, so a swap could relocate a stranger's card; the operator must
// resolve the duplicate instead.
// Inputs: a typeless spare with serial SN-1 against a NIC and an HBA that
// share it; a spare with asset tag AT-9 against two modules carrying it.
// Outputs: an error naming the serial, then one naming the asset tag.
// Data choice: the typeless spare is the shape import produces, where no
// type filter can narrow the serial.
func TestFindSpareByIdentityRefusesSeveralMatches(t *testing.T) {
	nicType, hbaType := uuid.New(), uuid.New()
	modules := []identityModule{{uuid.New(), "SN-1", "AT-9", nicType}, {uuid.New(), "SN-1", "AT-9", hbaType}}
	var calls []string
	e, cleanup := newExporterWithServer(t, identityListServer(nicType, modules, &calls))
	defer cleanup()

	_, err := e.findSpareByIdentity(context.Background(), &devicetypes.CaniModuleType{Serial: "SN-1"})
	if err == nil || !strings.Contains(err.Error(), `2 Nautobot modules share serial "SN-1"`) {
		t.Errorf("shared serial: error = %v, want a conflict naming the serial", err)
	}
	_, err = e.findSpareByIdentity(context.Background(), &devicetypes.CaniModuleType{Model: "NIC", AssetTag: "AT-9"})
	if err == nil || !strings.Contains(err.Error(), `2 Nautobot modules share asset tag "AT-9"`) {
		t.Errorf("shared asset tag: error = %v, want a conflict naming the asset tag", err)
	}
}
