#!/usr/bin/env python3
"""Seed a fresh Nautobot instance with a small spine-leaf topology.

Creates:
  - 1 Location (Site)
  - 1 Rack
  - 2 Device Types (hpe-proliant-dl325-gen11-8sff, hpe-aruba-8325-48y8c)
  - 4 DL325 devices at U1-U4 in the rack
  - 2 8325 switches at U40-U41
    - 2 interfaces and 1 cable between compute-001 and compute-002

Reads nautobot URL and token from the cani config file (CANI_CONF env var),
falling back to NAUTOBOT_URL / NAUTOBOT_TOKEN env vars if the config is
unavailable.
"""
import urllib.parse

from seed_nautobot_api import api, delete_all, find_or_create


def purge_all():
    """Remove all test objects from Nautobot so seeding starts fresh.

    Order matters: delete children before parents to avoid FK conflicts.
    """
    print("Purging existing Nautobot data...")
    for endpoint in [
        "dcim/cables/",
        "dcim/interfaces/",
        "dcim/devices/",
        "dcim/racks/",
    ]:
        delete_all(endpoint)
    print("Purge complete.")


def get_status_id(name, content_types=None):
    """Find a status by name and return its UUID."""
    data = api("GET", f"extras/statuses/?name={urllib.parse.quote(name)}")
    if data.get("results"):
        return data["results"][0]["id"]
    if content_types is None:
        content_types = ["dcim.device", "dcim.rack", "dcim.location"]
    # Create it
    result = api("POST", "extras/statuses/", {
        "name": name,
        "color": "4caf50",
        "content_types": content_types,
    })
    return result["id"]


def get_or_create_role(name):
    """Find or create a role."""
    data = api("GET", f"extras/roles/?name={urllib.parse.quote(name)}")
    if data.get("results"):
        return data["results"][0]["id"]
    result = api("POST", "extras/roles/", {
        "name": name,
        "color": "2196f3",
        "content_types": ["dcim.device"],
    })
    return result["id"]


def get_or_create_manufacturer(name):
    """Find or create a manufacturer."""
    data = api("GET", f"dcim/manufacturers/?name={urllib.parse.quote(name)}")
    if data.get("results"):
        return data["results"][0]["id"]
    result = api("POST", "dcim/manufacturers/", {"name": name})
    return result["id"]


def get_or_create_location_type(name):
    """Find or create a location type (e.g. 'Site')."""
    data = api("GET", f"dcim/location-types/?name={urllib.parse.quote(name)}")
    if data.get("results"):
        return data["results"][0]["id"]
    result = api("POST", "dcim/location-types/", {
        "name": name,
        "nestable": True,
        "content_types": ["dcim.rack", "dcim.device"],
    })
    return result["id"]


def create_interface(device_id, name, status_id):
    """Create an interface on a seeded device and return it."""
    return api("POST", "dcim/interfaces/", {
        "device": {"id": device_id},
        "name": name,
        "type": "1000base-t",
        "status": {"id": status_id},
    })


def create_cable(interface_a_id, interface_b_id, status_id):
    """Create a cable between two seeded interfaces and return it."""
    return api("POST", "dcim/cables/", {
        "label": "compute-001-to-compute-002",
        "status": {"id": status_id},
        "type": "cat6",
        "color": "00ff00",
        "termination_a_type": "dcim.interface",
        "termination_a_id": interface_a_id,
        "termination_b_type": "dcim.interface",
        "termination_b_id": interface_b_id,
    })


def main():
    purge_all()

    status_id = get_status_id("Active", [
        "dcim.device",
        "dcim.rack",
        "dcim.location",
        "dcim.interface",
    ])
    cable_status_id = get_status_id("Connected", ["dcim.cable"])
    role_id = get_or_create_role("Compute")
    switch_role_id = get_or_create_role("Network")
    mfg_id = get_or_create_manufacturer("HPE")
    loc_type_id = get_or_create_location_type("Site")

    # Location
    location = find_or_create(
        "dcim/locations/", "name", "test-dc",
        {
            "name": "test-dc",
            "status": {"id": status_id},
            "location_type": {"id": loc_type_id},
            "comments": "primary test location",
        },
    )
    location_id = location["id"]
    print(f"Location: {location_id}")

    # Device types — model names must match the cani device type library slugs
    # so that Validate() can find them after import.
    dl325 = find_or_create(
        "dcim/device-types/", "model", "ProLiant DL325 Gen11 8SFF",
        {
            "manufacturer": {"id": mfg_id},
            "model": "ProLiant DL325 Gen11 8SFF",
            "u_height": 1,
            "is_full_depth": True,
        },
    )
    dl325_id = dl325["id"]
    print(f"DL325 type: {dl325_id}")

    switch_type = find_or_create(
        "dcim/device-types/", "model", "HPE Aruba Networking 8325-48Y8C Power-to-Port Airflow 6 Fans 2 Power Supply Units Bundle",
        {
            "manufacturer": {"id": mfg_id},
            "model": "HPE Aruba Networking 8325-48Y8C Power-to-Port Airflow 6 Fans 2 Power Supply Units Bundle",
            "u_height": 1,
            "is_full_depth": False,
        },
    )
    switch_type_id = switch_type["id"]
    print(f"Switch type: {switch_type_id}")

    # Rack
    rack = find_or_create(
        "dcim/racks/", "name", "rack-01",
        {
            "name": "rack-01",
            "status": {"id": status_id},
            "location": {"id": location_id},
            "u_height": 42,
            "comments": "primary test rack",
        },
    )
    rack_id = rack["id"]
    print(f"Rack: {rack_id}")

    # Devices - 4 DL325s at U1-U4
    dl325_devices = []
    for i in range(1, 5):
        dev = find_or_create(
            "dcim/devices/", "name", f"compute-{i:03d}",
            {
                "name": f"compute-{i:03d}",
                "device_type": {"id": dl325_id},
                "role": {"id": role_id},
                "status": {"id": status_id},
                "location": {"id": location_id},
                "rack": {"id": rack_id},
                "position": i,
                "face": "front",
                "comments": "primary compute node" if i == 1 else f"compute node {i}",
            },
        )
        dl325_devices.append(dev)
        print(f"Device compute-{i:03d}: {dev['id']} at U{i}")

    # 2 switches at U40-U41
    for i, pos in enumerate([40, 41], start=1):
        dev = find_or_create(
            "dcim/devices/", "name", f"spine-{i:03d}",
            {
                "name": f"spine-{i:03d}",
                "device_type": {"id": switch_type_id},
                "role": {"id": switch_role_id},
                "status": {"id": status_id},
                "location": {"id": location_id},
                "rack": {"id": rack_id},
                "position": pos,
                "face": "front",
                "comments": f"network spine {i}",
            },
        )
        print(f"Device spine-{i:03d}: {dev['id']} at U{pos}")

    iface_a = create_interface(dl325_devices[0]["id"], "eth0", status_id)
    iface_b = create_interface(dl325_devices[1]["id"], "eth0", status_id)
    cable = create_cable(iface_a["id"], iface_b["id"], cable_status_id)
    print(f"Interface compute-001 eth0: {iface_a['id']}")
    print(f"Interface compute-002 eth0: {iface_b['id']}")
    print(f"Cable compute-001-to-compute-002: {cable['id']}")

    print("\nSeed complete: 1 location, 1 rack, 4 compute + 2 switch devices, 2 interfaces, 1 cable")


if __name__ == "__main__":
    main()
