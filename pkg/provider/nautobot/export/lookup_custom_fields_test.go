/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

// sentCustomField is the minimal view of the WritableCustomFieldRequest body
// POSTed to /api/extras/custom-fields/.
type sentCustomField struct {
	Label        string   `json:"label"`
	Key          string   `json:"key"`
	Type         string   `json:"type"`
	ContentTypes []string `json:"content_types"`
	Description  string   `json:"description"`
	Required     bool     `json:"required"`
	Weight       int      `json:"weight"`
}

// sentCustomFieldChoice is the minimal view of the CustomFieldChoiceRequest
// body POSTed to /api/extras/custom-field-choices/.
type sentCustomFieldChoice struct {
	CustomField struct {
		ID string `json:"id"`
	} `json:"custom_field"`
	Value  string `json:"value"`
	Weight int    `json:"weight"`
}

// TestEnsureCustomFields_CreatesFieldAndChoices verifies the full
// EnsureCustomFields flow: a text field is created (no choices), and a select
// field with choices triggers both a field create and choice creates.
func TestEnsureCustomFields_CreatesFieldAndChoices(t *testing.T) {
	textFieldID := uuid.New()
	selectFieldID := uuid.New()

	var fieldPosts [][]byte
	var choicePosts [][]byte

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// List custom fields — always return empty (field does not exist yet)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-fields"):
			fmt.Fprint(w, `{"count":0,"results":[]}`)

		// Create custom field
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/custom-fields"):
			body, _ := io.ReadAll(r.Body)
			fieldPosts = append(fieldPosts, body)

			// Return the appropriate ID based on order
			var id uuid.UUID
			if len(fieldPosts) == 1 {
				id = textFieldID
			} else {
				id = selectFieldID
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id":"%s","label":"test","content_types":[],"type":{"value":"text","label":"Text"}}`, id)

		// List custom field choices — return empty
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-field-choices"):
			fmt.Fprint(w, `{"count":0,"results":[]}`)

		// Create custom field choice
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/custom-field-choices"):
			body, _ := io.ReadAll(r.Body)
			choicePosts = append(choicePosts, body)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id":"`+uuid.New().String()+`","value":"x","custom_field":{"id":"`+selectFieldID.String()+`"}}`)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "xname",
		Label:        "XName",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
		Description:  "Hardware location name",
	})
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "tier",
		Label:        "Tier",
		Type:         "select",
		ContentTypes: []string{"dcim.location"},
		Choices:      []string{"edge", "core", "aggregation"},
	})

	err := e.EnsureCustomFields(context.Background(), inv)
	if err != nil {
		t.Fatalf("EnsureCustomFields() error: %v", err)
	}

	// Verify two field creates were issued
	if len(fieldPosts) != 2 {
		t.Fatalf("expected 2 field POSTs, got %d", len(fieldPosts))
	}

	// Verify first field payload
	var f1 sentCustomField
	if err := json.Unmarshal(fieldPosts[0], &f1); err != nil {
		t.Fatalf("unmarshal field[0]: %v", err)
	}
	if f1.Label != "XName" || f1.Key != "xname" || f1.Type != "text" {
		t.Errorf("field[0] = %+v, want label=XName key=xname type=text", f1)
	}
	if f1.Description != "Hardware location name" {
		t.Errorf("field[0].Description = %q, want %q", f1.Description, "Hardware location name")
	}

	// Verify second field payload
	var f2 sentCustomField
	if err := json.Unmarshal(fieldPosts[1], &f2); err != nil {
		t.Fatalf("unmarshal field[1]: %v", err)
	}
	if f2.Label != "Tier" || f2.Key != "tier" || f2.Type != "select" {
		t.Errorf("field[1] = %+v, want label=Tier key=tier type=select", f2)
	}

	// Verify 3 choice creates for the select field
	if len(choicePosts) != 3 {
		t.Fatalf("expected 3 choice POSTs, got %d", len(choicePosts))
	}
	expectedChoices := []string{"edge", "core", "aggregation"}
	for i, raw := range choicePosts {
		var c sentCustomFieldChoice
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("unmarshal choice[%d]: %v", i, err)
		}
		if c.Value != expectedChoices[i] {
			t.Errorf("choice[%d].Value = %q, want %q", i, c.Value, expectedChoices[i])
		}
		if c.CustomField.ID != selectFieldID.String() {
			t.Errorf("choice[%d].CustomField.ID = %q, want %q", i, c.CustomField.ID, selectFieldID)
		}
	}
}

// TestEnsureCustomFields_SkipsExistingField verifies that a custom field
// already present in Nautobot is not re-created.
func TestEnsureCustomFields_SkipsExistingField(t *testing.T) {
	existingID := uuid.New()
	var createCalls int

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-fields"):
			// Return one existing result
			fmt.Fprintf(w, `{"count":1,"results":[{"id":"%s","label":"XName","content_types":["dcim.device"],"type":{"value":"text","label":"Text"}}]}`, existingID)

		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/custom-fields"):
			createCalls++
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id":"`+uuid.New().String()+`","label":"x","content_types":[],"type":{"value":"text","label":"Text"}}`)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "xname",
		Label:        "XName",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
	})

	err := e.EnsureCustomFields(context.Background(), inv)
	if err != nil {
		t.Fatalf("EnsureCustomFields() error: %v", err)
	}
	if createCalls != 0 {
		t.Errorf("expected 0 create calls for existing field, got %d", createCalls)
	}
}

// TestEnsureCustomFields_SkipsExistingChoices verifies that choices already in
// Nautobot are not re-created while missing ones are.
func TestEnsureCustomFields_SkipsExistingChoices(t *testing.T) {
	fieldID := uuid.New()
	var createdChoices []string

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-fields"):
			fmt.Fprintf(w, `{"count":1,"results":[{"id":"%s","label":"Tier","content_types":["dcim.location"],"type":{"value":"select","label":"Selection"}}]}`, fieldID)

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-field-choices"):
			// "edge" already exists
			fmt.Fprintf(w, `{"count":1,"results":[{"id":"%s","value":"edge","custom_field":{"id":"%s"}}]}`, uuid.New(), fieldID)

		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/custom-field-choices"):
			body, _ := io.ReadAll(r.Body)
			var c sentCustomFieldChoice
			_ = json.Unmarshal(body, &c)
			createdChoices = append(createdChoices, c.Value)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id":"%s","value":"%s","custom_field":{"id":"%s"}}`, uuid.New(), c.Value, fieldID)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "tier",
		Label:        "Tier",
		Type:         "select",
		ContentTypes: []string{"dcim.location"},
		Choices:      []string{"edge", "core", "aggregation"},
	})

	err := e.EnsureCustomFields(context.Background(), inv)
	if err != nil {
		t.Fatalf("EnsureCustomFields() error: %v", err)
	}

	// Only "core" and "aggregation" should be created; "edge" was skipped
	if len(createdChoices) != 2 {
		t.Fatalf("expected 2 choice creates, got %d: %v", len(createdChoices), createdChoices)
	}
	if createdChoices[0] != "core" || createdChoices[1] != "aggregation" {
		t.Errorf("created choices = %v, want [core aggregation]", createdChoices)
	}
}

// TestEnsureCustomFields_EmptyDefinitions verifies a no-op when the inventory
// has no custom field definitions.
func TestEnsureCustomFields_EmptyDefinitions(t *testing.T) {
	var calls int
	handler := func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	err := e.EnsureCustomFields(context.Background(), inv)
	if err != nil {
		t.Fatalf("EnsureCustomFields() error: %v", err)
	}
	if calls != 0 {
		t.Errorf("expected 0 HTTP calls for empty definitions, got %d", calls)
	}
}

// TestEnsureCustomFields_CreateFieldError verifies the error is propagated
// when the field create API returns a non-201 status.
func TestEnsureCustomFields_CreateFieldError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-fields"):
			fmt.Fprint(w, `{"count":0,"results":[]}`)
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/custom-fields"):
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"detail":"server error"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "xname",
		Label:        "XName",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
	})

	err := e.EnsureCustomFields(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error when field create fails")
	}
	if !strings.Contains(err.Error(), "custom field \"xname\"") {
		t.Errorf("error = %q, want it to mention the field key", err)
	}
}

// TestEnsureCustomFields_CreateChoiceError verifies the error is propagated
// when the choice create API returns a non-201 status.
func TestEnsureCustomFields_CreateChoiceError(t *testing.T) {
	fieldID := uuid.New()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-fields"):
			fmt.Fprintf(w, `{"count":1,"results":[{"id":"%s","label":"Tier","content_types":["dcim.location"],"type":{"value":"select","label":"Selection"}}]}`, fieldID)
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-field-choices"):
			fmt.Fprint(w, `{"count":0,"results":[]}`)
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "extras/custom-field-choices"):
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"detail":"bad request"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()

	inv := devicetypes.NewInventory()
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "tier",
		Label:        "Tier",
		Type:         "select",
		ContentTypes: []string{"dcim.location"},
		Choices:      []string{"edge"},
	})

	err := e.EnsureCustomFields(context.Background(), inv)
	if err == nil {
		t.Fatal("expected error when choice create fails")
	}
	if !strings.Contains(err.Error(), "choices") {
		t.Errorf("error = %q, want it to mention choices", err)
	}
}
