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

# Custom field values set in Nautobot reach devices already in the datastore.
#
# A device new to the datastore is inserted whole, so its custom fields always
# survived; a device matched to an existing record went through a merge that
# skipped them, so a value set or changed in Nautobot never arrived on
# re-import (PR #281 review). This spec pins the re-import path.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# The key is spec-owned (cani_rt_ prefix) and dropped at the end, so no other
# spec sees it. Edits on the Nautobot side go through roundtrip.device_cf.
#
# Walkthrough:
#   a. Reset, seed, define cani_rt_reimport_note on devices and set
#      compute-001 to "before"; import into a fresh datastore.
#   b. In Nautobot change compute-001 to "after" and set compute-002 to "new".
#   c. Import again into the same datastore: the preview lists both custom
#      field changes and no parent change.
#   d. The datastore holds "after" on compute-001 and "new" on compute-002.

CF_KEY=cani_rt_reimport_note

#shellcheck disable=SC2317
rt_cf() { PYTHONPATH="$FIXTURES/nautobot" python3 -m roundtrip.device_cf "$@"; }

#shellcheck disable=SC2317
seed_with_value() {
  rt_begin && rt_reset_and_seed && rt_cf define "$CF_KEY" &&
    rt_cf set compute-001 "$CF_KEY=before"
}

#shellcheck disable=SC2317
edit_in_nautobot() {
  rt_cf set compute-001 "$CF_KEY=after" && rt_cf set compute-002 "$CF_KEY=new"
}

#shellcheck disable=SC2317
datastore_values() {
  rt_cf show "$CANI_DS" compute-001 "$CF_KEY" && rt_cf show "$CANI_DS" compute-002 "$CF_KEY"
}

#shellcheck disable=SC2317
drop_key() { rt_cf drop "$CF_KEY" >/dev/null 2>&1 || true; }

Describe 'INTEGRATION: Nautobot re-import carries device custom field values'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  AfterAll 'drop_key'

  # a. The first import inserts the device whole.
  It 'seeds Nautobot with a device custom field value'
    When call seed_with_value
    The status should equal 0
    The output should include "set compute-001 $CF_KEY=before"
  End

  It 'imports it into a fresh datastore'
    BeforeCall rt_write_config
    When call rt_import --default-location test-dc --default-status Active
    The status should equal 0
    The output should include 'Import completed successfully'
  End

  # b. Another operator edits the values in Nautobot.
  It 'changes one value and sets another in Nautobot'
    When call edit_in_nautobot
    The status should equal 0
    The output should include "set compute-002 $CF_KEY=new"
  End

  # c. The re-import merges onto the existing records.
  It 're-imports and previews the custom field changes'
    When call rt_import --default-location test-dc --default-status Active
    The status should equal 0
    The output should include 'Import completed successfully'
    The output should include "customFields.$CF_KEY: before --> after"
    The output should match pattern "*customFields.$CF_KEY: * --> new*"
    The output should not include 'parent:'
  End

  # d. Both values reached the existing devices.
  It 'stores the Nautobot values on the existing devices'
    When call datastore_values
    The status should equal 0
    The line 1 of output should equal "compute-001 $CF_KEY=after"
    The line 2 of output should equal "compute-002 $CF_KEY=new"
  End
End
