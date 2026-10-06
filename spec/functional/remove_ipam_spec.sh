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

# ── remove prefix / ip / vrf ────────────────────────────────────────

# Helper: CRUD env with a root prefix holding one address, and a tenant VRF.
#shellcheck disable=SC2317
setup_remove_ipam_env() {
  setup_crud_env
  bin/cani alpha add prefix 10.250.0.0/24 --config "$CANI_CONF" >/dev/null 2>&1
  bin/cani alpha add ip 10.250.0.5/24 --config "$CANI_CONF" >/dev/null 2>&1
  bin/cani alpha add vrf blue --namespace tenant-a --config "$CANI_CONF" >/dev/null 2>&1
}

Describe 'cani alpha remove prefix'
  Before 'setup_remove_ipam_env'

  It 'refuses a root prefix that holds addresses and names --cascade'
    When call bin/cani alpha remove prefix 10.250.0.0/24 --config "$CANI_CONF"
    The status should equal 1
    The stderr should include 'rerun with --cascade'
  End

  It 'removes the prefix and the addresses it holds with --cascade'
    When call bin/cani alpha remove prefix 10.250.0.0/24 --cascade --config "$CANI_CONF"
    The status should equal 0
    The stderr should include 'Removed prefix 10.250.0.0/24 in namespace Global'
    The stderr should include 'and 1 IP address(es) it held'
  End

  It 'rejects an unknown CIDR'
    When call bin/cani alpha remove prefix 10.251.0.0/24 --config "$CANI_CONF"
    The status should equal 1
    The stderr should include 'not found in namespace "Global"'
  End
End

Describe 'cani alpha remove ip'
  Before 'setup_remove_ipam_env'

  It 'removes an address written with its mask'
    When call bin/cani alpha remove ip 10.250.0.5/24 --config "$CANI_CONF"
    The status should equal 0
    The stderr should include 'Removed IP address 10.250.0.5/24 in namespace Global'
  End
End

Describe 'cani alpha remove vrf'
  Before 'setup_remove_ipam_env'

  It 'removes a VRF by namespace-qualified name'
    When call bin/cani alpha remove vrf tenant-a/blue --config "$CANI_CONF"
    The status should equal 0
    The stderr should include 'Removed VRF "blue" in namespace tenant-a'
  End

  It 'lists the interfaces that still name the removed VRF'
    bin/cani alpha add interface eth9 --device test-device --type 1000base-t --vrf tenant-a/blue --config "$CANI_CONF" >/dev/null 2>&1
    When call bin/cani alpha remove vrf tenant-a/blue --config "$CANI_CONF"
    The status should equal 0
    The stderr should include '1 interface(s) still name it (test-device/eth9)'
  End
End
