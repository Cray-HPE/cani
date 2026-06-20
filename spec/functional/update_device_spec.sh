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

# ── update device ───────────────────────────────────────────────────

Describe 'cani alpha update device'
  Before 'setup_crud_env'

  Describe 'valid name'
    It 'updates a device by name'
      When call bin/cani alpha update device test-device --description "updated description" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated device'
    End
  End

  Describe 'invalid name'
    It 'rejects an unknown name'
      When call bin/cani alpha update device nonexistent-name --description "x" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'no item found matching'
    End
  End

  # ── --primary-ipv4 flag ─────────────────────────────────────────

  Describe '--primary-ipv4'
    It 'rejects an IP not in inventory'
      When call bin/cani alpha update device test-device --primary-ipv4 "10.0.0.1/24" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'not found in inventory'
    End
  End

  # ── --primary-ipv6 flag ─────────────────────────────────────────

  Describe '--primary-ipv6'
    It 'rejects an IP not in inventory'
      When call bin/cani alpha update device test-device --primary-ipv6 "fd00::1/128" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'not found in inventory'
    End
  End

  # ── --swap flag ─────────────────────────────────────────────────

  Describe '--swap'
    It 'swaps two devices when the target slot is occupied'
      When call bin/cani alpha update device test-device-2 --position 1 --swap --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Swapped positions'
    End

    It 'rejects moving onto an occupied slot without --swap'
      When call bin/cani alpha update device test-device-2 --position 1 --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'use --swap'
    End
  End

  Describe '--swap on an unracked device'
    Before 'setup_orphan_env'
    It 'requires the device to be in a rack'
      When call bin/cani alpha update device orphan-device --position 1 --swap --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'is not assigned to a rack'
    End
  End

End
