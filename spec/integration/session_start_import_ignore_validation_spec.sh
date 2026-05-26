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

It 'attempt to start a session with failures'
	BeforeCall remove_datastore
	BeforeCall "curl -sk -X POST -F "sls_dump=@testdata/fixtures/sls/valid_hardware_networks.json" https://localhost:8443/apis/sls/v1/loadstate"
  When call bin/cani alpha --config "$CANI_CONF" import csm -S
  The status should not equal 0
  The stderr should include 'External inventory is unstable'
End

It 'start a session ignoring validation failures'
	BeforeCall remove_datastore
  When call bin/cani alpha --config "$CANI_CONF" import csm -S --ignore-validation
  The status should equal 0
  The stderr should include 'Cabinet x9000 does not exist in datastore at System:0->Cabinet:9000'
  The stderr should include 'Cabinet x9000 device type slug is hpe-ex2000'
  The stderr should include 'WRN Ignoring these failures'
End

It 'commit and reconcile ignoring validation failures'
  When call bin/cani alpha --config "$CANI_CONF" export csm -S --commit --ignore-validation
  The status should equal 0
  The stderr should include 'Export completed successfully'
  The stdout should include 'Summary:'
  The stdout should include '--------'
  The stdout should include 'ID  TYPE  STATUS'
  The stdout should include '0 new hardware item(s) are in the inventory'
End

End
