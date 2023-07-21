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
  When call bin/cani alpha session --config canitest.yml start csm -S
  The status should equal 0
  The line 1 of stderr should include 'Using simulation mode'
End

It 'import from SLS'
  When call bin/cani alpha session --config canitest.yml import
  The status should equal 0
  The line 1 of stderr should include 'Session is STOPPED'
  The line 2 of stderr should include 'Committing changes to session'
  The line 3 of stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The line 4 of stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The line 5 of stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'
End

It 'add ex4000 cabinet'
  When call bin/cani alpha --config canitest.yml add cabinet hpe-ex4000 --auto --accept
  The status should equal 0
  The line 1 of stderr should include 'Querying inventory to suggest cabinet number and VLAN ID'
  The line 2 of stderr should include 'Suggested cabinet number: 1000'
  The line 3 of stderr should include 'Suggested VLAN ID: 3001'
  The line 4 of stderr should include 'Cabinet 1000 was successfully staged to be added to the system'
End

It 'commit and reconcile'
  When call bin/cani alpha session --config canitest.yml stop --commit
  The status should equal 0
  The stderr should include 'Hardware added to the system'
  The stderr should include 'x1000            - Type: Cabinet, Class: Mountain, Networks: {"cn":{}}'
  The stdout should include 'Cabinet                         (staged)'
End

End
