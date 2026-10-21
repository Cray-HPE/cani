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

# ── device bays and derived slugs ───────────────────────────────────
#
# Uses the fixture library in testdata/fixtures/cani/types/bays: a two-bay
# "example-dual-chassis" and an "Example Half Switch" with no slug, which cani
# derives as "example-half-switch". The library is copied into $CANI_DIR so
# the per-directory types cache is never written into the repository.

# CRUD inventory (rack "test-rack") plus a private copy of the bay library.
setup_bays_env() {
  setup_crud_env
  cp -R "$FIXTURES/cani/types/bays" "$CANI_DIR/types"
}

# cani_bays runs bin/cani against the test config and fixture library.
cani_bays() {
  bin/cani alpha "$@" --config "$CANI_CONF" --types-dirs "$CANI_DIR/types"
}

add_chassis() {
  cani_bays add device example-dual-chassis --rack test-rack --position 30 --name ch-30 >/dev/null 2>&1
}

fill_chassis() {
  add_chassis
  cani_bays add device example-half-switch --parent ch-30 --qty 2 --prefix sw- >/dev/null 2>&1
}

# move_into_chassis adds an unplaced switch, installs it with update, and
# prints the rack tree so the bay label can be checked.
move_into_chassis() {
  add_chassis
  cani_bays add device example-half-switch --name sw-moved >/dev/null 2>&1
  cani_bays update device sw-moved --parent ch-30 >/dev/null 2>&1 || return 1
  cani_bays show rack test-rack -o tree --no-color 2>&1
}

Describe 'device bays'
  Before 'setup_bays_env'
  After 'teardown_test_env'

  Describe 'derived slug'
    It 'adds a device whose library file has no slug'
      When call cani_bays add device example-half-switch --name sw-loose
      The status should equal 0
      The stderr should include 'device(s) added'
    End
  End

  Describe 'add device --parent'
    Before 'add_chassis'

    It 'installs a child into the first free bay'
      When call cani_bays add device example-half-switch --parent ch-30 --name sw-a
      The status should equal 0
      The stderr should include 'in bay "Bay 1"'
    End

    It 'installs a child into the bay named by --bay'
      When call cani_bays add device example-half-switch --parent ch-30 --bay "Bay 2" --name sw-b
      The status should equal 0
      The stderr should include 'in bay "Bay 2"'
    End

    It 'rejects an unknown bay'
      When call cani_bays add device example-half-switch --parent ch-30 --bay "Bay 9"
      The status should equal 1
      The stderr should include 'has no bay "Bay 9"'
    End

    It 'rejects --bay without a device parent'
      When call cani_bays add device example-half-switch --rack test-rack --position 31 --bay "Bay 1"
      The status should equal 1
      The stderr should include '--bay requires --parent'
    End
  End

  Describe 'full chassis'
    Before 'fill_chassis'

    It 'rejects a third child'
      When call cani_bays add device example-half-switch --parent ch-30 --name sw-extra
      The status should equal 1
      The stderr should include 'no free bay on "ch-30"'
    End
  End

  Describe 'update device --parent'
    It 'installs an existing device into a bay and shows it in the tree'
      When call move_into_chassis
      The status should equal 0
      The output should include '[Bay 1] sw-moved'
    End
  End
End
