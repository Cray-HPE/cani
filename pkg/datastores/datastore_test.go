/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

// | Function       | Happy-path test           | Failure test                  |
// |----------------|---------------------------|-------------------------------|
// | SetDeviceStore | TestSetDeviceStoreJSON     | TestSetDeviceStoreUnsupported |

import (
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
)

// TestSetDeviceStoreJSON verifies the datastore factory selects JSON storage.
//
// Why it matters: command startup depends on SetDeviceStore wiring the global
// datastore implementation before CRUD commands load or save inventory data.
// Inputs: a cli root command whose datastore flag is json. Outputs: nil error
// and a global Datastore containing a *JSONStore.
// Data choice: json is the only implemented datastore type in this package.
func TestSetDeviceStoreJSON(t *testing.T) {
	original := config.Cfg
	config.Cfg = &config.Config{
		Path:      "/tmp/cani-test/config.yaml",
		Datastore: "inventory.json",
	}
	defer func() {
		config.Cfg = original
		Datastore = nil
	}()

	root := &cli.Command{}
	root.PersistentFlags().String("datastore", "json", "datastore type")

	if err := SetDeviceStore(root, nil); err != nil {
		t.Fatalf("SetDeviceStore() returned unexpected error: %v", err)
	}

	if Datastore == nil {
		t.Fatal("SetDeviceStore() did not set global Datastore")
	}

	if _, ok := Datastore.(*JSONStore); !ok {
		t.Errorf("expected Datastore to be *JSONStore, got %T", Datastore)
	}
}

// TestSetDeviceStoreUnsupported verifies unsupported datastore names fail.
//
// Why it matters: accepting an unknown datastore type would leave command CRUD
// operations without a reliable persistence backend.
// Inputs: a cli root command whose datastore flag is unsupported. Outputs: an
// error that names the unsupported type and no JSON datastore selection.
// Data choice: "unsupported" is outside the StoreType constants and exercises
// the default branch.
func TestSetDeviceStoreUnsupported(t *testing.T) {
	Datastore = nil
	defer func() { Datastore = nil }()

	root := &cli.Command{}
	root.PersistentFlags().String("datastore", "unsupported", "datastore type")

	err := SetDeviceStore(root, nil)
	if err == nil {
		t.Fatal("SetDeviceStore() expected error for unsupported type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported datastore type: unsupported") {
		t.Fatalf("SetDeviceStore() error = %v, want unsupported datastore context", err)
	}
	if Datastore != nil {
		t.Fatalf("Datastore = %T, want nil after unsupported type", Datastore)
	}
}
