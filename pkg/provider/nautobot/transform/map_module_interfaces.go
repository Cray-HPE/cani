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
package transform

import (
	"fmt"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// AttachModuleInterfaces appends each module-owned interface to its module's
// specs. The device grouping skips every interface with a module reference,
// including the ports of an installed module on which Nautobot 3.2 also sets
// the root device, so without this pass they would be lost. An interface
// whose module was not imported is left out and reported, in the same form as
// the other records the transform cannot place.
func AttachModuleInterfaces(modules map[uuid.UUID]*devicetypes.CaniModuleType, raw []nautobotapi.Interface) []string {
	index := nautobotIDIndex(modules, func(m *devicetypes.CaniModuleType) map[string]uuid.UUID { return m.ExternalIDs })
	var warnings []string
	for _, iface := range raw {
		moduleNBID := refID(iface.Module)
		if moduleNBID == uuid.Nil {
			continue
		}
		module := modules[index[moduleNBID]]
		if module == nil {
			warnings = append(warnings, fmt.Sprintf("interface %s: module %s not imported", iface.Name, moduleNBID))
			continue
		}
		module.Interfaces = append(module.Interfaces, interfaceSpecFromRaw(iface))
	}
	return warnings
}
