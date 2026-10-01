#!/usr/bin/env python3
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

"""Purge every DCIM object an export can create so a spec starts from a clean Nautobot.

Order matters: children before parents to avoid FK conflicts.
Statuses and roles are intentionally kept (they are shared fixtures).
"""
from seed_nautobot_api import delete_all


def main():
    for endpoint in [
        "dcim/cables/",
        "ipam/ip-addresses/",
        "ipam/prefixes/",
        "ipam/vlans/",
        "dcim/interfaces/",
        "dcim/modules/",
        "dcim/module-bays/",
        "dcim/devices/",
        "dcim/racks/",
        "dcim/locations/",
        "dcim/module-types/",
        "dcim/device-types/",
        "dcim/location-types/",
    ]:
        delete_all(endpoint)
    print("Purge complete")


if __name__ == "__main__":
    main()
