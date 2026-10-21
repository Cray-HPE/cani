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

# Integration test: device bays survive a Nautobot export/import round trip.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1
#
# Test flow:
#   1. Purge DCIM objects and build an inventory with the CLI using the
#      fixture library in testdata/fixtures/cani/types/library: a two-bay
#      "example-dual-chassis" in a rack, with one "example-half-switch" in
#      each bay (Bay 1 auto-assigned, Bay 2 named).
#   2. Export. Nautobot must hold both switches in the chassis bays, with no
#      rack position or face of their own.
#   3. Export again: both bays are already installed and are skipped.
#   4. Import into an empty datastore: the switches come back as children of
#      the chassis with their bay names.
#
# The same fixture adds an "example-server" with an "example-dual-port-nic" in
# module bay PCIe1 (position 1). The NIC's ports use forge's {module}
# placeholder, so Nautobot must receive them as "HSN 1/0" and "HSN 1/1".

Describe 'INTEGRATION: Nautobot device bays round trip'

  nautobot_unreachable() {
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
  }

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  TYPES_DIR="${CANI_DIR}/types"

  # --- helpers available inside this Describe scope ---

  cani_cfg() { cani alpha "$@" --config "$CANI_CONF" --types-dirs "$TYPES_DIR"; }

  build_inventory() {
    cp -R "$FIXTURES/cani/types/library" "$TYPES_DIR" &&
    cani_cfg add location dc --name bays-dc &&
    cani_cfg add location level --name bays-level --parent bays-dc &&
    cani_cfg add location section --name bays-section --parent bays-level &&
    cani_cfg add rack hpe-42u-800mmx1200mm-g2-enterprise-shock-rack \
      --location bays-section --name bays-rack &&
    cani_cfg add device example-dual-chassis --rack bays-rack --position 30 --name bays-chassis &&
    cani_cfg add device example-half-switch --parent bays-chassis --name bays-sw-a &&
    cani_cfg add device example-half-switch --parent bays-chassis --bay "Bay 2" --name bays-sw-b &&
    cani_cfg add device example-server --rack bays-rack --position 32 --name bays-srv &&
    cani_cfg add module example-dual-port-nic --device bays-srv --bay PCIe1 --name bays-nic
  }

  build_fixture() {
    python3 "$FIXTURES/nautobot/purge_nautobot_dcim.py" && build_inventory
  }

  run_export() { cani_cfg export nautobot 2>&1; }

  # Print "<bay>=<installed device>" for each bay on the chassis, sorted.
  nb_chassis_bays() {
    nb_list "dcim/device-bays/?device=bays-chassis&depth=1" | python3 -c "
import json, sys
bays = json.load(sys.stdin)['results']
print(','.join(sorted(b['name'] + '=' + ((b.get('installed_device') or {}).get('name') or '') for b in bays)))
"
  }

  # Print "<position>/<face>" for a Nautobot device ("None/None" when unset).
  nb_device_slot() {
    nb_list "dcim/devices/?name=$1&depth=1" | python3 -c "
import json, sys
d = json.load(sys.stdin)['results'][0]
face = d.get('face')
print(f\"{d.get('position')}/{face.get('value') if isinstance(face, dict) else face}\")
"
  }

  # Print a device's Nautobot "HSN" interface names, sorted and comma-separated
  # (the fixture server has no ports of its own, so export adds defaults).
  nb_interface_names() {
    nb_list "dcim/interfaces/?device=$1&limit=100" | python3 -c "
import json, sys
print(','.join(sorted(i['name'] for i in json.load(sys.stdin)['results'] if i['name'].startswith('HSN'))))
"
  }

  # Import into a fresh datastore, keeping the exported one aside.
  import_fresh() {
    mv "$CANI_DS" "$CANI_DS.exported" &&
    cani_cfg import nautobot --no-color </dev/null 2>&1
  }

  # Print "<parent name>/<bay>" for a device in the cani datastore.
  cani_device_bay() {
    DS="$CANI_DS" NAME="$1" python3 -c "
import json, os
inv = json.load(open(os.environ['DS']))
devices = inv.get('devices', {})
d = next(x for x in devices.values() if x.get('name') == os.environ['NAME'])
parent = devices.get(d.get('parent'), {}).get('name', '')
print(f\"{parent}/{d.get('parentBay', '')}\")
"
  }

  Describe 'fixture'
    It 'purges Nautobot and builds a chassis with a switch in each bay'
      BeforeCall setup_nautobot_env
      When call build_fixture
      The status should equal 0
      The output should include 'Purge complete'
      The stderr should include 'in bay "Bay 1"'
      The stderr should include 'in bay "Bay 2"'
    End
  End

  Describe 'export'
    It 'installs both switches into the chassis bays'
      When call run_export
      The status should equal 0
      The output should include 'Installed devices in bays: 2'
      The output should not include 'Child device types cannot be assigned'
    End

    It 'shows each switch in its bay in Nautobot'
      When call nb_chassis_bays
      The output should equal 'Bay 1=bays-sw-a,Bay 2=bays-sw-b'
    End

    It 'leaves the switches without their own rack position or face'
      When call nb_device_slot bays-sw-a
      The output should equal 'None/None'
    End

    It 'resolves {module} in module port names to the bay position'
      When call nb_interface_names bays-srv
      The output should equal 'HSN 1/0,HSN 1/1'
    End

    It 'skips bays that are already installed on a second export'
      When call run_export
      The status should equal 0
      The output should include 'Skipped device bays (already installed): 2'
    End
  End

  Describe 'import'
    It 'imports the bays into an empty datastore'
      When call import_fresh
      The status should equal 0
      The output should include '(2 in device bays)'
    End

    It 'restores the first switch to Bay 1 of the chassis'
      When call cani_device_bay bays-sw-a
      The output should equal 'bays-chassis/Bay 1'
    End

    It 'restores the second switch to Bay 2 of the chassis'
      When call cani_device_bay bays-sw-b
      The output should equal 'bays-chassis/Bay 2'
    End
  End
End
