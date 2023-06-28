package ipam

import (
	"errors"
	"fmt"

	sls_client "github.com/Cray-HPE/cani/pkg/sls-client"
	"github.com/rs/zerolog/log"
	"inet.af/netaddr"
)

// IsSupernetHacked will determine if a subnet has the supernet hack applied
func IsSupernetHacked(network sls_client.Network, subnet sls_client.NetworkIpv4Subnet) (*netaddr.IPPrefix, error) {
	// Args:
	//     network_address (netaddr.IP): Address of the network the subnet is in
	//     subnet (sls_client.NetworkIpv4Subnet): The subnet in question
	// Returns:
	//     None if the supernet hack has not been applied, or an "unhacked" subnet IPv4 network.
	//
	// This is some mix of heuristics from cray-site-init and black magic.  The supernet hack was
	// applied to subnets in order for NCNs to be on the same "subnet" as the network hardware.  The
	// hack is to apply the network prefix (CIDR mask) and gateway to the subnet.
	//
	// Once the supernet hack is applied there is a fundamental loss of information, so detecting and
	// correcting a supernet hack in a subnet is very difficult unless other information can be found.
	//
	// Additional information is found from cray-site-init:
	// * The supernet hack is only applied to the HMN, NMN, CMN and MTL networks.
	// * A supernet-like hack is applied to the CAN.
	// * The supernet hack is only applied to bootstrap_dhcp, network_hardware, can_metallb_static_pool,
	//  and can_metallb_address_pool subnets
	// * default network hardware netmask = /24
	// * default bootstrap dhcp netmask = /24
	//
	// The most important heuristic indicator of the supernet hack is if a subnet has the same netmask
	// as its containing network then the supernet.

	// A clear clue as to the application of the supernet hack is where the subnet
	// mask is the same as the network mask.
	networkCIDR, err := netaddr.ParseIPPrefix(network.ExtraProperties.CIDR)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("unable to parse network CIDR (%s)", network.ExtraProperties.CIDR), err)
	}
	subnetCIDR, err := netaddr.ParseIPPrefix(subnet.CIDR)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("unable to parse subnet CIDR (%s)", network.ExtraProperties.CIDR), err)
	}
	if subnetCIDR.Bits() != networkCIDR.Bits() {
		return nil, nil
	}

	usedAddressesBuilder := netaddr.IPSetBuilder{}
	for _, ipReservation := range subnet.IPReservations {
		ip, err := netaddr.ParseIP(ipReservation.IPAddress)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to parse IPReservation IP (%s)", ipReservation.IPAddress), err)
		}
		usedAddressesBuilder.Add(ip)
	}
	if subnet.DHCPStart != "" {
		ip, err := netaddr.ParseIP(subnet.DHCPStart)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to parse DHCP Start IP (%s)", subnet.DHCPStart), err)
		}
		usedAddressesBuilder.Add(ip)
	}
	if subnet.DHCPEnd != "" {
		ip, err := netaddr.ParseIP(subnet.DHCPEnd)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to parse DHCP End IP (%s)", subnet.DHCPEnd), err)
		}
		usedAddressesBuilder.Add(ip)
	}
	usedAddresses, err := usedAddressesBuilder.IPSet()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to build used IP addresses set"), err)
	}

	usedIPRanges := usedAddresses.Ranges()
	if len(usedIPRanges) == 0 {
		return nil, nil
	}
	minIP := usedIPRanges[0].From()
	maxIP := usedIPRanges[len(usedIPRanges)-1].To()

	// TODO
	// print("ORIG SUBNET: ", subnet.name(), subnet.ipv4_address(), subnet.ipv4_network())
	log.Info().Msgf("ORIG SUBNET: %s CIDR: %v", subnet.Name, subnetCIDR)
	log.Info().Msgf("MIN IP: %v MAX IP: %v", minIP, maxIP)
	log.Info().Msgf("PREFIXES SAME: %v", subnetCIDR.Bits() != networkCIDR.Bits())

	// The following are from cray-site-init where the supernet hack is applied.
	coreSubnets := map[string]bool{
		// Core Subnets
		"bootstrap_dhcp":   true,
		"network_hardware": true,
	}
	staticPoolSubnets := map[string]bool{
		// Static bool Subnets
		"can_metallb_static_pool": true,
	}
	dynamicPoolSubnets := map[string]bool{
		// Dynamic pool subnets
		"can_metallb_address_pool": true,
	}

	// Do not apply the reverse hackology for subnets in CSI it is not applied
	if !(coreSubnets[subnet.Name] || staticPoolSubnets[subnet.Name] || dynamicPoolSubnets[subnet.Name]) {
		log.Info().Msgf("Subnet %s is present in supernet hacked subnet", subnet.Name)
		return nil, nil
	}

	// Subnet masks found in CSI for different subnets. This prevents reverse
	// engineering very small subnets based on number of hosts and dhcp ranges alone.
	prefixDiff := 30
	if coreSubnets[subnet.Name] {
		prefixDiff = 24
	} else if staticPoolSubnets[subnet.Name] {
		prefixDiff = 28
	} else if dynamicPoolSubnets[subnet.Name] {
		prefixDiff = 27
	}
	log.Info().Msgf("PREFIXLEN: %v", prefixDiff)
	prefixDiff -= int(networkCIDR.Bits())

	for level := prefixDiff; level > 0; level-- {

		blocks, err := SplitNetwork(networkCIDR, uint8(level))
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to split network CIDR (%v) with subnet size /%d", networkCIDR, level), err)
		}

		for _, block := range blocks {
			if block.Contains(minIP) && block.Contains(maxIP) {
				log.Info().Msgf("CATCH: %v in %v", block, blocks)
				log.Info().Msgf("    Address: %v", block)
				log.Info().Msgf("    Gateway: %v", block.Range().From())
				return &block, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to determine prefix length")
}
