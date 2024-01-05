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
Describe 'INTEGRATION:'

It 'start a session'
  BeforeCall use_inactive_session
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config
  When call bin/cani alpha session --config "$CANI_CONF" init csm -S
  The status should equal 0
  The line 1 of stderr should include 'Using simulation mode'
  The stderr should include 'Validated CANI inventory'
  The stderr should include 'Validated external inventory provider'
  # Verify the import logic reached out to SLS
  The stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'

  # Verify the import logic pushed changes into SLS
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

  # Verify the session has started
  The stderr should include 'Session is now ACTIVE with provider csm and datastore'
End

It 'add ex2000 cabinet'
  When call bin/cani alpha --config "$CANI_CONF" add cabinet csm hpe-ex2000 --auto --accept
  The status should equal 0
  The line 1 of stderr should include 'Querying inventory to suggest Cabinet'
  The line 2 of stderr should include 'Suggested cabinet number: 9001'
  The line 3 of stderr should include 'Suggested VLAN ID: 3001'
  The line 4 of stderr should include 'Cabinet 9001 was successfully staged to be added to the system'
End

It 'commit and reconcile'
  When call bin/cani alpha session --config "$CANI_CONF" apply --commit --dryrun
  The status should equal 0
  The stderr should include 'Performing dryrun no changes will be applied to the system!'
  The stderr should include 'Hardware added to the system'
  The stderr should include 'x9001            - Type: Cabinet, Class: Hill, Networks: {"cn":{"HMN":{"CIDR":"10.104.4.0/22","Gateway":"10.104.4.1","VLan":3001},"NMN":{"CIDR":"10.100.4.0/22","Gateway":"10.100.4.1","VLan":2001}}}'
  The stdout should include 'Cabinet                         (staged)'
  The stderr should include 'Dryrun enabled, no changes performed!'
End

End
