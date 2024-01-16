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
package canu

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Cray-HPE/cani/pkg/hardwaretypes"
	"github.com/netbox-community/go-netbox/v3"
)

type Paddle struct {
	Architecture string               `json:"architecture" yaml:"architecture" mapstructure:"architecture"`
	CanuVersion  string               `json:"canu_version" yaml:"canu_version" mapstructure:"canu_version"`
	ShcdFile     string               `json:"shcd_file" yaml:"shcd_file" mapstructure:"shcd_file"`
	Topology     []PaddleTopologyElem `json:"topology" yaml:"topology" mapstructure:"topology"`
	UpdatedAt    *string              `json:"updated_at,omitempty" yaml:"updated_at,omitempty" mapstructure:"updated_at,omitempty"`
}

type PaddleTopologyElem struct {
	Architecture  *string                       `json:"architecture,omitempty" yaml:"architecture,omitempty" mapstructure:"architecture,omitempty"`
	CommonName    *string                       `json:"common_name,omitempty" yaml:"common_name,omitempty" mapstructure:"common_name,omitempty"`
	Id            *int                          `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`
	Location      *PaddleTopologyElemLocation   `json:"location,omitempty" yaml:"location,omitempty" mapstructure:"location,omitempty"`
	Model         *string                       `json:"model,omitempty" yaml:"model,omitempty" mapstructure:"model,omitempty"`
	Ports         []PaddleTopologyElemPortsElem `json:"ports,omitempty" yaml:"ports,omitempty" mapstructure:"ports,omitempty"`
	RackElevation *string                       `json:"rack_elevation,omitempty" yaml:"rack_elevation,omitempty" mapstructure:"rack_elevation,omitempty"`
	RackNumber    *string                       `json:"rack_number,omitempty" yaml:"rack_number,omitempty" mapstructure:"rack_number,omitempty"`
	Type          *string                       `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type,omitempty"`
	Vendor        *string                       `json:"vendor,omitempty" yaml:"vendor,omitempty" mapstructure:"vendor,omitempty"`
}

type PaddleTopologyElemLocation struct {
	Elevation   *string `json:"elevation,omitempty" yaml:"elevation,omitempty" mapstructure:"elevation,omitempty"`
	Parent      *string `json:"parent,omitempty" yaml:"parent,omitempty" mapstructure:"parent,omitempty"`
	Rack        *string `json:"rack,omitempty" yaml:"rack,omitempty" mapstructure:"rack,omitempty"`
	SubLocation *string `json:"sub_location,omitempty" yaml:"sub_location,omitempty" mapstructure:"sub_location,omitempty"`
}

type PaddleTopologyElemPortsElem struct {
	DestinationNodeId *int    `json:"destination_node_id,omitempty" yaml:"destination_node_id,omitempty" mapstructure:"destination_node_id,omitempty"`
	DestinationPort   *int    `json:"destination_port,omitempty" yaml:"destination_port,omitempty" mapstructure:"destination_port,omitempty"`
	DestinationSlot   *string `json:"destination_slot,omitempty" yaml:"destination_slot,omitempty" mapstructure:"destination_slot,omitempty"`
	Port              *int    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port,omitempty"`
	Slot              *string `json:"slot,omitempty" yaml:"slot,omitempty" mapstructure:"slot,omitempty"`
	Speed             *int    `json:"speed,omitempty" yaml:"speed,omitempty" mapstructure:"speed,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (ccj *Paddle) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if v, ok := raw["architecture"]; !ok || v == nil {
		return fmt.Errorf("field architecture in Paddle: required")
	}
	if v, ok := raw["canu_version"]; !ok || v == nil {
		return fmt.Errorf("field canu_version in Paddle: required")
	}
	if v, ok := raw["shcd_file"]; !ok || v == nil {
		return fmt.Errorf("field shcd_file in Paddle: required")
	}
	if v, ok := raw["topology"]; !ok || v == nil {
		return fmt.Errorf("field topology in Paddle: required")
	}
	type Plain Paddle
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*ccj = Paddle(plain)
	return nil
}

func LoadPaddle(path string) (ccj *Paddle, err error) {
	p, err := os.ReadFile(path)
	if err != nil {
		return ccj, err
	}

	err = json.Unmarshal(p, &ccj)
	if err != nil {
		return ccj, err
	}

	return ccj, nil
}

// // ElevationToLocationPath converts a CANU elevation to a CANI LocationPath
// func (top PaddleTopologyElem) ElevationToLocationPath(b []byte) error {
// 	// HPCM        --->   CANI
// 	// ------------------------------
// 	// Rack        --->   Cabinet
// 	// Chassis     --->   Chassis
// 	// Tray        --->   NodeBlade/SwitchBlade
// 	// Controller  --->   NodeController
// 	// Node        --->   Node
// 	var system, cabinet, chassis, blade, controller, node inventory.LocationToken
// 	// rack and chassis map to cabinet and chassis
// 	system = inventory.LocationToken{HardwareType: hardwaretypes.System, Ordinal: 0}
// 	cabinet = inventory.LocationToken{HardwareType: hardwaretypes.Cabinet, Ordinal: int(hpcmLoc.Rack)}
// 	chassis = inventory.LocationToken{HardwareType: hardwaretypes.Chassis, Ordinal: int(hpcmLoc.Chassis)}

// 	// HPCM's Tray could be one of NodeBlade, ManagementSwitchEnclosure, or HighSpeedSwitchEnclosure
// 	switch top.Type {
// 	case hardwaretypes.System:
// 		log.Debug().Msgf("LocationPath for %+v is currently limited to a single system", hardwaretypes.System)
// 		caniLoc = inventory.LocationPath{system}
// 	case hardwaretypes.Cabinet:
// 		caniLoc = inventory.LocationPath{system, cabinet}
// 	case hardwaretypes.Chassis:
// 		caniLoc = inventory.LocationPath{system, cabinet, chassis}
// 	case hardwaretypes.NodeBlade:
// 		blade = inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: int(hpcmLoc.Tray)}
// 		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade}
// 	case hardwaretypes.ManagementSwitchEnclosure:
// 		blade = inventory.LocationToken{HardwareType: hardwaretypes.ManagementSwitchEnclosure, Ordinal: int(hpcmLoc.Tray)}
// 		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade}
// 	case hardwaretypes.HighSpeedSwitchEnclosure:
// 		blade = inventory.LocationToken{HardwareType: hardwaretypes.HighSpeedSwitchEnclosure, Ordinal: int(hpcmLoc.Tray)}
// 		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade}
// 	case hardwaretypes.Node:
// 		blade = inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: int(hpcmLoc.Tray)}
// 		controller = inventory.LocationToken{HardwareType: hardwaretypes.NodeController, Ordinal: int(hpcmLoc.Controller)}
// 		node = inventory.LocationToken{HardwareType: hardwaretypes.Node, Ordinal: int(hpcmLoc.Node)}
// 		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade, controller, node}
// 	default:
// 		// assume a node
// 		blade = inventory.LocationToken{HardwareType: hardwaretypes.NodeBlade, Ordinal: int(hpcmLoc.Tray)}
// 		controller = inventory.LocationToken{HardwareType: hardwaretypes.NodeController, Ordinal: int(hpcmLoc.Controller)}
// 		node = inventory.LocationToken{HardwareType: hardwaretypes.Node, Ordinal: int(hpcmLoc.Node)}
// 		caniLoc = inventory.LocationPath{system, cabinet, chassis, blade, controller, node}
// 		// log.Warn().Msgf("Unable to get LocationPath from hardware type: %v", caniHwType)
// 	}

// 	log.Debug().Msgf("Set LocationPath via HPCM geo location values: %+v -> %v", hpcmLoc, caniLoc)

// 	return caniLoc, nil
// 	return nil
// }

// PaddleTypeToCaniHardwareType converts a PADDLE type into a CANI hardwaretype
func (top PaddleTopologyElem) PaddleTypeToCaniHardwareType() (t hardwaretypes.HardwareType, err error) {
	switch *top.Type {
	case "switch":
		switch *top.Architecture {
		case "spine", "river_bmc_leaf":
			t = hardwaretypes.ManagementSwitch
		case "slingshot_hsn_switch":
			t = hardwaretypes.HighSpeedSwitch
		}

	case "server":
		switch *top.Architecture {
		case "river_ncn_node_2_port_gigabyte", "river_ncn_node_4_port_gigabyte", "mountain_compute_leaf":
			t = hardwaretypes.NodeBlade
		}

	case "none":
		switch *top.Architecture {
		case "pdu":
			t = hardwaretypes.CabinetPDU
		case "kvm":
			t = "FIXME hardwaretypes.CabinetKVM"
		}

	case "node":
		t = hardwaretypes.Node

	default:
		err = fmt.Errorf("unable to map Paddle type to CANI hardwaretype: %v", *top.Type)
	}
	if err != nil {
		return t, err
	}

	return t, nil
}

// func (port PaddleTopologyElemPortsElem) PaddleToNetbox() (nb map[string]interface{}, err error) {
// 	dt := hardwaretypes.DeviceType{}

// 	nb = make(map[string]interface{}, 0)
// 	nb["name"] = fmt.Sprintf("port-%d", *port.Port)
// 	nb["destination"] = *port.DestinationPort
// 	nb["type"] = "1000base-t"
// 	return nb, nil
// }

// PaddleToNetboxDeviceType converts a PADDLE type into a CANI hardwaretype
func (top PaddleTopologyElem) PaddleToNetboxDeviceType() (dt hardwaretypes.DeviceType, err error) {
	dt = hardwaretypes.DeviceType{}
	if top.RackElevation != nil {
		racku, err := strconv.Atoi(*top.RackElevation)
		if err != nil {
			return dt, err
		}
		u := float64(racku)
		dt.UHeight = &u
	}

	for _, p := range top.Ports {
		frontport := netbox.FrontPort{}
		ifcfg := netbox.Interface{}
		frontport.Name = fmt.Sprintf("port-%d", *p.DestinationPort)
		if p.DestinationNodeId != nil {
			// log.Info().Msgf("destid %+v", *p.DestinationNodeId)
			frontport.Id = int32(*p.DestinationNodeId)
		}
		if p.DestinationPort != nil {

			// log.Info().Msgf("destport %+v", *p.DestinationPort)
		}
		if p.DestinationSlot != nil {
			// log.Info().Msgf("destslot %+v", *p.DestinationSlot)
		}
		if p.Port != nil {
			// log.Info().Msgf("port %+v", *p.Port)
		}
		if p.Slot != nil {
			// log.Info().Msgf("slot %+v", *p.Slot)
			t := netbox.FrontPortType{}
			l := netbox.FrontPortTypeLabel(*p.Slot)
			t.Label = &l
			frontport.Type = t
		}
		if p.Speed != nil {
			is := netbox.NullableInt32{}
			s := int32(*p.Speed)
			is.Set(&s)
			ifcfg.Speed = is
			// log.Info().Msgf("speed %+v", *p.Speed)
		}
		// port.Type = "1000base-t"
		// port.Destination = *p.DestinationPort
		// dt.FrontPorts = append(dt.FrontPorts, frontport)
		// dt.Interfaces = append(dt.Interfaces, ifcfg)
	}

	// log.Debug().Msgf("Set LocationPath via HPCM geo location values: %+v -> %v", top.Location, dt)

	return dt, nil
}
