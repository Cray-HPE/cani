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

		switch string(dev.Type) {
		case "node":
			payload.Nodes = append(payload.Nodes, buildEntry(dev, meta))
		case "bmc":
			payload.BMCs = append(payload.BMCs, buildEntry(dev, meta))
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

// effectiveMeta builds the metadata map an entry is written from: the
// device's xname, which is an identity key any provider's bucket may hold,
// and then everything in the OChaMi bucket, which wins. The network fields
// come from the OChaMi bucket alone; no other provider records them.
func effectiveMeta(dev *devicetypes.CaniDeviceType) map[string]any {
	result := map[string]any{}
	if xname, ok := dev.IdentityValue("xname"); ok {
		result["xname"] = xname
	}
	if ochami, ok := dev.GetProviderSubMap(providerName); ok {
		for k, v := range ochami {
			result[k] = v
		}
	}
	return result
}

func buildEntry(dev *devicetypes.CaniDeviceType, meta map[string]any) openChamiEntry {
	xname := extractString(meta, "xname")
	if xname == "" {
		xname = dev.Name
	}
	return openChamiEntry{
		Xname: xname,
		MAC:   effectiveMAC(meta),
		IP:    extractString(meta, "ip"),
	}
}

func effectiveMAC(meta map[string]any) string {
	if mac := extractString(meta, "mac"); mac != "" {
		return mac
	}
	return extractString(meta, "boot_mac")
}
