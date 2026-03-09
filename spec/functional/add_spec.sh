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

# ── add command ─────────────────────────────────────────────────────

Describe 'cani alpha add'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add --help
      The status should equal 0
      The stdout should include 'Add items to the inventory'
    End

    # subcommands
    It 'lists the location subcommand'
      When call bin/cani alpha add --help
      The stdout should include 'location'
    End

    It 'lists the rack subcommand'
      When call bin/cani alpha add --help
      The stdout should include 'rack'
    End

    It 'lists the device subcommand'
      When call bin/cani alpha add --help
      The stdout should include 'device'
    End

    It 'lists the module subcommand'
      When call bin/cani alpha add --help
      The stdout should include 'module'
    End

    It 'lists the cable subcommand'
      When call bin/cani alpha add --help
      The stdout should include 'cable'
    End

    # persistent flags
    It 'has --auto / -a flag'
      When call bin/cani alpha add --help
      The stdout should include '--auto'
    End

    It 'has --accept / -y flag'
      When call bin/cani alpha add --help
      The stdout should include '--accept'
    End

    It 'has --list-supported-types / -L flag'
      When call bin/cani alpha add --help
      The stdout should include '--list-supported-types'
    End

    It 'has --qty / -q flag'
      When call bin/cani alpha add --help
      The stdout should include '--qty'
    End

    It 'has --parent / -p flag'
      When call bin/cani alpha add --help
      The stdout should include '--parent'
    End
  End

  # ── add location help & flags ───────────────────────────────────

  Describe 'location --help'
    It 'exits 0 and describes adding a location'
      When call bin/cani alpha add location --help
      The status should equal 0
      The stdout should include 'Add a location'
    End

    It 'has --type flag'
      When call bin/cani alpha add location --help
      The stdout should include '--type'
    End

    It 'has --parent flag'
      When call bin/cani alpha add location --help
      The stdout should include '--parent'
    End
  End

  # ── add rack help & flags ───────────────────────────────────────

  Describe 'rack --help'
    It 'exits 0 and describes adding racks'
      When call bin/cani alpha add rack --help
      The status should equal 0
      The stdout should include 'Add one or more racks'
    End

    It 'has --location flag'
      When call bin/cani alpha add rack --help
      The stdout should include '--location'
    End
  End

  # ── add device help & flags ─────────────────────────────────────

  Describe 'device --help'
    It 'exits 0 and describes adding devices'
      When call bin/cani alpha add device --help
      The status should equal 0
      The stdout should include 'Add one or more devices'
    End

    It 'has --rack flag'
      When call bin/cani alpha add device --help
      The stdout should include '--rack'
    End

    It 'has --position flag'
      When call bin/cani alpha add device --help
      The stdout should include '--position'
    End

    It 'has --face flag'
      When call bin/cani alpha add device --help
      The stdout should include '--face'
    End
  End

  # ── add module help & flags ─────────────────────────────────────

  Describe 'module --help'
    It 'exits 0 and describes adding modules'
      When call bin/cani alpha add module --help
      The status should equal 0
      The stdout should include 'Add one or more modules'
    End

    It 'has --device flag'
      When call bin/cani alpha add module --help
      The stdout should include '--device'
    End

    It 'has --bay flag'
      When call bin/cani alpha add module --help
      The stdout should include '--bay'
    End
  End

  # ── add cable help & flags ──────────────────────────────────────

  Describe 'cable --help'
    It 'exits 0 and describes adding cables'
      When call bin/cani alpha add cable --help
      The status should equal 0
      The stdout should include 'Add one or more cables'
    End

    It 'has --a-device flag'
      When call bin/cani alpha add cable --help
      The stdout should include '--a-device'
    End

    It 'has --a-port flag'
      When call bin/cani alpha add cable --help
      The stdout should include '--a-port'
    End

    It 'has --b-device flag'
      When call bin/cani alpha add cable --help
      The stdout should include '--b-device'
    End

    It 'has --b-port flag'
      When call bin/cani alpha add cable --help
      The stdout should include '--b-port'
    End

    It 'has --label flag'
      When call bin/cani alpha add cable --help
      The stdout should include '--label'
    End

    It 'has --color flag'
      When call bin/cani alpha add cable --help
      The stdout should include '--color'
    End
  End

  # ── argument validation ─────────────────────────────────────────
  # These trigger setupDomain (loads device type libraries).

  Describe 'argument validation'
    It 'add device with no arg fails with slug required'
      When call bin/cani alpha add device
      The status should equal 1
      The stderr should include 'slug or part number required'
    End

    It 'add device with unknown slug fails'
      When call bin/cani alpha add device nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown device slug or part number: nonexistent-slug'
    End

    It 'add rack with unknown slug fails'
      When call bin/cani alpha add rack nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown rack slug or part number'
    End

    It 'add module with unknown slug fails'
      When call bin/cani alpha add module nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown module slug or part number'
    End

    It 'add cable with unknown slug fails'
      When call bin/cani alpha add cable nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown cable slug or part number'
    End

    It 'add location with no arg fails'
      When call bin/cani alpha add location
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End
  End

  # ── CRUD: add location ──────────────────────────────────────────

  Describe 'CRUD'

    It 'adds a location successfully'
      When call bin/cani alpha add location CrudTestSite --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added location'
      The stderr should include 'CrudTestSite'
    End
  End

End
