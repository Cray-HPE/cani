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

# Integration test: location types survive a Nautobot export/import round
# trip by their canidb key, and export never invents a location.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1
#
# Test flow:
#   1. Purge DCIM objects, build dc > level > section with one rack and one
#      device, export, import into a fresh datastore and read the key back.
#   2. Place a rack directly under a container-only `dc` location and export:
#      the run fails naming the rack, and Nautobot gains no synthetic location.

Describe 'INTEGRATION: Nautobot location types'

  nautobot_unreachable() {
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
  }

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  cani_cfg() { cani alpha --config "$CANI_CONF" "$@"; }

  purge_nautobot() { python3 "$FIXTURES/nautobot/purge_nautobot_dcim.py"; }

  # Print the locationType key stored for a location in the datastore.
  cani_location_type() {
    NAME="$1" DS="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
for loc in inventory.get("locations", {}).values():
    if loc.get("name") == os.environ["NAME"]:
        print(loc.get("locationType", ""))
        break
PY
  }

  # Print the Nautobot location names that have no datastore counterpart.
  nb_synthetic_locations() {
    DS="$CANI_DS" python3 - <<'PY'
import json
import os
import urllib.request

with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
local_names = {loc["name"] for loc in inventory.get("locations", {}).values()}

base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"]}
request = urllib.request.Request(base_url + "dcim/locations/?limit=500", headers=headers)
with urllib.request.urlopen(request) as response:
    remote_names = {loc["name"] for loc in json.load(response)["results"]}
print("synthetic=" + ",".join(sorted(remote_names - local_names)))
PY
  }

  Describe 'round trip by key'
    build_hierarchy() {
      purge_nautobot &&
      cani_cfg add location dc --name rt-dc &&
      cani_cfg add location level --name rt-level --parent rt-dc &&
      cani_cfg add location section --name rt-section --parent rt-level &&
      cani_cfg add rack hpe-42u-800mmx1200mm-g2-enterprise-shock-rack \
        --location rt-section --name rt-rack &&
      cani_cfg add device hpe-dl380-gen-11 --rack rt-rack --position 10 --name rt-dl380
    }

    It 'purges Nautobot and builds a dc > level > section hierarchy'
      BeforeCall setup_nautobot_env
      When call build_hierarchy
      The status should equal 0
      The output should include 'Purge complete'
      The stderr should include 'device(s) added'
    End

    It 'exports to Nautobot'
      When call cani_cfg export nautobot
      The status should equal 0
      The stderr should include 'Export completed successfully'
    End

    It 'imports the exported locations into a fresh datastore'
      BeforeCall remove_datastore
      When call cani_cfg import nautobot --default-location rt-section --default-status Active
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'round-trips the dc location type by its canidb key'
      When call cani_location_type rt-dc
      The output should equal 'dc'
    End

    It 'round-trips the section location type by its canidb key'
      When call cani_location_type rt-section
      The output should equal 'section'
    End
  End

  Describe 'rack directly under a container-only dc location'
    build_flat_rack() {
      purge_nautobot &&
      cani_cfg add location dc --name flat-dc &&
      cani_cfg add rack hpe-42u-800mmx1200mm-g2-enterprise-shock-rack \
        --location flat-dc --name flat-rack
    }

    It 'purges Nautobot and builds a rack under a dc location'
      BeforeCall setup_nautobot_env
      When call build_flat_rack
      The status should equal 0
      The output should include 'Purge complete'
      The stderr should include 'rack(s) added'
    End

    It 'fails the export naming the rack instead of inventing a location'
      When call cani_cfg export nautobot
      The status should equal 1
      The stderr should include 'flat-rack has no location that accepts racks'
      The stderr should not include 'Creating location: Default'
    End

    It 'creates only locations that exist in the datastore'
      When call nb_synthetic_locations
      The output should equal 'synthetic='
    End
  End

End
