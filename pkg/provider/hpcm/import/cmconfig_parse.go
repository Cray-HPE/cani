package import_

import (
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/provider/hpcm/import/cmconfig"
)

// sectionKeyTemplate is the key holding a NIC template definition.
const sectionKeyTemplate = "template"

// ParseCmConfig parses an HPCM cm.config file into a CmConfig struct.
// The cm.config format repeats keys (name=, hostname1=, template=) once per
// entry; parseINI preserves those shadow values so every entry is captured.
func ParseCmConfig(data []byte) (*CmConfig, error) {
	cfg := parseINI(data)
	return &CmConfig{
		Templates:    importTemplatesSection(cfg),
		NicTemplates: importNicTemplatesSection(cfg),
		Discover:     importDiscoverSection(cfg),
	}, nil
}

// parseSubkeys splits a cm.config shadow value into "key=value" pairs and
// invokes apply for each. The first bare (non key=value) token is passed to
// setName; subsequent bare keywords are ignored.
func parseSubkeys(v string, apply func(key, val string), setName func(name string)) {
	for i, subkey := range strings.Split(v, ", ") {
		kvp := strings.SplitN(subkey, "=", 2)
		if len(kvp) == 2 {
			sk := strings.TrimSpace(kvp[0])
			sv := strings.Trim(kvp[1], `"`)
			apply(sk, sv)
		} else if i == 0 {
			setName(strings.TrimSpace(kvp[0]))
		}
	}
}

// importDiscoverSection parses the [discover] section into Discover entries.
// Ported from upstream validate_hpcm_config.go importCfgDiscoverSection.
func importDiscoverSection(cfg *iniFile) map[string]cmconfig.Discover {
	discover := map[string]cmconfig.Discover{}
	for _, section := range cfg.sections {
		if section.name() != "discover" {
			continue
		}
		secName := discoverKeyName(section)
		if secName == "" {
			continue
		}
		for _, v := range section.valuesFor(secName) {
			d := cmconfig.Discover{}
			parseSubkeys(v,
				func(k, val string) { applyDiscoverField(&d, k, val) },
				func(name string) { d.Hostname1 = name },
			)
			if d.Hostname1 != "" {
				discover[d.Hostname1] = d
			}
		}
	}
	return discover
}

// discoverKeyName determines the primary key for [discover] entries.
// Different cm.config vintages use different key names.
func discoverKeyName(section *iniSection) string {
	for _, name := range []string{"hostname1", "internal_name", "alias1", "temponame"} {
		if section.hasKey(name) {
			return name
		}
	}
	return ""
}

// applyDiscoverField sets a single field on a Discover by key name.
func applyDiscoverField(d *cmconfig.Discover, key, val string) {
	switch key {
	case "hostname1":
		d.Hostname1 = val
	case "internal_name":
		d.InternalName = val
	case "template_name":
		d.TemplateName = val
	case "rack_nr":
		d.RackNr = atoiDefault(val, 0)
	case "chassis":
		d.Chassis = atoiDefault(val, 0)
	case "tray":
		d.Tray = atoiDefault(val, 0)
	case "node_nr":
		d.NodeNr = atoiDefault(val, 0)
	case "controller_nr":
		d.ControllerNr = atoiDefault(val, 0)
	case "alias_groups":
		d.AliasGroups = append(d.AliasGroups, val)
	case "card_type":
		d.CardType = val
	case "architecture":
		d.Architecture = val
	case "type":
		d.Type = val
	case "node_controller":
		d.NodeController = val
	case "image":
		d.Image = val
	case "kernel":
		d.Kernel = val
	case "mgmt_net_macs":
		d.MgmtNetMacs = appendUnique(d.MgmtNetMacs, splitTrim(val)...)
	case "mgmt_net_ip":
		d.MgmtNetIp = val
	case "mgmt_bmc_net_macs":
		d.MgmtBmcNetMacs = appendUnique(d.MgmtBmcNetMacs, splitTrim(val)...)
	case "mgmt_bmc_net_ip":
		d.MgmtBmcNetIp = appendUnique(d.MgmtBmcNetIp, splitTrim(val)...)
	case "mgmt_net_name":
		d.MgmtNetName = val
	case "cmcinventory_managed":
		d.CmcInventoryManaged = parseBool(val)
	}
}

// importTemplatesSection parses the [templates] section into Template entries.
// Ported from upstream importCfgTemplatesSection.
func importTemplatesSection(cfg *iniFile) map[string]cmconfig.Template {
	templates := map[string]cmconfig.Template{}
	for _, section := range cfg.sections {
		if section.name() != "templates" {
			continue
		}
		if !section.hasKey("name") {
			continue
		}
		for _, v := range section.valuesFor("name") {
			t := cmconfig.Template{}
			parseSubkeys(v,
				func(k, val string) { applyTemplateField(&t, k, val) },
				func(name string) { t.Name = name },
			)
			if t.Name != "" {
				templates[t.Name] = t
			}
		}
	}
	return templates
}

// applyTemplateField sets a single field on a Template by key name.
func applyTemplateField(t *cmconfig.Template, key, val string) {
	switch key {
	case "ctrl_model":
		t.CtrlModel = val
	case "card_type":
		t.CardType = val
	case "architecture":
		t.Architecture = val
	case "mgmt_net_interfaces":
		t.MgmtNetInterfaces = appendUnique(t.MgmtNetInterfaces, splitTrim(val)...)
	case "data1_net_name":
		t.Data1NetName = val
	case "data1_net_interfaces":
		t.Data1NetInterfaces = appendUnique(t.Data1NetInterfaces, splitTrim(val)...)
	case "data2_net_name":
		t.Data2NetName = val
	case "data2_net_interfaces":
		t.Data2NetInterfaces = appendUnique(t.Data2NetInterfaces, splitTrim(val)...)
	case "rootfs":
		t.RootFs = val
	case "transport":
		t.Transport = val
	case "username":
		t.Username = val
	case "image":
		t.Image = val
	}
}

// importNicTemplatesSection parses the [nic_templates] section.
// Ported from upstream importCfgNicTemplatesSection.
func importNicTemplatesSection(cfg *iniFile) map[string]cmconfig.NicTemplate {
	templates := map[string]cmconfig.NicTemplate{}
	for _, section := range cfg.sections {
		if section.name() != "nic_templates" {
			continue
		}
		if !section.hasKey(sectionKeyTemplate) {
			continue
		}
		for _, v := range section.valuesFor(sectionKeyTemplate) {
			n := cmconfig.NicTemplate{}
			parseSubkeys(v,
				func(k, val string) { applyNicTemplateField(&n, k, val) },
				func(name string) { n.Template = name },
			)
			if n.Template != "" && n.Network != "" {
				key := n.Template + "/" + n.Network
				templates[key] = n
			}
		}
	}
	return templates
}

// applyNicTemplateField sets a single field on a NicTemplate by key name.
func applyNicTemplateField(n *cmconfig.NicTemplate, key, val string) {
	switch key {
	case sectionKeyTemplate:
		n.Template = val
	case "network":
		n.Network = val
	case "bonding_master":
		n.BondingMaster = val
	case "bonding_mode":
		n.BondingMode = val
	case "net_ifs":
		n.NetIfs = appendUnique(n.NetIfs, splitTrim(val)...)
	case "br_name":
		n.BrName = val
	}
}

// --- helpers ---

// parseBool converts common bool-ish strings to bool.
func parseBool(val string) bool {
	switch strings.ToLower(val) {
	case "1", "t", "true", "yes", "y":
		return true
	}
	return false
}

// containS checks if a slice contains a string.
func containS(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

// appendUnique appends values to a slice, skipping duplicates.
func appendUnique(slice []string, vals ...string) []string {
	for _, v := range vals {
		if !containS(slice, v) {
			slice = append(slice, v)
		}
	}
	return slice
}

// splitTrim splits a comma-separated value and trims quotes from each part.
func splitTrim(val string) []string {
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// atoiDefault converts a string to int, returning def on error.
func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// parseAliasGroups splits entries like "key1:val1,key2:val2" into a map.
// Each element in raw is a comma-separated list of key:value pairs.
func parseAliasGroups(raw []string) map[string]string {
	result := make(map[string]string)
	for _, entry := range raw {
		if entry == "" {
			continue
		}
		groups := strings.Split(entry, ",")
		for _, g := range groups {
			parts := strings.SplitN(g, ":", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				if k != "" && v != "" {
					result[k] = v
				}
			}
		}
	}
	return result
}
