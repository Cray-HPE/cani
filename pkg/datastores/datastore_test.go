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
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/spf13/cobra"
)

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

	root := &cobra.Command{}
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

func TestSetDeviceStoreUnsupported(t *testing.T) {
	defer func() { Datastore = nil }()

	root := &cobra.Command{}
	root.PersistentFlags().String("datastore", "unsupported", "datastore type")

	err := SetDeviceStore(root, nil)
	if err == nil {
		t.Error("SetDeviceStore() expected error for unsupported type, got nil")
	}
}
