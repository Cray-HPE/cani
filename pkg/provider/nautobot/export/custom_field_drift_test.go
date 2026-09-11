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
