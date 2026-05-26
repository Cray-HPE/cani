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

# ── Root command ────────────────────────────────────────────────────

Describe 'cani'

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani --help
      The status should equal 0
      The stdout should include 'Continious And Never-ending Inventory'
    End

    Describe 'subcommands'
      Parameters:value alpha init
      It "lists the $1 subcommand"
        When call bin/cani --help
        The stdout should include "$1"
      End
    End

    Describe 'flags'
      Parameters:value --config --debug --datastore --datastore-path --types-dirs --types-repos --types-repo-clone --types-repo-pull --strict --version
      It "has $1 flag"
        When call bin/cani --help
        The stdout should include "$1"
      End
    End
  End

  Describe '(no args)'
    It 'exits 0 and prints help'
      When call bin/cani
      The status should equal 0
      The stdout should include 'Usage:'
    End
  End

  Describe '--version'
    It 'exits 0 and prints a version string'
      When call bin/cani --version
      The status should equal 0
      The stdout should include 'cani version'
    End
  End

End

# ── init command ────────────────────────────────────────────────────

Describe 'cani init'

  Describe '--help'
    It 'exits 0 and describes provider scaffold generation'
      When call bin/cani init --help
      The status should equal 0
      The stdout should include 'Generate a new provider scaffold'
    End

    It 'lists the --output / -o flag'
      When call bin/cani init --help
      The stdout should include '--output'
    End

    It 'lists the --force / -f flag'
      When call bin/cani init --help
      The stdout should include '--force'
    End
  End

  Describe 'argument validation'
    It 'fails with no arguments'
      When call bin/cani init
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'rejects an invalid provider name'
      When call bin/cani init 'INVALID!'
      The status should equal 1
      The stderr should include 'must start with a lowercase letter'
    End

    It 'rejects reserved provider names'
      When call bin/cani init main
      The status should equal 1
      The stderr should include "is reserved"
    End
  End

End
