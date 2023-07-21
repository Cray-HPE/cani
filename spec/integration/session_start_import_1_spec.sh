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
  # The line 2 of stderr should include 'canidb.json does not exist, creating default datastore'
  # The line 3 of stderr should include 'Session is now ACTIVE with provider csm and datastore canidb.json'
  # The line 4 of stderr should include 'Validated external inventory provider'
End

It 'import from SLS'
  When call bin/cani alpha session --config canitest.yml import
  The status should equal 0
  The line 1 of stderr should include 'Session is STOPPED'
  The line 2 of stderr should include 'Committing changes to session'
  The line 3 of stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The line 4 of stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The line 5 of stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'
  The line 6 of stderr should include 'Cabinet x9000 does not exist in datastore at System:0->Cabinet:9000'
  The line 7 of stderr should include 'Cabinet x9000 device type slug is hpe-ex2000'
End

It 'commit and reconcile'
  When call bin/cani alpha session --config canitest.yml stop --commit
  The status should equal 0
  The line 1 of stderr should include 'Session is STOPPED'
  The line 2 of stderr should include 'Committing changes to session'
  The line 1 of stdout should include 'Summary:'
  The line 2 of stdout should include '--------'
  The line 3 of stdout should include 'ID  TYPE  STATUS'
  The line 5 of stdout should include '0 new hardware item(s) are in the inventory'
End

End
