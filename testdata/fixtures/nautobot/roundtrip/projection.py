#
# MIT License
#
# (C) Copyright 2026 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#

"""Project Nautobot and a cani datastore onto one comparable shape.

Both projections produce {family: {natural key: {field: value}}}. A family is
only as wide as the rows pkg/devicetypes/NAUTOBOT_MAPPING.md marks **Mapped**
in both directions, so an equality between the two sides is a statement about
what cani promises to carry, not about everything Nautobot stores:

  location   §2.1   key name            parent, status, description, tags
  rack       §2.2   key name            location, status, u_height, comments, tags
  device     §2.3   key name            device_type, status, role, location, rack,
                                        position, face, serial, comments, tags,
                                        primary_ip4
  interface  §2a.1  key device|name     type, mac, label, description, status, tags
  cable      §2.5   key endpoints       label, status, color
  namespace  §2a    key name            (membership only: names some record uses)
  vrf        §2a.2  key namespace|name  rd, description, prefixes, devices, tags
  prefix     §2a.4  key namespace|cidr  status, type, role, description, vlan,
                                        locations, parent, tags
  ip         §2a.5  key namespace|addr  status, type, role, dns_name, description,
                                        parent, tags
  vlan       §2a.3  key vid|location    name, status, role, description, tags

A device's location is its rack's when it has one and its own otherwise, on
both sides. A VLAN is keyed the way MergeVLANs matches it, by VID within its
location, so its name is compared rather than relied on. Not projected on
purpose: location type keys
(nautobot_location_types_spec.sh owns the key<->display rule), IP->interface
assignments (import leaves them unset until FORGE-566), Cable.Type (export
derives it from CableType, CableCategory, ConnectorType or Slug, so a library
cable gains it only after a round trip), tenants and custom fields. Lists are
sorted so set-valued fields compare as values.
"""

from roundtrip.cani_source import is_set

GLOBAL = "Global"


def text(value):
    return "" if value is None else str(value)


def ref(obj, key="name"):
    """Name (or another key) of a nested Nautobot reference, "" when null."""
    return text((obj or {}).get(key))


def choice(value):
    """Value of a Nautobot choice field, which depth=1 renders as a dict."""
    return text(value.get("value")) if isinstance(value, dict) else text(value)


def tags(obj):
    """Sorted tag names of a Nautobot object, whose depth=1 tags carry names."""
    return sorted(ref(tag) for tag in obj.get("tags") or [])


def local_tags(record):
    """Sorted tag names of a datastore record."""
    return sorted(record.get("tags") or [])


def vlan_key(vid, locations):
    return f"{vid}|{','.join(locations)}"


# --- Nautobot ----------------------------------------------------------------

def project_nautobot(raw):
    interfaces = {i["id"]: f"{ref(i.get('device'))}|{i['name']}" for i in raw["interfaces"]}
    prefix_ns = {p["id"]: ref(p.get("namespace")) or GLOBAL for p in raw["prefixes"]}
    members = _group(raw["vrf_prefixes"], "vrf", "prefix", "prefix")
    assigned = _group(raw["vrf_devices"], "vrf", "device", "name")
    vlan_locs = _group(raw["vlan_locations"], "vlan", "location", "name")
    prefix_locs = _group(raw["prefix_locations"], "prefix", "location", "name")
    out = {
        "location": {loc["name"]: {
            "parent": ref(loc.get("parent")), "status": ref(loc.get("status")),
            "description": text(loc.get("description")), "tags": tags(loc),
        } for loc in raw["locations"]},
        "rack": {rack["name"]: {
            "location": ref(rack.get("location")), "status": ref(rack.get("status")),
            "u_height": rack.get("u_height"), "comments": text(rack.get("comments")), "tags": tags(rack),
        } for rack in raw["racks"]},
        "device": {dev["name"]: {
            "device_type": ref(dev.get("device_type"), "model"), "status": ref(dev.get("status")),
            "role": ref(dev.get("role")), "location": ref(dev.get("location")),
            "rack": ref(dev.get("rack")), "position": dev.get("position") or 0,
            "face": choice(dev.get("face")), "serial": text(dev.get("serial")),
            "comments": text(dev.get("comments")), "tags": tags(dev),
            "primary_ip4": ref(dev.get("primary_ip4"), "address"),
        } for dev in raw["devices"]},
        "interface": {interfaces[i["id"]]: {
            "type": choice(i.get("type")), "mac": text(i.get("mac_address")).lower(),
            "label": text(i.get("label")), "description": text(i.get("description")),
            "status": ref(i.get("status")), "tags": tags(i),
        } for i in raw["interfaces"]},
        "cable": {_endpoints(interfaces.get(c.get("termination_a_id")), interfaces.get(c.get("termination_b_id"))): {
            "label": text(c.get("label")), "status": ref(c.get("status")), "color": text(c.get("color")),
        } for c in raw["cables"]},
        "vrf": {f"{ref(v.get('namespace')) or GLOBAL}|{v['name']}": {
            "rd": text(v.get("rd")), "description": text(v.get("description")),
            "prefixes": members.get(v["id"], []), "devices": assigned.get(v["id"], []), "tags": tags(v),
        } for v in raw["vrfs"]},
        "prefix": {f"{prefix_ns[p['id']]}|{p['prefix']}": {
            "status": ref(p.get("status")), "type": choice(p.get("type")), "role": ref(p.get("role")),
            "description": text(p.get("description")),
            "vlan": vlan_key(p["vlan"]["vid"], vlan_locs.get(p["vlan"]["id"], [])) if p.get("vlan") else "",
            "locations": prefix_locs.get(p["id"], []), "parent": ref(p.get("parent"), "prefix"), "tags": tags(p),
        } for p in raw["prefixes"]},
        "ip": {f"{prefix_ns.get(ref(ip.get('parent'), 'id'), GLOBAL)}|{ip['address']}": {
            "status": ref(ip.get("status")), "type": choice(ip.get("type")), "role": ref(ip.get("role")),
            "dns_name": text(ip.get("dns_name")), "description": text(ip.get("description")),
            "parent": ref(ip.get("parent"), "prefix"), "tags": tags(ip),
        } for ip in raw["ip_addresses"]},
        "vlan": {vlan_key(v["vid"], vlan_locs.get(v["id"], [])): {
            "name": text(v.get("name")), "status": ref(v.get("status")), "role": ref(v.get("role")),
            "description": text(v.get("description")), "tags": tags(v),
        } for v in raw["vlans"]},
    }
    out["namespace"] = _namespaces(out)
    return out


def _group(assignments, owner, member, member_key):
    """{owner id: sorted [member key]} from an assignment table."""
    grouped = {}
    for row in assignments:
        grouped.setdefault(ref(row.get(owner), "id"), []).append(ref(row.get(member), member_key))
    return {owner_id: sorted(values) for owner_id, values in grouped.items()}


def _endpoints(a, b):
    return " <-> ".join(sorted([text(a), text(b)]))


def _namespaces(out):
    """Namespaces are modelled only through the records that name them."""
    names = {key.split("|", 1)[0] for family in ("vrf", "prefix", "ip") for key in out[family]}
    return {name: {} for name in sorted(names)}


# --- cani --------------------------------------------------------------------

def project_cani(inv):
    loc_name = _namer(inv["locations"], "name")
    dev_name = _namer(inv["devices"], "name")
    vlans = inv["vlans"]
    prefixes = inv["prefixes"]

    def rack_of(dev):
        return inv["racks"].get(dev.get("parent") or "", {})

    def ns(record):
        return record.get("namespace") or GLOBAL

    def parent_cidr(record):
        return text(prefixes.get(record.get("parent") or "", {}).get("prefix"))

    def scope(record):
        return [loc_name(record["location"])] if is_set(record.get("location")) else []

    def vlan_of(prefix):
        vlan = vlans.get(prefix.get("vlan") or "")
        return vlan_key(vlan["vid"], scope(vlan)) if vlan else ""

    def endpoint(cable, side):
        return f"{dev_name(cable.get(f'termination{side}Device'))}|{cable.get(f'termination{side}Port')}"

    def primary_ip4(dev):
        return text(inv["ip_addresses"].get(dev.get("primaryIpv4") or "", {}).get("address"))

    out = {
        "location": {loc["name"]: {
            "parent": loc_name(loc.get("parent")), "status": text(loc.get("status")),
            "description": text(loc.get("description")), "tags": local_tags(loc),
        } for loc in inv["locations"].values()},
        "rack": {rack["name"]: {
            "location": loc_name(rack.get("location")), "status": text(rack.get("status")),
            "u_height": rack.get("uHeight"), "comments": text(rack.get("comments")), "tags": local_tags(rack),
        } for rack in inv["racks"].values()},
        "device": {dev["name"]: {
            "device_type": text(dev.get("model")), "status": text(dev.get("status")),
            "role": text(dev.get("role")),
            "location": loc_name(rack_of(dev).get("location") or dev.get("location")),
            "rack": text(rack_of(dev).get("name")), "position": dev.get("rackPosition") or 0,
            "face": text(dev.get("face")), "serial": text(dev.get("serial")),
            "comments": text(dev.get("comments")), "tags": local_tags(dev),
            "primary_ip4": primary_ip4(dev),
        } for dev in inv["devices"].values()},
        "interface": {f"{dev_name(i.get('deviceId'))}|{i['name']}": {
            "type": text(i.get("interfaceType")), "mac": text(i.get("macAddress")).lower(),
            "label": text(i.get("label")), "description": text(i.get("description")),
            "status": text(i.get("status")), "tags": local_tags(i),
        } for i in inv["interfaces"].values()},
        "cable": {_endpoints(endpoint(c, "A"), endpoint(c, "B")): {
            "label": text(c.get("label")), "status": text(c.get("status")), "color": text(c.get("color")),
        } for c in inv["cables"].values()},
        "vrf": {f"{ns(v)}|{v['name']}": {
            "rd": text(v.get("rd")), "description": text(v.get("description")),
            "prefixes": sorted(p["prefix"] for p in prefixes.values() if vrf_id in (p.get("vrfs") or [])),
            "devices": sorted(dev_name(d) for d in v.get("devices") or []), "tags": local_tags(v),
        } for vrf_id, v in inv["vrfs"].items()},
        "prefix": {f"{ns(p)}|{p['prefix']}": {
            "status": text(p.get("status")), "type": text(p.get("type")), "role": text(p.get("role")),
            "description": text(p.get("description")), "vlan": vlan_of(p),
            "locations": scope(p), "parent": parent_cidr(p), "tags": local_tags(p),
        } for p in prefixes.values()},
        "ip": {f"{ns(ip)}|{ip['address']}": {
            "status": text(ip.get("status")), "type": text(ip.get("type")), "role": text(ip.get("ipRole")),
            "dns_name": text(ip.get("dnsName")), "description": text(ip.get("description")),
            "parent": parent_cidr(ip), "tags": local_tags(ip),
        } for ip in inv["ip_addresses"].values()},
        "vlan": {vlan_key(v["vid"], scope(v)): {
            "name": text(v.get("name")), "status": text(v.get("status")), "role": text(v.get("role")),
            "description": text(v.get("description")), "tags": local_tags(v),
        } for v in vlans.values()},
    }
    out["namespace"] = _namespaces(out)
    return out


def _namer(records, field):
    """UUID -> record[field], "" for nil/unknown references."""
    return lambda uuid_value: text(records.get(uuid_value, {}).get(field)) if is_set(uuid_value) else ""


# --- comparison --------------------------------------------------------------

def diff(a, b):
    """Human-readable differences between two projections; empty when equal."""
    lines = []
    for family in sorted(set(a) | set(b)):
        left, right = a.get(family, {}), b.get(family, {})
        for key in sorted(set(left) | set(right)):
            if key not in right:
                lines.append(f"{family}/{key}: only in A")
            elif key not in left:
                lines.append(f"{family}/{key}: only in B")
            else:
                lines.extend(f"{family}/{key}: {field}: {left[key].get(field)!r} != {right[key].get(field)!r}"
                             for field in sorted(set(left[key]) | set(right[key]))
                             if left[key].get(field) != right[key].get(field))
    return lines
