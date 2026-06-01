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

# ── IPAM commands (vlan, prefix, ip) ────────────────────────────────

Describe 'cani alpha add vlan'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add vlan --help
      The status should equal 0
      The stdout should include 'Add a VLAN'
    End

    Describe 'flags'
      Parameters:value --name --location --description --status
      It "has $1 flag"
        When call bin/cani alpha add vlan --help
        The stdout should include "$1"
      End
    End
  End

  # ── argument validation (fail tests) ───────────────────────────

  Describe 'validation'
    It 'fails without a VID argument'
      When call bin/cani alpha add vlan
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End

    It 'fails with a non-integer VID'
      When call bin/cani alpha add vlan abc --name "Test"
      The status should equal 1
      The stderr should include 'invalid VLAN ID'
    End

    It 'fails with VID out of range (0)'
      When call bin/cani alpha add vlan 0 --name "Test"
      The status should equal 1
      The stderr should include 'VLAN ID must be between 1 and 4094'
    End

    It 'fails with VID out of range (4095)'
      When call bin/cani alpha add vlan 4095 --name "Test"
      The status should equal 1
      The stderr should include 'VLAN ID must be between 1 and 4094'
    End

    It 'fails without --name'
      When call bin/cani alpha add vlan 100
      The status should equal 1
      The stderr should include '--name is required'
    End
  End
End

Describe 'cani alpha add prefix'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add prefix --help
      The status should equal 0
      The stdout should include 'Add an IP prefix'
    End

    Describe 'flags'
      Parameters:value --type --role --vlan --vrf --location --description
      It "has $1 flag"
        When call bin/cani alpha add prefix --help
        The stdout should include "$1"
      End
    End
  End

  # ── argument validation (fail tests) ───────────────────────────

  Describe 'validation'
    It 'fails without a CIDR argument'
      When call bin/cani alpha add prefix
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End
  End
End

Describe 'cani alpha add ip'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add ip --help
      The status should equal 0
      The stdout should include 'Add an IP address'
    End

    Describe 'flags'
      Parameters:value --interface --type --role --dns-name --description
      It "has $1 flag"
        When call bin/cani alpha add ip --help
        The stdout should include "$1"
      End
    End
  End

  # ── argument validation (fail tests) ───────────────────────────

  Describe 'validation'
    It 'fails without a CIDR argument'
      When call bin/cani alpha add ip
      The status should equal 1
      The stderr should include 'accepts 1 arg(s), received 0'
    End
  End
End

Describe 'cani alpha show vlan'

  Describe '--help'
    It 'exits 0 and describes listing VLANs'
      When call bin/cani alpha show vlan --help
      The status should equal 0
      The stdout should include 'VLAN'
    End
  End
End

Describe 'cani alpha show prefix'

  Describe '--help'
    It 'exits 0 and describes listing prefixes'
      When call bin/cani alpha show prefix --help
      The status should equal 0
      The stdout should include 'prefix'
    End
  End

  Describe 'flags'
    It 'has --tree flag'
      When call bin/cani alpha show prefix --help
      The stdout should include '--tree'
    End
  End
End

Describe 'cani alpha show ip'

  Describe '--help'
    It 'exits 0 and describes listing IPs'
      When call bin/cani alpha show ip --help
      The status should equal 0
      The stdout should include 'IP address'
    End
  End

  Describe 'flags'
    It 'has --prefix flag'
      When call bin/cani alpha show ip --help
      The stdout should include '--prefix'
    End
  End
End
