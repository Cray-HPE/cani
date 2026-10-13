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

# A change made on one side reaches the other: additions do, removals do not.
#
# Additions propagate in both directions; removals propagate in neither
# (export never deletes, import never prunes). Each asymmetry is pinned with
# the exact projection difference and carries a placeholder example for the
# behaviour that is not designed yet.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims (spec/support/nautobot_roundtrip.sh wraps the oracle):
#   parity    = Nautobot and the datastore projected to one shape and compared
#               field by field; "projections equal" or one line per difference,
#               where "only in A" means only in Nautobot and "only in B" only
#               in the datastore
#   ledger    = Nautobot's object-change count, marked before a command and
#               read after it; changes=0 means the command wrote nothing
#   rt_py ... = an edit made on the Nautobot side, as another operator would
#
# Walkthrough:
#   a. Reset Nautobot, seed the fixture and import it.
#   b. In Nautobot add tenant-b 10.0.0.7/24; import; parity: the addition
#      arrived in the datastore.
#   c. In Nautobot delete that address; import: the command succeeds.
#   d. Parity fails with exactly one line, the address "only in B": the local
#      copy survived because import merges additively (pinned; placeholder).
#   e. Remove the address locally; parity is restored.
#   f. Locally add VLAN 3300 scoped to test-dc and prefix 10.33.0.0/24 on it;
#      export: one VLAN and one prefix created; parity.
#   g. Locally remove compute-002, which takes its cable with it. Mark the
#      ledger and export: success and a ledger that grew by zero.
#   h. Parity fails with exactly three lines, the cable, the device and its
#      interface "only in A": export deleted nothing (pinned; placeholder).
#   i. Import: the device and its cable come back; parity.

Describe 'INTEGRATION: Nautobot reconciliation of edits made on one side'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  seed_and_import() {
    rt_begin && rt_reset_and_seed && rt_import --default-location test-dc --default-status Active
  }

  Describe 'the provider adds and then deletes an address'
    # a. Both sides start equal.
    It 'seeds Nautobot and imports it'
      When call seed_and_import
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # b. An addition on the Nautobot side arrives with the next import.
    It 'adds tenant-b 10.0.0.7/24 in Nautobot'
      When call rt_py create-ip tenant-b 10.0.0.7/24 'added in Nautobot'
      The status should equal 0
      The output should include 'created tenant-b|10.0.0.7/24'
    End

    It 'imports the addition'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End

    # c. A deletion on the Nautobot side is not an error for the next import.
    It 'deletes tenant-b 10.0.0.7/24 in Nautobot'
      When call rt_py delete-ip tenant-b 10.0.0.7/24
      The status should equal 0
      The output should include 'deleted tenant-b|10.0.0.7/24'
    End

    It 'imports without error'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # d. Import merges additively, so a record the provider deleted survives
    # locally; pinned until pruning is designed.
    It 'keeps the local copy of the deleted address'
      When call rt_assert_parity
      The status should equal 1
      The output should equal 'ip/tenant-b|10.0.0.7/24: only in B'
    End

    It 'drops a record the provider deleted'
    End

    # e. Until then, the operator removes it by hand.
    It 'restores parity by removing the address locally'
      When call rt_cani remove ip tenant-b/10.0.0.7/24
      The status should equal 0
      The stderr should include 'Removed IP address 10.0.0.7/24'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

  Describe 'the datastore adds a VLAN and a prefix'
    # f. Additions on the cani side arrive with the next export.
    It 'adds VLAN 3300 scoped to test-dc'
      When call rt_cani add vlan 3300 --name rt-vlan-3 --status Active --location test-dc
      The status should equal 0
      The stderr should include 'Added VLAN 3300'
    End

    It 'adds 10.33.0.0/24 on that VLAN'
      When call rt_cani add prefix 10.33.0.0/24 --type network --status Active --vlan rt-vlan-3 --location test-dc
      The status should equal 0
      The stderr should include 'Added prefix 10.33.0.0/24'
    End

    It 'exports both'
      When call rt_export
      The status should equal 0
      The output should include 'Created VLANs: 1'
      The output should include 'Created prefixes: 1'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

  Describe 'the datastore removes a device'
    # g. A removal on the cani side: the export has nothing to write.
    It 'removes compute-002 and, with it, the cable'
      When call rt_cani remove device compute-002 --force
      The status should equal 0
      The stderr should include 'Removed device'
    End

    It 'exports without deleting anything'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End

    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    # h. Export never deletes, so Nautobot still holds the device, its
    # interface and the cable; pinned until deletion is designed.
    It 'leaves the device, its interface and the cable in Nautobot'
      When call rt_assert_parity
      The status should equal 1
      The lines of output should equal 3
      The output should include 'cable/compute-001|eth0 <-> compute-002|eth0: only in A'
      The output should include 'device/compute-002: only in A'
      The output should include 'interface/compute-002|eth0: only in A'
    End

    It 'deletes a device the datastore removed'
    End

    # i. Nautobot is the source again: the next import restores the device.
    It 'brings the device and cable back on the next import'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

End
