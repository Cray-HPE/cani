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

# ── update command ──────────────────────────────────────────────────

Describe 'cani alpha update'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha update --help
      The status should equal 0
      The stdout should include 'Update items in the inventory'
    End

    Describe 'subcommands'
      Parameters:value location rack device module cable interface orphans
      It "lists the $1 subcommand"
        When call bin/cani alpha update --help
        The stdout should include "$1"
      End
    End

    Describe 'persistent flags'
      Parameters:value --set --tag --metadata
      It "has $1 flag"
        When call bin/cani alpha update --help
        The stdout should include "$1"
      End
    End
  End

  # ── update location help & flags ────────────────────────────────

  Describe 'location --help'
    It 'exits 0 and describes updating a location'
      When call bin/cani alpha update location --help
      The status should equal 0
      The stdout should include 'Update a location'
    End

    It 'shows uuid-or-name in usage'
      When call bin/cani alpha update location --help
      The stdout should include '<uuid-or-name>'
    End

    Describe 'flags'
      Parameters:value --name --content-types --parent --description
      It "has $1 flag"
        When call bin/cani alpha update location --help
        The stdout should include "$1"
      End
    End
  End

  # ── update rack help & flags ────────────────────────────────────

  Describe 'rack --help'
    It 'exits 0 and describes updating a rack'
      When call bin/cani alpha update rack --help
      The status should equal 0
      The stdout should include 'Update a rack'
    End

    Describe 'flags'
      Parameters:value --name --status --role --description --u-height --location
      It "has $1 flag"
        When call bin/cani alpha update rack --help
        The stdout should include "$1"
      End
    End
  End

  # ── update device help & flags ──────────────────────────────────

  Describe 'device --help'
    It 'exits 0 and describes updating a device'
      When call bin/cani alpha update device --help
      The status should equal 0
      The stdout should include 'Update a device'
    End

    Describe 'flags'
      Parameters:value --name --status --role --description --position --face --swap --parent --nid --alias
      It "has $1 flag"
        When call bin/cani alpha update device --help
        The stdout should include "$1"
      End
    End
  End

  # ── update module help & flags ──────────────────────────────────

  Describe 'module --help'
    It 'exits 0 and describes updating a module'
      When call bin/cani alpha update module --help
      The status should equal 0
      The stdout should include 'Update a module'
    End

    Describe 'flags'
      Parameters:value --name --status --role --description --bay
      It "has $1 flag"
        When call bin/cani alpha update module --help
        The stdout should include "$1"
      End
    End
  End

  # ── update cable help & flags ───────────────────────────────────

  Describe 'cable --help'
    It 'exits 0 and describes updating a cable'
      When call bin/cani alpha update cable --help
      The status should equal 0
      The stdout should include 'Update a cable'
    End

    Describe 'flags'
      Parameters:value --label --status --color --description
      It "has $1 flag"
        When call bin/cani alpha update cable --help
        The stdout should include "$1"
      End
    End
  End

  # ── update orphans help & flags ─────────────────────────────────

  Describe 'orphans --help'
    It 'exits 0 and describes assigning orphans'
      When call bin/cani alpha update orphans --help
      The status should equal 0
      The stdout should include 'orphan'
    End

    It 'has --dry-run flag'
      When call bin/cani alpha update orphans --help
      The stdout should include '--dry-run'
    End

    It 'has --apply-plan flag'
      When call bin/cani alpha update orphans --help
      The stdout should include '--apply-plan'
    End
  End

  # ── argument validation ─────────────────────────────────────────

  Describe 'argument validation'
    Parameters:value location rack device module cable
    It "update $1 with no arg fails"
      When call bin/cani alpha update "$1"
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End
  End

  # ── CRUD: add then update location ──────────────────────────────

  Describe 'CRUD'
    It 'updates a location name'
      # first add a location
      bin/cani alpha add location dc --name UpdateTestSite --config "$CANI_CONF" >/dev/null 2>&1
      When call bin/cani alpha update location UpdateTestSite --name RenamedSite --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated location'
      The stderr should include 'RenamedSite'
    End
  End

End
