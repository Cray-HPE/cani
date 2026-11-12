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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// reconcileCustomFieldDrift
// -----------------------------------------------------------------------------

func TestReconcileCustomFieldDrift_PatchesWhenDrifted(t *testing.T) {
	cfID := uuid.New()
	var patchCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "extras/custom-fields") {
			patchCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.Merge = true

	remote := nautobotapi.CustomField{
		Label:        "Old Label",
		ContentTypes: []string{"dcim.device"},
	}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{
		Key:          "my_field",
		Label:        "New Label",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
	}

	err := e.reconcileCustomFieldDrift(context.Background(), cfID, remote, local)
	if err != nil {
		t.Fatalf("reconcileCustomFieldDrift() error = %v", err)
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH call, got %d", patchCalls)
	}
}

func TestReconcileCustomFieldDrift_SkipsWhenIdentical(t *testing.T) {
	cfID := uuid.New()
	var patchCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalls++
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.Merge = true

	remote := nautobotapi.CustomField{
		Label:        "My Field",
		ContentTypes: []string{"dcim.device"},
	}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{
		Key:          "my_field",
		Label:        "My Field",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
	}

	err := e.reconcileCustomFieldDrift(context.Background(), cfID, remote, local)
	if err != nil {
		t.Fatalf("reconcileCustomFieldDrift() error = %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("expected 0 PATCH calls for identical fields, got %d", patchCalls)
	}
}

func TestReconcileCustomFieldDrift_DryRunDoesNotPatch(t *testing.T) {
	cfID := uuid.New()
	var patchCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalls++
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.Merge = true
	e.Options.DryRun = true

	remote := nautobotapi.CustomField{
		Label:        "Old",
		ContentTypes: []string{"dcim.device"},
	}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{
		Key:          "my_field",
		Label:        "New",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
	}

	err := e.reconcileCustomFieldDrift(context.Background(), cfID, remote, local)
	if err != nil {
		t.Fatalf("reconcileCustomFieldDrift() error = %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("dry-run should not PATCH, got %d calls", patchCalls)
	}
}

func TestReconcileCustomFieldDrift_DetectsContentTypeDrift(t *testing.T) {
	cfID := uuid.New()
	var patchCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.Merge = true

	remote := nautobotapi.CustomField{
		Label:        "Field",
		ContentTypes: []string{"dcim.device"},
	}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{
		Key:          "field",
		Label:        "Field",
		Type:         "text",
		ContentTypes: []string{"dcim.device", "dcim.rack"},
	}

	err := e.reconcileCustomFieldDrift(context.Background(), cfID, remote, local)
	if err != nil {
		t.Fatalf("reconcileCustomFieldDrift() error = %v", err)
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH for content type drift, got %d", patchCalls)
	}
}

func TestReconcileCustomFieldDrift_DetectsWeightDrift(t *testing.T) {
	cfID := uuid.New()
	var patchCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalls++
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.Merge = true

	remoteWeight := 100
	remote := nautobotapi.CustomField{
		Label:        "Priority",
		ContentTypes: []string{"dcim.device"},
		Weight:       &remoteWeight,
	}
	setNBValue(&remote.Type, "integer")
	local := devicetypes.CustomFieldDefinition{
		Key:          "priority",
		Label:        "Priority",
		Type:         "integer",
		ContentTypes: []string{"dcim.device"},
		Weight:       200,
	}

	err := e.reconcileCustomFieldDrift(context.Background(), cfID, remote, local)
	if err != nil {
		t.Fatalf("reconcileCustomFieldDrift() error = %v", err)
	}
	if patchCalls != 1 {
		t.Errorf("expected 1 PATCH for weight drift, got %d", patchCalls)
	}
}

// -----------------------------------------------------------------------------
// buildCustomFieldPatch / mergedContentTypes — Nautobot's immutability rules
// -----------------------------------------------------------------------------

// TestBuildCustomFieldPatch_NeverSendsType verifies that a type mismatch
// between the catalog and Nautobot produces no PATCH field and, on its own,
// no drift.
//
// Why it matters: Nautobot rejects "Type cannot be changed once created", so
// a PATCH carrying type fails the whole definition reconciliation; the
// mismatch is reported instead and the operator decides.
// Inputs: a remote text field and a local integer definition, otherwise
// identical.
// Outputs: drifted == false and req.Type == nil.
// Data choice: every other attribute matches so the type is the only
// candidate for drift.
func TestBuildCustomFieldPatch_NeverSendsType(t *testing.T) {
	remote := nautobotapi.CustomField{Label: "Slots", ContentTypes: []string{"dcim.device"}}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{Key: "slots", Label: "Slots", Type: "integer", ContentTypes: []string{"dcim.device"}}

	req, drifted := buildCustomFieldPatch(remote, local)

	if drifted {
		t.Error("drifted = true, want false: a type difference must not be reconciled")
	}
	if req.Type != nil {
		t.Errorf("req.Type = %v, want nil", *req.Type)
	}
}

// TestBuildCustomFieldPatch_ContentTypesOnlyGrow verifies that a content type
// Nautobot has and the catalog lacks is kept, and that one the catalog adds
// is sent as the union.
//
// Why it matters: removing a content type from a definition makes Nautobot
// delete the field's values on every object of that type; the export must
// never shrink the list under --merge.
// Inputs: remote {device, rack} with local {device} (nothing to add); then
// remote {device} with local {rack}.
// Outputs: no drift in the first case; ContentTypes {device, rack} in the
// second.
// Data choice: the two directions of asymmetry, each in isolation.
func TestBuildCustomFieldPatch_ContentTypesOnlyGrow(t *testing.T) {
	remote := nautobotapi.CustomField{Label: "Owner", ContentTypes: []string{"dcim.device", "dcim.rack"}}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{Key: "owner", Label: "Owner", Type: "text", ContentTypes: []string{"dcim.device"}}

	if _, drifted := buildCustomFieldPatch(remote, local); drifted {
		t.Error("drifted = true for a remote-only content type, want false (never shrink)")
	}

	remote.ContentTypes = []string{"dcim.device"}
	local.ContentTypes = []string{"dcim.rack"}
	req, drifted := buildCustomFieldPatch(remote, local)
	if !drifted || req.ContentTypes == nil {
		t.Fatalf("drifted = %v, ContentTypes = %v; want drift with the union", drifted, req.ContentTypes)
	}
	if got := *req.ContentTypes; len(got) != 2 || got[0] != "dcim.device" || got[1] != "dcim.rack" {
		t.Errorf("ContentTypes = %v, want [dcim.device dcim.rack]", got)
	}
}

// TestReconcileCustomFieldDrift_TypeMismatchAloneDoesNotPatch verifies the
// HTTP path: a definition that differs only in type issues no PATCH.
//
// Why it matters: the reconciliation runs under --merge before any object is
// written; a rejected PATCH here would abort the export for a change
// Nautobot can never accept.
// Inputs: an exporter against a fake server counting PATCHes, remote type
// text, local type integer.
// Outputs: nil error and zero PATCH calls.
// Data choice: the fake answers every other request with an empty list so
// only a PATCH could register.
func TestReconcileCustomFieldDrift_TypeMismatchAloneDoesNotPatch(t *testing.T) {
	var patchCalls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPatch {
			patchCalls++
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptyListJSON)
	}
	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.Merge = true

	remote := nautobotapi.CustomField{Label: "Slots", ContentTypes: []string{"dcim.device"}}
	setNBValue(&remote.Type, "text")
	local := devicetypes.CustomFieldDefinition{Key: "slots", Label: "Slots", Type: "integer", ContentTypes: []string{"dcim.device"}}

	if err := e.reconcileCustomFieldDrift(context.Background(), uuid.New(), remote, local); err != nil {
		t.Fatalf("reconcileCustomFieldDrift() error = %v", err)
	}
	if patchCalls != 0 {
		t.Errorf("PATCH calls = %d, want 0", patchCalls)
	}
}
