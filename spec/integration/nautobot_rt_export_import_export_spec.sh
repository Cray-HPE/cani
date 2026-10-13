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

# Round trip that starts in cani: export -> import -> export.
#
# An inventory built with the CLI must export faithfully into an empty
# Nautobot, come back whole when imported into a fresh datastore, and export
# a second time without writing anything.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims (spec/support/nautobot_roundtrip.sh wraps the oracle):
#   parity   = Nautobot and the datastore projected to one shape and compared
#              field by field; "projections equal" or one line per difference
#   ledger   = Nautobot's object-change count, marked before a command and read
#              after it; changes=0 means the command wrote nothing
#
# Walkthrough:
#   a. Reset Nautobot to empty.
#   b. Build the fixture with `cani alpha add ...`: rt-dc > rt-level >
#      rt-section, rt-rack, rt-node-1 and rt-node-2 (DL325s) joined by a
#      cable, tenants tenant-a/tenant-b with overlapping 10.0.0.0/24 and
#      10.0.0.5/24, VRF blue in both, VLAN 3100 and 10.31.0.0/24 on it.
#   c. Export into the empty Nautobot: two devices and three prefixes created,
#      no interface skipped for a type Nautobot lacks.
#   d. Parity: Nautobot now projects identically to the datastore.
#   e. Snapshot the datastore projection, delete canidb.json and import what
#      was exported into the empty datastore.
#   f. The new datastore projects identically to the snapshot (nothing was lost
#      on the way out and back) and parity holds.
#   g. Mark the ledger and export again: no "Created" or "Updated" line, and
#      the ledger grew by zero entries.

Describe 'INTEGRATION: Nautobot round trip export -> import -> export'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  Describe 'export'
    # a. Start from an empty Nautobot.
    It 'resets Nautobot'
      BeforeCall rt_begin
      When call rt_reset
      The status should equal 0
      The output should include 'Reset complete'
    End

    # b. The CLI counterpart of the seed; its last command adds 10.31.0.0/24.
    It 'builds the fixture inventory with the CLI'
      When call rt_build_inventory
      The status should equal 0
      The stderr should include 'Added prefix 10.31.0.0/24'
    End

    # c. Everything is new to Nautobot, so the summary counts creates only.
    It 'exports into the empty Nautobot'
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
      The output should include 'Created devices: 2'
      The output should include 'Created prefixes: 3'
      The output should not include 'Skipped interfaces (type not in Nautobot)'
    End

    # d. What Nautobot holds now is exactly what the datastore holds.
    It 'projects Nautobot and the datastore identically'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End
  End

  Describe 'import into a fresh datastore'
    # e. Remember the exported datastore, throw it away, and import from Nautobot.
    It 'snapshots the exported datastore'
      When call rt_snapshot cani exported
      The status should equal 0
    End

    It 'imports what was exported into an empty datastore'
      BeforeCall remove_datastore
      When call rt_import --default-location rt-section --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # f. The import rebuilt the datastore the export started from.
    It 'reproduces the exported datastore'
      When call rt_assert_cani_unchanged exported
      The status should equal 0
      The output should equal 'projections equal'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End
  End

  Describe 'second export'
    # g. Exporting the re-imported datastore has nothing left to do.
    It 'writes nothing to Nautobot'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
      The output should not include 'Created '
      The output should not include 'Updated '
    End

    It 'leaves the change ledger where it was'
      When call rt_changes_since_mark
      The status should equal 0
      The line 1 of output should equal 'changes=0'
    End
  End

End
