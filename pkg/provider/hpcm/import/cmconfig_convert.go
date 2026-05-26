package import_

import (
	"log"
	"strings"

	"github.com/Cray-HPE/cani/pkg/provider/hpcm/import/cmconfig"
)

// CmConfigToNodes converts parsed cm.config discover entries into Nodes
// suitable for the existing classify -> transform pipeline.
func CmConfigToNodes(cfg *CmConfig) []Node {
	nodes := make([]Node, 0, len(cfg.Discover))
	for _, d := range cfg.Discover {
		n := discoverToNode(d, cfg)
		nodes = append(nodes, n)
	}
	log.Printf("Converted %d cm.config discover entries to nodes", len(nodes))
	return nodes
}

// discoverToNode converts a single Discover entry into a Node.
func discoverToNode(d cmconfig.Discover, cfg *CmConfig) Node {
	return Node{
		Name:           d.Hostname1,
		InternalName:   d.InternalName,
		TemplateName:   d.TemplateName,
		Type:           inferNodeType(d),
		NodeController: d.NodeController,
		Aliases:        buildAliases(d, cfg),
		Location:       buildLocation(d),
	}
}

// inferNodeType determines the CANI-compatible node type from discover fields.
func inferNodeType(d cmconfig.Discover) string {
	switch strings.ToLower(d.Type) {
	case "spine", "leaf":
		return "mgmt_switch"
	}
	if strings.ToLower(d.InternalName) == "admin" {
		return "admin"
	}
	if strings.HasPrefix(strings.ToLower(d.InternalName), "mgmtsw") {
		return "mgmt_switch"
	}
	return "compute"
}

// buildAliases constructs the aliases map from discover alias_groups
// and enriches with template ctrl_model and template_name.
func buildAliases(d cmconfig.Discover, cfg *CmConfig) map[string]string {
	aliases := parseAliasGroups(d.AliasGroups)

	// Inject ctrl_model from the referenced template.
	if d.TemplateName != "" {
		if tpl, ok := cfg.Templates[d.TemplateName]; ok {
			if tpl.CtrlModel != "" {
				aliases["ctrl_model"] = tpl.CtrlModel
			}
		}
	}

	// Inject template_name as an alias for lookup fallback.
	if d.TemplateName != "" {
		aliases["template_name"] = d.TemplateName
	}

	// Inject card_type for vendor resolution.
	if d.CardType != "" {
		aliases["card_type"] = d.CardType
	}

	if len(aliases) == 0 {
		return nil
	}
	return aliases
}

// buildLocation constructs a LocationSettings from discover fields.
// Returns nil if no location fields are set (standalone nodes).
func buildLocation(d cmconfig.Discover) *LocationSettings {
	hasRack := d.RackNr != 0
	hasChassis := d.Chassis != 0 || d.CmcInventoryManaged
	hasTray := d.Tray != 0
	hasNode := d.NodeNr != 0
	hasCtrl := d.ControllerNr != 0

	// If nothing at all is set, return nil.
	if !hasRack && !hasChassis && !hasTray && !hasNode && !hasCtrl {
		return nil
	}

	loc := &LocationSettings{}
	if hasRack {
		loc.Rack = Int32Ptr(int32(d.RackNr))
	}
	if hasChassis {
		loc.Chassis = Int32Ptr(int32(d.Chassis))
	}
	// Tray and node: set even if 0 when deeper fields exist,
	// because tray=0 is a valid tray number.
	if hasTray || hasNode || hasCtrl {
		loc.Tray = Int32Ptr(int32(d.Tray))
	}
	// Node: set if hasTray (node_nr=0 is valid within a tray) or hasNode/hasCtrl.
	if hasTray || hasNode || hasCtrl {
		loc.Node = Int32Ptr(int32(d.NodeNr))
	}
	if hasCtrl {
		loc.Controller = Int32Ptr(int32(d.ControllerNr))
	}
	return loc
}
