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

"""Read a canidb.json datastore into the shape the projection consumes."""

import json

NIL_UUID = "00000000-0000-0000-0000-000000000000"

# Datastore section -> projection collection; every section is a map keyed by UUID.
SECTIONS = {
    "locations": "locations",
    "racks": "racks",
    "devices": "devices",
    "interfaces": "interfaces",
    "cables": "cables",
    "prefixes": "prefixes",
    "ipAddresses": "ip_addresses",
    "vlans": "vlans",
    "vrfs": "vrfs",
}


def load_cani(path):
    """Return {collection: {uuid: record}} with absent sections as empty maps."""
    with open(path, encoding="utf-8") as src:
        inventory = json.load(src)
    return {name: dict(inventory.get(section) or {}) for section, name in SECTIONS.items()}


def is_set(uuid_value):
    """True for a UUID reference that points at something."""
    return bool(uuid_value) and uuid_value != NIL_UUID


def stamp_count(inv):
    """Records in any collection that carry a Nautobot external ID."""
    return sum(
        1
        for records in inv.values()
        for record in records.values()
        if is_set((record.get("externalIDs") or {}).get("nautobot"))
    )
