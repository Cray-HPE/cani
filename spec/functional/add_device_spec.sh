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

# ── add device ──────────────────────────────────────────────────────

Describe 'cani alpha add device'

  Describe 'valid slug'
    It 'adds a device with a known slug'
      When call bin/cani alpha add device cray-xd225v --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
      The stderr should include 'device(s) added'
    End
  End

  Describe 'invalid slug'
    It 'rejects an unknown slug'
      When call bin/cani alpha add device nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown device slug or part number: nonexistent-slug'
    End
  End

  # ── --qty bulk add ──────────────────────────────────────────────

  Describe '--qty flag'
    It 'adds multiple devices with --qty'
      When call bin/cani alpha add device cray-xd225v --qty 3 --config "$CANI_CONF"
      The status should equal 0
      The stderr should include '3 device(s) added'
    End
  End

  # ── --dry-run ───────────────────────────────────────────────────

  Describe '--dry-run flag'
    Before 'setup_crud_env'

    It 'exits 0 with --dry-run'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --dry-run --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'device(s) added'
    End
  End

  # ── sequential naming (--prefix, --start, --pad-width) ──────────

  Describe 'sequential naming'
    It 'names devices with --prefix and --start'
      When call bin/cani alpha add device cray-xd225v --qty 2 --prefix node --start 5 --pad-width 3 --config "$CANI_CONF"
      The status should equal 0
      The stderr should include '2 device(s) added'
    End
  End

  # ── --zone flag ─────────────────────────────────────────────────

  Describe '--zone flag'
    Before 'setup_crud_env'

    It 'accepts a valid zone'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --zone top --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
    End
  End

  # ── name expansion patterns ─────────────────────────────────────

  Describe 'name expansion'
    Before 'setup_crud_env'

    It 'expands %{SEQ} in device name'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --name "%{SEQ}-compute" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
    End
  End

  # ── --custom-field and stray arguments ──────────────────────────
  # A key=value flag with no value used to swallow the next token (here
  # `--status`) and every pair was then dropped silently; the parser now
  # refuses a flag token as a value and a pair without key=value. The key
  # itself must be a custom field the inventory defines for devices.

  Describe '--custom-field flag'
    setup_custom_field_env() {
      setup_crud_env
      bin/cani alpha add metadata custom-field site_code --label "Site code" --type text --content-types dcim.device --config "$CANI_CONF" >/dev/null 2>&1
    }
    Before 'setup_custom_field_env'

    # Prints "<status> <role> <site_code>" for the named device straight from
    # the datastore, so the assertion does not depend on show's table layout.
    device_status_role_and_field() {
      DEVICE_NAME="$1" DS="$CANI_DS" python3 - <<'PY'
import json, os
with open(os.environ["DS"]) as fh:
    devices = json.load(fh)["devices"].values()
for device in devices:
    if device["name"] == os.environ["DEVICE_NAME"]:
        print(device["status"], device.get("role"), (device.get("customFields") or {}).get("site_code"))
PY
    }

    It 'applies a role, a custom field and a status together'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --name meta-01 --role SpineSwitch --custom-field site_code=dc1 --status Active --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
    End

    It 'persists the role, the custom field and the status on the device'
      bin/cani alpha add device cray-xd225v --rack test-rack --name meta-01 --role SpineSwitch --custom-field site_code=dc1 --status Active --config "$CANI_CONF" >/dev/null 2>&1
      When call device_status_role_and_field meta-01
      The output should equal 'Active SpineSwitch dc1'
    End

    It 'refuses a custom field the inventory does not define'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --custom-field site_cod=dc1 --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--custom-field: custom field "site_cod" is not defined in the inventory'
      The stderr should include "define it with 'add metadata custom-field'"
    End

    It 'rejects a --custom-field that swallowed the next flag'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --custom-field site_code=dc1 --custom-field --status Active --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'flag needs an argument: --custom-field'
    End

    It 'rejects a --custom-field value without key=value'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --custom-field bogus --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'invalid argument "bogus" for --custom-field: expected key=value'
    End

    It 'rejects a --custom-field key that is itself a flag'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --custom-field --status=Active --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'flag needs an argument: --custom-field'
    End

    It 'no longer accepts --metadata'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --metadata role=SpineSwitch --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'unknown flag: --metadata'
    End
  End

  # ── --location ──────────────────────────────────────────────────
  # With a literal rack, or no rack, --location is where the device is
  # placed; it used to be read only to filter racks for a placement strategy.

  Describe '--location flag'
    Before 'setup_crud_env'

    # Prints the name of the location the named device records, straight
    # from the datastore.
    device_location() {
      DEVICE_NAME="$1" DS="$CANI_DS" python3 - <<'PY'
import json, os
with open(os.environ["DS"]) as fh:
    inventory = json.load(fh)
for device in inventory["devices"].values():
    if device["name"] == os.environ["DEVICE_NAME"]:
        print(inventory["locations"].get(device.get("location"), {}).get("name"))
PY
    }

    It 'places a device that has no rack in the location'
      bin/cani alpha add device cray-xd225v --name loose-01 --location removable-location --config "$CANI_CONF" >/dev/null 2>&1
      When call device_location loose-01
      The output should equal 'removable-location'
    End

    It 'refuses a location that does not hold the rack'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --name stray-01 --location removable-location --config "$CANI_CONF"
      The status should equal 1
      The stderr should include "its location must be the rack's or one that holds it"
    End
  End

  Describe 'positional arguments'
    Before 'setup_crud_env'

    It 'rejects a second positional argument'
      When call bin/cani alpha add device cray-xd225v Active --rack test-rack --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'unexpected argument(s) ["Active"]; add device takes a single slug or part number'
    End
  End

End
