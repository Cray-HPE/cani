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

# ── add device ──────────────────────────────────────────────────────

Describe 'cani alpha add device'

  Describe 'valid slug'
    It 'adds a device with a known slug'
      When call bin/cani alpha add device cray-xd225v --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
      The stderr should include 'device(s) added'
    End
  End

  Describe 'invalid slug'
    It 'rejects an unknown slug'
      When call bin/cani alpha add device nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown device slug or part number: nonexistent-slug'
    End
  End

  # ── --qty bulk add ──────────────────────────────────────────────

  Describe '--qty flag'
    It 'adds multiple devices with --qty'
      When call bin/cani alpha add device cray-xd225v --qty 3 --config "$CANI_CONF"
      The status should equal 0
      The stderr should include '3 device(s) added'
    End
  End

  # ── --dry-run ───────────────────────────────────────────────────

  Describe '--dry-run flag'
    Before 'setup_crud_env'

    It 'exits 0 with --dry-run'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --dry-run --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'device(s) added'
    End
  End

  # ── sequential naming (--prefix, --start, --pad-width) ──────────

  Describe 'sequential naming'
    It 'names devices with --prefix and --start'
      When call bin/cani alpha add device cray-xd225v --qty 2 --prefix node --start 5 --pad-width 3 --config "$CANI_CONF"
      The status should equal 0
      The stderr should include '2 device(s) added'
    End
  End

  # ── --zone flag ─────────────────────────────────────────────────

  Describe '--zone flag'
    Before 'setup_crud_env'

    It 'accepts a valid zone'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --zone top --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
    End
  End

  # ── name expansion patterns ─────────────────────────────────────

  Describe 'name expansion'
    Before 'setup_crud_env'

    It 'expands %{SEQ} in device name'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --name "%{SEQ}-compute" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
    End
  End

End
