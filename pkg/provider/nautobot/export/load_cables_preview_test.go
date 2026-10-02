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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestPreviewPlannedCableTreatsPlannedModuleEndAsPlanned verifies a dry run
// previews a cable whose port belongs to a module the run only planned, and
// never filters an interface list on that module's persisted external ID;
// a module the run neither planned nor found leaves the preview to the
// create path.
//
// Why it matters: a module deleted in Nautobot keeps its external ID in the
// datastore; Nautobot answers an interface filter on that ID with 400, so the
// dry run reported cable errors the real run, which re-creates the module
// first, never produces.
// Inputs: DryRun on, a server that fails any request filtering on a module,
// a cable from a planned module's port to a device port; then the same cable
// with the module unplanned. Outputs: previewed once with CablesCreated 1 and
// no module-filtered request; not previewed the second time.
// Data choice: the stale external ID is set on the module so the test fails
// if the preview ever reads it instead of the run's own record.
func TestPreviewPlannedCableTreatsPlannedModuleEndAsPlanned(t *testing.T) {
	var requests []string
	e, cleanup := newExporterWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path+"?"+r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.RawQuery, "module=") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"module":["Select a valid choice."]}`)
			return
		}
		_, _ = io.WriteString(w, emptyListJSON)
	})
	defer cleanup()
	e.Options.DryRun = true
	device := &devicetypes.CaniDeviceType{Name: "node-01"}
	module := &devicetypes.CaniModuleType{ID: uuid.New(), Name: "nic-1",
		ObjectMeta: devicetypes.ObjectMeta{ExternalIDs: map[string]uuid.UUID{externalIDKeyNautobot: uuid.New()}}}
	cable := &devicetypes.CaniCableType{Label: "nic-to-node"}
	a, b := cableEnd{device: device, module: module, port: "p1"}, cableEnd{device: device, port: "eth0"}
	e.resolveEndModules(&a, &b)

	e.Cache.PlanModule(module.ID)
	result := &LoadResult{}
	if !e.previewPlannedCable(cable, a, b, result) || result.CablesCreated != 1 {
		t.Fatalf("planned module end: previewed = false or CablesCreated = %d, want a preview counted once", result.CablesCreated)
	}
	for _, req := range requests {
		if strings.Contains(req, "module=") {
			t.Errorf("preview filtered on a module: %s", req)
		}
	}

	e.Cache = NewLookupCache(e.Client)
	if e.previewPlannedCable(cable, a, b, &LoadResult{}) {
		t.Error("a module the run neither planned nor found must not be previewed as planned")
	}
}
