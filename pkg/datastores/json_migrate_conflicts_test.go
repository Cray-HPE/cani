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
package datastores

import (
	"bytes"
	"encoding/json"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

// TestLoadLogsMigrationConflictsOnce verifies a migration that leaves a
// conflict logs it when the datastore is migrated, and a later load does not.
//
// Why it matters: loads are silent, so the migration is the one moment a
// read-only command can tell the operator what it could not fold.
// Inputs: a v1alpha6 datastore with two Global VRFs named red and a prefix
// whose legacy VRF name is red, loaded twice.
// Outputs: the first load logs the unresolved legacy name once; the second,
// after the migration was saved, does not.
// Data choice: two VRFs with exactly that name keep the legacy name ambiguous.
func TestLoadLogsMigrationConflictsOnce(t *testing.T) {
	// Arrange.
	path := filepath.Join(t.TempDir(), "inventory.json")
	inventory := devicetypes.NewInventory()
	inventory.SchemaVersion = devicetypes.SchemaVersionV1Alpha6
	redOne, redTwo, prefix := uuid.New(), uuid.New(), uuid.New()
	inventory.VRFs[redOne] = &devicetypes.CaniVRF{ID: redOne, Name: "red"}
	inventory.VRFs[redTwo] = &devicetypes.CaniVRF{ID: redTwo, Name: "red"}
	inventory.Prefixes[prefix] = &devicetypes.CaniPrefix{ID: prefix, Prefix: "10.0.0.0/24", VRF: "red"}
	data, err := json.Marshal(inventory)
	if err != nil {
		t.Fatalf("marshalling v1alpha6 inventory: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("writing v1alpha6 inventory: %v", err)
	}
	var out bytes.Buffer
	writer := stdlog.Writer()
	stdlog.SetOutput(&out)
	t.Cleanup(func() { stdlog.SetOutput(writer) })

	// Act.
	_, firstErr := (&JSONStore{Path: path}).Load()
	first := out.String()
	out.Reset()
	_, secondErr := (&JSONStore{Path: path}).Load()
	second := out.String()

	// Assert.
	if firstErr != nil || secondErr != nil {
		t.Fatalf("Load errors = %v, %v", firstErr, secondErr)
	}
	want := `legacy VRF name "red" does not resolve`
	if got := strings.Count(first, want); got != 1 {
		t.Errorf("migrating load logged the legacy name %d times, want once; log:\n%s", got, first)
	}
	if strings.Contains(second, want) {
		t.Errorf("load after the migration logged it again; log:\n%s", second)
	}
}
