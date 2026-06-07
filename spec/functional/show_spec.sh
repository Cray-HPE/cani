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

# ── show command ────────────────────────────────────────────────────

Describe 'cani alpha show'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha show --help
      The status should equal 0
      The stdout should include 'Show items from the inventory'
    End

    Describe 'subcommands'
      Parameters:value location rack device module cable interface fru metadata vlan prefix ip
      It "lists the $1 subcommand"
        When call bin/cani alpha show --help
        The stdout should include "$1"
      End
    End

    Describe 'persistent flags'
      Parameters:value --sort --format --no-color --file --with
      It "has $1 flag"
        When call bin/cani alpha show --help
        The stdout should include "$1"
      End
    End
  End

  # ── show cable help ─────────────────────────────────────────────

  Describe 'cable --help'
    It 'exits 0 and describes listing cables'
      When call bin/cani alpha show cable --help
      The status should equal 0
      The stdout should include 'cable'
    End
  End

  # ── show interface help ─────────────────────────────────────────

  Describe 'interface --help'
    It 'exits 0 and describes listing interfaces'
      When call bin/cani alpha show interface --help
      The status should equal 0
      The stdout should include 'interface'
    End
  End

  # ── show fru help ───────────────────────────────────────────────

  Describe 'fru --help'
    It 'exits 0 and describes listing FRUs'
      When call bin/cani alpha show fru --help
      The status should equal 0
      The stdout should include 'FRU'
    End
  End

  # ── show metadata help ──────────────────────────────────────────

  Describe 'metadata --help'
    It 'exits 0 and describes showing metadata definitions'
      When call bin/cani alpha show metadata --help
      The status should equal 0
      The stdout should include 'metadata'
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
    Parameters:value location rack device module cable
    It "show $1 exits 0"
      When call bin/cani alpha show "$1" --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'Total:'
    End
  End

  # ── CRUD: show with empty inventory ─────────────────────────────

  Describe 'CRUD'
    It 'shows inventory summary'
      When call bin/cani alpha show --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'Total:'
    End
  End

  # ── show rack visual modes ──────────────────────────────────────

  Describe 'rack visual modes'
    Before 'setup_crud_env'

    It 'show rack with --columns exits 0'
      When call bin/cani alpha show rack test-rack --columns 2 --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End

    It 'show rack with -V (verbose) exits 0'
      When call bin/cani alpha show rack test-rack -V --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End

    It 'show rack with -VV (extra verbose) exits 0'
      When call bin/cani alpha show rack test-rack -VV --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End

    It 'show rack with --labels exits 0'
      When call bin/cani alpha show rack test-rack --labels --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End
  End

  # ── --with tree detail ──────────────────────────────────────────

  Describe '--with flag'
    Before 'setup_crud_env'

    It 'includes modules in tree output'
      When call bin/cani alpha show --with modules --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End

    It 'includes interfaces in tree output'
      When call bin/cani alpha show --with interfaces --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End

    It 'includes cables in tree output'
      When call bin/cani alpha show --with cables --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End

    It 'includes empty-us in tree output'
      When call bin/cani alpha show --with empty-us --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End
  End

  # ── show prefix --tree ──────────────────────────────────────────

  Describe 'prefix --tree'
    Before 'setup_crud_env'

    It 'exits 0 with --tree flag'
      When call bin/cani alpha show prefix --tree --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End
  End

  # ── show ip --prefix ────────────────────────────────────────────

  Describe 'ip --prefix filter'
    Before 'setup_crud_env'

    It 'exits 0 with --prefix filter'
      When call bin/cani alpha show ip --prefix "10.0.0.0/24" --config "$CANI_CONF"
      The status should equal 0
      The stdout should be present
    End
  End

End
