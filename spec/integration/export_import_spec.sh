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

It 'add cabinet hpe-ex2500-1-liquid-cooled-chassis'
  When call bin/cani alpha --config "$CANI_CONF" add rack hpe-ex2500-1-liquid-cooled-chassis --auto --accept
  The status should equal 0
  The stderr should include 'Added rack'
  The stderr should include '1 rack(s) added'
End

It 'add blade hpe-crayex-ex235n-compute-blade (orphan)'
  When call bin/cani alpha --config "$CANI_CONF" add hpe-crayex-ex235n-compute-blade
  The status should equal 0
  The stderr should include "1 device(s) added"
End


It 'export'
  When call bin/cani alpha --config "$CANI_CONF" export csm
  The status should equal 0
  The stdout should include "ID"
  The output should include "Cabinet,3001"
  The output should include "Node,,"
  The stderr should include 'Export completed successfully'
End

It 'export id,role,subrole'
  When call bin/cani alpha --config "$CANI_CONF" export csm --headers id,role,subrole
  The status should equal 0
  The line 1 of stdout should equal "ID,Role,SubRole"
  The stderr should include 'Export completed successfully'
End

It 'import vlan change'
  When run command sh -c '\
      cani alpha --config "$CANI_CONF" export csm | \
      sed s/Cabinet,3001/Cabinet,3002/ | \
      cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The stderr should include 'Success: Wrote 1 records of a total'
End

It 'import vlan changes again'
  When run command sh -c '\
      cani alpha --config "$CANI_CONF" export csm > canitest_export.csv; \
      cat canitest_export.csv | \
        sed s/Cabinet,3001/Cabinet,3002/ | \
        cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The stderr should include 'Success: Wrote 0 records of a total'
End

It 'export after cabinets import'
  When call bin/cani alpha --config "$CANI_CONF" export csm --type cabinet --headers id,Type,vlan
  The status should equal 0
  The line 1 of stdout should equal "ID,Type,Vlan"
  The output should include "Cabinet,3002"
  The output should include "Cabinet,3000"
  The output should not include "Node,"
  The stderr should include 'Export completed successfully'
End

It 'import changes to existing compute nodes'
  When run command sh -c '\
      cani alpha --config "$CANI_CONF" export csm --headers type,vlan,role,subrole,nid,alias,name,id > canitest_export.csv; \
      cat canitest_export.csv | \
        sed "0,/Compute,,1000,nid001000/s//Compute,Worker,10000,nid10000/" | \
        sed "0,/Compute,,1001,nid001001/s//Compute,Worker,20000,nid20000/" | \
        cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The stderr should include 'Success: Wrote 2 records of a total'
End

It 'export after node import'
  When call bin/cani alpha --config "$CANI_CONF" export csm --type NODE --headers 'id,Type,role,subrole,nid,alias'
  The status should equal 0
  The line 1 of stdout should equal "ID,Type,Role,SubRole,Nid,Alias"
  The output should include "Node,Compute,Worker,10000,nid10000"
  The output should include "Node,Compute,Worker,20000,nid20000"
  The output should not include "Cabinet,"
  The stderr should include 'Export completed successfully'
End

It 'import changes to nodes that already have metadata'
  When run command sh -c '\
      cani alpha --config "$CANI_CONF" export csm --headers type,vlan,role,subrole,nid,alias,name,id > canitest_export.csv;
      cat canitest_export.csv | \
        sed "0,/Compute,Worker,10000,nid10000/s//Compute,Master,10000,nid10000/" | \
        cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The stderr should include 'Success: Wrote 1 records of a total'
End

It 'export after node second import'
  When call bin/cani alpha --config "$CANI_CONF" export csm --type NODE,nodeblade --headers 'id,Type,role,subrole,nid,alias'
  The status should equal 0
  The line 1 of stdout should equal "ID,Type,Role,SubRole,Nid,Alias"
  The output should include "Node,Compute,Master,10000,nid10000"
  The output should include "Node,Compute,Worker,20000,nid20000"
  The output should not include "Cabinet,"
  The output should include "NodeBlade,"
  The stderr should include 'Export completed successfully'
End

It 'export to simulator with commit'
  When call bin/cani alpha --config "$CANI_CONF" export csm -S --commit
  The status should equal 0
  The stdout should not equal ''
  The stderr should include 'Fetching current SLS state'
  The stderr should include 'Reconcile'
End

End
