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

# The IPAM update and remove verbs against an export.
#
# The namespace-aware `show` filters must pick one tenant's records; membership
# additions must reach Nautobot as assignments, while membership and record
# removals do not (pinned, each with a placeholder for the undesigned removal).
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims (spec/support/nautobot_roundtrip.sh wraps the oracle):
#   parity = Nautobot and the datastore projected to one shape and compared
#            field by field; "projections equal" or one line per difference,
#            where "only in A" means only in Nautobot
#   ledger = Nautobot's object-change count, marked before a command and read
#            after it; the newest entries name what was written
#
# Walkthrough:
#   a. Reset Nautobot, seed the fixture and import it.
#   b. `show prefix --namespace tenant-a`, `show vrf --namespace tenant-b` and
#      `show ip --namespace tenant-b` each list exactly one record, the one
#      from the namespace asked for.
#   c. Put tenant-b 10.0.0.0/24 into VRF tenant-b/blue and assign compute-002
#      to it; export; parity: both assignments exist in Nautobot.
#   d. Take both memberships away again locally; mark the ledger and export:
#      success and a ledger that grew by zero.
#   e. Parity fails with exactly two lines: Nautobot's VRF still lists the
#      device and the prefix. Export only adds assignments (pinned; two
#      placeholders).
#   f. Add both memberships back; mark and export: zero ledger entries, since
#      Nautobot already had them; parity.
#   g. Add interface eth9 on compute-001 naming VRF tenant-a/blue, then remove
#      that VRF: the command names the interface that still refers to it.
#   h. Mark and export: the export creates the interface and warns that it
#      skipped the unresolved VRF reference; the ledger shows exactly one
#      entry, the interface create.
#   i. Parity fails with exactly one line, the VRF "only in A" (pinned;
#      placeholder).
#   j. Remove tenant-b 10.0.0.5/24, then tenant-a 10.0.0.0/24 with --cascade,
#      which also removes the address it holds and clears compute-001's
#      primary IPv4. Mark and export: zero ledger entries.
#   k. Parity shows the primary IPv4, both addresses and the prefix still in
#      Nautobot (pinned; placeholder).

Describe 'INTEGRATION: Nautobot IPAM update and remove verbs'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  seed_and_import() {
    rt_begin && rt_reset_and_seed && rt_import --default-location test-dc --default-status Active
  }

  Describe 'show --namespace'
    # a. Both sides start equal.
    It 'seeds Nautobot and imports it'
      When call seed_and_import
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    # b. Each tenant holds one of each kind; the filter must return only that one.
    It 'lists only the prefix of the namespace asked for'
      When call rt_cani show prefix --namespace tenant-a
      The status should equal 0
      The output should include '10.0.0.0/24'
      The output should include 'Total: 1 prefix(es)'
    End

    It 'lists only the VRF of the namespace asked for'
      When call rt_cani show vrf --namespace tenant-b
      The status should equal 0
      The output should include 'blue'
      The output should include 'Total: 1 VRF(s)'
    End

    It 'lists only the address of the namespace asked for'
      When call rt_cani show ip --namespace tenant-b
      The status should equal 0
      The output should include '10.0.0.5/24'
      The output should include 'Total: 1 IP address(es)'
    End
  End

  Describe 'adding memberships'
    # c. Memberships added locally become VRF assignments in Nautobot.
    It 'puts tenant-b 10.0.0.0/24 into VRF blue'
      When call rt_cani update prefix tenant-b/10.0.0.0/24 --vrf tenant-b/blue
      The status should equal 0
      The stderr should include 'Updated prefix 10.0.0.0/24 in namespace tenant-b'
    End

    It 'assigns compute-002 to VRF blue'
      When call rt_cani update vrf tenant-b/blue --add-device compute-002
      The status should equal 0
      The stderr should include 'Updated VRF "blue" in namespace tenant-b'
    End

    It 'exports both assignments'
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

  Describe 'removing memberships'
    # d. The same memberships taken away locally give the export nothing to write.
    It 'takes tenant-b 10.0.0.0/24 out of VRF blue'
      When call rt_cani update prefix tenant-b/10.0.0.0/24 --remove-vrf tenant-b/blue
      The status should equal 0
      The stderr should include 'Updated prefix 10.0.0.0/24 in namespace tenant-b'
    End

    It 'unassigns compute-002 from VRF blue'
      When call rt_cani update vrf tenant-b/blue --remove-device compute-002
      The status should equal 0
      The stderr should include 'Updated VRF "blue" in namespace tenant-b'
    End

    It 'exports without a write'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End

    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    # e. Export only adds assignments, so Nautobot keeps both memberships;
    # pinned until removal is designed.
    It 'leaves both memberships in Nautobot'
      When call rt_assert_parity
      The status should equal 1
      The output should equal "vrf/tenant-b|blue: devices: ['compute-002'] != []
vrf/tenant-b|blue: prefixes: ['10.0.0.0/24'] != []"
    End

    It 'removes a VRF-prefix assignment the datastore dropped'
    End

    It 'removes a VRF-device assignment the datastore dropped'
    End

    # f. Re-adding what Nautobot still has is a no-op for the export.
    It 'restores parity by adding the memberships back'
      When call rt_cani update prefix tenant-b/10.0.0.0/24 --vrf tenant-b/blue
      The status should equal 0
      The stderr should include 'Updated prefix'
    End

    It 'adds the device back'
      When call rt_cani update vrf tenant-b/blue --add-device compute-002
      The status should equal 0
      The stderr should include 'Updated VRF'
    End

    It 'exports idempotently'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End

    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    It 'keeps parity'
      When call rt_assert_parity
      The output should equal 'projections equal'
    End
  End

  Describe 'remove vrf while an interface still names it'
    # g. An interface keeps naming the VRF after it is removed; the command says so.
    It 'adds an interface that names tenant-a/blue'
      When call rt_cani add interface eth9 --device compute-001 --type 1000base-t --vrf tenant-a/blue
      The status should equal 0
      The stderr should include 'Added interface "eth9"'
    End

    It 'removes the VRF and says which interface still names it'
      When call rt_cani remove vrf tenant-a/blue
      The status should equal 0
      The stderr should include 'Removed VRF "blue" in namespace tenant-a'
      The stderr should include '1 interface(s) still name it (compute-001/eth9)'
    End

    # h. The interface is created without the reference it cannot resolve.
    It 'exports the interface and skips the VRF reference'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'unresolved VRF reference "tenant-a/blue", skipping'
      The output should include 'Export completed successfully'
    End

    It 'created only the interface'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=1'
      The line 2 of output should equal 'create dcim.interface eth9'
    End

    # i. Export never deletes, so the VRF stays in Nautobot.
    It 'leaves the VRF in Nautobot'
      When call rt_assert_parity
      The status should equal 1
      The output should equal 'vrf/tenant-a|blue: only in A'
    End

    It 'deletes a VRF the datastore removed'
    End
  End

  Describe 'remove ip and remove prefix --cascade'
    # j. Record removals on the cani side give the export nothing to write.
    It 'removes tenant-b 10.0.0.5/24'
      When call rt_cani remove ip tenant-b/10.0.0.5/24
      The status should equal 0
      The stderr should include 'Removed IP address 10.0.0.5/24 in namespace tenant-b'
    End

    It 'removes tenant-a 10.0.0.0/24 with the address it holds'
      When call rt_cani remove prefix tenant-a/10.0.0.0/24 --cascade
      The status should equal 0
      The stderr should include 'Removed prefix 10.0.0.0/24 in namespace tenant-a'
      The stderr should include 'and 1 IP address(es) it held'
    End

    It 'exports without a write'
      BeforeCall rt_mark_changes
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End

    It 'writes nothing to Nautobot'
      When call rt_changes_since_mark
      The line 1 of output should equal 'changes=0'
    End

    # k. The removed address was compute-001's primary IPv4; the cascade cleared
    # it locally, and the export leaves it in Nautobot like the address. The
    # VRF and namespace lines from the earlier blocks are still present too.
    It 'leaves the prefix, both addresses and the primary IP in Nautobot'
      When call rt_assert_parity
      The status should equal 1
      The output should include "device/compute-001: primary_ip4: '10.0.0.5/24' != ''"
      The output should include 'ip/tenant-a|10.0.0.5/24: only in A'
      The output should include 'ip/tenant-b|10.0.0.5/24: only in A'
      The output should include 'prefix/tenant-a|10.0.0.0/24: only in A'
    End

    It 'deletes a prefix and address the datastore removed'
    End
  End

End
