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

# The export create_* options: off refuses and names the missing object, the
# matching --create-* flag unblocks the same export.
#
# Each create_<kind>=false must refuse an export that needs an object Nautobot
# lacks before anything is written, name the object, and the matching
# --create-<kind> flag must unblock the identical export.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims: rt_count asks a Nautobot list endpoint how many
# objects it holds, so a count of 0 after a refusal means nothing was written.
#
# Walkthrough, once per parameter row (device types, location types, module
# types, statuses, roles):
#   a. Reset Nautobot, build the CLI fixture and write a config with
#      create_<kind>: false; for the status and role rows, point rt-node-2 at
#      a status (cani-rt-status) or role (cani-rt-role) Nautobot does not have;
#      for the module-types row, put a power supply (rt-psu-1) in its PSU1 bay.
#   b. Export: exit 1 (refused_exit=1).
#   c. The log names the gap, e.g. "status not found: cani-rt-status" or
#      'location type "Section" not found in Nautobot ...' (refusal_named=yes).
#   d. Nautobot holds none of that kind (created_while_off=0) and no prefixes
#      either (written_while_off=0): prefixes are exported late, so a refusal
#      that came after earlier phases would have left them behind.
#   e. Export again with --create-<kind>: exit 0, exactly the expected number
#      of that kind created (created_with_flag) and both devices present.
# Single examples outside the table:
#   f. With create_locations=false the export still reports three created
#      locations: the option governs only names resolved through the lookup
#      cache (default_location), not the inventory's own locations.
#   g. With the config saying create_device_types: true,
#      --create-device-types=false refuses the export and names the device
#      type: a flag on the command line overrides the file in both directions.

Describe 'INTEGRATION: Nautobot export create_* options'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  # a. prepare_row KEY NEED: empty Nautobot, fresh fixture, KEY off in the
  # config, and a device that needs a status, role or module type Nautobot
  # does not have.
  prepare_row() {
    rt_begin && rt_reset && rt_build_inventory && rt_write_config "create_$1=false" &&
    case "$2" in
      status) rt_cani update device rt-node-2 --status cani-rt-status ;;
      role)   rt_cani update device rt-node-2 --role cani-rt-role ;;
      module) rt_cani add module hpe-1600w-flex-slot-platinum-hot-plug-psu \
                --device rt-node-2 --bay PSU1 --name rt-psu-1 ;;
      *) : ;;
    esac
  }

  # Same start with every create_* key on, the shipped default.
  prepare_default() { rt_begin && rt_reset && rt_build_inventory && rt_write_config; }

  # run_row KEY FLAG NEED REFUSAL ENDPOINT: steps a-e for one kind, reported
  # as key=value lines. ShellSpec repeats each example per parameter row
  # rather than each row per group, so the steps cannot be separate examples;
  # the report keeps every step's result readable in a failure.
  run_row() {
    prepare_row "$1" "$3" >/dev/null 2>&1 || { echo 'prepare=failed'; return 1; }
    # b. The export with the key off.
    rt_export >/dev/null 2>&1
    echo "refused_exit=$?"
    # c. Its log must name the missing object.
    if grep -qF "$4" "$RT_DIR/export.log"; then echo 'refusal_named=yes'; else echo 'refusal_named=no'; fi
    # d. Nothing of that kind, and nothing at all, reached Nautobot.
    echo "created_while_off=$(rt_count "$5")"
    echo "written_while_off=$(rt_count 'ipam/prefixes/')"
    # e. The same export with the flag on.
    rt_export "$2" >/dev/null 2>&1
    echo "unblocked_exit=$?"
    echo "created_with_flag=$(rt_count "$5")"
    echo "devices=$(rt_count 'dcim/devices/')"
  }

  Describe 'a key set to false in the config file'
    Parameters
      # key              flag                       needs    refusal                                                                        count endpoint                           created by the flag
      device_types   --create-device-types    none   'device type not found: hpe-proliant-dl325-gen11-8sff'                        'dcim/device-types/'                   1
      location_types --create-location-types  none   'location type "Section" not found in Nautobot (enable create_location_types)' 'dcim/location-types/'                 3
      module_types   --create-module-types    module 'module type "HPE 1600W Flex Slot Platinum Hot Plug Power Supply" not in Nautobot (enable create_module_types)' 'dcim/module-types/' 1
      statuses       --create-statuses        status 'status not found: cani-rt-status'                                              'extras/statuses/?name=cani-rt-status' 1
      roles          --create-roles           role   'role not found: cani-rt-role'                                                  'extras/roles/?name=cani-rt-role'      1
    End

    # a-e. One row = one kind; the lines come out in step order. The prefix
    # count stands in for "anything at all": prefixes are exported late, so a
    # non-atomic export would have created them before refusing.
    It "create_$1=false refuses naming the missing object and $2 unblocks it"
      When call run_row "$1" "$2" "$3" "$4" "$5"
      The status should equal 0
      The line 1 of output should equal 'refused_exit=1'
      The line 2 of output should equal 'refusal_named=yes'
      The line 3 of output should equal 'created_while_off=0'
      The line 4 of output should equal 'written_while_off=0'
      The line 5 of output should equal 'unblocked_exit=0'
      The line 6 of output should equal "created_with_flag=$6"
      The line 7 of output should equal 'devices=2'
    End
  End

  Describe 'create_locations=false'
    # f. Same clean start, with only create_locations off.
    It 'starts from an empty Nautobot and a fresh fixture'
      When call prepare_row locations none
      The status should equal 0
      The output should include 'Reset complete'
      The stderr should include 'Added prefix 10.31.0.0/24'
    End

    # The option governs locations resolved by name only (default_location);
    # the inventory's own locations are exported by the location phase.
    It 'still exports the three inventory locations'
      When call rt_export
      The status should equal 0
      The output should include 'Created locations: 3'
    End
  End

  Describe 'a --create-*=false flag against a config file that says true'
    # g. Every create_* key on in the config, then turn one off by flag.
    It 'starts from an empty Nautobot and a fresh fixture'
      When call prepare_default
      The status should equal 0
      The output should include 'Reset complete'
      The stderr should include 'Added prefix 10.31.0.0/24'
    End

    # The command line overrides the config file in both directions, so the
    # export refuses the way a create_device_types: false file would.
    It 'disables creation from the command line'
      When call rt_export --create-device-types=false
      The status should equal 1
      The output should include 'device type not found: hpe-proliant-dl325-gen11-8sff'
    End

    It 'created no device type'
      When call rt_count 'dcim/device-types/'
      The output should equal 0
    End
  End

End
