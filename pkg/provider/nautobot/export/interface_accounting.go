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

import "github.com/Cray-HPE/cani/pkg/devicetypes"

// noteExistingModuleInterface accounts for a module port that already exists
// on the parent device: a port the device spec itself declares is a
// duplicate-name skip (the device spec takes precedence), anything else was
// created by an earlier run.
func noteExistingModuleInterface(result *LoadResult, module *devicetypes.CaniModuleType,
	parent *devicetypes.CaniDeviceType, name string) {
	if deviceDeclaresInterface(parent, name) {
		clog.Skipped("Skipped interface %s on module %s: device %s already declares this port (device spec takes precedence)",
			name, module.Name, parent.Name)
		result.IfacesSkippedDuplicate++
		return
	}
	result.IfacesSkipped++
}

// deviceDeclaresInterface reports whether the device's own specs name the port.
func deviceDeclaresInterface(device *devicetypes.CaniDeviceType, name string) bool {
	if device == nil {
		return false
	}
	for _, iface := range device.Interfaces {
		if iface.Name == name {
			return true
		}
	}
	return false
}

// lossyExportCount totals the objects that did not reach Nautobot as
// authored without being errors.
func lossyExportCount(result *LoadResult) int {
	return result.IfacesSkippedDuplicate + result.IfacesUnsupported +
		result.IfacesUnresolvedRefs + result.CablesConflicted
}

// printInterfaceAccounting adds the duplicate-name skip line to the sync
// summary; the unsupported-type line is printed with the interface counts.
func printInterfaceAccounting(result *LoadResult) {
	if result.IfacesSkippedDuplicate > 0 {
		clog.Skipped("Skipped interfaces (name already on device): %d", result.IfacesSkippedDuplicate)
	}
}

// printLossyExportWarning closes the summary with an explicit warning when
// anything was skipped, so exit status 0 cannot hide it.
func printLossyExportWarning(result *LoadResult) {
	if count := lossyExportCount(result); count > 0 {
		clog.Warn("%d object(s) were not exported as authored; review the skipped lines above", count)
	}
}
