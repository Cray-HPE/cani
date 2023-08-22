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

  # Verify the import logic pushed changes into SLS
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

  # Verify the session has started
  The stderr should include 'Session is now ACTIVE with provider csm and datastore'
End

It 'verify empty blade slot'
  When call bin/cani alpha list blade --config canitest.yml
  The status should equal 0
  The line 2 of output should include 'empty		System:0->Cabinet:9000->Chassis:1->NodeBlade:0'
  The line 3 of output should include 'empty		System:0->Cabinet:9000->Chassis:1->NodeBlade:1'
End

It 'verify empty nodes'
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
End

It 'add ex425 blade'
  When call bin/cani alpha --config canitest.yml add blade hpe-crayex-ex425-compute-blade --auto --accept
  The status should equal 0
  The line 1 of stderr should include 'Querying inventory to suggest cabinet, chassis, and blade for this NodeBlade'
  The line 2 of stderr should include 'Suggested Cabinet number: 9000'
  The line 3 of stderr should include 'Suggested Chassis number: 1'
  The line 4 of stderr should include 'Suggested NodeBlade number: 0'
  The line 6 of stderr should include 'NodeBlade was successfully staged to be added to the system'
  The line 7 of stderr should include 'UUID: '
  The line 8 of stderr should include 'Cabinet: 9000'
  The line 9 of stderr should include 'Chassis: 1'
  The line 10 of stderr should include 'Blade: 0'
End

It 'verify staged blade slot'
  When call bin/cani alpha list blade --config canitest.yml
  The status should equal 0
  # four nodes should be added
  The line 2 of stdout should include 'staged'
  The line 2 of stdout should include 'hpe-crayex-ex425-compute-blade'
  The line 2 of stdout should include 'System:0->Cabinet:9000->Chassis:1->NodeBlade:0'
End

It 'verify staged nodes'
  When call bin/cani alpha list node --config canitest.yml
  The status should equal 0
  # four nodes should be added
  The line 2 of stdout should include 'staged'
  The line 2 of stdout should include 'hpe-crayex-ex425-compute-node'
  The line 2 of stdout should include 'Compute'
  The line 2 of stdout should include '[nid001000]'
  The line 2 of stdout should include '1000'
  The line 2 of stdout should include 'System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:0->Node:0'

  The line 3 of stdout should include 'staged'
  The line 3 of stdout should include 'hpe-crayex-ex425-compute-node'
  The line 3 of stdout should include 'Compute'
  The line 3 of stdout should include '[nid001001]'
  The line 3 of stdout should include '1001'
  The line 3 of stdout should include 'System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:0->Node:1'
  
  The line 4 of stdout should include 'staged'
  The line 4 of stdout should include 'hpe-crayex-ex425-compute-node'
  The line 4 of stdout should include 'Compute'
  The line 4 of stdout should include '[nid001002]'
  The line 4 of stdout should include '1002'
  The line 4 of stdout should include 'System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:1->Node:0'

  The line 5 of stdout should include 'staged'
  The line 5 of stdout should include 'hpe-crayex-ex425-compute-node'
  The line 5 of stdout should include 'Compute'
  The line 5 of stdout should include '[nid001003]'
  The line 5 of stdout should include '1003'
  The line 5 of stdout should include 'System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:1->Node:1'
End

It 'commit and reconcile'
  When call bin/cani alpha session --config canitest.yml apply --commit --dryrun
  # committing without node metadata should fail
  The status should equal 0
  The line 1 of stderr should include 'Session is STOPPED'
  The line 2 of stderr should include 'Performing dryrun no changes will be applied to the system!'
  The line 3 of stderr should include 'Committing changes to session'
  The stderr should include 'x9000c1s0b0n0    - Type: Node, Class: Hill, Aliases: [nid001000], Role: Compute, NID: 1000'
  The stderr should include 'x9000c1s0b0n1    - Type: Node, Class: Hill, Aliases: [nid001001], Role: Compute, NID: 1001'
  The stderr should include 'x9000c1s0b1n0    - Type: Node, Class: Hill, Aliases: [nid001002], Role: Compute, NID: 1002'
  The stderr should include 'x9000c1s0b1n1    - Type: Node, Class: Hill, Aliases: [nid001003], Role: Compute, NID: 1003'
  The stdout should include 'Node            (staged)'
End

End
