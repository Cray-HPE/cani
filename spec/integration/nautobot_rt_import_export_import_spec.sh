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

# Round trip that starts in Nautobot: import -> export -> import.
#
# Importing a Nautobot instance, exporting the untouched result and importing
# it again must change nothing in Nautobot and drift nothing in the datastore.
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
#   a. Reset Nautobot to empty and seed the fixture: test-dc > rack-01 holding
#      compute-001 (tagged, primary IPv4 10.0.0.5 and IPv6 2001:db8:a::5, both
#      on its eth0) and compute-002 joined by a cable, test-edge holding
#      rack-less edge-001 (tenant-a's 10.0.0.5 is on its eth0 too), tenants
#      tenant-a/tenant-b each with 10.0.0.0/24 and 10.0.0.5/24, tenant-a also
#      with 2001:db8:a::/64, VRF blue in both, VLAN 3100.
#   b. Import into an empty datastore; the command reports success.
#   c. Parity: every location, rack, device (edge-001's location, the tag and
#      both primary addresses included), interface, cable, VRF, prefix,
#      address (its port assignments included) and VLAN matches on both sides.
#   d. Mark the ledger and the two tables the ledger cannot see (address
#      assignments, relationship associations) and export the untouched
#      datastore: exit 0, and the summary has no "Created" or "Updated" line.
#   e. The ledger grew by zero entries and both tables hold the same number of
#      rows: the export wrote nothing, not even a duplicate assignment.
#   f. The export log counts all three devices as skipped, names no field
#      that would change, and counts the three assignments as already there.
#   g. Parity still holds after the export.
#   h. Snapshot the datastore projection and import again into the populated
#      datastore; the command reports success.
#   i. The datastore projection equals the snapshot (a re-import is a no-op)
#      and parity still holds.
#   j. Negative control: take tenant-a's host off edge-001 eth0 in Nautobot;
#      parity must fail on exactly that address's interfaces, so the oracle
#      is shown to see an assignment. Put it back; parity holds again.

Describe 'INTEGRATION: Nautobot round trip import -> export -> import'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  Describe 'import'
    # a. Empty Nautobot, then the seeded fixture: both scripts report completion.
    It 'resets Nautobot and seeds the fixture'
      BeforeCall rt_begin
      When call rt_reset_and_seed
      The status should equal 0
      The output should include 'Reset complete'
      The output should include 'Seed complete'
    End

    # b. First import into an empty datastore.
    It 'imports the fixture into an empty datastore'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # c. Both sides project identically; a difference would print its key and field.
    It 'projects the datastore and Nautobot identically'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End
  End

  Describe 'export of an unchanged datastore'
    # d. Export what was just imported: nothing to create or update.
    It 'writes nothing to Nautobot'
      BeforeCall rt_mark_changes
      BeforeCall rt_mark_joins
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
      The output should not include 'Created '
      The output should not include 'Updated '
    End

    # e. The ledger proves it for every change-logged model, and the two
    # tables Nautobot does not change-log are counted directly.
    It 'leaves the change ledger where it was'
      When call rt_changes_since_mark
      The status should equal 0
      The line 1 of output should equal 'changes=0'
    End

    It 'adds no assignment or association row'
      When call rt_joins_unchanged
      The status should equal 0
      The output should equal 'joins unchanged'
    End

    # f. An up-to-date device is counted under "conflicts" with no field detail;
    # the label is misleading but the count is the contract this pins. The
    # three seeded assignments are read before any write and reported as
    # already present.
    It 'counts all three devices as skipped with nothing to change'
      When call grep -E 'Skipped devices|would change' "$RT_DIR/export.log"
      The lines of output should equal 1
      The output should include 'Skipped devices (conflicts): 3'
    End

    It 'counts the three assignments as already in Nautobot'
      When call grep -E 'IP address assignments' "$RT_DIR/export.log"
      The lines of output should equal 1
      The output should include 'Skipped IP address assignments (already exist): 3'
    End

    # g. The export changed neither side.
    It 'keeps parity'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End
  End

  Describe 'second import'
    # h. Remember the datastore projection, then import on top of it.
    It 'snapshots the datastore before importing again'
      When call rt_snapshot cani second-import
      The status should equal 0
    End

    It 'imports again into the populated datastore'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # i. The re-import was a no-op for the datastore, and parity still holds.
    It 'leaves the datastore projection unchanged'
      When call rt_assert_cani_unchanged second-import
      The status should equal 0
      The output should equal 'projections equal'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End
  End

  Describe 'the oracle sees an assignment'
    # j. Take one of the two ports off tenant-a's host in Nautobot: parity must
    # fail on exactly that field, or the comparison above proved nothing about
    # assignments.
    It 'takes tenant-a 10.0.0.5/24 off edge-001 eth0 in Nautobot'
      When call rt_py unassign-ip tenant-a 10.0.0.5/24 edge-001 eth0
      The status should equal 0
      The output should include 'unassigned tenant-a|10.0.0.5/24 from edge-001|eth0'
    End

    It 'fails parity on that address alone'
      When call rt_assert_parity
      The status should equal 1
      The lines of output should equal 1
      The output should equal "ip/tenant-a|10.0.0.5/24: interfaces: ['compute-001|eth0'] != ['compute-001|eth0', 'edge-001|eth0']"
    End

    It 'puts the assignment back'
      When call rt_py assign-ip tenant-a 10.0.0.5/24 edge-001 eth0
      The status should equal 0
      The output should include 'assigned tenant-a|10.0.0.5/24 to edge-001|eth0'
    End

    It 'keeps parity again'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End
  End

End
