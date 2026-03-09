package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/csm/client"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
)

// hmnMTNKey is the SLS network name for the mountain HMN.
const hmnMTNKey = "HMN_MTN"

// nmnMTNKey is the SLS network name for the mountain NMN.
const nmnMTNKey = "NMN_MTN"

// enrichCabinetNetworks sets ExtraProperties.Networks on each new
// cabinet entry in expected so the PUT to SLS includes network data.
func enrichCabinetNetworks(
	expected map[string]import_.SlsHardware,
	networks map[string]import_.SlsNetwork,
	inventory devicetypes.Inventory,
) {
	hmnNet := networks[hmnMTNKey]
	nmnNet := networks[nmnMTNKey]

	for _, dev := range inventory.Devices {
		if dev == nil || dev.GetType() != devicetypes.TypeCabinet {
			continue
		}
		if dev.Status != "staged" {
			continue
		}
		xname := extractXname(dev)
		hw, ok := expected[xname]
		if !ok {
			continue
		}
		sub, _ := dev.GetProviderSubMap("csm")
		hmnVlan := intFromMeta(sub, "hmnVlan")
		if hmnVlan == 0 {
			continue
		}

		nmnVlan := deriveNMNVlan(hmnVlan, &hmnNet, &nmnNet)

		hmnSub, err := computeSubnet(&hmnNet, hmnVlan, xname)
		if err != nil {
			continue
		}
		nmnSub, err := computeSubnet(&nmnNet, nmnVlan, xname)
		if err != nil {
			continue
		}

		setNetworkMetadata(&hw, hmnSub, nmnSub, hmnVlan, nmnVlan)
		expected[xname] = hw
	}
}

// reconcileNetworks pushes updated SLS network definitions when new
// cabinet subnets need to be added.
func reconcileNetworks(
	c *client.Client,
	networks map[string]import_.SlsNetwork,
	changes hardwareChanges,
	inventory devicetypes.Inventory,
	stats *reconcileStats,
) error {
	hmnNet := networks[hmnMTNKey]
	nmnNet := networks[nmnMTNKey]

	var hmnDirty, nmnDirty bool

	for _, hw := range changes.Added {
		if hw.TypeString != "Cabinet" {
			continue
		}
		dev := findDeviceByXname(hw.Xname, inventory)
		if dev == nil {
			continue
		}
		sub, _ := dev.GetProviderSubMap("csm")
		hmnVlan := intFromMeta(sub, "hmnVlan")
		if hmnVlan == 0 {
			continue
		}
		nmnVlan := deriveNMNVlan(hmnVlan, &hmnNet, &nmnNet)
		ordinal := cabinetOrdinalFromXname(hw.Xname)

		hmnSub, err := computeSubnet(&hmnNet, hmnVlan, hw.Xname)
		if err != nil {
			return fmt.Errorf("unable to allocate subnet for cabinet (%s) in network (%s): %w", hw.Xname, hmnMTNKey, err)
		}
		nmnSub, err := computeSubnet(&nmnNet, nmnVlan, hw.Xname)
		if err != nil {
			return fmt.Errorf("unable to allocate subnet for cabinet (%s) in network (%s): %w", hw.Xname, nmnMTNKey, err)
		}

		if appendSubnet(&hmnNet, cabinetSubnetName(ordinal), hmnSub, hmnVlan) {
			hmnDirty = true
		}
		if appendSubnet(&nmnNet, cabinetSubnetName(ordinal), nmnSub, nmnVlan) {
			nmnDirty = true
		}
	}

	if hmnDirty {
		if err := putNetwork(c, hmnNet); err != nil {
			return err
		}
		stats.PutCount++
	}
	if nmnDirty {
		if err := putNetwork(c, nmnNet); err != nil {
			return err
		}
		stats.PutCount++
	}
	return nil
}

// deriveNMNVlan computes the NMN VLAN from an HMN VLAN using the
// base-VLAN offset between the two mountain networks.
func deriveNMNVlan(
	hmnVlan int,
	hmn *import_.SlsNetwork,
	nmn *import_.SlsNetwork,
) int {
	hmnBase := baseVlan(hmn)
	nmnBase := baseVlan(nmn)
	return nmnBase + (hmnVlan - hmnBase)
}

// computeSubnet derives the /22 subnet parameters from a mountain
// network and a VLAN number. It returns an error if the computed
// subnet falls outside the network's CIDR.
func computeSubnet(
	network *import_.SlsNetwork,
	vlan int,
	xname string,
) (subnetInfo, error) {
	netCIDR := networkBaseCIDR(network)
	baseIP, parentNet, err := net.ParseCIDR(netCIDR)
	if err != nil {
		return subnetInfo{}, fmt.Errorf("failed to parse network CIDR (%s): %w", netCIDR, err)
	}
	base := baseVlan(network)
	offset := vlan - base

	ip := make(net.IP, len(baseIP.To4()))
	copy(ip, baseIP.To4())
	// Compute the byte offset across octets 2 and 3.
	byteOffset := offset * 4
	ip[2] += byte(byteOffset & 0xFF)
	if byteOffset > 0xFF {
		ip[1] += byte(byteOffset >> 8)
	}

	if !parentNet.Contains(ip) {
		return subnetInfo{}, fmt.Errorf("failed to allocate cabinet subnet for (%s) in CIDR (%s): %w",
			xname, netCIDR, errors.New("network space has been exhausted"))
	}

	cidr := fmt.Sprintf("%s/22", ip.String())
	gateway := net.IP(make([]byte, 4))
	copy(gateway, ip)
	gateway[3] = 1

	dhcpStart := net.IP(make([]byte, 4))
	copy(dhcpStart, ip)
	dhcpStart[3] = 10

	dhcpEnd := net.IP(make([]byte, 4))
	copy(dhcpEnd, ip)
	dhcpEnd[2] += 3
	dhcpEnd[3] = 254

	return subnetInfo{
		CIDR:      cidr,
		Gateway:   gateway.String(),
		DHCPStart: dhcpStart.String(),
		DHCPEnd:   dhcpEnd.String(),
	}, nil
}

// subnetInfo holds computed subnet parameters.
type subnetInfo struct {
	CIDR      string
	Gateway   string
	DHCPStart string
	DHCPEnd   string
}

// setNetworkMetadata writes the Networks map into a cabinet's ExtraProperties.
func setNetworkMetadata(
	hw *import_.SlsHardware,
	hmn subnetInfo, nmn subnetInfo,
	hmnVlan, nmnVlan int,
) {
	if hw.ExtraProperties == nil {
		hw.ExtraProperties = make(map[string]any)
	}
	hw.ExtraProperties["Networks"] = map[string]any{
		"cn": map[string]any{
			"HMN": map[string]any{
				"CIDR":    hmn.CIDR,
				"Gateway": hmn.Gateway,
				"VLan":    hmnVlan,
			},
			"NMN": map[string]any{
				"CIDR":    nmn.CIDR,
				"Gateway": nmn.Gateway,
				"VLan":    nmnVlan,
			},
		},
	}
}

// appendSubnet adds a cabinet subnet to a network if it doesn't exist.
// Returns true if the network was modified.
func appendSubnet(
	network *import_.SlsNetwork,
	name string,
	info subnetInfo,
	vlan int,
) bool {
	if network.ExtraProperties == nil {
		return false
	}
	for _, sub := range network.ExtraProperties.Subnets {
		if sub.Name == name {
			return false // already exists
		}
	}
	network.ExtraProperties.Subnets = append(network.ExtraProperties.Subnets, import_.SlsSubnet{
		Name:      name,
		FullName:  "",
		CIDR:      info.CIDR,
		Gateway:   info.Gateway,
		VlanID:    vlan,
		DHCPStart: info.DHCPStart,
		DHCPEnd:   info.DHCPEnd,
	})
	return true
}

// putNetwork PUTs an updated SLS network definition.
func putNetwork(c *client.Client, net import_.SlsNetwork) error {
	url := c.BaseURLSLS + "/networks/" + net.Name
	data, err := json.Marshal(net)
	if err != nil {
		return fmt.Errorf("marshaling network %s: %w", net.Name, err)
	}
	log.Printf("PUT %s", url)
	if _, err := c.Put(url, data); err != nil {
		return fmt.Errorf("PUT %s: %w", url, err)
	}
	return nil
}

// baseVlan extracts the first VLAN from a network's VlanRange.
func baseVlan(net *import_.SlsNetwork) int {
	if net.ExtraProperties == nil || len(net.ExtraProperties.VlanRange) == 0 {
		return 0
	}
	return net.ExtraProperties.VlanRange[0]
}

// networkBaseCIDR returns the super-network CIDR from ExtraProperties.
func networkBaseCIDR(net *import_.SlsNetwork) string {
	if net.ExtraProperties == nil {
		return ""
	}
	return net.ExtraProperties.CIDR
}

// cabinetSubnetName returns the SLS subnet name for a cabinet ordinal.
func cabinetSubnetName(ordinal int) string {
	return fmt.Sprintf("cabinet_%d", ordinal)
}

// cabinetOrdinalFromXname extracts the numeric ordinal from a cabinet xname.
func cabinetOrdinalFromXname(xname string) int {
	var n int
	if _, err := fmt.Sscanf(xname, "x%d", &n); err != nil {
		return 0
	}
	return n
}

// findDeviceByXname locates a device in the inventory by its CSM xname.
func findDeviceByXname(
	xname string,
	inventory devicetypes.Inventory,
) *devicetypes.CaniDeviceType {
	for _, dev := range inventory.Devices {
		if extractXname(dev) == xname {
			return dev
		}
	}
	return nil
}

// intFromMeta extracts an integer from a metadata map.
func intFromMeta(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}
