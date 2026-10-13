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

# Who wins when both sides change: --merge, provider edits, identity conflicts.
#
# Without --merge an export must list every field a merge would overwrite and
# write nothing; with it the datastore wins and only the differing objects are
# touched. An import that meets the same natural key under a new Nautobot UUID
# must refuse with an exact remedy instead of guessing.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims (spec/support/nautobot_roundtrip.sh wraps the oracle):
#   parity      = Nautobot and the datastore projected to one shape and compared
#                 field by field; "projections equal" or one line per difference
#   ledger      = Nautobot's object-change count, marked before a command and
#                 read after it; the newest entries name what was written
#   fingerprint = a checksum of canidb.json taken before the command
#   rt_py ...   = an edit made on the Nautobot side, as another operator would
#
# Walkthrough:
#   a. Reset Nautobot, seed the fixture and import it.
#   b. Locally change compute-001's role to Network; in Nautobot change
#      compute-002's role to Network and its comments.
#   c. Mark the ledger and export without --merge: the summary lists
#      compute-001 with 1 field (role) and compute-002 with 2 fields (role and
#      comments) that would change, counts all three devices as skipped, and
#      still reports success.
#   d. The ledger grew by zero: nothing was written without --merge.
#   e. Mark the ledger and export --merge: both devices are updated.
#   f. Parity: Nautobot now matches the datastore, so compute-002's Nautobot
#      comment was overwritten, which step c announced.
#   g. The ledger shows exactly two entries, the two device updates: the
#      unchanged interfaces were not re-sent.
#   h. In Nautobot change compute-002's serial and comments again, import, and
#      parity: the re-import carried both onto the existing device.
#   i. In Nautobot delete and re-create tenant-b 10.0.0.5/24 (same address,
#      new UUID). Mark the fingerprint and import: exit 1, the message names
#      the address and namespace and prints "cani alpha remove ip <uuid>".
#   j. The datastore is byte-identical (the refused import saved nothing); run
#      the printed command, import again, and parity holds.
#   k. In Nautobot delete and re-create tenant-a 10.0.0.0/24 together with the
#      address it holds. Import: exit 1 naming the prefix, with a remedy that
#      ends in --cascade.
#   l. The datastore is untouched; run the remedy (it removes the prefix and
#      the one address it held), import again, and parity holds: Nautobot
#      dropped the VRF membership and compute-001's primary IPv4 with the old
#      prefix, so both sides agree they are gone.

Describe 'INTEGRATION: Nautobot merge and conflict resolution'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  seed_and_import() {
    rt_begin && rt_reset_and_seed && rt_import --default-location test-dc --default-status Active
  }

  Describe 'both sides edit a device'
    # a. Both sides start equal.
    It 'seeds Nautobot and imports it'
      When call seed_and_import
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # b. One edit per side, on different devices.
    It 'changes the role of compute-001 in the datastore'
      When call rt_cani update device compute-001 --role Network
      The status should equal 0
      The stderr should include 'Updated device'
    End

    It 'changes the role and comments of compute-002 in Nautobot'
      When call rt_py patch-device compute-002 role=Network comments='edited in Nautobot'
      The status should equal 0
      The output should equal 'patched compute-002: role=Network comments=edited in Nautobot'
    End

    # c. compute-002 differs in role and comments; both count, because --merge
    # will overwrite both.
    It 'reports both devices as pending without --merge'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'Devices with pending changes (use --merge to apply)'
      The output should include 'compute-001: 1 field(s) would change'
      The output should include 'compute-002: 2 field(s) would change'
      The output should include 'Skipped devices (conflicts): 3'
      The output should include 'Export completed successfully'
    End

    # d. Reporting is not writing.
    It 'writes nothing without --merge'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    # e. With --merge the datastore's values land on both devices.
    It 'applies the datastore to both devices with --merge'
      BeforeCall rt_mark_changes
      When call rt_export --merge
      The status should equal 0
      The output should include 'Updated device: compute-001'
      The output should include 'Updated device: compute-002'
      The output should include 'Updated devices: 2'
    End

    # f. The datastore won, comments included.
    It 'makes Nautobot match the datastore, comments included'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End

    # g. Exactly the two device PATCHes; interfaces that already matched were skipped.
    It 'patched only the two devices'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=2'
      The output should include 'update dcim.device compute-001'
      The output should include 'update dcim.device compute-002'
    End
  End

  Describe 'the provider edits a device'
    # h. A Nautobot-side edit of an already-imported device reaches cani on re-import.
    It 'changes the serial and comments of compute-002 in Nautobot'
      When call rt_py patch-device compute-002 serial=SN-EDITED comments='edited again'
      The status should equal 0
      The output should include 'patched compute-002'
    End

    It 'imports the provider values'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'takes the serial and the comments'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

  Describe 'the provider recreates an address under the same key'
    # i. Same natural key, new Nautobot UUID: a source-identity conflict.
    It 'deletes and re-adds tenant-b 10.0.0.5/24 in Nautobot'
      When call rt_py recreate-ip tenant-b 10.0.0.5/24
      The status should equal 0
      The output should include 'recreated tenant-b|10.0.0.5/24'
    End

    It 'refuses the import and names the remove command'
      BeforeCall rt_mark_datastore
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 1
      The output should include 'IP address 10.0.0.5 in namespace "tenant-b"'
      The output should include 'remove the record and import again: cani alpha remove ip '
    End

    # j. A refused import saves nothing; the printed remedy is enough to recover.
    It 'leaves the datastore untouched by the refused import'
      When call rt_datastore_unchanged
      The output should equal 'datastore unchanged'
    End

    It 'removes the record with the printed command'
      When call rt_run_import_remedy
      The status should equal 0
      The line 1 of output should include 'remedy: cani alpha remove ip '
      The stderr should include 'Removed IP address'
    End

    It 'imports the recreated address'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

  Describe 'the provider recreates a prefix and the address it holds'
    # k. The same conflict one level up: the prefix and the address inside it.
    It 'deletes and re-adds tenant-a 10.0.0.0/24 with its address in Nautobot'
      When call rt_py recreate-prefix tenant-a 10.0.0.0/24
      The status should equal 0
      The output should include 'recreated tenant-a|10.0.0.0/24 with 1 address(es)'
    End

    It 'refuses the import and names remove prefix --cascade'
      BeforeCall rt_mark_datastore
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 1
      The output should include 'prefix 10.0.0.0/24 in namespace "tenant-a"'
      The output should include 'remove the record and import again: cani alpha remove prefix '
      The output should include ' --cascade'
    End

    # l. The remedy removes the prefix and the address it held in one step.
    It 'leaves the datastore untouched by the refused import'
      When call rt_datastore_unchanged
      The output should equal 'datastore unchanged'
    End

    It 'removes the prefix and its address with the printed command'
      When call rt_run_import_remedy
      The status should equal 0
      The line 1 of output should include 'remedy: cani alpha remove prefix '
      The stderr should include 'and 1 IP address(es) it held'
    End

    It 'imports the recreated prefix and address'
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # Nautobot dropped the VRF membership and compute-001's primary IPv4 with
    # the old prefix, so both sides now agree they are gone.
    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

End
