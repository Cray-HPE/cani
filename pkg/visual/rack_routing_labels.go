/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024-2026 Hewlett Packard Enterprise Development LP
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
package visual

import (
	"fmt"
	"sort"
	"strings"
)

// connEndpoint describes one cable's contribution at a specific U.
type connEndpoint struct {
	localPort  string // abbreviated port at this U
	remotePort string // abbreviated port at other end
	remoteU    int    // remote U position (always set)
	remoteRack string // non-empty for inter-rack
	group      cableGroup
}

// collectAnnotationEndpoints gathers the non-dimmed cable endpoints that should
// be annotated at row u. Inter-rack cables are only annotated at the local
// device's U position.
func collectAnnotationEndpoints(u int, cables []routingCable) []connEndpoint {
	var endpoints []connEndpoint
	for _, c := range cables {
		if c.dimmed {
			continue
		}
		if u != c.topU && u != c.botU {
			continue
		}
		// For inter-rack cables, only annotate at the local device's U.
		if c.interRack && u != c.localU {
			continue
		}
		var ep connEndpoint
		ep.group = c.group
		if u == c.topU {
			ep.localPort = abbreviatePort(c.portAtTop)
			ep.remotePort = abbreviatePort(c.portAtBot)
			ep.remoteU = c.botU
		} else {
			ep.localPort = abbreviatePort(c.portAtBot)
			ep.remotePort = abbreviatePort(c.portAtTop)
			ep.remoteU = c.topU
		}
		if c.interRack {
			ep.remoteRack = abbreviateRack(c.remoteRack)
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints
}

// buildEndpointAnnotation produces a compact annotation for cables at row u.
// Intra-rack:  →U(localPort:remotePort, ...)
// Inter-rack:  ⇢rack(localPort:remotePort, ...)
func buildEndpointAnnotation(u int, cables []routingCable) string {
	endpoints := collectAnnotationEndpoints(u, cables)
	if len(endpoints) == 0 {
		return ""
	}
	return formatEndpoints(endpoints, colorFuncs{})
}

// buildColoredAnnotation is like buildEndpointAnnotation but applies ANSI
// colors: intra-rack labels use the cable group color, inter-rack labels
// are dimmed gray, and local U references are bold.
func buildColoredAnnotation(u int, cables []routingCable, cf colorFuncs) string {
	endpoints := collectAnnotationEndpoints(u, cables)
	if len(endpoints) == 0 {
		return ""
	}
	return formatEndpoints(endpoints, cf)
}

// uHasEndpoint returns true if any non-dimmed cable starts or ends at u.
func uHasEndpoint(u int, cables []routingCable) bool {
	for _, c := range cables {
		if c.dimmed {
			continue
		}
		if u == c.topU || u == c.botU {
			return true
		}
	}
	return false
}

// groupKey identifies a unique destination for grouping cables.
type groupKey struct {
	remoteU    int
	remoteRack string     // empty = intra-rack
	group      cableGroup // semantic classification for coloring
}

// destGroup collects the port pairs that share a single destination key.
type destGroup struct {
	key   groupKey
	pairs []portPair
}

// formatEndpoints groups endpoints by destination and formats them.
// When cf has color functions (non-zero), labels are colorized:
// intra-rack uses the cable group color, inter-rack is dimmed gray,
// and local U numbers are bold.
func formatEndpoints(eps []connEndpoint, cf colorFuncs) string {
	groups, order := groupEndpointsByDest(eps)
	sortDestKeys(order)
	hasColor := cf.bold != nil
	var parts []string
	for _, k := range order {
		parts = append(parts, formatDestLabel(k, groups[k].pairs, cf, hasColor))
	}
	return strings.Join(parts, " ")
}

// groupEndpointsByDest buckets endpoints by destination, preserving the
// first-seen order of each unique destination key.
func groupEndpointsByDest(eps []connEndpoint) (map[groupKey]*destGroup, []groupKey) {
	groups := make(map[groupKey]*destGroup)
	var order []groupKey
	for _, ep := range eps {
		k := groupKey{remoteRack: ep.remoteRack, group: ep.group}
		if ep.remoteRack == "" {
			k.remoteU = ep.remoteU // only group by U for intra-rack
		}
		g, ok := groups[k]
		if !ok {
			g = &destGroup{key: k}
			groups[k] = g
			order = append(order, k)
		}
		g.pairs = append(g.pairs, portPair{local: ep.localPort, remote: ep.remotePort})
	}
	return groups, order
}

// sortDestKeys orders destinations: intra-rack first (by U desc), then
// inter-rack (by rack name asc).
func sortDestKeys(order []groupKey) {
	sort.Slice(order, func(i, j int) bool {
		iInter := order[i].remoteRack != ""
		jInter := order[j].remoteRack != ""
		if iInter != jInter {
			return !iInter
		}
		if iInter {
			return order[i].remoteRack < order[j].remoteRack
		}
		return order[i].remoteU > order[j].remoteU
	})
}

// formatDestLabel formats one destination group's label.
func formatDestLabel(k groupKey, pairs []portPair, cf colorFuncs, hasColor bool) string {
	compressed := compressPortPairs(pairs)
	if k.remoteRack != "" {
		return formatInterRackLabel(k.remoteRack, compressed, cf, hasColor)
	}
	return formatIntraRackLabel(k, compressed, cf, hasColor)
}

// formatInterRackLabel renders a dimmed inter-rack destination label.
func formatInterRackLabel(rack, compressed string, cf colorFuncs, hasColor bool) string {
	label := fmt.Sprintf("⇢%s(%s)", rack, compressed)
	if hasColor {
		label = cf.gray(label)
	}
	return label
}

// formatIntraRackLabel renders a colorized intra-rack destination label.
func formatIntraRackLabel(k groupKey, compressed string, cf colorFuncs, hasColor bool) string {
	if !hasColor {
		return fmt.Sprintf("→%d(%s)", k.remoteU, compressed)
	}
	color := groupColorKey(k.group)
	uStr := cf.bold(cf.colorize(fmt.Sprintf("%d", k.remoteU), color))
	body := cf.colorize(fmt.Sprintf("(%s)", compressed), color)
	return "→" + uStr + body
}

// printRoutingLegend prints the symbol and cable legend.
func printRoutingLegend(opts CompactRenderOptions) {
	cf := newColorFuncs(opts.NoColor)
	fmt.Println("  Symbols: " +
		cf.green("D") + "=device " +
		cf.green("d") + "=half " +
		cf.green("M") + "=modules " +
		cf.green("*") + "=cont " +
		cf.gray("·") + "=empty")
	fmt.Println("  Cables:  " +
		cf.yellow("■") + "=mgmt " +
		cf.green("■") + "=hsn " +
		cf.magenta("■") + "=net " +
		cf.gray("■") + "=dimmed")
	fmt.Println("  Glyphs:  " +
		"┐┘=intra ┌│┘=outgoing ┐│┘=incoming │=route ─=same-U")
	fmt.Println("  Labels:  " +
		"→U(local:remote)=intra  ⇢rack(local:remote)=inter")
	if opts.Verbose >= 2 {
		fmt.Println("  (showing all cables including intra-rack)")
	} else {
		fmt.Println("  (inter-rack only; -VV for all)")
	}
	fmt.Println()
}
