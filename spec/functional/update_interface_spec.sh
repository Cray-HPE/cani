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

# ── update interface ────────────────────────────────────────────────

Describe 'cani alpha update interface'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha update interface --help
      The status should equal 0
      The stdout should include 'Update one or more interfaces'
    End

    It 'shows examples in the long description'
      When call bin/cani alpha update interface --help
      The stdout should include 'List interfaces on a device'
    End

    Describe 'flags'
      Parameters:value --device --name --role --label --list
      It "has $1 flag"
        When call bin/cani alpha update interface --help
        The stdout should include "$1"
      End
    End

    It 'has -L shorthand for --list'
      When call bin/cani alpha update interface --help
      The stdout should include '-L'
    End
  End

  # ── argument validation ─────────────────────────────────────────

  Describe 'validation'
    It 'requires --role or --label'
      When call bin/cani alpha update interface --device foo --name bar --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'at least one of --role or --label must be specified'
    End

    It 'requires --device when no positional UUID'
      When call bin/cani alpha update interface --name bar --role hsn --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--device'
    End

    It 'requires --name when using --device'
      When call bin/cani alpha update interface --device foo --role hsn --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--name is required'
    End

    It '-L requires --device'
      When call bin/cani alpha update interface -L --config "$CANI_CONF"
      The status should equal 1
      The stderr should include '--device is required'
    End

    It 'rejects an invalid UUID'
      When call bin/cani alpha update interface not-a-uuid --role hsn --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'invalid interface UUID'
    End
  End

  # ── CRUD with real fixture ──────────────────────────────────────

  Describe 'CRUD'
    Before 'setup_crud_env'

    It 'sets role on an interface by device + name'
      When call bin/cani alpha update interface --device test-device --name "GigabitEthernet0/0/1" --role DataInterface --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated interface'
    End

    It 'sets label on an interface'
      When call bin/cani alpha update interface --device test-device --name "GigabitEthernet0/0/1" --label "Primary Uplink" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Updated interface'
    End

    It 'warns on unknown role but succeeds'
      When call bin/cani alpha update interface --device test-device --name "GigabitEthernet0/0/1" --role MyCustomRole --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'not a well-known role'
      The stderr should include 'Updated interface'
    End

    It 'lists interfaces with -L'
      When call bin/cani alpha update interface --device test-device -L --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'GigabitEthernet0/0/1'
      The stdout should include 'NAME'
      The stdout should include 'TYPE'
      The stdout should include 'ROLE'
      The stdout should include 'SOURCE'
    End

    It 'rejects unknown device with -L'
      When call bin/cani alpha update interface --device nonexistent -L --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'resolving --device'
    End

    It 'returns error when name pattern matches nothing'
      When call bin/cani alpha update interface --device test-device --name "NoSuchInterface*" --role hsn --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'no interfaces matched'
    End
  End

End
