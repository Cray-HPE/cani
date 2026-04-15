#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2023, 2026 Hewlett Packard Enterprise Development LP
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

    Describe 'subcommands'
      Parameters:value location rack device module cable metadata
      It "lists the $1 subcommand"
        When call bin/cani alpha add --help
        The stdout should include "$1"
      End
    End

    Describe 'persistent flags'
      Parameters:value --auto --accept --list-supported-types --qty --parent --prefix --start --pad-width --tag --metadata --status --serial
      It "has $1 flag"
        When call bin/cani alpha add --help
        The stdout should include "$1"
      End
    End
  End

  # ── add location help & flags ───────────────────────────────────

  Describe 'location --help'
    It 'exits 0 and describes adding a location'
      When call bin/cani alpha add location --help
      The status should equal 0
      The stdout should include 'Add a location'
    End

    Describe 'flags'
      Parameters:value --type --parent --description --content-types
      It "has $1 flag"
        When call bin/cani alpha add location --help
        The stdout should include "$1"
      End
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

    Describe 'flags'
      Parameters:value --rack --position --face --zone --dry-run --location
      It "has $1 flag"
        When call bin/cani alpha add device --help
        The stdout should include "$1"
      End
    End
  End

  # ── add module help & flags ─────────────────────────────────────

  Describe 'module --help'
    It 'exits 0 and describes adding modules'
      When call bin/cani alpha add module --help
      The status should equal 0
      The stdout should include 'Add one or more modules'
    End

    Describe 'flags'
      Parameters:value --device --bay --bay-filter --dry-run --location
      It "has $1 flag"
        When call bin/cani alpha add module --help
        The stdout should include "$1"
      End
    End
  End

  # ── add cable help & flags ──────────────────────────────────────

  Describe 'cable --help'
    It 'exits 0 and describes adding cables'
      When call bin/cani alpha add cable --help
      The status should equal 0
      The stdout should include 'Add one or more cables'
    End

    Describe 'flags'
      Parameters:value --a-device --a-port --b-device --b-port --label --color
      It "has $1 flag"
        When call bin/cani alpha add cable --help
        The stdout should include "$1"
      End
    End
  End

  # ── add metadata help & flags ────────────────────────────────────

  Describe 'metadata --help'
    It 'exits 0 and describes creating metadata definitions'
      When call bin/cani alpha add metadata --help
      The status should equal 0
      The stdout should include 'metadata definitions'
    End

    Describe 'subcommands'
      Parameters:value role status tag
      It "lists the $1 subcommand"
        When call bin/cani alpha add metadata --help
        The stdout should include "$1"
      End
    End

    Describe 'flags'
      Parameters:value --content-types --color --description
      It "has $1 flag"
        When call bin/cani alpha add metadata --help
        The stdout should include "$1"
      End
    End
  End

  # ── add metadata noun help ──────────────────────────────────────

  Describe 'metadata noun help'
    Parameters:value role status tag
    It "metadata $1 --help exits 0"
      When call bin/cani alpha add metadata "$1" --help
      The status should equal 0
      The stdout should include "$1"
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

    It 'add location with unknown slug fails'
      When call bin/cani alpha add location nonexistent-slug
      The status should equal 1
      The stderr should include 'unknown location type slug'
    End

    It 'add metadata role with no arg fails'
      When call bin/cani alpha add metadata role
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'add metadata status with no arg fails'
      When call bin/cani alpha add metadata status
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'add metadata tag with no arg fails'
      When call bin/cani alpha add metadata tag
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End
  End

  # ── CRUD: add location ──────────────────────────────────────────

  Describe 'CRUD'

    It 'adds a location successfully'
      When call bin/cani alpha add location dc --name CrudTestSite --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added location'
      The stderr should include 'CrudTestSite'
    End
  End

End
