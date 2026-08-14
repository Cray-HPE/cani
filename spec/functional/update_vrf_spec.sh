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

# ── update vrf ───────────────────────────────────────────────────────

# Helper: set up CRUD env and seed a VRF named "test-vrf".
#shellcheck disable=SC2317
setup_vrf_env() {
  setup_crud_env
  bin/cani alpha add vrf test-vrf --rd 65000:1 --description "Test VRF" --config "$CANI_CONF" >/dev/null 2>&1
}

# Helper: set up CRUD env, add VRF, and assign test-device to it.
#shellcheck disable=SC2317
setup_vrf_with_device_env() {
  setup_crud_env
  bin/cani alpha add vrf test-vrf --rd 65000:1 --description "Test VRF" --config "$CANI_CONF" >/dev/null 2>&1
  bin/cani alpha update vrf test-vrf --add-device test-device --config "$CANI_CONF" >/dev/null 2>&1
}

Describe 'cani alpha update vrf'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha update vrf --help
      The status should equal 0
      The stdout should include 'Update a VRF'
    End

    Describe 'flags'
      Parameters:value --description --rd --add-device --remove-device
      It "has $1 flag"
        When call bin/cani alpha update vrf --help
        The stdout should include "$1"
      End
    End
  End

  # ── argument validation (fail tests) ───────────────────────────

  Describe 'validation'
    It 'fails without a name argument'
      When call bin/cani alpha update vrf
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    Describe 'with fixture'
      Before 'setup_crud_env'

      It 'fails with unknown VRF'
        When call bin/cani alpha update vrf nonexistent --description "x" --config "$CANI_CONF"
        The status should equal 1
        The stderr should include 'not found'
      End
    End

    Describe 'no-change rejection'
      Before 'setup_vrf_env'

      It 'fails with no change flags'
        When call bin/cani alpha update vrf test-vrf --config "$CANI_CONF"
        The status should equal 1
        The stderr should include 'no changes specified'
      End
    End
  End

  # ── CRUD with fixture ──────────────────────────────────────────

  Describe 'CRUD'
    Before 'setup_vrf_env'

    It 'updates description'
      When call bin/cani alpha update vrf test-vrf --description "New desc" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated VRF'
    End

    It 'adds a device'
      When call bin/cani alpha update vrf test-vrf --add-device test-device --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated VRF'
    End

    It 'rejects --add-device with unknown device'
      When call bin/cani alpha update vrf test-vrf --add-device nonexistent --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'resolving --add-device'
    End

    It 'rejects --remove-device with unknown device'
      When call bin/cani alpha update vrf test-vrf --remove-device nonexistent --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'resolving --remove-device'
    End

    It 'updates rd'
      When call bin/cani alpha update vrf test-vrf --rd 65000:9999 --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated VRF'
    End
  End

  Describe 'remove-device'
    Before 'setup_vrf_with_device_env'

    It 'removes a device from the VRF'
      When call bin/cani alpha update vrf test-vrf --remove-device test-device --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated VRF'
    End
  End

End
