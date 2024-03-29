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
package ipam

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/rs/zerolog/log"
	"inet.af/netaddr"
)

func ExistingIPAddresses(slsSubnet sls_client.NetworkIpv4Subnet) (*netaddr.IPSet, error) {
	var existingIPAddresses netaddr.IPSetBuilder
	gatewayIP, err := netaddr.ParseIP(slsSubnet.Gateway)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to parse gateway IP (%v)", slsSubnet.Gateway), err)
	}
	existingIPAddresses.Add(gatewayIP)

	for _, ipReservation := range slsSubnet.IPReservations {
		ip, err := netaddr.ParseIP(ipReservation.IPAddress)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to parse IPReservation IP (%v)", ipReservation.IPAddress), err)
		}
		existingIPAddresses.Add(ip)
	}

	return existingIPAddresses.IPSet()
}

func FindNextAvailableIP(slsNetwork sls_client.Network, slsSubnet sls_client.NetworkIpv4Subnet) (netaddr.IP, error) {
	// TODO this function should have guardrails to ensure that the IPs are within the static range.

	subnet, err := netaddr.ParseIPPrefix(slsSubnet.CIDR)
	if err != nil {
		return netaddr.IP{}, errors.Join(fmt.Errorf("failed to parse subnet CIDR (%v)", slsSubnet.CIDR), err)
	}

	// If the subnet has been supernet hacked, then the unhacked CIDR will be returned. Otherwise none will be returned.
	if correctedSubnet, correctedGateway, err := IsSupernetHacked(slsNetwork, slsSubnet); correctedSubnet != nil {
		log.Info().Msgf("Info the %s subnet in the %s network has been supernet hacked! Changing CIDR from %v to %v for IP address calculation", slsSubnet.Name, slsNetwork.Name, subnet, correctedSubnet)
		subnet = *correctedSubnet
		// This won't alter the incoming SLS object, as it is passed in by value
		slsSubnet.CIDR = correctedSubnet.String()
		slsSubnet.Gateway = correctedGateway.String()
	} else if err != nil {
		return netaddr.IP{}, errors.Join(fmt.Errorf("failed to detect supernet hack on subnet %s in network %s", slsSubnet.Name, slsNetwork.Name), err)
	}

	existingIPAddressesSet, err := ExistingIPAddresses(slsSubnet)
	if err != nil {
		return netaddr.IP{}, err
	}

	startingIP := subnet.Range().From().Next() // Start at the first usable available IP in the subnet.
	endingIP := subnet.Range().To()            // This is the broadcast IP

	for ip := startingIP; ip.Less(endingIP); ip = ip.Next() {
		if !existingIPAddressesSet.Contains(ip) {
			return ip, nil
		}
	}

	return netaddr.IP{}, fmt.Errorf("subnet has no available IPs")
}

func AdvanceIP(ip netaddr.IP, n uint32) (netaddr.IP, error) {
	if ip.Is6() {
		return netaddr.IP{}, fmt.Errorf("IPv6 is not supported")
	}
	if ip.IsZero() {
		return netaddr.IP{}, fmt.Errorf("empty IP address provided")
	}

	// This is kind of crude hack, but if it works it works.
	ipOctets := ip.As4()
	ipRaw := binary.BigEndian.Uint32(ipOctets[:])

	// Advance the IP by n
	ipRaw += n

	// Now put it back into an netaddr.IP
	var updatedIPOctets [4]byte
	binary.BigEndian.PutUint32(updatedIPOctets[:], ipRaw)

	return netaddr.IPFrom4(updatedIPOctets), nil
}

func SplitNetwork(network netaddr.IPPrefix, subnetMaskOneBits uint8) ([]netaddr.IPPrefix, error) {
	// TODO why only allow this range?
	if subnetMaskOneBits < 1 || 30 < subnetMaskOneBits {
		return nil, fmt.Errorf("invalid subnet mask provided /%d", subnetMaskOneBits)
	}

	// Verify that the network can be split (this is allowing a split of the same size)
	if subnetMaskOneBits < network.Bits() {
		return nil, fmt.Errorf("provided subnet mask bits /%d is larger than starting network subnet mask /%d", subnetMaskOneBits, network.Bits())
	}

	subnetStartIP := network.Range().From()

	// TODO add a counter to prevent this loop from going in forever!
	var subnets []netaddr.IPPrefix
	for {
		subnets = append(subnets, netaddr.IPPrefixFrom(subnetStartIP, subnetMaskOneBits))

		advanceBy := uint32(math.Pow(2, float64(32-subnetMaskOneBits)))

		// Now advance!
		var err error
		subnetStartIP, err = AdvanceIP(subnetStartIP, advanceBy)
		if err != nil {
			return nil, err
		}

		if network.Range().To().Less(subnetStartIP) {
			break
		}
	}

	return subnets, nil
}

func FindNextAvailableSubnet(slsNetwork sls_client.NetworkExtraProperties, desiredSubnetMaskBits uint8) (netaddr.IPPrefix, error) {
	// TODO make the /22 configurable
	var existingSubnets netaddr.IPSetBuilder
	for _, slsSubnet := range slsNetwork.Subnets {
		subnet, err := netaddr.ParseIPPrefix(slsSubnet.CIDR)
		if err != nil {
			return netaddr.IPPrefix{}, errors.Join(fmt.Errorf("failed to parse subnet CIDR (%v)", slsSubnet.CIDR), err)
		}

		existingSubnets.AddPrefix(subnet)
	}

	existingSubnetsSet, err := existingSubnets.IPSet()
	if err != nil {
		return netaddr.IPPrefix{}, err
	}

	network, err := netaddr.ParseIPPrefix(slsNetwork.CIDR)
	if err != nil {
		return netaddr.IPPrefix{}, errors.Join(fmt.Errorf("failed to parse network CIDR (%s)", slsNetwork.CIDR), err)
	}

	availableSubnets, err := SplitNetwork(network, desiredSubnetMaskBits)
	if err != nil {
		return netaddr.IPPrefix{}, errors.Join(fmt.Errorf("failed to split network CIDR (%s)", slsNetwork.CIDR), err)
	}
	for _, subnet := range availableSubnets {
		if existingSubnetsSet.Contains(subnet.IP()) {
			continue
		}

		return subnet, nil
	}

	return netaddr.IPPrefix{}, fmt.Errorf("network space has been exhausted")
}

func AllocateCabinetSubnet(networkName string, slsNetwork sls_client.NetworkExtraProperties, xname xnames.Cabinet, desiredSubnetMaskBits uint8, vlanOverride *int32) (sls_client.NetworkIpv4Subnet, error) {
	cabinetSubnet, err := FindNextAvailableSubnet(slsNetwork, desiredSubnetMaskBits)
	if err != nil {
		return sls_client.NetworkIpv4Subnet{}, errors.Join(fmt.Errorf("failed to allocate cabinet subnet for (%s) in CIDR (%s)", xname.String(), slsNetwork.CIDR), err)
	}

	// Verify this subnet is new
	subnetName := fmt.Sprintf("cabinet_%d", xname.Cabinet)
	for _, otherSubnet := range slsNetwork.Subnets {
		if otherSubnet.Name == subnetName {
			return sls_client.NetworkIpv4Subnet{}, fmt.Errorf("subnet (%s) already exists", subnetName)
		}
	}

	// Calculate VLAN if one was not provided
	// TODO This needs to be updated to calculate only the NMN_MTN network for Cray EX cabinets. River both HMN_RVR and NMN_RVR networks
	vlan := int32(-1)
	if vlanOverride != nil {
		vlan = *vlanOverride
	} else {
		// Look at other cabinets in the subnet and pick one.

		// Determine the current vlans in use by other cabinets
		vlansInUse := map[int32]bool{}
		for _, existingSubnet := range slsNetwork.Subnets {
			vlansInUse[existingSubnet.VlanID] = true
		}

		// Now lest find the smallest free Vlan!
		var vlanLow int32 = -1
		var vlanHigh int32 = -1

		if networkName == "HMN_RVR" {
			// The following values are defined here in CSI: https://github.com/Cray-HPE/cray-site-init/blob/4ead6fccd0ba0710e7250357f1c3a2525996d293/cmd/init.go#L160
			vlanLow = 1513
			vlanHigh = 1769
		} else if networkName == "NMN_RVR" {
			// The following values are defined here in CSI: https://github.com/Cray-HPE/cray-site-init/blob/4ead6fccd0ba0710e7250357f1c3a2525996d293/cmd/init.go#L189
			vlanLow = 1770
			vlanHigh = 1999
		} else if networkName == "HMN_MTN" {
			// The following values are defined here in CSI: https://github.com/Cray-HPE/cray-site-init/blob/4ead6fccd0ba0710e7250357f1c3a2525996d293/cmd/init.go#L148
			vlanLow = 3000
			vlanHigh = 3999
		} else if networkName == "NMN_MTN" {
			// The following values are defined here in CSI: https://github.com/Cray-HPE/cray-site-init/blob/4ead6fccd0ba0710e7250357f1c3a2525996d293/cmd/init.go#L176
			vlanLow = 2000
			vlanHigh = 2999

		} else {
			return sls_client.NetworkIpv4Subnet{}, fmt.Errorf("unsupported network (%s) unable to allocate vlan for cabinet subnet", networkName)
		}

		for vlanCandidate := vlanLow; vlanCandidate <= vlanHigh; vlanCandidate++ {
			if vlansInUse[vlanCandidate] {
				// currently in use
				continue
			}

			vlan = vlanCandidate
			break
		}
	}

	if vlan == -1 {
		return sls_client.NetworkIpv4Subnet{}, fmt.Errorf("failed to allocate cabinet subnet for (%s) no subnets available", subnetName)
	}

	// DHCP starts 10 into the subnet
	dhcpStart, err := AdvanceIP(cabinetSubnet.Range().From(), 10)
	if err != nil {
		return sls_client.NetworkIpv4Subnet{}, fmt.Errorf("failed to determine DHCP start in CIDR (%s)", cabinetSubnet.String())
	}

	return sls_client.NetworkIpv4Subnet{
		Name:      subnetName,
		CIDR:      cabinetSubnet.String(),
		VlanID:    vlan,
		Gateway:   cabinetSubnet.Range().From().Next().IPAddr().IP.String(),
		DHCPStart: dhcpStart.IPAddr().IP.String(),
		DHCPEnd:   cabinetSubnet.Range().To().Prior().IPAddr().IP.String(),
	}, nil
}

func AllocateIP(slsNetwork sls_client.Network, slsSubnet sls_client.NetworkIpv4Subnet, xname xnames.Xname, alias string) (sls_client.NetworkIpReservation, error) {
	ip, err := FindNextAvailableIP(slsNetwork, slsSubnet)
	if err != nil {
		return sls_client.NetworkIpReservation{}, errors.Join(
			fmt.Errorf("failed to allocate ip for hardware (%s) in subnet (%s)", xname.String(), slsSubnet.CIDR),
			err,
		)
	}

	// Verify this switch is unique within the subnet
	for _, ipReservation := range slsSubnet.IPReservations {
		matchingAlias := ipReservation.Name == alias
		matchingXName := ipReservation.Comment == xname.String()

		if matchingAlias && matchingXName {
			// IP reservation already exists
			return sls_client.NetworkIpReservation{}, fmt.Errorf("ip reservation with name (%v) and xname (%v) already exists with IP (%s)", ipReservation.Name, ipReservation.Comment, ipReservation.IPAddress)
		} else if matchingAlias {
			return sls_client.NetworkIpReservation{}, fmt.Errorf("ip reservation with name (%v) already exists on (%v) with IP (%s)", alias, ipReservation.Comment, ipReservation.IPAddress)
		} else if matchingXName {
			return sls_client.NetworkIpReservation{}, fmt.Errorf("ip reservation with xname (%v) already exists with name (%v) with IP (%s)", xname.String(), ipReservation.Name, ipReservation.IPAddress)
		}
	}

	// TODO Move this outside this function? So this function just gives back IP within the subnet, and then have outside logic
	// Verify that the IP is actually valid ie within the DHCP range, and if not in the DHCP range expand it and verify nothing is
	// using the IP address.

	// Verify IP is within the static IP range. The static range
	if slsSubnet.DHCPStart != "" {
		dhcpStart, err := netaddr.ParseIP(slsSubnet.DHCPStart)
		if err != nil {
			return sls_client.NetworkIpReservation{}, errors.Join(fmt.Errorf("failed to parse DHCP Start IP (%s) address", slsSubnet.DHCPStart), err)
		}

		if !ip.Less(dhcpStart) {
			return sls_client.NetworkIpReservation{}, fmt.Errorf("ip reservation with xname (%v) and IP %s is outside the static IP address range, with starting DHCP IP of %s", xname.String(), ip.String(), slsSubnet.DHCPStart)
		}
	}

	return sls_client.NetworkIpReservation{
		Comment:   xname.String(),
		IPAddress: ip.IPAddr().IP.String(),
		Name:      alias,
	}, nil
}

//
// The following functions may be needed when adding hardware like application or management nodes
//

// func FreeIPsInStaticRange(slsSubnet sls_client.NetworkIpv4Subnet) (uint32, error) {
// 	// Probably need to steal some of the logic for allocate IP. Need to share the logic between the two
//
// 	subnet, err := netaddr.ParseIPPrefix(slsSubnet.CIDR)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to parse subnet CIDR (%v): %w", slsSubnet.CIDR, err)
// 	}
//
// 	existingIPAddressesSet, err := ExistingIPAddresses(slsSubnet)
// 	if err != nil {
// 		return 0, err
// 	}
//
// 	startingIP := subnet.Range().From().Next() // Start at the first usable available IP in the subnet.
// 	endingIP, err := netaddr.ParseIP(slsSubnet.DHCPStart)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to convert DHCP Start IP address to netaddr struct")
// 	}
//
// 	var count uint32
// 	for ip := startingIP; ip.Less(endingIP); ip = ip.Next() {
// 		if existingIPAddressesSet.Contains(ip) {
// 			// IP address currently in use
// 			continue
// 		}
// 		count++
// 	}
//
// 	return count, nil
// }

// func ExpandSubnetStaticRange(slsSubnet *sls_common.IPV4Subnet, count uint32) error {
// 	if slsSubnet.DHCPStart == nil || slsSubnet.DHCPEnd == nil {
// 		return fmt.Errorf("subnet does not have DHCP range")
// 	}
//
// 	dhcpStart, ok := netaddr.FromStdIP(slsSubnet.DHCPStart)
// 	if !ok {
// 		return fmt.Errorf("failed to convert DHCP Start IP address to netaddr struct")
// 	}
//
// 	dhcpEnd, ok := netaddr.FromStdIP(slsSubnet.DHCPEnd)
// 	if !ok {
// 		return fmt.Errorf("failed to convert DHCP END IP address to netaddr struct")
// 	}
//
// 	// Move it forward!
// 	dhcpStart, err := AdvanceIP(dhcpStart, count)
// 	if err != nil {
// 		return fmt.Errorf("failed to advice DHCP Start IP address: %w", err)
// 	}
//
// 	// Verify the DHCP Start address is smaller than the end address
// 	if !dhcpStart.Less(dhcpEnd) {
// 		return fmt.Errorf("new DHCP Start address %v is equal or larger then the DHCP End address %v", dhcpStart, dhcpEnd)
// 	}
//
// 	// Now update the SLS subnet
// 	slsSubnet.DHCPStart = dhcpStart.IPAddr().IP
// 	return nil
// }
