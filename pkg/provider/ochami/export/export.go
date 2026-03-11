package export

import (
	"fmt"
	"io"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"gopkg.in/yaml.v3"
)

const providerName = "ochami"

// Export iterates the CANI inventory and writes an openChamiPayload YAML
// document containing BMC and Node entries to the given writer.
func Export(inventory devicetypes.Inventory, w io.Writer) error {
	var payload openChamiPayload

	for _, dev := range inventory.Devices {
		if dev == nil {
			continue
		}
		meta := effectiveMeta(dev)

		switch dev.HardwareType {
		case "node":
			payload.Nodes = append(payload.Nodes, buildNode(dev, meta))
		case "bmc":
			payload.BMCs = append(payload.BMCs, buildBMC(dev, meta))
		}
	}

	sort.Slice(payload.Nodes, func(i, j int) bool {
		return payload.Nodes[i].Xname < payload.Nodes[j].Xname
	})
	sort.Slice(payload.BMCs, func(i, j int) bool {
		return payload.BMCs[i].Xname < payload.BMCs[j].Xname
	})

	out, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal openchami payload: %w", err)
	}

	_, err = w.Write(out)
	return err
}

// effectiveMeta builds a merged metadata map for the given device.
// Ochami-specific keys take priority; CSM keys are used as fallback.
// CSM stores aliases as "aliases" while ochami expects "host_aliases"
// and "hostname", so the mapping is performed here.
func effectiveMeta(dev *devicetypes.CaniDeviceType) map[string]any {
	result := map[string]any{}

	// Start with CSM metadata as the base.
	if csm, ok := dev.GetProviderSubMap("csm"); ok {
		copyCsmKey(result, csm, "xname")
		copyCsmKey(result, csm, "ip")
		copyCsmKey(result, csm, "mac")
		copyCsmKey(result, csm, "boot_mac")
		copyCsmKey(result, csm, "nid")
		mapCsmAliases(result, csm)
	}

	// Overlay ochami-specific keys (always win).
	if ochami, ok := dev.GetProviderSubMap(providerName); ok {
		for k, v := range ochami {
			result[k] = v
		}
	}

	return result
}

// copyCsmKey copies a single key from CSM metadata into the result map.
func copyCsmKey(dst, csm map[string]any, key string) {
	if v, ok := csm[key]; ok {
		dst[key] = v
	}
}

// mapCsmAliases converts CSM "aliases" into ochami "hostname" and
// "host_aliases" fields when they are not already set.
func mapCsmAliases(dst map[string]any, csm map[string]any) {
	aliases := extractStringSlice(csm, "aliases")
	if len(aliases) == 0 {
		return
	}
	dst["hostname"] = aliases[0]
	dst["host_aliases"] = aliases
}

func buildNode(dev *devicetypes.CaniDeviceType, meta map[string]any) openChamiNode {
	xname := extractString(meta, "xname")
	if xname == "" {
		xname = dev.Name
	}
	return openChamiNode{
		Xname:       xname,
		IP:          extractString(meta, "ip"),
		BootMAC:     extractString(meta, "boot_mac"),
		NID:         extractIntPtr(meta, "nid"),
		Hostname:    extractString(meta, "hostname"),
		HostAliases: extractStringSlice(meta, "host_aliases"),
	}
}

func buildBMC(dev *devicetypes.CaniDeviceType, meta map[string]any) openChamiBMC {
	xname := extractString(meta, "xname")
	if xname == "" {
		xname = dev.Name
	}
	return openChamiBMC{
		Xname: xname,
		IP:    extractString(meta, "ip"),
		MAC:   extractString(meta, "mac"),
	}
}
