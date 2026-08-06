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
package devicetypes

import (
	"embed"
	"io/fs"
	"strings"
	"testing"
)

// TestEmbeddedLibraryHasNoProviderDefaults verifies that no embedded hardware
// type YAML carries a provider_defaults key.
//
// Why it matters: the embedded library is the portable, Nautobot/NetBox-shaped
// source of truth shared by every provider. Provider configuration (CSM cabinet
// ordinals, HMN VLAN ranges) placed here leaks one provider's policy into data
// that all providers consume, and lets a provider hook fire on an inventory it
// does not own.
//
// Inputs: every YAML file in the embedded device-, module-, cable-, rack- and
// location-type filesystems. Output: a failure naming each offending file.
//
// Data choice: a raw substring scan rather than a decode, because the guard
// must catch the key regardless of which type it is attached to or whether the
// corresponding Go field still exists.
func TestEmbeddedLibraryHasNoProviderDefaults(t *testing.T) {
	filesystems := map[string]embed.FS{
		"device-types":   embeddedDeviceTypes,
		"module-types":   embeddedModuleTypes,
		"cable-types":    embeddedCableTypes,
		"rack-types":     embeddedRackTypes,
		"location-types": embeddedLocationTypes,
	}

	for name, fsys := range filesystems {
		err := fs.WalkDir(fsys, name, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
				return nil
			}
			data, readErr := fsys.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if strings.Contains(string(data), "provider_defaults") {
				t.Errorf("%s contains provider_defaults; provider config belongs in the provider package", path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", name, err)
		}
	}
}
