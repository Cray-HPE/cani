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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func TestEnsureCustomFields_DryRunSkipsCreation(t *testing.T) {
	var postCalls int

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "extras/custom-fields"):
			fmt.Fprint(w, `{"count":0,"results":[]}`)
		case r.Method == http.MethodPost:
			postCalls++
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e, cleanup := newExporterWithServer(t, handler)
	defer cleanup()
	e.Options.DryRun = true

	inv := devicetypes.NewInventory()
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "xname",
		Label:        "XName",
		Type:         "text",
		ContentTypes: []string{"dcim.device"},
	})
	_ = inv.AddCustomField(devicetypes.CustomFieldDefinition{
		Key:          "tier",
		Label:        "Tier",
		Type:         "select",
		ContentTypes: []string{"dcim.location"},
		Choices:      []string{"edge", "core"},
	})

	err := e.EnsureCustomFields(context.Background(), inv)
	if err != nil {
		t.Fatalf("EnsureCustomFields(dry-run) error: %v", err)
	}
	if postCalls != 0 {
		t.Errorf("expected 0 POST calls during dry-run, got %d", postCalls)
	}
}
