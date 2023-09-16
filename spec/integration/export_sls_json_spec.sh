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

It 'init session'
  BeforeCall use_inactive_session
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config
  When call bin/cani alpha session --config "$CANI_CONF" init csm -S
  The status should equal 0
  The line 1 of stderr should include 'Using simulation mode'
End

It 'add cabinet hpe-ex2500-1-liquid-cooled-chassis'
  When call bin/cani alpha --config "$CANI_CONF" add cabinet hpe-ex2500-1-liquid-cooled-chassis --auto --accept
  The status should equal 0
  The line 1 of stderr should include 'Querying inventory to suggest cabinet number and VLAN ID'
  The line 2 of stderr should include 'Suggested cabinet number: 8000'
  The line 3 of stderr should include 'Suggested VLAN ID: 3001'
  The line 4 of stderr should include 'Cabinet 8000 was successfully staged to be added to the system'
End

It 'export sls json'
  When call bin/cani alpha --config "$CANI_CONF" export --format sls-json --validate
  The status should equal 0
  The stderr should include 'GET http'
  The stderr should include 'sls/v1/dumpstate'
  The output should include '"x8000": {'
End

It 'export sls json and parse the json'
  When call sh -c 'bin/cani alpha --config "$CANI_CONF" export --format sls-json | jq'
  The status should equal 0
  The stderr should include 'GET http'
  The stderr should include 'sls/v1/dumpstate'
  The output should include '"x8000": {'
End

It 'add blade --config "$CANI_CONF" hpe-crayex-ex235n-compute-blade --cabinet 8000 --chassis 0 --blade 0'
  When call bin/cani alpha add blade --config "$CANI_CONF" hpe-crayex-ex235n-compute-blade --cabinet 8000 --chassis 0 --blade 0
  The status should equal 0
  The line 2 of stderr should include "NodeBlade was successfully staged to be added to the system"
  The line 3 of stderr should include "UUID: "
  The line 4 of stderr should include "Cabinet: 8000"
  The line 5 of stderr should include "Chassis: 0"
  The line 6 of stderr should include "Blade: 0"
End

It 'export invalid sls data but with valid json'
  When call sh -c 'bin/cani alpha --config "$CANI_CONF" export --format sls-json | jq'
  The status should equal 0
  The stderr should include 'GET http'
  The stderr should include 'sls/v1/dumpstate'
  The output should include '"x8000": {'
End

It 'export invalid sls data with the validate option'
  When call bin/cani alpha --config "$CANI_CONF" export --format sls-json --validate
  The status should equal 1
  The stderr should include 'GET http'
  The stderr should include 'sls/v1/dumpstate'
End

End
