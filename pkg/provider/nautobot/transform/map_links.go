/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package transform

import (
	"reflect"
	"sort"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
	"github.com/google/uuid"
)

// BuildTagNameMap indexes the tag catalog by UUID so the tag references on
// every other object resolve to names.
func BuildTagNameMap(tags []nautobotapi.Tag) map[uuid.UUID]string {
	names := make(map[uuid.UUID]string, len(tags))
	for _, tag := range tags {
		names[directUUID(tag.Id)] = tag.Name
	}
	return names
}

// tagNames resolves a generated `Tags *[]struct{Id ...}` slice to the sorted
// names of the tags the catalog knows; nil when there are none.
func tagNames(refs any, names map[uuid.UUID]string) []string {
	v := reflect.ValueOf(refs)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return nil
	}
	var out []string
	for i := 0; i < v.Len(); i++ {
		if name, ok := names[refID(v.Index(i).Interface())]; ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// LinkTags copies each raw object's tags onto the mapped record that carries
// its Nautobot UUID, for every family the round trip compares. Mappers keep
// their signatures; the link runs once over the finished result.
func LinkTags(result *devicetypes.TransformResult, raw *RawData) {
	names := BuildTagNameMap(raw.Tags)
	linkTags(raw.Locations, result.Locations, names,
		func(l *nautobotapi.Location) (uuid.UUID, any) { return directUUID(l.Id), l.Tags },
		func(c *devicetypes.CaniLocationType) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkTags(raw.Racks, result.Racks, names,
		func(r *nautobotapi.Rack) (uuid.UUID, any) { return directUUID(r.Id), r.Tags },
		func(c *devicetypes.CaniRackType) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkTags(raw.Devices, result.Devices, names,
		func(d *nautobotapi.Device) (uuid.UUID, any) { return directUUID(d.Id), d.Tags },
		func(c *devicetypes.CaniDeviceType) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkTags(raw.VLANs, result.VLANs, names,
		func(v *nautobotapi.VLAN) (uuid.UUID, any) { return directUUID(v.Id), v.Tags },
		func(c *devicetypes.CaniVLAN) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkTags(raw.Prefixes, result.Prefixes, names,
		func(p *nautobotapi.Prefix) (uuid.UUID, any) { return directUUID(p.Id), p.Tags },
		func(c *devicetypes.CaniPrefix) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkTags(raw.IPAddresses, result.IPAddresses, names,
		func(a *nautobotapi.IPAddress) (uuid.UUID, any) { return directUUID(a.Id), a.Tags },
		func(c *devicetypes.CaniIPAddress) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkTags(raw.VRFs, result.VRFs, names,
		func(v *nautobotapi.VRF) (uuid.UUID, any) { return directUUID(v.Id), v.Tags },
		func(c *devicetypes.CaniVRF) *devicetypes.ObjectMeta { return &c.ObjectMeta })
	linkInterfaceTags(raw.Interfaces, result.Devices, names)
}

// linkTags sets Tags on every record of one family from the raw object with
// the same Nautobot UUID.
func linkTags[R any, C any](
	raws []R,
	records map[uuid.UUID]*C,
	names map[uuid.UUID]string,
	fields func(*R) (uuid.UUID, any),
	meta func(*C) *devicetypes.ObjectMeta,
) {
	index := nautobotIDIndex(records, func(c *C) map[string]uuid.UUID { return meta(c).ExternalIDs })
	for i := range raws {
		nbID, refs := fields(&raws[i])
		if record := records[index[nbID]]; record != nil {
			meta(record).Tags = tagNames(refs, names)
		}
	}
}

// linkInterfaceTags sets Tags on the device interface specs, which are
// persisted on the device rather than as records of their own.
func linkInterfaceTags(raws []nautobotapi.Interface, devices map[uuid.UUID]*devicetypes.CaniDeviceType, names map[uuid.UUID]string) {
	index := nautobotIDIndex(devices, func(d *devicetypes.CaniDeviceType) map[string]uuid.UUID { return d.ExternalIDs })
	for i := range raws {
		device := devices[index[refID(raws[i].Device)]]
		if device == nil {
			continue
		}
		ifaceID := directUUID(raws[i].Id)
		for s := range device.Interfaces {
			if device.Interfaces[s].ID == ifaceID {
				device.Interfaces[s].Tags = tagNames(raws[i].Tags, names)
			}
		}
	}
}

// LinkPrimaryIPs sets each device's primary IPv4 and IPv6 from the raw device,
// resolved through the imported addresses' Nautobot UUIDs. An address the
// import left out leaves the field unset rather than dangling.
func LinkPrimaryIPs(result *devicetypes.TransformResult, raw []nautobotapi.Device) {
	devices := nautobotIDIndex(result.Devices, func(d *devicetypes.CaniDeviceType) map[string]uuid.UUID { return d.ExternalIDs })
	addresses := nautobotIDIndex(result.IPAddresses, addressExternalIDs)
	for i := range raw {
		device := result.Devices[devices[directUUID(raw[i].Id)]]
		if device == nil {
			continue
		}
		if id, ok := addresses[refID(raw[i].PrimaryIp4)]; ok {
			device.PrimaryIPv4 = id
		}
		if id, ok := addresses[refID(raw[i].PrimaryIp6)]; ok {
			device.PrimaryIPv6 = id
		}
	}
}
