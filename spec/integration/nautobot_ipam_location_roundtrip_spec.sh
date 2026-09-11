#!/usr/bin/env sh
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

Describe 'INTEGRATION: Nautobot scoped IPAM round-trip'

  nautobot_unreachable() {
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
  }

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  cani_ipam_scope_summary() {
    VLAN_NAME="$1" PREFIX_CIDR="$2" CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)

location = next(
    (item for item in inventory.get("locations", {}).values() if item.get("name") == "test-dc"),
    {},
)
vlan = next(
    (item for item in inventory.get("vlans", {}).values() if item.get("name") == os.environ["VLAN_NAME"]),
    {},
)
prefix = next(
    (item for item in inventory.get("prefixes", {}).values() if item.get("prefix") == os.environ["PREFIX_CIDR"]),
    {},
)

print("vlan_location_matches=" + str(vlan.get("location") == location.get("id")))
print("prefix_location_matches=" + str(prefix.get("location") == location.get("id")))
print("prefix_vlan_matches=" + str(prefix.get("vlan") == vlan.get("id")))
PY
  }

  nautobot_ipam_scope_summary() {
    VLAN_NAME="$1" PREFIX_CIDR="$2" python3 - <<'PY'
import json
import os
import urllib.parse
import urllib.request

base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"]}


def get(path, **params):
    query = urllib.parse.urlencode(params)
    request = urllib.request.Request(base_url + path + ("?" + query if query else ""), headers=headers)
    with urllib.request.urlopen(request) as response:
        return json.load(response)


def ref_id(item, key):
    return str((item.get(key) or {}).get("id", ""))


location = get("dcim/locations/", name="test-dc")["results"][0]
vlan = get("ipam/vlans/", name=os.environ["VLAN_NAME"])["results"][0]
prefix = get("ipam/prefixes/", prefix=os.environ["PREFIX_CIDR"])["results"][0]
vlan_assignments = get("ipam/vlan-location-assignments/", vlan=vlan["id"])["results"]
prefix_assignments = get("ipam/prefix-location-assignments/", prefix=prefix["id"])["results"]

print("vlan_location_matches=" + str(any(
    ref_id(item, "location") == location["id"] for item in vlan_assignments
)))
print("prefix_location_matches=" + str(any(
    ref_id(item, "location") == location["id"] for item in prefix_assignments
)))
print("prefix_vlan_matches=" + str(ref_id(prefix, "vlan") == vlan["id"]))
PY
  }

  It 'seeds scoped Nautobot IPAM data'
    BeforeCall setup_nautobot_env
    When call python3 "$FIXTURES/nautobot/seed_nautobot_ipam.py"
    The status should equal 0
    The output should include 'IPAM seed complete'
  End

  It 'imports scoped IPAM data from Nautobot'
    When call cani alpha --config "$CANI_CONF" import nautobot \
      --default-location test-dc --default-status Active
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End

  It 'preserves imported VLAN prefix and location relationships'
    When call cani_ipam_scope_summary cani-seeded-vlan 10.31.0.0/24
    The status should equal 0
    The output should include 'vlan_location_matches=True'
    The output should include 'prefix_location_matches=True'
    The output should include 'prefix_vlan_matches=True'
  End

  It 'adds a locally scoped VLAN'
    When call cani alpha add vlan 3101 --name cani-exported-vlan \
      --location test-dc --status Active --config "$CANI_CONF"
    The status should equal 0
    The stderr should include 'Added VLAN 3101'
  End

  It 'adds a locally scoped prefix'
    When call cani alpha add prefix 10.31.1.0/24 --type network \
      --vlan cani-exported-vlan --location test-dc --status Active \
      --config "$CANI_CONF"
    The status should equal 0
    The stderr should include 'Added prefix 10.31.1.0/24'
  End

  It 'exports the local scoped IPAM data to Nautobot'
    When call cani alpha --config "$CANI_CONF" export nautobot
    The status should equal 0
    The stderr should include 'Export completed successfully'
  End

  It 'creates live Nautobot location assignments'
    When call nautobot_ipam_scope_summary cani-exported-vlan 10.31.1.0/24
    The status should equal 0
    The output should include 'vlan_location_matches=True'
    The output should include 'prefix_location_matches=True'
    The output should include 'prefix_vlan_matches=True'
  End

  It 're-imports the exported scoped IPAM data'
    When call cani alpha --config "$CANI_CONF" import nautobot \
      --default-location test-dc --default-status Active
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End

  It 'keeps scoped relationships after re-import'
    When call cani_ipam_scope_summary cani-exported-vlan 10.31.1.0/24
    The status should equal 0
    The output should include 'vlan_location_matches=True'
    The output should include 'prefix_location_matches=True'
    The output should include 'prefix_vlan_matches=True'
  End

End