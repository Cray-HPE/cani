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

# ── show command ────────────────────────────────────────────────────

Describe 'cani alpha show'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha show --help
      The status should equal 0
      The stdout should include 'Show items from the inventory'
    End

    # subcommands
    It 'lists the location subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'location'
    End

    It 'lists the rack subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'rack'
    End

    It 'lists the device subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'device'
    End

    It 'lists the module subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'module'
    End

    It 'lists the cable subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'cable'
    End

    It 'lists the cables subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'cables'
    End

    It 'lists the interfaces subcommand'
      When call bin/cani alpha show --help
      The stdout should include 'interfaces'
    End

    # persistent flags
    It 'has --sort / -s flag'
      When call bin/cani alpha show --help
      The stdout should include '--sort'
    End

    It 'has --format / -o flag'
      When call bin/cani alpha show --help
      The stdout should include '--format'
    End

    It 'has --visual / -v flag'
      When call bin/cani alpha show --help
      The stdout should include '--visual'
    End

    It 'has --rack flag'
      When call bin/cani alpha show --help
      The stdout should include '--rack'
    End

    It 'has --no-color flag'
      When call bin/cani alpha show --help
      The stdout should include '--no-color'
    End

    It 'has --file / -f flag'
      When call bin/cani alpha show --help
      The stdout should include '--file'
    End

    It 'has --show-cables flag'
      When call bin/cani alpha show --help
      The stdout should include '--show-cables'
    End

    It 'has --rack-view flag'
      When call bin/cani alpha show --help
      The stdout should include '--rack-view'
    End

    It 'has --columns flag'
      When call bin/cani alpha show --help
      The stdout should include '--columns'
    End

    It 'has --verbose / -V flag'
      When call bin/cani alpha show --help
      The stdout should include '--verbose'
    End

    It 'has --show-routing flag'
      When call bin/cani alpha show --help
      The stdout should include '--show-routing'
    End

    It 'has --cable-type flag'
      When call bin/cani alpha show --help
      The stdout should include '--cable-type'
    End
  End

  # ── show cables help ────────────────────────────────────────────

  Describe 'cables --help'
    It 'exits 0 and describes listing cables'
      When call bin/cani alpha show cables --help
      The status should equal 0
      The stdout should include 'List cables in the inventory'
    End

    It 'has --unconnected flag'
      When call bin/cani alpha show cables --help
      The stdout should include '--unconnected'
    End
  End

  # ── show interfaces help ────────────────────────────────────────

  Describe 'interfaces --help'
    It 'exits 0 and describes listing interfaces'
      When call bin/cani alpha show interfaces --help
      The status should equal 0
      The stdout should include 'List interfaces'
    End

    It 'has --device flag'
      When call bin/cani alpha show interfaces --help
      The stdout should include '--device'
    End

    It 'has --type flag'
      When call bin/cani alpha show interfaces --help
      The stdout should include '--type'
    End
  End

  # ── validation ──────────────────────────────────────────────────

  Describe 'validation'
    It 'rejects an invalid sort key'
      When call bin/cani alpha show --sort invalid --config "$CANI_CONF"
      The status should equal 1
      The stderr should include "invalid sort key"
    End
  End

  # ── show noun subcommands ─────────────────────────────────────

  Describe 'noun subcommands'
    It 'show location exits 0'
      When call bin/cani alpha show location --config "$CANI_CONF"
      The status should equal 0
      The stdout should include '['
    End

    It 'show rack exits 0'
      When call bin/cani alpha show rack --config "$CANI_CONF"
      The status should equal 0
      The stdout should include '['
    End

    It 'show device exits 0'
      When call bin/cani alpha show device --config "$CANI_CONF"
      The status should equal 0
      The stdout should include '['
    End

    It 'show module exits 0'
      When call bin/cani alpha show module --config "$CANI_CONF"
      The status should equal 0
      The stdout should include '['
    End

    It 'show cable exits 0'
      When call bin/cani alpha show cable --config "$CANI_CONF"
      The status should equal 0
      The stdout should include '['
    End
  End

  # ── CRUD: show with empty inventory ─────────────────────────────

  Describe 'CRUD'
    It 'shows inventory as JSON'
      When call bin/cani alpha show --config "$CANI_CONF"
      The status should equal 0
      The stdout should include '{'
    End
  End

End
