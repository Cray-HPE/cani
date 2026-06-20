package transform

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/devicetypes/connections"
	import_ "github.com/Cray-HPE/cani/pkg/provider/example/import"
)

// transformSystemConnections resolves connection records into cables.
func transformSystemConnections(data *import_.SystemCSV, result *devicetypes.TransformResult, inv *devicetypes.Inventory) error {
	if len(data.Connections) == 0 {
		return nil
	}

	// Build a ConnectionMap from system CSV connection records
	cm := &connections.ConnectionMap{
		Version: "v1",
	}

	// Determine defaults from _defaults and connection_defaults
	globalDefaults := data.Defaults
	sectionDefaults, hasSectionDefaults := data.SectionDefaults["connection"]
	if globalDefaults.Status != "" || hasSectionDefaults {
		cd := &connections.CableDefaults{}
		if globalDefaults.Status != "" {
			cd.Status = globalDefaults.Status
		}
		if hasSectionDefaults {
			if sectionDefaults.Status != "" {
				cd.Status = sectionDefaults.Status
			}
			if sectionDefaults.Color != "" {
				cd.Color = sectionDefaults.Color
			}
			if sectionDefaults.LengthUnit != "" {
				cd.LengthUnit = sectionDefaults.LengthUnit
			}
		}
		cm.CableDefaults = cd
	}

	for _, rec := range data.Connections {
		rec = data.ApplyDefaults(rec)

		entry := connections.ConnectionEntry{
			A: connections.Endpoint{Device: rec.ADevice, Port: rec.APort},
			B: connections.Endpoint{Device: rec.BDevice, Port: rec.BPort},
		}

		cable := &connections.CableProps{}
		hasCableProps := false

		if rec.PartNumber != "" {
			cable.Type = rec.PartNumber
			hasCableProps = true
		}
		if rec.Color != "" {
			cable.Color = rec.Color
			hasCableProps = true
		}
		if rec.Length != "" {
			if l, err := strconv.ParseFloat(rec.Length, 64); err == nil {
				cable.Length = &l
				hasCableProps = true
			}
		}
		if rec.LengthUnit != "" {
			cable.LengthUnit = rec.LengthUnit
			hasCableProps = true
		}
		if rec.Status != "" {
			cable.Status = rec.Status
			hasCableProps = true
		}

		if hasCableProps {
			entry.Cable = cable
		}

		cm.Connections = append(cm.Connections, entry)
	}

	// Resolve patterns and device names
	resolved, errs := connections.ResolveConnectionMap(cm, inv)
	if len(errs) > 0 {
		for _, e := range errs {
			log.Printf("WARN: connection resolution: %v", e)
		}
		if len(resolved) == 0 {
			return fmt.Errorf("no connections resolved; %d errors", len(errs))
		}
	}

	// Create cable objects
	for _, conn := range resolved {
		cable := devicetypes.NewCable(conn.Cable.Type, conn.Cable.Label)
		cable.TerminationADevice = conn.ADevice
		cable.TerminationAPort = conn.APort
		cable.TerminationBDevice = conn.BDevice
		cable.TerminationBPort = conn.BPort

		if conn.Cable.Color != "" {
			cable.Color = conn.Cable.Color
		}
		if conn.Cable.Length != nil {
			cable.Length = conn.Cable.Length
		}
		if conn.Cable.LengthUnit != "" {
			cable.LengthUnit = conn.Cable.LengthUnit
		}
		if conn.Cable.Status != "" {
			cable.Status = conn.Cable.Status
		}

		result.Cables[cable.ID] = cable
	}

	return nil
}
