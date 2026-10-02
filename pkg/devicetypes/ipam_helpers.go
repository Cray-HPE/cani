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
	p.Namespace = strings.TrimSpace(p.Namespace)

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
	addr.Namespace = strings.TrimSpace(addr.Namespace)

	if ip.To4() != nil {
		addr.IPVersion = 4
	} else {
		addr.IPVersion = 6
	}
	return nil
}

// FindParentPrefix finds the most-specific prefix in the target's namespace
// that contains the given prefix. Returns uuid.Nil if no parent exists.
// Equal-length candidates (legacy duplicates) tie-break on UUID so the result
// is stable across map iteration order.
func FindParentPrefix(target *CaniPrefix, prefixes map[uuid.UUID]*CaniPrefix) uuid.UUID {
	if target == nil {
		return uuid.Nil
	}
	_, targetNet, err := net.ParseCIDR(target.Prefix)
	if err != nil {
		return uuid.Nil
	}
	return findContainingPrefix(targetNet.IP, target.PrefixLen, target.EffectiveNamespace(), target.ID, prefixes)
}

// FindParentPrefixForIP finds the most-specific prefix that contains the
// given IP address within the address's intended namespace. Returns uuid.Nil
// if no parent exists.
func FindParentPrefixForIP(addr *CaniIPAddress, prefixes map[uuid.UUID]*CaniPrefix) uuid.UUID {
	if addr == nil {
		return uuid.Nil
	}
	ip := net.ParseIP(addr.Host)
	if ip == nil {
		return uuid.Nil
	}
	return findContainingPrefix(ip, maxPrefixLen(ip), NamespaceOrDefault(addr.Namespace), uuid.Nil, prefixes)
}

// findContainingPrefix returns the longest prefix in namespace that contains
// ip and is strictly shorter than maxLen, skipping the excluded ID.
func findContainingPrefix(
	ip net.IP, maxLen int, namespace string, exclude uuid.UUID, prefixes map[uuid.UUID]*CaniPrefix,
) uuid.UUID {
	var bestID uuid.UUID
	bestLen := -1
	for id, p := range prefixes {
		if id == exclude {
			continue
		}
		ones, ok := containingPrefixLen(p, ip, maxLen, namespace)
		if ok && (ones > bestLen || (ones == bestLen && id.String() < bestID.String())) {
			bestLen = ones
			bestID = id
		}
	}
	return bestID
}

// containingPrefixLen returns the mask length of p when p is in namespace,
// contains ip, and is shorter than maxLen.
func containingPrefixLen(p *CaniPrefix, ip net.IP, maxLen int, namespace string) (int, bool) {
	if p == nil || p.EffectiveNamespace() != namespace {
		return 0, false
	}
	_, candidateNet, err := net.ParseCIDR(p.Prefix)
	if err != nil {
		return 0, false
	}
	ones, _ := candidateNet.Mask.Size()
	if ones >= maxLen || !candidateNet.Contains(ip) {
		return 0, false
	}
	return ones, true
}

// maxPrefixLen returns the exclusive upper bound on a containing prefix
// length for a host address, i.e. one past the full mask of its family.
func maxPrefixLen(ip net.IP) int {
	if ip.To4() != nil {
		return 33
	}
	return 129
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
