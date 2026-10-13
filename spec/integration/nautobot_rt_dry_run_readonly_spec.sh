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

# A dry run never changes anything on either side.
#
# Every dry-run path must preview its work without writing to Nautobot,
# changing the datastore file, or stamping a Nautobot external ID.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims (spec/support/nautobot_roundtrip.sh wraps the oracle):
#   ledger      = Nautobot's object-change count, marked before a command and
#                 read after it; changes=0 means the command wrote nothing
#   projection  = a snapshot of Nautobot taken before the command, compared
#                 with Nautobot after it
#   fingerprint = a checksum of canidb.json taken before the command
#
# Walkthrough:
#   a. Reset Nautobot and build the CLI fixture (rt-dc > rt-level > rt-section,
#      rt-rack, rt-node-1 and rt-node-2 joined by a cable, the tenant IPAM).
#   b. Mark ledger and fingerprint, snapshot the empty Nautobot, then
#      `export --dry-run`: exit 0 and a "Would create" line for the location,
#      the device, the cable between the two planned devices and the prefix,
#      ending in "Dry run: inventory left unchanged".
#   c. The ledger grew by zero, Nautobot still projects empty, the datastore
#      is byte-identical and no record carries a Nautobot ID.
#   d. Export for real, then edit locally: rt-node-1's role, a new VLAN 3200
#      and a new address 10.0.0.9/24 in tenant-a.
#   e. Mark everything and snapshot the populated Nautobot, then
#      `export --merge --dry-run`: it previews the device update and the two
#      creates.
#   f. Ledger zero, Nautobot projection unchanged, datastore byte-identical.
#   g. Put `dry_run: true` in the config file instead of the flag and run
#      `export --merge`: the same preview, the same three detectors unchanged.
#   h. Restore the config, delete the datastore, mark the ledger and run
#      `import --phase et`: two devices transformed, success reported.
#   i. No canidb.json was created and the ledger grew by zero.

Describe 'INTEGRATION: Nautobot dry runs are read-only'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  Describe 'export --dry-run into an empty Nautobot'
    # a. Empty Nautobot, CLI-built datastore; the last add is the VLAN's prefix.
    It 'resets Nautobot and builds the fixture inventory'
      BeforeCall rt_begin
      When call rt_reset_and_build
      The status should equal 0
      The output should include 'Reset complete'
      The stderr should include 'Added prefix 10.31.0.0/24'
    End

    # b. Every object is a planned create, the cable included: its devices do
    # not exist yet, so the plan stands in for the interface lookup.
    It 'plans every create, the cable between two planned devices included'
      BeforeCall 'rt_mark_changes && rt_mark_datastore && rt_snapshot nautobot empty'
      When call rt_export --dry-run
      The status should equal 0
      The output should include 'Would create location: rt-dc'
      The output should include 'Would create device: rt-node-1'
      The output should include 'Would create cable: rt-node-1-to-rt-node-2'
      The output should include 'Would create prefix: 10.0.0.0/24'
      The output should include 'Dry run: inventory left unchanged'
    End

    # c. The three detectors, plus the external-ID count: nothing moved.
    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    It 'leaves Nautobot empty'
      When call rt_assert_nautobot_unchanged empty
      The output should equal 'projections equal'
    End

    It 'leaves the datastore byte-identical'
      When call rt_datastore_unchanged
      The output should equal 'datastore unchanged'
    End

    It 'stamps no Nautobot external ID'
      When call rt_stamp_count
      The output should equal '0'
    End
  End

  Describe 'export --merge --dry-run against a populated Nautobot'
    # d. Populate Nautobot for real, then make the datastore differ in three ways.
    It 'exports for real and then edits the datastore'
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End

    It 'changes a device role locally'
      When call rt_cani update device rt-node-1 --role Network
      The status should equal 0
      The stderr should include 'Updated device'
    End

    It 'adds a VLAN locally'
      When call rt_cani add vlan 3200 --name rt-vlan-2 --status Active --location rt-section
      The status should equal 0
      The stderr should include 'Added VLAN 3200'
    End

    It 'adds an address locally'
      When call rt_cani add ip 10.0.0.9/24 --namespace tenant-a --type host --status Active
      The status should equal 0
      The stderr should include 'Added IP address 10.0.0.9/24'
    End

    # e. One update and two creates are previewed, none performed.
    It 'previews the update and the creates'
      BeforeCall 'rt_mark_changes && rt_mark_datastore && rt_snapshot nautobot populated'
      When call rt_export --merge --dry-run
      The status should equal 0
      The output should include 'Would update device: rt-node-1'
      The output should include 'Would create VLAN 3200: rt-vlan-2'
      The output should include 'Would create IP address: 10.0.0.9/24'
      The output should include 'Dry run: inventory left unchanged'
    End

    # f. Same three detectors against the populated snapshot.
    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    It 'leaves the Nautobot projection unchanged'
      When call rt_assert_nautobot_unchanged populated
      The output should equal 'projections equal'
    End

    It 'leaves the datastore byte-identical'
      When call rt_datastore_unchanged
      The output should equal 'datastore unchanged'
    End
  End

  Describe 'dry_run from the config file only'
    # g. The config file alone turns a plain --merge export into a preview.
    It 'previews without the flag'
      BeforeCall 'rt_write_config dry_run=true && rt_mark_changes && rt_mark_datastore'
      When call rt_export --merge
      The status should equal 0
      The output should include 'Would create VLAN 3200: rt-vlan-2'
      The output should include 'Dry run: inventory left unchanged'
    End

    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    It 'leaves the Nautobot projection unchanged'
      When call rt_assert_nautobot_unchanged populated
      The output should equal 'projections equal'
    End

    It 'leaves the datastore byte-identical'
      When call rt_datastore_unchanged
      The output should equal 'datastore unchanged'
    End
  End

  Describe 'import --phase et'
    # h. Extract and transform only: the fixture is read from Nautobot but never loaded.
    It 'transforms without loading into an empty datastore'
      BeforeCall 'rt_write_config && remove_datastore && rt_mark_changes'
      When call rt_import --phase et --default-location rt-section --default-status Active
      The status should equal 0
      The output should include 'Transformed 2 devices'
      The output should include 'Import completed successfully'
    End

    # i. No datastore appeared and Nautobot was only read.
    It 'creates no datastore'
      When call rt_ds_checksum
      The output should equal 'absent'
    End

    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End
  End

End
