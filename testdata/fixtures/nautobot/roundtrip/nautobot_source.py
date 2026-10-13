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

"""Read every collection the projection needs from the Nautobot REST API."""

import urllib.parse

from seed_nautobot_api import api

PAGE_SIZE = 200

# Collection name -> list endpoint. depth=1 nests one level of related objects
# (name, model, vid, prefix) so the projection never needs a second request.
COLLECTIONS = {
    "locations": "dcim/locations/",
    "racks": "dcim/racks/",
    "devices": "dcim/devices/",
    "interfaces": "dcim/interfaces/",
    "cables": "dcim/cables/",
    "namespaces": "ipam/namespaces/",
    "vrfs": "ipam/vrfs/",
    "prefixes": "ipam/prefixes/",
    "ip_addresses": "ipam/ip-addresses/",
    "vlans": "ipam/vlans/",
    "vlan_locations": "ipam/vlan-location-assignments/",
    "prefix_locations": "ipam/prefix-location-assignments/",
    "vrf_prefixes": "ipam/vrf-prefix-assignments/",
    "vrf_devices": "ipam/vrf-device-assignments/",
}


def _relative(next_url):
    """Turn the absolute `next` link Nautobot returns into an api() path."""
    parsed = urllib.parse.urlsplit(next_url)
    path = parsed.path.split("/api/", 1)[1]
    return f"{path}?{parsed.query}" if parsed.query else path


def fetch_all(path):
    """Return every result of a list endpoint, following pagination."""
    results = []
    page = api("GET", f"{path}?depth=1&limit={PAGE_SIZE}")
    while True:
        results.extend(page.get("results", []))
        if not page.get("next"):
            return results
        page = api("GET", _relative(page["next"]))


def fetch_nautobot():
    return {name: fetch_all(path) for name, path in COLLECTIONS.items()}
