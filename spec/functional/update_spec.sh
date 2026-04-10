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

# ── update command ──────────────────────────────────────────────────

Describe 'cani alpha update'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha update --help
      The status should equal 0
      The stdout should include 'Update items in the inventory'
    End

    # subcommands
    It 'lists the location subcommand'
      When call bin/cani alpha update --help
      The stdout should include 'location'
    End

    It 'lists the rack subcommand'
      When call bin/cani alpha update --help
      The stdout should include 'rack'
    End

    It 'lists the device subcommand'
      When call bin/cani alpha update --help
      The stdout should include 'device'
    End

    It 'lists the module subcommand'
      When call bin/cani alpha update --help
      The stdout should include 'module'
    End

    It 'lists the cable subcommand'
      When call bin/cani alpha update --help
      The stdout should include 'cable'
    End

    # persistent flags
    It 'has --set flag'
      When call bin/cani alpha update --help
      The stdout should include '--set'
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

    It 'has --name flag'
      When call bin/cani alpha update location --help
      The stdout should include '--name'
    End

    It 'has --content-types flag'
      When call bin/cani alpha update location --help
      The stdout should include '--content-types'
    End

    It 'has --parent flag'
      When call bin/cani alpha update location --help
      The stdout should include '--parent'
    End

    It 'has --description flag'
      When call bin/cani alpha update location --help
      The stdout should include '--description'
    End
  End

  # ── update rack help & flags ────────────────────────────────────

  Describe 'rack --help'
    It 'exits 0 and describes updating a rack'
      When call bin/cani alpha update rack --help
      The status should equal 0
      The stdout should include 'Update a rack'
    End

    It 'has --name flag'
      When call bin/cani alpha update rack --help
      The stdout should include '--name'
    End

    It 'has --status flag'
      When call bin/cani alpha update rack --help
      The stdout should include '--status'
    End

    It 'has --role flag'
      When call bin/cani alpha update rack --help
      The stdout should include '--role'
    End

    It 'has --description flag'
      When call bin/cani alpha update rack --help
      The stdout should include '--description'
    End

    It 'has --u-height flag'
      When call bin/cani alpha update rack --help
      The stdout should include '--u-height'
    End
  End

  # ── update device help & flags ──────────────────────────────────

  Describe 'device --help'
    It 'exits 0 and describes updating a device'
      When call bin/cani alpha update device --help
      The status should equal 0
      The stdout should include 'Update a device'
    End

    It 'has --name flag'
      When call bin/cani alpha update device --help
      The stdout should include '--name'
    End

    It 'has --status flag'
      When call bin/cani alpha update device --help
      The stdout should include '--status'
    End

    It 'has --role flag'
      When call bin/cani alpha update device --help
      The stdout should include '--role'
    End

    It 'has --description flag'
      When call bin/cani alpha update device --help
      The stdout should include '--description'
    End

    It 'has --position flag'
      When call bin/cani alpha update device --help
      The stdout should include '--position'
    End

    It 'has --face flag'
      When call bin/cani alpha update device --help
      The stdout should include '--face'
    End
  End

  # ── update module help & flags ──────────────────────────────────

  Describe 'module --help'
    It 'exits 0 and describes updating a module'
      When call bin/cani alpha update module --help
      The status should equal 0
      The stdout should include 'Update a module'
    End

    It 'has --name flag'
      When call bin/cani alpha update module --help
      The stdout should include '--name'
    End

    It 'has --status flag'
      When call bin/cani alpha update module --help
      The stdout should include '--status'
    End

    It 'has --role flag'
      When call bin/cani alpha update module --help
      The stdout should include '--role'
    End

    It 'has --description flag'
      When call bin/cani alpha update module --help
      The stdout should include '--description'
    End

    It 'has --bay flag'
      When call bin/cani alpha update module --help
      The stdout should include '--bay'
    End
  End

  # ── update cable help & flags ───────────────────────────────────

  Describe 'cable --help'
    It 'exits 0 and describes updating a cable'
      When call bin/cani alpha update cable --help
      The status should equal 0
      The stdout should include 'Update a cable'
    End

    It 'has --label flag'
      When call bin/cani alpha update cable --help
      The stdout should include '--label'
    End

    It 'has --status flag'
      When call bin/cani alpha update cable --help
      The stdout should include '--status'
    End

    It 'has --color flag'
      When call bin/cani alpha update cable --help
      The stdout should include '--color'
    End

    It 'has --description flag'
      When call bin/cani alpha update cable --help
      The stdout should include '--description'
    End
  End

  # ── argument validation ─────────────────────────────────────────

  Describe 'argument validation'
    It 'update location with no arg fails'
      When call bin/cani alpha update location
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'update rack with no arg fails'
      When call bin/cani alpha update rack
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'update device with no arg fails'
      When call bin/cani alpha update device
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'update module with no arg fails'
      When call bin/cani alpha update module
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'update cable with no arg fails'
      When call bin/cani alpha update cable
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
