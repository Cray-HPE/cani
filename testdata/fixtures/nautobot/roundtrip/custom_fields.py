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


"""Custom fields in the round-trip oracle: definitions, values and projection.

The base seed carries none, so the specs that share it are unaffected; the
fresh-seed spec layers these on with `seed-custom-fields`. Every definition
key starts with cani_rt_, which `reset` deletes, so an export into a reset
Nautobot has to create the definitions itself (Phase 0a) before the values
can arrive. The values include false, 0 and "" because Nautobot fills an
omitted key from the definition's default, so omitting a falsy value and
sending it are different writes.

The projection keeps only cani_rt_ keys, so definitions other fixtures leave
behind add nothing, and drops nulls: Nautobot lists every defined key on
every object of a content type, null when unset, while the datastore may
leave an unset key out.
"""

import urllib.parse

from seed_nautobot_api import api

PREFIX = "cani_rt_"
GLOBAL = "Global"
CONTENT_TYPES = ["dcim.device", "dcim.location", "ipam.ipaddress", "ipam.prefix", "ipam.vlan"]

# key -> (type, choices); choice weights are the 100, 200, ... an export
# writes back, so the import reports no lost attribute.
DEFINITIONS = {
    "cani_rt_owner": ("text", []),
    "cani_rt_count": ("integer", []),
    "cani_rt_flag": ("boolean", []),
    "cani_rt_tier": ("select", ["gold", "silver"]),
}

# (list endpoint, filter query, values) for objects of the base seed.
VALUES = (
    ("dcim/devices/", "name=compute-001",
     {"cani_rt_owner": "", "cani_rt_count": 0, "cani_rt_flag": False, "cani_rt_tier": "gold"}),
    ("dcim/devices/", "name=compute-002", {"cani_rt_owner": "ops", "cani_rt_count": 7, "cani_rt_flag": True}),
    ("dcim/locations/", "name=test-dc", {"cani_rt_owner": "facilities", "cani_rt_flag": False}),
    ("ipam/vlans/", "vid=3100", {"cani_rt_count": 0, "cani_rt_tier": "silver"}),
    ("ipam/prefixes/", "prefix=10.31.0.0/24&namespace=Global", {"cani_rt_owner": "", "cani_rt_flag": True}),
    ("ipam/ip-addresses/", "address=10.0.0.5/24&namespace=tenant-a", {"cani_rt_owner": "net", "cani_rt_count": 0}),
)


def seed():
    """Create the definitions and choices, then set the values."""
    for key, (field_type, choices) in DEFINITIONS.items():
        field = _find_field(key) or api("POST", "extras/custom-fields/", {
            "key": key, "label": key, "type": field_type, "content_types": CONTENT_TYPES,
        })
        existing = {c["value"] for c in api("GET", f"extras/custom-field-choices/?custom_field={field['id']}")["results"]}
        for index, value in enumerate(choices, start=1):
            if value not in existing:
                api("POST", "extras/custom-field-choices/", {
                    "custom_field": {"id": field["id"]}, "value": value, "weight": index * 100,
                })
    for path, query, values in VALUES:
        _patch_one(path, query, values)
    print(f"Custom fields seeded: {len(DEFINITIONS)} definitions, {len(VALUES)} objects")


def _patch_one(path, query, values):
    found = api("GET", f"{path}?{urllib.parse.quote(query, safe='=&/')}").get("results", [])
    if len(found) != 1:
        raise SystemExit(f"seed-custom-fields: {path}?{query} matched {len(found)} objects, want 1")
    api("PATCH", f"{path}{found[0]['id']}/", {"custom_fields": values})


def _values(custom_fields):
    """The non-null cani_rt_ values of one object."""
    return {k: v for k, v in (custom_fields or {}).items() if k.startswith(PREFIX) and v is not None}


def _find_field(key):
    """The definition with this key; Nautobot's listing filters by q, not key."""
    found = api("GET", f"extras/custom-fields/?q={urllib.parse.quote(key)}").get("results", [])
    return next((field for field in found if field.get("key") == key), None)


def _family(pairs):
    """{key: values} for the objects that carry at least one value."""
    return {key: values for key, values in pairs if values}


def project_nautobot(raw):
    """The custom-field family of the Nautobot projection."""
    prefix_ns = {p["id"]: _ref(p.get("namespace")) or GLOBAL for p in raw["prefixes"]}
    return _family(
        [(f"device|{d['name']}", _values(d.get("custom_fields"))) for d in raw["devices"]]
        + [(f"location|{loc['name']}", _values(loc.get("custom_fields"))) for loc in raw["locations"]]
        + [(f"vlan|{v['vid']}|{v['name']}", _values(v.get("custom_fields"))) for v in raw["vlans"]]
        + [(f"prefix|{prefix_ns[p['id']]}|{p['prefix']}", _values(p.get("custom_fields"))) for p in raw["prefixes"]]
        + [(f"ip|{prefix_ns.get(_ref(ip.get('parent'), 'id'), GLOBAL)}|{ip['address']}", _values(ip.get("custom_fields")))
           for ip in raw["ip_addresses"]]
    )


def project_cani(inv):
    """The custom-field family of the datastore projection."""
    def ns(record):
        return record.get("namespace") or GLOBAL

    return _family(
        [(f"device|{d['name']}", _values(d.get("customFields"))) for d in inv["devices"].values()]
        + [(f"location|{loc['name']}", _values(loc.get("customFields"))) for loc in inv["locations"].values()]
        + [(f"vlan|{v['vid']}|{v['name']}", _values(v.get("customFields"))) for v in inv["vlans"].values()]
        + [(f"prefix|{ns(p)}|{p['prefix']}", _values(p.get("customFields"))) for p in inv["prefixes"].values()]
        + [(f"ip|{ns(ip)}|{ip['address']}", _values(ip.get("customFields"))) for ip in inv["ip_addresses"].values()]
    )


def _ref(obj, key="name"):
    return str((obj or {}).get(key) or "")
