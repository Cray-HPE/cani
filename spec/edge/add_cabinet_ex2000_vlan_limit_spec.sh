#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
Describe 'EDGE:'

It 'import from csm'
	BeforeCall remove_datastore
  BeforeCall "curl -sk -X POST -F "sls_dump=@testdata/fixtures/sls/valid_hardware_networks_giant_mnt_networks.json" https://localhost:8443/apis/sls/v1/loadstate"
  When call bin/cani alpha --config "$CANI_CONF" import csm -S
  The status should equal 0
  The stderr should include 'Imported SLS:'
  The stderr should include 'Transform:'
End

End

Describe 'EDGE:'

Parameters:dynamic
  for i in $(seq 1 999); do
    #shellcheck disable=SC2004
    %data "$((9000+$i))" "$((3000+$i))"
  done
End

It 'Add ex2000 cabinet to reach the vlan limit'
  When call bin/cani alpha --config "$CANI_CONF" add rack hpe-ex2000 --auto --accept
  The status should equal 0
  The stderr should include 'Querying inventory to suggest Cabinet'
  The stderr should include "Suggested cabinet number: $1"
  The stderr should include "Suggested VLAN ID: $2"
  The stderr should include "Cabinet was successfully staged to be added to the system"
  The stderr should include "Cabinet Number: $1"
End

End

Describe 'EDGE:'

It 'Add ex2000 cabinet to exceed the vlan limit'
  When call bin/cani alpha --config "$CANI_CONF" add rack hpe-ex2000 --auto --accept
  The status should equal 1
  The stderr should include 'Querying inventory to suggest Cabinet'
  The stderr should include "Suggested cabinet number: 10000"
  The stderr should include "Suggested VLAN ID: 4000"
  The stderr should include "Error: VLAN exceeds the provider's maximum range (3999).  Please choose a valid VLAN"
End

It 'commit and reconcile'
  When call bin/cani alpha --config "$CANI_CONF" export csm -S --commit
  The status should equal 0
  The stderr should include 'Export completed successfully'
  The stdout should include '2997 new hardware item(s) are in the inventory'
End

End
