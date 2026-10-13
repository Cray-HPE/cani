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

"""Seed the deterministic fixture the import-first round-trip specs start from.

DCIM: location test-dc, rack rack-01, compute-001 (tagged cani-rt-tag) and
compute-002 (one eth0 each) joined by one cable; location test-edge holding
edge-001 with no rack. IPAM: namespaces tenant-a and tenant-b that each hold
10.0.0.0/24 and host 10.0.0.5/24; VRF "blue" in both tenants, assigned to
tenant-a's prefix and to compute-001; tenant-a's host is assigned to
compute-001 eth0 and is its primary IPv4; VLAN 3100 and 10.31.0.0/24 in
Global scoped to test-dc. Run after `reset` so every POST creates exactly one
object.
"""

from seed_nautobot import (
    create_cable,
    create_interface,
    get_or_create_location_type,
    get_or_create_manufacturer,
    get_or_create_role,
    get_status_id,
)
from seed_nautobot_api import api, find_or_create
from seed_nautobot_ipam import ensure_content_types

LOCATION = "test-dc"
EDGE_LOCATION = "test-edge"
RACK = "rack-01"
DEVICE_TYPE = "ProLiant DL325 Gen11 8SFF"
TAG = "cani-rt-tag"
TENANTS = ("tenant-a", "tenant-b")
SHARED_PREFIX = "10.0.0.0/24"
SHARED_HOST = "10.0.0.5/24"

ACTIVE_CONTENT_TYPES = [
    "dcim.device", "dcim.interface", "dcim.location", "dcim.rack",
    "ipam.ipaddress", "ipam.prefix", "ipam.vlan", "ipam.vrf",
]


def ref(obj_id):
    return {"id": obj_id}


def seed_dcim(status_id):
    """Create the locations, rack, three devices, their interfaces and a cable.

    Returns the test-dc location, the devices by name and compute-001's eth0.
    """
    location_type_id = get_or_create_location_type("Site")
    ensure_content_types("dcim/location-types/", location_type_id, ["ipam.prefix", "ipam.vlan"])
    location = find_or_create("dcim/locations/", "name", LOCATION, {
        "name": LOCATION, "status": ref(status_id), "location_type": ref(location_type_id),
    })
    edge = find_or_create("dcim/locations/", "name", EDGE_LOCATION, {
        "name": EDGE_LOCATION, "status": ref(status_id), "location_type": ref(location_type_id),
    })
    device_type = find_or_create("dcim/device-types/", "model", DEVICE_TYPE, {
        "manufacturer": ref(get_or_create_manufacturer("HPE")),
        "model": DEVICE_TYPE, "u_height": 1, "is_full_depth": True,
    })
    tag = find_or_create("extras/tags/", "name", TAG, {"name": TAG, "content_types": ["dcim.device"]})
    rack = api("POST", "dcim/racks/", {
        "name": RACK, "status": ref(status_id), "location": ref(location["id"]),
        "u_height": 42, "comments": "primary test rack",
    })
    role_id = get_or_create_role("Compute")
    get_or_create_role("Network")
    common = {"device_type": ref(device_type["id"]), "role": ref(role_id), "status": ref(status_id)}
    devices = {}
    for position in (1, 2):
        name = f"compute-{position:03d}"
        devices[name] = api("POST", "dcim/devices/", dict(common, **{
            "name": name, "location": ref(location["id"]), "rack": ref(rack["id"]),
            "position": position, "face": "front", "serial": f"SN-{position:03d}",
            "comments": "primary compute node" if position == 1 else "compute node 2",
            "tags": [ref(tag["id"])] if position == 1 else [],
        }))
    # A device placed directly in a location, as Nautobot allows.
    devices["edge-001"] = api("POST", "dcim/devices/", dict(common, **{
        "name": "edge-001", "location": ref(edge["id"]), "serial": "SN-EDGE", "comments": "edge node",
    }))
    eth0 = {name: create_interface(dev["id"], "eth0", status_id) for name, dev in devices.items()}
    create_cable(eth0["compute-001"]["id"], eth0["compute-002"]["id"], get_status_id("Connected", ["dcim.cable"]))
    return location, devices, eth0["compute-001"]


def seed_tenants(status_id, compute_001_id, compute_001_eth0_id):
    """Create the overlapping tenant networks, the VRF memberships and the
    primary address of compute-001."""
    for index, tenant in enumerate(TENANTS, start=1):
        namespace = api("POST", "ipam/namespaces/", {"name": tenant})
        vrf = api("POST", "ipam/vrfs/", {
            "name": "blue", "namespace": ref(namespace["id"]),
            "rd": f"65000:{index}", "description": f"blue VRF in {tenant}",
        })
        prefix = api("POST", "ipam/prefixes/", {
            "prefix": SHARED_PREFIX, "namespace": ref(namespace["id"]),
            "status": ref(status_id), "type": "network", "description": f"network in {tenant}",
        })
        host = api("POST", "ipam/ip-addresses/", {
            "address": SHARED_HOST, "namespace": ref(namespace["id"]), "status": ref(status_id),
            "type": "host", "dns_name": f"host.{tenant}.example.test", "description": f"host in {tenant}",
        })
        if tenant == "tenant-a":
            api("POST", "ipam/vrf-prefix-assignments/", {"vrf": ref(vrf["id"]), "prefix": ref(prefix["id"])})
            api("POST", "ipam/vrf-device-assignments/", {"vrf": ref(vrf["id"]), "device": ref(compute_001_id)})
            api("POST", "ipam/ip-address-to-interface/", {
                "ip_address": ref(host["id"]), "interface": ref(compute_001_eth0_id),
            })
            api("PATCH", f"dcim/devices/{compute_001_id}/", {"primary_ip4": ref(host["id"])})


def seed_global_scope(status_id, location_id):
    """Create a location-scoped VLAN and prefix in the Global namespace."""
    namespace = find_or_create("ipam/namespaces/", "name", "Global", {"name": "Global"})
    vlan = api("POST", "ipam/vlans/", {"vid": 3100, "name": "cani-seeded-vlan", "status": ref(status_id)})
    api("POST", "ipam/vlan-location-assignments/", {"location": ref(location_id), "vlan": ref(vlan["id"])})
    prefix = api("POST", "ipam/prefixes/", {
        "prefix": "10.31.0.0/24", "namespace": ref(namespace["id"]), "status": ref(status_id),
        "type": "network", "vlan": ref(vlan["id"]),
    })
    api("POST", "ipam/prefix-location-assignments/", {"location": ref(location_id), "prefix": ref(prefix["id"])})


def seed():
    status_id = get_status_id("Active", ACTIVE_CONTENT_TYPES)
    ensure_content_types("extras/statuses/", status_id, ACTIVE_CONTENT_TYPES)
    location, devices, compute_001_eth0 = seed_dcim(status_id)
    seed_tenants(status_id, devices["compute-001"]["id"], compute_001_eth0["id"])
    seed_global_scope(status_id, location["id"])
    print("Seed complete: 2 locations, 1 rack, 3 devices, 1 cable, 2 tenants, 3 prefixes, 2 addresses, 2 VRFs, 1 VLAN")


if __name__ == "__main__":
    seed()
