/*
 *
 *  MIT License
 *
 *  (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
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
package sls

import (
	"fmt"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/antihax/optional"
	"github.com/mitchellh/mapstructure"
)

func DecodeNetworkExtraProperties(extraPropertiesRaw interface{}, extraProperties *sls_common.NetworkExtraProperties) error {
	// Map this network to a usable structure.
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToIPHookFunc(),
		Result:     extraProperties,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(extraPropertiesRaw)
}

func Networks(state sls_common.SLSState) (networks sls_common.NetworkArray) {
	for _, network := range state.Networks {
		networks = append(networks, network)
	}

	return networks
}

var ErrSubnetNotFound = fmt.Errorf("subnet not found")

// LookupSubnet returns a subnet by name
// Note the return index value is useful to put modifications back into the subnet slice of a network's extra properties
func LookupSubnet(network sls_client.Network, subnetName string) (sls_client.NetworkIpv4Subnet, int, error) {
	return LookupSubnetInEP(network.ExtraProperties, subnetName)
}

// LookupSubnet returns a subnet by name
// Note the return index value is useful to put modifications back into the subnet slice of a network's extra properties
func LookupSubnetInEP(networkEP *sls_client.NetworkExtraProperties, subnetName string) (sls_client.NetworkIpv4Subnet, int, error) {
	var found []sls_client.NetworkIpv4Subnet
	if networkEP == nil || len(networkEP.Subnets) == 0 {
		return sls_client.NetworkIpv4Subnet{}, 0, ErrSubnetNotFound
	}
	var index int
	for i, v := range networkEP.Subnets {
		if v.Name == subnetName {
			index = i
			found = append(found, v)
		}
	}
	if len(found) == 1 {
		// The Index is valid since, only one match was found!
		return found[0], index, nil
	}
	if len(found) > 1 {
		return found[0], 0, fmt.Errorf("found %v subnets instead of just one", len(found))
	}
	return sls_client.NetworkIpv4Subnet{}, 0, ErrSubnetNotFound
}

// ReservationsByName presents the IPReservations in a map by name
func ReservationsByName(subnet sls_client.NetworkIpv4Subnet) map[string]sls_client.NetworkIpReservation {
	reservations := make(map[string]sls_client.NetworkIpReservation)
	for _, v := range subnet.IPReservations {
		reservations[v.Name] = v
	}
	return reservations
}

func NewNetworkApiNetworksNetworkPutOpts(network sls_client.Network) *sls_client.NetworkApiNetworksNetworkPutOpts {
	return &sls_client.NetworkApiNetworksNetworkPutOpts{
		Body: optional.NewInterface(network),
	}
}
