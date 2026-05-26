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
Describe 'INTEGRATION:'

It 'import from simulator'
	BeforeCall remove_datastore
	BeforeCall "curl -sk -X POST -F "sls_dump=@testdata/fixtures/sls/valid_hardware_networks.json" https://localhost:8443/apis/sls/v1/loadstate"
  When call bin/cani alpha --config "$CANI_CONF" import csm -S --ignore-validation
  The status should equal 0
  The stderr should include 'Import completed successfully using provider csm'
End

It 'add ex4000 cabinet'
  When call bin/cani alpha --config "$CANI_CONF" add hpe-ex4000 --auto --accept
  The status should equal 0
  The stderr should include 'Querying inventory to suggest Cabinet'
  The stderr should include 'Suggested cabinet number: 1000'
  The stderr should include 'Suggested VLAN ID: 3001'
  The stderr should include 'Cabinet was successfully staged to be added to the system'
  The stderr should include "Cabinet Number: 1000"
End

It 'export to simulator'
  When call bin/cani alpha --config "$CANI_CONF" export csm -S --commit
  The status should equal 0
  The stderr should include 'Export completed successfully'
  The stdout should include 'Cabinet                         (Staged)'
End

End
