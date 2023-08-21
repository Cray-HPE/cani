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
  When call bin/cani alpha session --config canitest.yml init csm -S
  The status should equal 0
  The line 1 of stderr should include 'Using simulation mode'
  The stderr should include 'Validated CANI inventory'
  The stderr should include 'Validated external inventory provider'
  # Verify the import logic reached out to SLS
  The stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'
  The stderr should include 'Cabinet x9000 does not exist in datastore at System:0->Cabinet:9000'
  The stderr should include 'Cabinet x9000 device type slug is hpe-ex2000'

  # Verify the import logic pushed changes into SLS
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

  # Verify the session has started
  The stderr should include 'Session is now ACTIVE with provider csm and datastore'
End

It 'Verify imported cabinets'
  When call bin/cani alpha list cabinet --config canitest.yml
  The status should equal 0
  The line 2 of output should include 'provisioned	hpe-ex2000	3000		System:0->Cabinet:9000'
End

It 'Verify imported chassis'
  When call bin/cani alpha list chassis --config canitest.yml
  The status should equal 0
  The line 2 of output should include 'provisioned	hpe-crayex-chassis	System:0->Cabinet:9000->Chassis:1'
  The line 3 of output should include 'provisioned	hpe-crayex-chassis	System:0->Cabinet:9000->Chassis:3'
End

It 'Verify imported blades'
  When call bin/cani alpha list blade --config canitest.yml
  The status should equal 0
  The line 2 of output should include 'empty		System:0->Cabinet:9000->Chassis:1->NodeBlade:0'
  The line 3 of output should include 'empty		System:0->Cabinet:9000->Chassis:1->NodeBlade:1'
  # Note there are more nodes present in CANI, but only checking the first 2
End

It 'Verify imported empty nodes'
  When call bin/cani alpha list node --config canitest.yml
  The status should equal 0
  The line 2 of output should include 'empty		Compute		[nid001000]	1000	System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:0->Node:0'
  The line 3 of output should include 'empty		Compute		[nid001001]	1001	System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:0->Node:1'
  The line 4 of output should include 'empty		Compute		[nid001002]	1002	System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:1->Node:0'
  The line 5 of output should include 'empty		Compute		[nid001003]	1003	System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:1->Node:1'
  The line 6 of output should include 'empty		Compute		[nid001004]	1004	System:0->Cabinet:9000->Chassis:1->NodeBlade:1->NodeCard:0->Node:0'
  The line 7 of output should include 'empty		Compute		[nid001005]	1005	System:0->Cabinet:9000->Chassis:1->NodeBlade:1->NodeCard:0->Node:1'
  The line 8 of output should include 'empty		Compute		[nid001006]	1006	System:0->Cabinet:9000->Chassis:1->NodeBlade:1->NodeCard:1->Node:0'
  The line 9 of output should include 'empty		Compute		[nid001007]	1007	System:0->Cabinet:9000->Chassis:1->NodeBlade:1->NodeCard:1->Node:1'
  # Note there are more nodes present in CANI, but only checking the first 8
End

It 'commit and reconcile'
  When call bin/cani alpha session --config canitest.yml apply --commit
  The status should equal 0
  The line 1 of stderr should include 'Session is STOPPED'
  The line 2 of stderr should include 'Committing changes to session'
  The line 1 of stdout should include 'Summary:'
  The line 2 of stdout should include '--------'
  The line 3 of stdout should include 'ID  TYPE  STATUS'
  The line 5 of stdout should include '0 new hardware item(s) are in the inventory'
End

End
