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

"""Return Nautobot to the state a round-trip spec expects to start from.

Everything an export can create is deleted; statuses, roles, tags,
manufacturers and custom-field definitions are shared fixtures and stay,
except the ones a spec made up under the cani-rt- prefix. The DCIM part
reuses purge_nautobot_dcim so the two purges cannot drift apart.
"""

import purge_nautobot_dcim
from seed_nautobot_api import NAUTOBOT_URL, _delete, api, delete_all

GLOBAL_NAMESPACE = "Global"
SPEC_OWNED_PREFIX = "cani-rt-"

# Assignment tables reference both a VRF and a prefix/device, so they go first.
VRF_ASSIGNMENTS = (
    "ipam/vrf-prefix-assignments/",
    "ipam/vrf-device-assignments/",
)


def delete_where(path, keep):
    """Delete every object at a list endpoint except those `keep` accepts."""
    while True:
        results = api("GET", f"{path}?limit=50").get("results", [])
        doomed = [obj for obj in results if not keep(obj)]
        if not doomed:
            return
        for obj in doomed:
            # api() expects a JSON body, which a 204 DELETE has none of.
            _delete(f"{NAUTOBOT_URL}/{path}{obj['id']}/")


def delete_namespaces_except_global():
    """Delete every namespace but Global, which Nautobot refuses to drop."""
    delete_where("ipam/namespaces/", lambda ns: ns["name"] == GLOBAL_NAMESPACE)


def delete_spec_owned_metadata():
    """Drop the statuses, roles and tags a spec invented."""
    for path in ("extras/statuses/", "extras/roles/", "extras/tags/"):
        delete_where(path, lambda obj: not obj["name"].startswith(SPEC_OWNED_PREFIX))


def reset():
    for endpoint in VRF_ASSIGNMENTS:
        delete_all(endpoint)
    purge_nautobot_dcim.main()
    delete_all("ipam/vrfs/")
    delete_namespaces_except_global()
    delete_spec_owned_metadata()
    print("Reset complete")


if __name__ == "__main__":
    reset()
