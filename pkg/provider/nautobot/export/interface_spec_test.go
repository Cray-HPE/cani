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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestSpecFromInterfaceCarriesEveryAttribute verifies specFromInterface copies
// label, MAC, mgmt_only, tags, description and the switchport fields from a
// persisted spec and derives type, speed and role.
//
// Why it matters: device and module ports are built through this one function
// so module-originated interfaces reach Nautobot with the same attribute set
// as device-level ones.
// Inputs: an InterfaceSpec with every exportable field set and no explicit
// role. Outputs: an interfaceSpec with each value copied and role inferred.
// Data choice: a management-only 1G port with a MAC and label is the typical
// BMC adapter shape and exercises the inferred-role branch.
func TestSpecFromInterfaceCarriesEveryAttribute(t *testing.T) {
	mgmt := true
	spec := specFromInterface(devicetypes.InterfaceSpec{
		Name: "ocp1-p1", Type: devicetypes.InterfacesElemTypeA1000BaseT, Label: "BMC",
		MacAddress: "aa:bb:cc:dd:ee:01", MgmtOnly: &mgmt, Tags: []string{"oob"},
		Lag: "bond0", Mode: "access", UntaggedVLAN: 10, TaggedVLANs: []int{20}, VRF: "mgmt",
		Description: "module port",
	})

	if spec.Label != "BMC" || spec.Mac != "aa:bb:cc:dd:ee:01" || !spec.MgmtOnly || spec.Description != "module port" {
		t.Fatalf("scalar attributes not carried: %+v", spec)
	}
	if len(spec.Tags) != 1 || spec.Lag != "bond0" || spec.Mode != "access" || spec.UntaggedVLAN != 10 || len(spec.TaggedVLANs) != 1 || spec.VRF != "mgmt" {
		t.Fatalf("switchport attributes not carried: %+v", spec)
	}
	if spec.Type != "1000base-t" || spec.Speed != 1000000 || spec.Role == "" {
		t.Fatalf("derived fields wrong: type=%q speed=%d role=%q", spec.Type, spec.Speed, spec.Role)
	}
}

// capturingInterfaceServer records the JSON body of the interface-create POST
// and answers every other request with an empty list.
func capturingInterfaceServer(body *map[string]any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "dcim/interfaces") {
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, body)
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprintf(w, `{"id":%q,"name":"Port 1","display":"Port 1"}`, uuid.NewString())
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	}
}

// TestCreateModuleInterfacesSendsPortAttributes verifies the interface POST
// issued for a module port carries mac_address, label, description and
// mgmt_only from the module's spec.
//
// Why it matters: before this, module ports were created with name, type and
// role only, so a MAC recorded on a NIC port never reached Nautobot.
// Inputs: a module with one 25G port carrying all four attributes and a server
// that captures the POST body. Outputs: the body holds each value.
// Data choice: the 2-port NIC "Port 1" case mirrors the live reproduction.
func TestCreateModuleInterfacesSendsPortAttributes(t *testing.T) {
	var body map[string]any
	e, cleanup := newExporterWithServer(t, capturingInterfaceServer(&body))
	defer cleanup()
	seedActiveStatus(t, e)
	e.Cache.roles["management"] = &CachedItem{ID: uuid.New(), Name: "management"}
	mgmt := true
	module := &devicetypes.CaniModuleType{Name: "nic-a", Interfaces: []devicetypes.InterfaceSpec{{
		Name: "Port 1", Type: devicetypes.InterfacesElemTypeA25GbaseXSfp28, Label: "Fabric A",
		MacAddress: "aa:bb:cc:dd:ee:01", MgmtOnly: &mgmt, Description: "module port",
	}}}

	if err := e.createModuleInterfaces(context.Background(), module, &devicetypes.CaniDeviceType{Name: "node"}, uuid.New(), &LoadResult{}); err != nil {
		t.Fatalf("createModuleInterfaces() error = %v", err)
	}

	if body["mac_address"] != "aa:bb:cc:dd:ee:01" || body["label"] != "Fabric A" || body["description"] != "module port" {
		t.Fatalf("POST body missing attributes: %v", body)
	}
	if body["mgmt_only"] != true {
		t.Fatalf("mgmt_only = %v, want true", body["mgmt_only"])
	}
}

// TestUpdateInterfaceClearsLabel verifies the interface PATCH always carries
// label so an emptied local label clears the remote one.
//
// Why it matters: the inventory is authoritative on reconcile; omitting the
// field would leave a stale label in Nautobot after it was removed locally.
// Inputs: an interfaceSpec with an empty label and a server capturing the
// PATCH body. Outputs: the body contains "label" set to "".
// Data choice: an empty label is the only case where omit-when-empty and
// send-always differ on the wire.
func TestUpdateInterfaceClearsLabel(t *testing.T) {
	var body map[string]any
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":0,"results":[]}`))
	})
	defer cleanup()
	seedActiveStatus(t, e)

	err := e.updateInterface(context.Background(), uuid.New(), uuid.New(), interfaceSpec{Name: "eth0", Type: "1000base-t"}, &LoadResult{})
	if err != nil {
		t.Fatalf("updateInterface() error = %v", err)
	}
	label, present := body["label"]
	if !present || label != "" {
		t.Fatalf("PATCH body label = %v (present=%v), want an explicit empty string", label, present)
	}
}
