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

# ── Interface Roles integration ────────────────────────────────────
#
# Adds a switch device, assigns interface roles via glob patterns,
# verifies roles persist, and exercises the -L (list) flag.

IFACE_DS="/tmp/.cani/interface_roles_integration_test.json"

Describe 'Interface roles integration'

  # ── setup: add a device with interfaces ──────────────────────────

  Describe 'setup'
    It 'adds an HPE Aruba 8325-32C switch'
      When call bin/cani alpha add device hpe-aruba-8325-32c \
        --name "sw-test-leaf-001" \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'Added device'
    End
  End

  # ── list interfaces (-L) ────────────────────────────────────────

  Describe 'list interfaces'
    It 'lists all interfaces on the switch'
      When call bin/cani alpha update interface --device sw-test-leaf-001 -L \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stdout should include 'mgmt'
      The stdout should include '1/1/1'
      The stdout should include '1/1/32'
    End

    It 'shows table headers'
      When call bin/cani alpha update interface --device sw-test-leaf-001 -L \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The stdout should include 'NAME'
      The stdout should include 'TYPE'
      The stdout should include 'ROLE'
    End
  End

  # ── assign roles ────────────────────────────────────────────────

  Describe 'assign roles'
    It 'sets management role on mgmt interface'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "mgmt" --role ManagementInterface \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'Updated interface'
    End

    It 'sets uplink role on a single port'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "1/1/31" --role UplinkInterface \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'Updated interface'
    End

    It 'sets role on multiple ports via glob pattern'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "1/1/[1-4]" --role HSNInterface \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'Updated 4 interfaces'
    End

    It 'sets role on a wildcard range'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "1/1/1*" --role DataInterface \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'Updated 11 interfaces'
    End
  End

  # ── verify roles persist ────────────────────────────────────────

  Describe 'verify persistence'
    It 'shows ManagementInterface role on mgmt'
      When call bin/cani alpha update interface --device sw-test-leaf-001 -L \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stdout should include 'ManagementInterface'
    End

    It 'shows UplinkInterface role on port 31'
      When call bin/cani alpha update interface --device sw-test-leaf-001 -L \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The stdout should include 'UplinkInterface'
    End

    It 'shows DataInterface role on wildcard ports'
      When call bin/cani alpha update interface --device sw-test-leaf-001 -L \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The stdout should include 'DataInterface'
    End
  End

  # ── labels ──────────────────────────────────────────────────────

  Describe 'labels'
    It 'sets a label on an interface'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "mgmt" --label "Out-of-band Management" \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'Updated interface'
    End
  End

  # ── error cases ─────────────────────────────────────────────────

  Describe 'error handling'
    It 'rejects unknown device name'
      When call bin/cani alpha update interface \
        --device nonexistent-device --name "1/1/1" --role hsn \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 1
      The stderr should include 'resolving --device'
    End

    It 'reports when no interfaces match a pattern'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "99/99/*" --role hsn \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 1
      The stderr should include 'no interfaces matched'
    End

    It 'warns on non-standard role but succeeds'
      When call bin/cani alpha update interface \
        --device sw-test-leaf-001 --name "1/1/5" --role CustomRole \
        --config "$CANI_CONF" --datastore-path "$IFACE_DS"
      The status should equal 0
      The stderr should include 'not a well-known role'
    End
  End

End
