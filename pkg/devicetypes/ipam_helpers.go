package devicetypes

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
)

// ParsePrefix parses a CIDR string and populates the derived fields
// (Network, Broadcast, PrefixLen, IPVersion) on the given CaniPrefix.
func ParsePrefix(p *CaniPrefix) error {
	if p == nil {
		return fmt.Errorf("cannot parse nil prefix")
	}
	ip, ipNet, err := net.ParseCIDR(p.Prefix)
	if err != nil {
		return fmt.Errorf("invalid prefix %q: %w", p.Prefix, err)
	}

	ones, _ := ipNet.Mask.Size()
	p.PrefixLen = ones
	p.Network = ipNet.IP.String()

	if ip.To4() != nil {
		p.IPVersion = 4
		p.Broadcast = broadcastIPv4(ipNet)
	} else {
		p.IPVersion = 6
		p.Broadcast = broadcastIPv6(ipNet)
	}
	return nil
}

// ParseIPAddress parses the Address (CIDR) field and populates derived
// fields (Host, MaskLength, IPVersion) on the given CaniIPAddress.
func ParseIPAddress(addr *CaniIPAddress) error {
	if addr == nil {
		return fmt.Errorf("cannot parse nil IP address")
	}

	// Accept either "10.0.0.1/24" or bare "10.0.0.1".
	var ip net.IP
	var maskLen int

	if strings.Contains(addr.Address, "/") {
		parsedIP, ipNet, err := net.ParseCIDR(addr.Address)
		if err != nil {
			return fmt.Errorf("invalid address %q: %w", addr.Address, err)
		}
		ip = parsedIP
		ones, _ := ipNet.Mask.Size()
		maskLen = ones
	} else {
		ip = net.ParseIP(addr.Address)
		if ip == nil {
			return fmt.Errorf("invalid IP %q", addr.Address)
		}
		if ip.To4() != nil {
			maskLen = 32
		} else {
			maskLen = 128
		}
	}

	addr.Host = ip.String()
	addr.MaskLength = maskLen
	addr.Address = fmt.Sprintf("%s/%d", ip.String(), maskLen)

	if ip.To4() != nil {
		addr.IPVersion = 4
	} else {
		addr.IPVersion = 6
	}
	return nil
}

// FindParentPrefix finds the most-specific prefix that contains the
// given prefix. Returns uuid.Nil if no parent exists.
func FindParentPrefix(target *CaniPrefix, prefixes map[uuid.UUID]*CaniPrefix) uuid.UUID {
	if target == nil {
		return uuid.Nil
	}
	_, targetNet, err := net.ParseCIDR(target.Prefix)
	if err != nil {
		return uuid.Nil
	}

	var bestID uuid.UUID
	bestLen := -1

	for id, p := range prefixes {
		if id == target.ID || p == nil {
			continue
		}
		_, candidateNet, err := net.ParseCIDR(p.Prefix)
		if err != nil {
			continue
		}
		ones, _ := candidateNet.Mask.Size()
		if ones >= target.PrefixLen {
			continue // not less-specific
		}
		if !candidateNet.Contains(targetNet.IP) {
			continue
		}
		if ones > bestLen {
			bestLen = ones
			bestID = id
		}
	}
	return bestID
}

// FindParentPrefixForIP finds the most-specific prefix that contains
// the given IP address. Returns uuid.Nil if no parent exists.
func FindParentPrefixForIP(addr *CaniIPAddress, prefixes map[uuid.UUID]*CaniPrefix) uuid.UUID {
	if addr == nil {
		return uuid.Nil
	}
	ip := net.ParseIP(addr.Host)
	if ip == nil {
		return uuid.Nil
	}

	var bestID uuid.UUID
	bestLen := -1

	for id, p := range prefixes {
		if p == nil {
			continue
		}
		_, candidateNet, err := net.ParseCIDR(p.Prefix)
		if err != nil {
			continue
		}
		ones, _ := candidateNet.Mask.Size()
		if !candidateNet.Contains(ip) {
			continue
		}
		if ones > bestLen {
			bestLen = ones
			bestID = id
		}
	}
	return bestID
}

// broadcastIPv4 computes the broadcast address of an IPv4 network.
func broadcastIPv4(n *net.IPNet) string {
	ip := n.IP.To4()
	if ip == nil {
		return ""
	}
	broadcast := make(net.IP, 4)
	for i := range ip {
		broadcast[i] = ip[i] | ^n.Mask[i]
	}
	return broadcast.String()
}

// broadcastIPv6 computes the last address of an IPv6 network.
func broadcastIPv6(n *net.IPNet) string {
	ip := n.IP.To16()
	if ip == nil {
		return ""
	}
	broadcast := make(net.IP, 16)
	for i := range ip {
		broadcast[i] = ip[i] | ^n.Mask[i]
	}
	return broadcast.String()
}
