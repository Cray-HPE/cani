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

# ── add interface ────────────────────────────────────────────────────

# Helper: set up CRUD env with lag256 pre-created for duplicate/persistence tests.
#shellcheck disable=SC2317
setup_interface_env() {
  setup_crud_env
  bin/cani alpha add interface lag256 --device test-device --type lag --config "$CANI_CONF" >/dev/null 2>&1
}

Describe 'cani alpha add interface'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add interface --help
      The status should equal 0
      The stdout should include 'Add a standalone interface'
    End

    Describe 'flags'
      Parameters:value --device --type --role --label --mac --mode --untagged-vlan --tagged-vlan --vrf --description
      It "has $1 flag"
        When call bin/cani alpha add interface --help
        The stdout should include "$1"
      End
    End
  End

  # ── argument validation (fail tests) ───────────────────────────

  Describe 'validation'
    It 'fails without a name argument'
      When call bin/cani alpha add interface
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'fails without --device'
      When call bin/cani alpha add interface lag1 --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--device is required'
    End

    It 'fails with unknown interface type'
      When call bin/cani alpha add interface lag1 --device test-device --type bogus --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'unknown interface type'
    End

    Describe 'post-load validation'
      Before 'setup_crud_env'

      It 'fails with invalid mode'
        When call bin/cani alpha add interface lag1 --device test-device --type lag --mode invalid --config "$CANI_CONF"
        The status should equal 1
        The stderr should include 'invalid interface mode'
      End

      It 'fails with untagged-vlan out of range'
        When call bin/cani alpha add interface lag1 --device test-device --type lag --untagged-vlan 9999 --config "$CANI_CONF"
        The status should equal 1
        The stderr should include 'invalid VLAN ID'
      End

      It 'fails with non-integer tagged-vlan'
        When call bin/cani alpha add interface lag1 --device test-device --type lag --tagged-vlan abc --config "$CANI_CONF"
        The status should equal 1
        The stderr should include 'invalid tagged VLAN'
      End
    End
  End

  # ── CRUD with real fixture ──────────────────────────────────────

  Describe 'CRUD'
    Before 'setup_crud_env'

    It 'creates a LAG interface on a device'
      When call bin/cani alpha add interface lag256 --device test-device --type lag --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added interface'
      The stderr should include 'lag256'
    End

    It 'creates an interface with mode and VLANs'
      When call bin/cani alpha add interface lag512 --device test-device --type lag --mode tagged --untagged-vlan 100 --tagged-vlan 200,300 --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added interface'
    End

    It 'creates an interface with a VRF'
      When call bin/cani alpha add interface vrf-port --device test-device --type virtual --vrf LEGACY --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added interface'
    End

    It 'sets description and persists it to the datastore'
      When call bin/cani alpha add interface lag768 --device test-device --type lag --description "ISL uplink to spine" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added interface'
      The contents of file "$CANI_DS" should include 'ISL uplink to spine'
    End

    It 'resolves --device as unknown correctly'
      When call bin/cani alpha add interface lag1 --device nonexistent --type lag --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'resolving --device'
    End
  End

  Describe 'duplicate rejection'
    Before 'setup_interface_env'

    It 'rejects a duplicate interface name on the same device'
      When call bin/cani alpha add interface lag256 --device test-device --type lag --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'already exists'
    End
  End

  Describe 'persistence'
    Before 'setup_interface_env'

    It 'persists across reload'
      When call bin/cani alpha update interface --device test-device -L --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'lag256'
    End
  End

End
