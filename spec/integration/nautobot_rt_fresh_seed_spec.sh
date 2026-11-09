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

# Fresh seed: import from one Nautobot, export into an empty one.
#
# The inventory an import produces must seed a Nautobot that has never seen
# it: every object and value CANI carries arrives, and the custom-field
# definitions the values need are created before the first object. Nothing
# else in the suite exports an imported inventory into a reset Nautobot, which
# is where stale source IDs and missing definitions show.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# How to read the claims (spec/support/nautobot_roundtrip.sh wraps the oracle):
#   parity   = Nautobot and the datastore projected to one shape and compared
#              field by field; "projections equal" or one line per difference
#   source   = the Nautobot projection taken before the reset, compared with
#              the destination's after the export
#
# Walkthrough:
#   a. Reset and seed the base fixture, then add cani_rt_ custom-field
#      definitions (text, integer, boolean, select) and values on two devices,
#      a location, a VLAN, a prefix and an address, false, 0 and "" included.
#   b. Import into an empty datastore; parity holds, custom fields included,
#      and the source projection is kept.
#   c. Reset Nautobot: no object and no cani_rt_ definition survives.
#   d. Export into the empty Nautobot. The definitions are created first and
#      every value arrives; the racks, whose recorded IDs belong to the source,
#      are created. compute-001's primary IPv4 is refused, because import does
#      not yet carry the address's interface assignment, which a primary
#      address needs.
#   e. The destination projects as the source did except for that primary
#      IPv4.

# Reset, seed the base fixture and layer the custom fields on it.
#shellcheck disable=SC2317
rt_seed_with_custom_fields() { rt_reset_and_seed && rt_py seed-custom-fields; }

# Reset, then count the cani_rt_ definitions left in Nautobot.
#shellcheck disable=SC2317
rt_reset_to_empty() {
  rt_reset &&
  printf 'cani_rt_ definitions=%s\n' "$(rt_count 'extras/custom-fields/?q=cani_rt_')"
}

#shellcheck disable=SC2317
rt_diff_against_source() { rt_snapshot nautobot destination && rt_diff source destination; }

Describe 'INTEGRATION: Nautobot fresh seed: import from one Nautobot, export into an empty one'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  Describe 'the source'
    # a. The base fixture with custom fields on five families.
    It 'resets Nautobot and seeds the fixture with custom fields'
      BeforeCall rt_begin
      When call rt_seed_with_custom_fields
      The status should equal 0
      The output should include 'Seed complete'
      The output should include 'Custom fields seeded: 4 definitions, 6 objects'
    End

    # b. Import carries every value, falsy ones included.
    It 'imports it into an empty datastore'
      BeforeCall rt_write_config
      When call rt_import --default-location test-dc --default-status Active
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'projects the datastore and Nautobot identically, custom fields included'
      When call rt_assert_parity
      The status should equal 0
      The output should equal 'projections equal'
    End

    It 'keeps the source projection'
      When call rt_snapshot nautobot source
      The status should equal 0
    End
  End

  Describe 'a fresh destination'
    # c. Nothing of the source survives, its definitions included.
    It 'empties Nautobot'
      When call rt_reset_to_empty
      The status should equal 0
      The output should include 'Reset complete'
      The output should include 'cani_rt_ definitions=0'
    End

    # d. Definitions first, then every object; only the primary IPv4 fails.
    It 'exports the imported inventory into it'
      When call rt_export
      The status should equal 1
      The output should include 'Created custom field: cani_rt_flag'
      The output should include 'Created custom field: cani_rt_tier'
      The output should not include 'did not arrive'
      The output should not include 'Skipped rack (already exported)'
      The output should include 'compute-001: primary IP patch: status 400'
    End

    # e. Everything else arrived as the source had it.
    It 'matches the source but for the primary IPv4'
      When call rt_diff_against_source
      The status should equal 1
      The output should equal "device/compute-001: primary_ip4: '10.0.0.5/24' != ''"
    End
  End

End
