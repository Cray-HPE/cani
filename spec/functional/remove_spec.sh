#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

# ── remove command ──────────────────────────────────────────────────

Describe 'cani alpha remove'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha remove --help
      The status should equal 0
      The stdout should include 'Remove items from the inventory'
    End

    # subcommands
    It 'lists the location subcommand'
      When call bin/cani alpha remove --help
      The stdout should include 'location'
    End

    It 'lists the rack subcommand'
      When call bin/cani alpha remove --help
      The stdout should include 'rack'
    End

    It 'lists the device subcommand'
      When call bin/cani alpha remove --help
      The stdout should include 'device'
    End

    It 'lists the module subcommand'
      When call bin/cani alpha remove --help
      The stdout should include 'module'
    End

    It 'lists the cable subcommand'
      When call bin/cani alpha remove --help
      The stdout should include 'cable'
    End

    # persistent flags
    It 'has --force / -f flag'
      When call bin/cani alpha remove --help
      The stdout should include '--force'
    End
  End

  # ── remove location help ────────────────────────────────────────

  Describe 'location --help'
    It 'exits 0 and describes removing a location'
      When call bin/cani alpha remove location --help
      The status should equal 0
      The stdout should include 'Remove a location by UUID or name'
    End

    It 'shows uuid-or-name in usage'
      When call bin/cani alpha remove location --help
      The stdout should include '<uuid-or-name>'
    End
  End

  # ── remove rack help ────────────────────────────────────────────

  Describe 'rack --help'
    It 'exits 0 and describes removing a rack'
      When call bin/cani alpha remove rack --help
      The status should equal 0
      The stdout should include 'Remove a rack by UUID or name'
    End
  End

  # ── remove device help ──────────────────────────────────────────

  Describe 'device --help'
    It 'exits 0 and describes removing a device'
      When call bin/cani alpha remove device --help
      The status should equal 0
      The stdout should include 'Remove a device by UUID or name'
    End

    It 'mentions cascade behavior'
      When call bin/cani alpha remove device --help
      The stdout should include 'Cascades'
    End
  End

  # ── remove module help ──────────────────────────────────────────

  Describe 'module --help'
    It 'exits 0 and describes removing a module'
      When call bin/cani alpha remove module --help
      The status should equal 0
      The stdout should include 'Remove a module by UUID or name'
    End
  End

  # ── remove cable help ───────────────────────────────────────────

  Describe 'cable --help'
    It 'exits 0 and describes removing a cable'
      When call bin/cani alpha remove cable --help
      The status should equal 0
      The stdout should include 'Remove a cable by UUID or label'
    End
  End

  # ── argument validation ─────────────────────────────────────────

  Describe 'argument validation'
    It 'remove location with no arg fails'
      When call bin/cani alpha remove location
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'remove rack with no arg fails'
      When call bin/cani alpha remove rack
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'remove device with no arg fails'
      When call bin/cani alpha remove device
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'remove module with no arg fails'
      When call bin/cani alpha remove module
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'remove cable with no arg fails'
      When call bin/cani alpha remove cable
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End
  End

  # ── CRUD: add then remove location ──────────────────────────────

  Describe 'CRUD'
    It 'removes a location by name'
      # first add a location
      bin/cani alpha add location dc --name RemoveTestSite --config "$CANI_CONF" >/dev/null 2>&1
      When call bin/cani alpha remove location RemoveTestSite --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Removed location'
    End
  End

End
