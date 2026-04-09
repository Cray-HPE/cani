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
#   3. Verify the imported device count and DL360 positions
#   4. Swap the rack U positions of two DL360 compute nodes
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

  # It 'exports to Nautobot with merge'
  #   When call cani alpha --config "$CANI_CONF" export nautobot --merge
  #   The status should equal 0
  #   The stderr should include 'Export completed successfully'
  # End

  # It 'validates compute-001 is at U2 in Nautobot'
  #   When call nb_device_position compute-001
  #   The output should equal '2'
  # End

  # It 'validates compute-002 is at U1 in Nautobot'
  #   When call nb_device_position compute-002
  #   The output should equal '1'
  # End

  # It 're-imports from Nautobot (idempotency check)'
  #   # Keep the existing datastore — the import should merge changes.
  #   When call cani alpha --config "$CANI_CONF" import nautobot \
  #     --default-location test-dc --default-role Compute --default-status Active
  #   The status should equal 0
  #   The stderr should include 'Import completed successfully'
  # End

  # It 'still has 6 devices after re-import'
  #   When call cani_device_count
  #   The output should equal '6'
  # End

  # It 'compute-001 is still at U2 after re-import'
  #   When call cani_device_position compute-001
  #   The output should equal '2'
  # End

  # It 'compute-002 is still at U1 after re-import'
  #   When call cani_device_position compute-002
  #   The output should equal '1'
  # End

End
