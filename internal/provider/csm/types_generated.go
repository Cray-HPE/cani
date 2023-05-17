// MIT License
//
// (C) Copyright [2023] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

// WARNING GENERATED FILE DO NOT EDIT!

package csm

import (
	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

func BuildXname(locationPath inventory.LocationPath) xnames.Xname {
	switch GetHMSType(locationPath) {
	case xnametypes.CDU:
		return xnames.CDU{
			CDU: locationPath[0].Ordinal,
		}
	case xnametypes.CDUMgmtSwitch:
		return xnames.CDUMgmtSwitch{
			CDU:           locationPath[0].Ordinal,
			CDUMgmtSwitch: locationPath[1].Ordinal,
		}
	case xnametypes.Cabinet:
		return xnames.Cabinet{
			Cabinet: locationPath[0].Ordinal,
		}
	case xnametypes.CEC:
		return xnames.CEC{
			Cabinet: locationPath[0].Ordinal,
			CEC:     locationPath[1].Ordinal,
		}
	case xnametypes.CabinetPDUController:
		return xnames.CabinetPDUController{
			Cabinet:              locationPath[0].Ordinal,
			CabinetPDUController: locationPath[1].Ordinal,
		}
	case xnametypes.CabinetPDU:
		return xnames.CabinetPDU{
			Cabinet:              locationPath[0].Ordinal,
			CabinetPDUController: locationPath[1].Ordinal,
			CabinetPDU:           locationPath[2].Ordinal,
		}
	case xnametypes.Chassis:
		return xnames.Chassis{
			Cabinet: locationPath[0].Ordinal,
			Chassis: locationPath[1].Ordinal,
		}
	case xnametypes.ChassisBMC:
		return xnames.ChassisBMC{
			Cabinet:    locationPath[0].Ordinal,
			Chassis:    locationPath[1].Ordinal,
			ChassisBMC: locationPath[2].Ordinal,
		}
	case xnametypes.ComputeModule:
		return xnames.ComputeModule{
			Cabinet:       locationPath[0].Ordinal,
			Chassis:       locationPath[1].Ordinal,
			ComputeModule: locationPath[2].Ordinal,
		}
	case xnametypes.NodeBMC:
		return xnames.NodeBMC{
			Cabinet:       locationPath[0].Ordinal,
			Chassis:       locationPath[1].Ordinal,
			ComputeModule: locationPath[2].Ordinal,
			NodeBMC:       locationPath[3].Ordinal,
		}
	case xnametypes.Node:
		return xnames.Node{
			Cabinet:       locationPath[0].Ordinal,
			Chassis:       locationPath[1].Ordinal,
			ComputeModule: locationPath[2].Ordinal,
			NodeBMC:       locationPath[3].Ordinal,
			Node:          locationPath[4].Ordinal,
		}
	case xnametypes.RouterModule:
		return xnames.RouterModule{
			Cabinet:      locationPath[0].Ordinal,
			Chassis:      locationPath[1].Ordinal,
			RouterModule: locationPath[2].Ordinal,
		}
	case xnametypes.RouterBMC:
		return xnames.RouterBMC{
			Cabinet:      locationPath[0].Ordinal,
			Chassis:      locationPath[1].Ordinal,
			RouterModule: locationPath[2].Ordinal,
			RouterBMC:    locationPath[3].Ordinal,
		}
	}
	return nil
}
