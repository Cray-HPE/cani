#!/usr/bin/env python3
"""Seed scoped Nautobot IPAM objects for import/export integration tests."""

from seed_nautobot import (
    get_or_create_location_type,
    get_status_id,
    purge_all,
)
from seed_nautobot_api import api, delete_all, find_or_create


def ensure_content_types(path, object_id, required):
    """Ensure an object permits the content types used by this fixture."""
    obj = api("GET", f"{path}{object_id}/")
    content_types = sorted(set(obj.get("content_types", [])) | set(required))
    if content_types != sorted(obj.get("content_types", [])):
        api("PATCH", f"{path}{object_id}/", {"content_types": content_types})


def purge_ipam():
    """Delete IPAM children before their parents for repeatable test runs."""
    for endpoint in [
        "ipam/prefix-location-assignments/",
        "ipam/prefixes/",
        "ipam/vlan-location-assignments/",
        "ipam/vlans/",
    ]:
        delete_all(endpoint)


def main():
    purge_all()
    purge_ipam()

    status_id = get_status_id("Active", [
        "dcim.location",
        "ipam.prefix",
        "ipam.vlan",
    ])
    ensure_content_types(
        "extras/statuses/",
        status_id,
        ["dcim.location", "ipam.prefix", "ipam.vlan"],
    )

    location_type_id = get_or_create_location_type("Site")
    ensure_content_types(
        "dcim/location-types/",
        location_type_id,
        ["ipam.prefix", "ipam.vlan"],
    )
    location = find_or_create(
        "dcim/locations/",
        "name",
        "test-dc",
        {
            "name": "test-dc",
            "status": {"id": status_id},
            "location_type": {"id": location_type_id},
        },
    )

    namespace = find_or_create(
        "ipam/namespaces/",
        "name",
        "Global",
        {"name": "Global"},
    )
    vlan = api("POST", "ipam/vlans/", {
        "vid": 3100,
        "name": "cani-seeded-vlan",
        "status": {"id": status_id},
    })
    api("POST", "ipam/vlan-location-assignments/", {
        "location": {"id": location["id"]},
        "vlan": {"id": vlan["id"]},
    })

    prefix = api("POST", "ipam/prefixes/", {
        "prefix": "10.31.0.0/24",
        "namespace": {"id": namespace["id"]},
        "status": {"id": status_id},
        "type": "network",
        "vlan": {"id": vlan["id"]},
    })
    api("POST", "ipam/prefix-location-assignments/", {
        "location": {"id": location["id"]},
        "prefix": {"id": prefix["id"]},
    })

    print(
        "IPAM seed complete: location=test-dc, vlan=3100, "
        "prefix=10.31.0.0/24"
    )


if __name__ == "__main__":
    main()