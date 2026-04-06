#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2025 Hewlett Packard Enterprise Development LP
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

Describe 'cani ochami tests'

  # ── help & flags ────────────────────────────────────────────────

  Describe 'base command'
    It 'exits 0 and shows help message'
        When call bin/cani alpha import ochami
        The status should equal 0
    End
  End

  Describe 'import empty'
    It 'exits 0 and shows import completed successfully'
      When call bin/cani --datastore-path /tmp/tmpcanidb.json alpha import ochami -f testdata/fixtures/ochami/empty.json
      The status should equal 0
      The stderr should include 'No valid records found in JSON'
      The stderr should include 'Import completed successfully using provider ochami'
    End

    It 'exits 1 and shows no such file or directory'
      When call bin/cani --datastore-path /tmp/tmpcanidb.json alpha import ochami -f badpath
      The status should equal 1
      The stderr should include 'no such file or directory'
    End
  End


  Describe 'import inventory'
    It 'calls the basic inventory import'
      When call bin/cani --datastore-path /tmp/tmpcanidb.json alpha import ochami -f testdata/fixtures/ochami/upload_request.json
      The status should equal 0
      The stderr should include '0 racks'
      The stderr should include '3 devices'
      The stderr should include '0 cables'
      The stderr should include 'Import completed successfully using provider ochami'
    End
  End
End

