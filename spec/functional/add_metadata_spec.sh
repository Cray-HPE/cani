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

# ── add metadata ────────────────────────────────────────────────────

Describe 'cani alpha add metadata'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add metadata --help
      The status should equal 0
      The stdout should include 'metadata'
    End

    Describe 'flags'
      Parameters:value --content-types --color --description
      It "has $1 flag"
        When call bin/cani alpha add metadata --help
        The stdout should include "$1"
      End
    End

    Describe 'subcommands'
      Parameters:value role status tag
      It "lists $1 subcommand"
        When call bin/cani alpha add metadata --help
        The stdout should include "$1"
      End
    End
  End

  # ── add metadata role ───────────────────────────────────────────

  Describe 'role'
    Before 'setup_crud_env'

    Describe 'success'
      It 'adds a role'
        When call bin/cani alpha add metadata role my-test-role --config "$CANI_CONF"
        The status should equal 0
        The stderr should include 'Added role'
      End

      It 'adds a role with flags'
        When call bin/cani alpha add metadata role flagged-role --content-types dcim.device --color aa1409 --description "A test role" --config "$CANI_CONF"
        The status should equal 0
        The stderr should include 'Added role'
      End
    End

    Describe 'validation'
      It 'fails without a name argument'
        When call bin/cani alpha add metadata role
        The status should equal 1
        The stderr should include 'accepts 1 arg(s), received 0'
      End
    End
  End

  # ── add metadata status ─────────────────────────────────────────

  Describe 'status'
    Before 'setup_crud_env'

    Describe 'success'
      It 'adds a valid status'
        When call bin/cani alpha add metadata status Maintenance --config "$CANI_CONF"
        The status should equal 0
        The stderr should include 'Added status'
      End
    End

    Describe 'validation'
      It 'fails without a name argument'
        When call bin/cani alpha add metadata status
        The status should equal 1
        The stderr should include 'accepts 1 arg(s), received 0'
      End

      It 'rejects an invalid status name'
        When call bin/cani alpha add metadata status invalid-status-name --config "$CANI_CONF"
        The status should equal 1
        The stderr should include 'invalid status'
      End
    End
  End

  # ── add metadata tag ────────────────────────────────────────────

  Describe 'tag'
    Before 'setup_crud_env'

    Describe 'success'
      It 'adds a tag'
        When call bin/cani alpha add metadata tag my-test-tag --config "$CANI_CONF"
        The status should equal 0
        The stderr should include 'Added tag'
      End

      It 'adds a tag with content-types and color'
        When call bin/cani alpha add metadata tag colored-tag --content-types dcim.device,dcim.rack --color 00ff00 --config "$CANI_CONF"
        The status should equal 0
        The stderr should include 'Added tag'
      End
    End

    Describe 'validation'
      It 'fails without a name argument'
        When call bin/cani alpha add metadata tag
        The status should equal 1
        The stderr should include 'accepts 1 arg(s), received 0'
      End
    End
  End

End
