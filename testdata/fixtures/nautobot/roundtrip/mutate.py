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

"""Edits made on the Nautobot side, the way another operator would make them.

The specs use these to diverge Nautobot from the datastore: patch a device
field, add or delete an address, or recreate an object under the same
natural key so it comes back with a new UUID (the source-identity conflict
an import must refuse with a remedy).
"""

import urllib.parse

from seed_nautobot_api import NAUTOBOT_URL, _delete, api

IP_FIELDS = ("address", "type", "dns_name", "description")
PREFIX_FIELDS = ("prefix", "type", "description")


def one(path, **filters):
    """The single object a filtered list returns, or exit with a message."""
    query = urllib.parse.urlencode(filters)
    results = api("GET", f"{path}?{query}&depth=1").get("results", [])
    if len(results) != 1:
        raise SystemExit(f"expected one match for {path}?{query}, found {len(results)}")
    return results[0]


def ref_id(obj, key):
    nested = obj.get(key) or {}
    return {"id": nested["id"]} if nested.get("id") else None


def status_ref(name="Active"):
    return {"id": one("extras/statuses/", name=name)["id"]}


def patch_device(name, assignments):
    """patch-device NAME key=value...; role and status resolve by name."""
    payload = {}
    for assignment in assignments:
        key, value = assignment.split("=", 1)
        if key == "role":
            payload[key] = {"id": one("extras/roles/", name=value)["id"]}
        elif key == "status":
            payload[key] = status_ref(value)
        else:
            payload[key] = value
    device = one("dcim/devices/", name=name)
    api("PATCH", f"dcim/devices/{device['id']}/", payload)
    print(f"patched {name}: {' '.join(assignments)}")


def create_ip(namespace, address, description=""):
    ns = one("ipam/namespaces/", name=namespace)
    created = api("POST", "ipam/ip-addresses/", {
        "address": address, "namespace": {"id": ns["id"]}, "status": status_ref(),
        "type": "host", "description": description,
    })
    print(f"created {namespace}|{address} as {created['id']}")


def delete_ip(namespace, address):
    ip = one("ipam/ip-addresses/", address=address, namespace=namespace)
    _delete(f"{NAUTOBOT_URL}/ipam/ip-addresses/{ip['id']}/")
    print(f"deleted {namespace}|{address} ({ip['id']})")


def _snapshot(obj, fields):
    payload = {field: obj.get(field) for field in fields if obj.get(field) is not None}
    if isinstance(payload.get("type"), dict):
        payload["type"] = payload["type"]["value"]
    payload["status"] = ref_id(obj, "status")
    payload["role"] = ref_id(obj, "role")
    return {key: value for key, value in payload.items() if value is not None}


def recreate_ip(namespace, address):
    """Delete an address and create it again: same key, new UUID."""
    ns = one("ipam/namespaces/", name=namespace)
    ip = one("ipam/ip-addresses/", address=address, namespace=namespace)
    payload = _snapshot(ip, IP_FIELDS)
    payload["namespace"] = {"id": ns["id"]}
    _delete(f"{NAUTOBOT_URL}/ipam/ip-addresses/{ip['id']}/")
    created = api("POST", "ipam/ip-addresses/", payload)
    print(f"recreated {namespace}|{address}: {ip['id']} -> {created['id']}")


def recreate_prefix(namespace, cidr):
    """Delete a prefix with the addresses it holds and create them all again."""
    ns = one("ipam/namespaces/", name=namespace)
    prefix = one("ipam/prefixes/", prefix=cidr, namespace=namespace)
    children = api("GET", f"ipam/ip-addresses/?parent={prefix['id']}&depth=1").get("results", [])
    child_payloads = [dict(_snapshot(ip, IP_FIELDS), namespace={"id": ns["id"]}) for ip in children]
    for ip in children:
        _delete(f"{NAUTOBOT_URL}/ipam/ip-addresses/{ip['id']}/")
    _delete(f"{NAUTOBOT_URL}/ipam/prefixes/{prefix['id']}/")
    payload = dict(_snapshot(prefix, PREFIX_FIELDS), namespace={"id": ns["id"]})
    created = api("POST", "ipam/prefixes/", payload)
    for child in child_payloads:
        api("POST", "ipam/ip-addresses/", child)
    print(f"recreated {namespace}|{cidr} with {len(children)} address(es): {prefix['id']} -> {created['id']}")
