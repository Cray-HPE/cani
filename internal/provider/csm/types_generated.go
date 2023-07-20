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
	"fmt"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

func BuildXname(cHardware inventory.Hardware, locationPath inventory.LocationPath) (xnames.Xname, error) {
	hsmType, err := GetHMSType(cHardware, locationPath)
	if err != nil {
		return nil, err
	}

	switch hsmType {
	case xnametypes.HMSTypeInvalid:
		return nil, nil
	case xnametypes.CDU:
		return xnames.CDU{
			CDU: locationPath[1].Ordinal,
		}, nil
	case xnametypes.CDUMgmtSwitch:
		return xnames.CDUMgmtSwitch{
			CDU:           locationPath[1].Ordinal,
			CDUMgmtSwitch: locationPath[2].Ordinal,
		}, nil
	case xnametypes.Cabinet:
		return xnames.Cabinet{
			Cabinet: locationPath[1].Ordinal,
		}, nil
	case xnametypes.CEC:
		return xnames.CEC{
			Cabinet: locationPath[1].Ordinal,
			CEC:     locationPath[2].Ordinal,
		}, nil
	case xnametypes.CabinetPDUController:
		return xnames.CabinetPDUController{
			Cabinet:              locationPath[1].Ordinal,
			CabinetPDUController: locationPath[2].Ordinal,
		}, nil
	case xnametypes.CabinetPDU:
		return xnames.CabinetPDU{
			Cabinet:              locationPath[1].Ordinal,
			CabinetPDUController: locationPath[2].Ordinal,
			CabinetPDU:           locationPath[3].Ordinal,
		}, nil
	case xnametypes.Chassis:
		return xnames.Chassis{
			Cabinet: locationPath[1].Ordinal,
			Chassis: locationPath[2].Ordinal,
		}, nil
	case xnametypes.ChassisBMC:
		return xnames.ChassisBMC{
			Cabinet:    locationPath[1].Ordinal,
			Chassis:    locationPath[2].Ordinal,
			ChassisBMC: locationPath[3].Ordinal,
		}, nil
	case xnametypes.ComputeModule:
		return xnames.ComputeModule{
			Cabinet:       locationPath[1].Ordinal,
			Chassis:       locationPath[2].Ordinal,
			ComputeModule: locationPath[3].Ordinal,
		}, nil
	case xnametypes.NodeBMC:
		return xnames.NodeBMC{
			Cabinet:       locationPath[1].Ordinal,
			Chassis:       locationPath[2].Ordinal,
			ComputeModule: locationPath[3].Ordinal,
			NodeBMC:       locationPath[4].Ordinal,
		}, nil
	case xnametypes.Node:
		return xnames.Node{
			Cabinet:       locationPath[1].Ordinal,
			Chassis:       locationPath[2].Ordinal,
			ComputeModule: locationPath[3].Ordinal,
			NodeBMC:       locationPath[4].Ordinal,
			Node:          locationPath[5].Ordinal,
		}, nil
	case xnametypes.NodeEnclosure:
		return xnames.NodeEnclosure{
			Cabinet:       locationPath[1].Ordinal,
			Chassis:       locationPath[2].Ordinal,
			ComputeModule: locationPath[3].Ordinal,
			NodeEnclosure: locationPath[4].Ordinal,
		}, nil
	case xnametypes.MgmtHLSwitchEnclosure:
		return xnames.MgmtHLSwitchEnclosure{
			Cabinet:               locationPath[1].Ordinal,
			Chassis:               locationPath[2].Ordinal,
			MgmtHLSwitchEnclosure: locationPath[3].Ordinal,
		}, nil
	case xnametypes.MgmtHLSwitch:
		return xnames.MgmtHLSwitch{
			Cabinet:               locationPath[1].Ordinal,
			Chassis:               locationPath[2].Ordinal,
			MgmtHLSwitchEnclosure: locationPath[3].Ordinal,
			MgmtHLSwitch:          locationPath[4].Ordinal,
		}, nil
	case xnametypes.MgmtSwitch:
		return xnames.MgmtSwitch{
			Cabinet:    locationPath[1].Ordinal,
			Chassis:    locationPath[2].Ordinal,
			MgmtSwitch: locationPath[3].Ordinal,
		}, nil
	case xnametypes.RouterModule:
		return xnames.RouterModule{
			Cabinet:      locationPath[1].Ordinal,
			Chassis:      locationPath[2].Ordinal,
			RouterModule: locationPath[3].Ordinal,
		}, nil
	case xnametypes.RouterBMC:
		return xnames.RouterBMC{
			Cabinet:      locationPath[1].Ordinal,
			Chassis:      locationPath[2].Ordinal,
			RouterModule: locationPath[3].Ordinal,
			RouterBMC:    locationPath[4].Ordinal,
		}, nil
	}
	return nil, fmt.Errorf("unknown xnametype '%s'", hsmType.String())
}

func BuildXnameT[T xnames.Xname](cHardware inventory.Hardware, locationPath inventory.LocationPath) (*T, error) {
	xnameRaw, err := BuildXname(cHardware, locationPath)
	if err != nil {
		return nil, nil
	}

	xname, ok := xnameRaw.(T)
	if !ok {
		return nil, nil
	}

	return &xname, nil
}
