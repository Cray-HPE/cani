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

# ── IPAM CRUD integration ──────────────────────────────────────────
#
# Imports the example dcim.csv, then adds VLANs, prefixes, and IPs,
# verifying each step with "show" commands.

IPAM_DS="/tmp/.cani/ipam_integration_test.json"

Describe 'IPAM integration'

  # ── setup: import the dcim.csv ─────────────────────────────────

  Describe 'setup'
    It 'imports the dcim.csv fixture'
      When call bin/cani alpha import example \
        --csv testdata/fixtures/example/dcim.csv \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End
  End

  # ── add VLANs ───────────────────────────────────────────────────

  Describe 'add vlan'
    It 'adds VLAN 100 (Management)'
      When call bin/cani alpha add vlan 100 --name "Management" \
        --status active --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added VLAN 100'
      The stderr should include 'Management'
    End

    It 'adds VLAN 200 (BMC)'
      When call bin/cani alpha add vlan 200 --name "BMC" \
        --status active --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added VLAN 200'
      The stderr should include 'BMC'
    End

    It 'adds VLAN 300 (HSN)'
      When call bin/cani alpha add vlan 300 --name "HSN" \
        --status active --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added VLAN 300'
    End

    It 'rejects a VLAN without --name'
      When call bin/cani alpha add vlan 101 \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 1
      The stderr should include '--name is required'
    End
  End

  # ── add prefixes ────────────────────────────────────────────────

  Describe 'add prefix'
    It 'adds container prefix 10.0.0.0/16'
      When call bin/cani alpha add prefix 10.0.0.0/16 \
        --type container --role infrastructure \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added prefix 10.0.0.0/16'
    End

    It 'adds network prefix 10.0.1.0/24 with VLAN'
      When call bin/cani alpha add prefix 10.0.1.0/24 \
        --type network --role management --vlan "Management" \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added prefix 10.0.1.0/24'
    End

    It 'adds network prefix 10.0.2.0/24 for BMC'
      When call bin/cani alpha add prefix 10.0.2.0/24 \
        --type network --role bmc --vlan "BMC" \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added prefix 10.0.2.0/24'
    End
  End

  # ── add IP addresses ────────────────────────────────────────────

  Describe 'add ip'
    It 'adds an IP address 10.0.1.1/24'
      When call bin/cani alpha add ip 10.0.1.1/24 \
        --dns-name "switch1-mgmt.example.com" --status active \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added IP address 10.0.1.1/24'
    End

    It 'adds an IP address 10.0.1.10/24'
      When call bin/cani alpha add ip 10.0.1.10/24 \
        --dns-name "node1-mgmt.example.com" --status active \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added IP address 10.0.1.10/24'
    End

    It 'adds a BMC IP 10.0.2.10/24'
      When call bin/cani alpha add ip 10.0.2.10/24 \
        --dns-name "node1-bmc.example.com" --role loopback --status active \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stderr should include 'Added IP address 10.0.2.10/24'
    End
  End

  # ── show VLANs ──────────────────────────────────────────────────

  Describe 'show vlan'
    It 'lists VLANs including the ones we added'
      When call bin/cani alpha show vlan \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stdout should include 'Management'
      The stdout should include 'BMC'
      The stdout should include 'HSN'
    End
  End

  # ── show prefixes ───────────────────────────────────────────────

  Describe 'show prefix'
    It 'lists prefixes including the ones we added'
      When call bin/cani alpha show prefix \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stdout should include '10.0.0.0/16'
      The stdout should include '10.0.1.0/24'
      The stdout should include '10.0.2.0/24'
    End
  End

  # ── show IPs ────────────────────────────────────────────────────

  Describe 'show ip'
    It 'lists IP addresses including the ones we added'
      When call bin/cani alpha show ip \
        --config "$CANI_CONF" --datastore-path "$IPAM_DS"
      The status should equal 0
      The stdout should include '10.0.1.1/24'
      The stdout should include '10.0.1.10/24'
      The stdout should include '10.0.2.10/24'
    End
  End

  # ── cleanup ─────────────────────────────────────────────────────

  Describe 'cleanup'
    It 'removes the test datastore'
      When call rm -f "$IPAM_DS"
      The status should equal 0
    End
  End

End
