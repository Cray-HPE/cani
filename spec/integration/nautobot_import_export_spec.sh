#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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

# Integration test: Nautobot import → update → export round-trip.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - SKIP_EXTERNAL_TESTS != 1
#
# Test flow:
#   1. Seed Nautobot with a small spine-leaf topology via seed script
#   2. Import from Nautobot into cani
#   3. Verify imported counts, metadata, interfaces, cables, and positions
#   4. Swap the rack U positions of two DL325 compute nodes
#   5. Export back to Nautobot with --merge
#   6. Query the Nautobot API to validate the position swap
#   7. Re-import and verify idempotency

Describe 'INTEGRATION: Nautobot import/export round-trip'

  # Skip the entire block if external tests are disabled or Nautobot is down.
  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' \
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1

  # --- helpers available inside this Describe scope ---

  # Return the rack position (int) for a Nautobot device by name.
  nb_device_position() {
    curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/dcim/devices/?name=$1" | \
      python3 -c "import sys,json; d=json.load(sys.stdin)['results'][0]; print(d.get('position',''))"
  }

  # Return the number of Nautobot cables matching a label.
  nb_cable_count() {
    curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/dcim/cables/?label=$1" | \
      python3 -c "import sys,json; print(json.load(sys.stdin).get('count', 0))"
  }

  # Return the number of devices in the cani datastore.
  cani_device_count() {
    python3 -c "
import json, sys
with open('${CANI_DS}') as f:
    inv = json.load(f)
print(len(inv.get('devices', {})))
"
  }

  # Return the rack position of a device by name from the cani datastore.
  cani_device_position() {
    python3 -c "
import json, sys
with open('${CANI_DS}') as f:
    inv = json.load(f)
for d in inv.get('devices', {}).values():
    if d.get('name') == '$1':
        print(d.get('rackPosition', ''))
        sys.exit(0)
print('')
"
  }

  # Return the number of cables in the cani datastore.
  cani_cable_count() {
    python3 -c "
import json
with open('${CANI_DS}') as f:
    inv = json.load(f)
print(len(inv.get('cables', {})))
"
  }

  # Summarize imported Nautobot mapping fields that should survive transform.
  cani_nautobot_mapping_summary() {
    python3 -c "
import json
nil = '00000000-0000-0000-0000-000000000000'
with open('${CANI_DS}') as f:
    inv = json.load(f)

def by_name(section, name):
    for item in inv.get(section, {}).values():
        if item.get('name') == name:
            return item
    return {}

location = by_name('locations', 'test-dc')
rack = by_name('racks', 'rack-01')
compute = by_name('devices', 'compute-001')
spine = by_name('devices', 'spine-001')
cable = next(iter(inv.get('cables', {}).values()), {})
iface_ids = {
    iface.get('id')
    for device in inv.get('devices', {}).values()
    for iface in device.get('interfaces', [])
}
iface_ids.discard(None)

print('location_status=' + location.get('status', ''))
print('rack_status=' + rack.get('status', ''))
print('rack_location_matches=' + str(rack.get('location') == location.get('id')))
print('compute_status=' + compute.get('status', ''))
print('compute_role=' + compute.get('role', ''))
print('spine_role=' + spine.get('role', ''))
# device.Location is derived from the rack via Parent and not persisted; resolve
# the device location through its parent rack instead.
compute_rack = next((r for r in inv.get('racks', {}).values() if r.get('id') == compute.get('parent')), {})
print('compute_location_matches=' + str(compute_rack.get('location') == location.get('id')))
print('compute_comments=' + compute.get('comments', ''))
print('interface_count=' + str(len(iface_ids)))
print('interfaces_at_least_two=' + str(len(iface_ids) >= 2))
print('interfaces_non_nil=' + str(all(iface_id != nil for iface_id in iface_ids)))
print('cable_status=' + cable.get('status', ''))
print('cable_terminations_resolved=' + str(cable.get('terminationA') in iface_ids and cable.get('terminationB') in iface_ids))
"
  }

  It 'seeds Nautobot with fixture data'
    BeforeCall setup_nautobot_env
    When call python3 "$FIXTURES/nautobot/seed_nautobot.py"
    The status should equal 0
    The output should include 'Seed complete'
  End

  It 'imports from Nautobot'
    When call cani alpha --config "$CANI_CONF" import nautobot \
      --default-location test-dc --default-role Compute --default-status Active
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End

  It 'has 6 devices after import'
    When call cani_device_count
    The output should equal '6'
  End

  It 'has 1 cable after import'
    When call cani_cable_count
    The output should equal '1'
  End

  It 'imports Nautobot statuses roles locations comments interfaces and cables'
    When call cani_nautobot_mapping_summary
    The status should equal 0
    The output should include 'location_status=Active'
    The output should include 'rack_status=Active'
    The output should include 'rack_location_matches=True'
    The output should include 'compute_status=Active'
    The output should include 'compute_role=Compute'
    The output should include 'spine_role=Network'
    The output should include 'compute_location_matches=True'
    The output should include 'compute_comments=primary compute node'
    The output should include 'interface_count=2'
    The output should include 'interfaces_non_nil=True'
    The output should include 'cable_status=Connected'
    The output should include 'cable_terminations_resolved=True'
  End

  It 'compute-001 is at U1'
    When call cani_device_position compute-001
    The output should equal '1'
  End

  It 'compute-002 is at U2'
    When call cani_device_position compute-002
    The output should equal '2'
  End

  # Swap compute-001 and compute-002 positions in a single command.
  It 'swaps compute-001 to U2 (and compute-002 to U1)'
    When call cani alpha --config "$CANI_CONF" update device compute-001 --position 2 --swap
    The status should equal 0
    The stderr should include 'Swapped positions'
  End

  It 'verifies compute-001 is now at U2'
    When call cani_device_position compute-001
    The output should equal '2'
  End

  It 'verifies compute-002 is now at U1'
    When call cani_device_position compute-002
    The output should equal '1'
  End

  It 'exports to Nautobot with merge'
    When call cani alpha --config "$CANI_CONF" export nautobot --merge
    The status should equal 0
    The stderr should include 'Export completed successfully'
  End

  It 'validates compute-001 is at U2 in Nautobot'
    When call nb_device_position compute-001
    The output should equal '2'
  End

  It 'validates compute-002 is at U1 in Nautobot'
    When call nb_device_position compute-002
    The output should equal '1'
  End

  It 'keeps the seeded cable in Nautobot'
    When call nb_cable_count compute-001-to-compute-002
    The output should equal '1'
  End

  It 're-imports from Nautobot (idempotency check)'
    # Keep the existing datastore — the import should merge changes.
    When call cani alpha --config "$CANI_CONF" import nautobot \
      --default-location test-dc --default-role Compute --default-status Active
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End

  It 'still has 6 devices after re-import'
    When call cani_device_count
    The output should equal '6'
  End

  It 'compute-001 is still at U2 after re-import'
    When call cani_device_position compute-001
    The output should equal '2'
  End

  It 'compute-002 is still at U1 after re-import'
    When call cani_device_position compute-002
    The output should equal '1'
  End

  It 'keeps Nautobot mapping fields after re-import'
    When call cani_nautobot_mapping_summary
    The status should equal 0
    The output should include 'compute_role=Compute'
    The output should include 'spine_role=Network'
    The output should include 'compute_comments=primary compute node'
    The output should include 'interfaces_at_least_two=True'
    The output should include 'interfaces_non_nil=True'
  End

End
