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

# ── update device ───────────────────────────────────────────────────

Describe 'cani alpha update device'
  Before 'setup_crud_env'

  Describe 'valid name'
    It 'updates a device by name'
      When call bin/cani alpha update device test-device --description "updated description" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated device'
    End
  End

  Describe 'invalid name'
    It 'rejects an unknown name'
      When call bin/cani alpha update device nonexistent-name --description "x" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'no item found matching'
    End
  End

  # update shares add's key=value flag type, so the same typo fails the same way.
  Describe '--custom-field and --clear-custom-field'
    setup_custom_field_env() {
      setup_crud_env
      bin/cani alpha add metadata custom-field site_code --label "Site code" --type text --content-types dcim.device --config "$CANI_CONF" >/dev/null 2>&1
    }
    Before 'setup_custom_field_env'

    # Prints the custom fields of test-device as sorted JSON.
    device_custom_fields() {
      DS="$CANI_DS" python3 - <<'PY'
import json, os
with open(os.environ["DS"]) as fh:
    devices = json.load(fh)["devices"].values()
for device in devices:
    if device["name"] == "test-device":
        print(json.dumps(device.get("customFields"), sort_keys=True))
PY
    }

    It 'sets a custom field'
      bin/cani alpha update device test-device --custom-field site_code=dc1 --config "$CANI_CONF" >/dev/null 2>&1
      When call device_custom_fields
      The output should equal '{"site_code": "dc1"}'
    End

    It 'clears a custom field to an explicit null'
      bin/cani alpha update device test-device --custom-field site_code=dc1 --config "$CANI_CONF" >/dev/null 2>&1
      bin/cani alpha update device test-device --clear-custom-field site_code --config "$CANI_CONF" >/dev/null 2>&1
      When call device_custom_fields
      The output should equal '{"site_code": null}'
    End

    It 'refuses a custom field the inventory does not define'
      When call bin/cani alpha update device test-device --custom-field site_cod=dc1 --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--custom-field: custom field "site_cod" is not defined in the inventory'
    End

    It 'refuses to set and clear one key at once'
      When call bin/cani alpha update device test-device --custom-field site_code=dc1 --clear-custom-field site_code --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--custom-field and --clear-custom-field both name "site_code"'
    End

    It 'rejects a value without key=value'
      When call bin/cani alpha update device test-device --custom-field bogus --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'invalid argument "bogus" for --custom-field: expected key=value'
    End

    It 'rejects a --custom-field that swallowed the next flag'
      When call bin/cani alpha update device test-device --custom-field --status Active --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'flag needs an argument: --custom-field'
    End

    It 'no longer accepts --metadata'
      When call bin/cani alpha update device test-device --metadata role=SpineSwitch --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'unknown flag: --metadata'
    End
  End

  # ── --primary-ipv4 flag ─────────────────────────────────────────

  Describe '--primary-ipv4'
    It 'rejects an IP not in inventory'
      When call bin/cani alpha update device test-device --primary-ipv4 "10.0.0.1/24" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'not found in inventory'
    End
  End

  # ── --primary-ipv6 flag ─────────────────────────────────────────

  Describe '--primary-ipv6'
    It 'rejects an IP not in inventory'
      When call bin/cani alpha update device test-device --primary-ipv6 "fd00::1/128" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'not found in inventory'
    End
  End

  # ── --swap flag ─────────────────────────────────────────────────

  Describe '--swap'
    It 'swaps two devices when the target slot is occupied'
      When call bin/cani alpha update device test-device-2 --position 1 --swap --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Swapped positions'
    End

    It 'rejects moving onto an occupied slot without --swap'
      When call bin/cani alpha update device test-device-2 --position 1 --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'use --swap'
    End
  End

  Describe '--swap on an unracked device'
    Before 'setup_orphan_env'
    It 'requires the device to be in a rack'
      When call bin/cani alpha update device orphan-device --position 1 --swap --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'is not assigned to a rack'
    End
  End

End
