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

package common

import (
	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
)

type SlsStateExtended struct {
	SlsState       *sls_client.SlsState
	TypeToHardware map[string][]*sls_client.Hardware

	// This contains all the xnames that are specified as parents in the hardware entries.
	// This is currently treated as a set (a map with no meaningful value)
	// If something needs to look up hardware by parent, then change this to a map
	// for example change to: ParentToChildren map[string][]*sls_client.Hardware
	ParentHasChildren map[string]struct{}

	// todo remove these if they are not used
	AliasToHardware             map[string]*sls_client.Hardware
	IPReservationNameToNetworks map[string][]*sls_client.Network
	HardwareTypeToAlias         map[string]map[string]*sls_client.Hardware
	NetworkToIPReservations     map[string][]map[string]string
}

func NewSlsStateExtended(slsState *sls_client.SlsState) *SlsStateExtended {
	s := new(SlsStateExtended)
	s.SlsState = slsState

	s.TypeToHardware = make(map[string][]*sls_client.Hardware)
	s.AliasToHardware = make(map[string]*sls_client.Hardware)
	s.ParentHasChildren = make(map[string]struct{})

	s.HardwareTypeToAlias = make(map[string]map[string]*sls_client.Hardware)
	for _, hardware := range slsState.Hardware {
		h := hardware // create a copy, hardware var is reused
		t := hardware.TypeString.String()
		aliasToHardware, ok := s.HardwareTypeToAlias[t]
		if !ok {
			s.HardwareTypeToAlias[t] = make(map[string]*sls_client.Hardware)
			aliasToHardware = s.HardwareTypeToAlias[t]
		}
		// todo handle cases where a key is already there
		aliasToHardware[hardware.Xname] = &h

		s.ParentHasChildren[h.Parent] = struct{}{}

		// todo handle case where slice is not found
		aliases, _ := GetSliceOfStrings(hardware.ExtraProperties, "Aliases")
		for _, alias := range aliases {
			aliasToHardware[alias] = &h
		}
	}

	s.IPReservationNameToNetworks = make(map[string][]*sls_client.Network)
	for _, network := range slsState.Networks {
		n := network
		for _, subnet := range network.ExtraProperties.Subnets {
			for _, reservation := range subnet.IPReservations {
				list := s.IPReservationNameToNetworks[reservation.Name]
				list = append(list, &n)
				s.IPReservationNameToNetworks[reservation.Name] = list
			}
		}
	}

	for _, hardware := range slsState.Hardware {
		h := hardware // create a copy, hardware var is reused
		t := string(hardware.TypeString)
		list := append(s.TypeToHardware[t], &h)
		s.TypeToHardware[t] = list

		aliases, _ := GetSliceOfStrings(hardware.ExtraProperties, "Aliases")
		for _, alias := range aliases {
			s.AliasToHardware[alias] = &h
		}
	}
	return s
}
