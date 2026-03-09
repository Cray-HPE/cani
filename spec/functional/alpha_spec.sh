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

# ── alpha command ───────────────────────────────────────────────────

Describe 'cani alpha'

  Describe '--help'
    It 'exits 0 and shows unstable description'
      When call bin/cani alpha --help
      The status should equal 0
      The stdout should include 'unstable'
    End

    It 'lists the add subcommand'
      When call bin/cani alpha --help
      The stdout should include 'add'
    End

    It 'lists the remove subcommand'
      When call bin/cani alpha --help
      The stdout should include 'remove'
    End

    It 'lists the update subcommand'
      When call bin/cani alpha --help
      The stdout should include 'update'
    End

    It 'lists the show subcommand'
      When call bin/cani alpha --help
      The stdout should include 'show'
    End

    It 'lists the import subcommand'
      When call bin/cani alpha --help
      The stdout should include 'import'
    End

    It 'lists the export subcommand'
      When call bin/cani alpha --help
      The stdout should include 'export'
    End

    It 'lists the serve subcommand'
      When call bin/cani alpha --help
      The stdout should include 'serve'
    End
  End

  Describe '(no args)'
    It 'exits 0 and prints help'
      When call bin/cani alpha
      The status should equal 0
      The stdout should include 'Available Commands'
      The stderr should be defined
    End
  End

End

# ── serve (stub) ────────────────────────────────────────────────────

Describe 'cani alpha serve'

  Describe '--help'
    It 'exits 0 and describes the API server'
      When call bin/cani alpha serve --help
      The status should equal 0
      The stdout should include 'API server'
    End
  End

End
