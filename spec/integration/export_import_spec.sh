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
  When call bin/cani alpha --config "$CANI_CONF" add cabinet csm hpe-ex2500-1-liquid-cooled-chassis --auto --accept
  The status should equal 0
  The line 1 of stderr should include 'Querying inventory to suggest Cabinet'
  The line 2 of stderr should include 'Suggested cabinet number: 8000'
  The line 3 of stderr should include 'Suggested VLAN ID: 3001'
  The line 4 of stderr should include 'Cabinet was successfully staged to be added to the system'
  The line 6 of stderr should include "Cabinet Number: 8000"
End

It 'add blade csm --config "$CANI_CONF" hpe-crayex-ex235n-compute-blade --cabinet 8000 --chassis 0 --blade 0'
  When call bin/cani alpha add blade csm --config "$CANI_CONF" hpe-crayex-ex235n-compute-blade --cabinet 8000 --chassis 0 --blade 0
  The status should equal 0
  The line 2 of stderr should include "NodeBlade was successfully staged to be added to the system"
  The line 3 of stderr should include "UUID: "
  The line 4 of stderr should include "Cabinet: 8000"
  The line 5 of stderr should include "Chassis: 0"
  The line 6 of stderr should include "Blade: 0"
End


It 'export'
  When call bin/cani alpha --config "$CANI_CONF" export csm
  The status should equal 0
  The line 1 of stdout should include "ID"
  The output should include "Cabinet,3001"
  The output should include "Node,,,"
End

It 'export id,role,subrole'
  When call bin/cani alpha --config "$CANI_CONF" export csm --headers id,role,subrole
  The status should equal 0
  The line 1 of stdout should equal "ID,Role,SubRole"
End

It 'import vlan change'
  When call sh -c '\
      bin/cani alpha --config "$CANI_CONF" export csm | \
      sed s/Cabinet,3001/Cabinet,4000/ | \
      bin/cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The line 1 of stderr should include 'Success: Wrote 1 records of a total'
End

It 'import vlan changes again'
  When call sh -c '\
      bin/cani alpha --config "$CANI_CONF" export csm > canitest_export.csv; \
      cat canitest_export.csv | \
        sed s/Cabinet,3001/Cabinet,4000/ | \
        bin/cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The line 1 of stderr should include 'Success: Wrote 0 records of a total'
End

It 'export after cabinets import'
  When call bin/cani alpha --config "$CANI_CONF" export csm --type cabinet --headers id,Type,vlan
  The status should equal 0
  The line 1 of stdout should equal "ID,Type,Vlan"
  The output should include "Cabinet,4000"
  The output should include "Cabinet,3000"
  The output should not include "Node,"
End

It 'import changes to nodes that do not have metadata'
  When call sh -c '\
      bin/cani alpha --config "$CANI_CONF" export csm --headers type,vlan,role,subrole,nid,alias,name,id > canitest_export.csv; \
      cat canitest_export.csv | \
        sed "0,/Node,,,,,,,/s//Node,,Compute,,10000,nid10000,,/" | \
        sed "0,/Node,,,,,,,/s//Node,,Compute,,20000,nid20000,,/" | \
        bin/cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The line 1 of stderr should include 'Success: Wrote 2 records of a total'
End

It 'export after node import'
  When call bin/cani alpha --config "$CANI_CONF" export csm --type NODE --headers 'id,Type,role,nid,alias'
  The status should equal 0
  The line 1 of stdout should equal "ID,Type,Role,Nid,Alias"
  The output should include "Node,Compute,10000,nid10000"
  The output should include "Node,Compute,20000,nid20000"
  The output should not include "Cabinet,"
End

It 'import changes to nodes that already have metadata'
  When call sh -c '\
      bin/cani alpha --config "$CANI_CONF" export csm --headers type,vlan,role,subrole,nid,alias,name,id > canitest_export.csv;
      cat canitest_export.csv | \
        sed "0,/Node,,Compute,,10000,nid10000,,/s//Node,,Compute,Worker,10000,nid10000,,/" | \
        bin/cani alpha --config "$CANI_CONF" import csm'
  The status should equal 0
  The line 1 of stderr should include 'Success: Wrote 1 records of a total'
End

It 'export after node second import'
  When call bin/cani alpha --config "$CANI_CONF" export csm --type NODE,nodeblade --headers 'id,Type,role,subrole,nid,alias'
  The status should equal 0
  The line 1 of stdout should equal "ID,Type,Role,SubRole,Nid,Alias"
  The output should include "Node,Compute,Worker,10000,nid10000"
  The output should include "Node,Compute,,20000,nid20000"
  The output should not include "Cabinet,"
  The output should include "NodeBlade,"
End

It 'apply and reconcile session'
  When call bin/cani alpha session --config "$CANI_CONF" apply --commit
  The status should equal 0
  The line 1 of stderr should include 'Session is STOPPED'
  The line 2 of stderr should include 'Committing changes to session'
  The line 1 of stdout should include 'Summary:'
  The line 2 of stdout should include '--------'
  The line 3 of stdout should include 'ID'
End

End
