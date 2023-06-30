#!/usr/bin/env sh
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


Describe 'cani add cabinet'
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

# functions to deploy various fixtures with different scenarios
remove_config(){ rm -f canitest.yml; }
remove_datastore() { rm -f canitestdb.json; }
remove_log() { rm -f canitestdb.log; }

use_active_session(){ cp "$FIXTURES"/cani/configs/canitest_valid_active.yml canitest.yml; } # deploys a config with session.active = true
use_inactive_session(){ cp "$FIXTURES"/cani/configs/canitest_valid_inactive.yml canitest.yml; } # deploys a config with session.active = false
use_valid_datastore_system_only(){ cp "$FIXTURES"/cani/configs/canitestdb_valid_system_only.json canitestdb.json; } # deploys a datastore with one system only
use_valid_datastore_one_cabinet(){ cp "$FIXTURES"/cani/configs/canitestdb_valid_one_cabinet.json canitestdb.json; } # deploys a datastore with one cabinet (and child hardware)

# help output should succeed and match the fixture
# a config file should be created if one does not exist
It '--help'
  BeforeCall remove_config # Remove the config to start fresh
  When call bin/cani alpha add cabinet --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/cabinet/help'
  AfterCall The path canitest.yml should be exist
  AfterCall The path canitest.yml should be file
End

# Adding a cabinet withot a hardware type should fail
# it should list the available hardware types
It '--config canitest.yml'
  When call bin/cani alpha add cabinet --config canitest.yml
  The status should equal 1
  The line 1 of stderr should include 'Error: No hardware type provided: Choose from: hpe-eia-cabinet", "hpe-ex2000", "hpe-ex2500-1-liquid-cooled-chassis", "hpe-ex2500-2-liquid-cooled-chassis", "hpe-ex2500-3-liquid-cooled-chassis", "hpe-ex3000", "hpe-ex4000'
End

# Adding a cabinet with an invalid hardware type should fail
It '--config canitest.yml fake-hardware-type'
  When call bin/cani alpha add cabinet --config canitest.yml fake-hardware-type
  The status should equal 1
  The line 1 of stderr should equal 'Error: Invalid hardware type: fake-hardware-type'
End

# Listing hardware types should show available hardware types
It '--config canitest.yml -L'
  When call bin/cani alpha add cabinet --config canitest.yml -L
  The status should equal 0
  The line 1 of stderr should equal "- hpe-eia-cabinet"
  The line 2 of stderr should equal "- hpe-ex2000"
  The line 3 of stderr should equal "- hpe-ex2500-1-liquid-cooled-chassis"
  The line 4 of stderr should equal "- hpe-ex2500-2-liquid-cooled-chassis"
  The line 5 of stderr should equal "- hpe-ex2500-3-liquid-cooled-chassis"
  The line 6 of stderr should equal "- hpe-ex3000"
  The line 7 of stderr should equal "- hpe-ex4000"
End

# Adding a cabinet should fail if no session is active
It '--config canitest.yml hpe-ex2000'
  BeforeCall use_inactive_session # session is inactive
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000
  The status should equal 1
  The line 1 of stderr should equal "Error: No active session.  Run 'session start' to begin"
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore does not exist
It '--config canitest.yml hpe-ex2000'
  BeforeCall use_active_session # session is active
  BeforeCall remove_datastore # datastore does not exist
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000
  The status should equal 1
  The line 1 of stderr should equal "Error: Datastore './canitestdb.json' does not exist.  Run 'session start' to begin"
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is not set
#   - vlan-id flag is not set
It '--config canitest.yml hpe-ex2000'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "cabinet", "vlan-id" not set'
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is not set
It '--config canitest.yml hpe-ex2000 --cabinet 1234'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --cabinet 1234
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "vlan-id" not set'
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is not set
#   - vlan-id flag is set
It '--config canitest.yml hpe-ex2000 --vlan-id 1234'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --vlan-id 1234
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "cabinet" not set'
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is set
#   - vlan-id flag is not within an acceptable range
It '--config canitest.yml hpe-ex2000 --cabinet 1234 --vlan-id 12345678'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --cabinet 1234 --vlan-id 12345678
  The status should equal 1
  The line 1 of stderr should include "Inventory data validation errors encountered"
  The line 2 of stderr should include "System:0->Cabinet:1234"
  The line 3 of stderr should include "    - Specified HMN Vlan (12345678) is invalid, must be in range: 0-4094"
  The line 4 of stderr should equal "Error: data validation failure"
End

# Adding a cabinet should succeed if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is set
#   - the cabinet does not exist
It '--config canitest.yml hpe-ex2000 --cabinet 1234 --vlan-id 1234'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --cabinet 1234 --vlan-id 1234
  The status should equal 0
  The line 1 of stderr should include "Cabinet 1234 was successfully staged to be added to the system"
  The line 2 of stderr should include "UUID: "
  The line 3 of stderr should include "Cabinet Number: 1234"
  The line 4 of stderr should include "VLAN ID: 1234"
End

# Adding a cabinet should fail if (re-run the above command):
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is set
#   - the cabinet already exists
It '--config canitest.yml hpe-ex2000 --cabinet 1234 --vlan-id 1234'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --cabinet 1234 --vlan-id 1234
  The status should equal 1
  The line 1 of stderr should equal "Error: Cabinet number 1234 is already in use"
  The line 2 of stderr should equal "please re-run the command with an available Cabinet number"
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is set
#   - the cabinet does not exist
#   - the vlan already exists 
It '--config canitest.yml hpe-ex2000 --cabinet 4321 --vlan-id 1234'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --cabinet 4321 --vlan-id 1234
  The status should equal 1
  The line 1 of stderr should include "Inventory data validation errors encountered"
  The stderr should include "    - Specified HMN Vlan (1234) is not unique, shared by: "
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - cabinet flag is set
#   - vlan-id flag is not set 
It '--config canitest.yml hpe-ex2000 --auto --vlan-id 1234'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --auto --vlan-id 1234
  The status should equal 1
  The line 1 of stderr should equal "Error: if any flags in the group [cabinet vlan-id] are set they must all be set; missing [cabinet]"
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - cabinet flag is not set
#   - vlan-id flag is set 
It '--config canitest.yml hpe-ex2000 --auto --cabinet 4321'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --auto --cabinet 4321 
  The status should equal 1
  The line 1 of stderr should equal "Error: if any flags in the group [cabinet vlan-id] are set they must all be set; missing [vlan-id]"
End

# Adding a cabinet should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
It '--config canitest.yml hpe-ex2000 --auto --accept'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: "
  The line 7 of stderr should include " VLAN ID: "
End

# Adding a hpe-eia-cabinet should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 3000
#   - vlan id 1513
It '--config canitest.yml hpe-eia-cabinet --auto --accept'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-eia-cabinet --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 3000"
  The line 7 of stderr should include " VLAN ID: 1513"
End

# Adding another hpe-eia-cabinet should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 3001 (incremented by one)
#   - vlan id 1514 (incremented by one)
It '--config canitest.yml hpe-eia-cabinet --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-eia-cabinet --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 3001"
  The line 7 of stderr should include " VLAN ID: 1514"
End

# Adding a hpe-ex2000 should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 9000
#   - vlan id 3000
It '--config canitest.yml hpe-ex2000 --auto --accept'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 9000"
  The line 7 of stderr should include " VLAN ID: 3000"
End

# Adding another hpe-ex2000 should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 9001 (incremented by one)
#   - vlan id 3001 (incremented by one)
It '--config canitest.yml hpe-ex2000 --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 9001"
  The line 7 of stderr should include " VLAN ID: 3001"
End

# Adding a hpe-ex2500-1-liquid-cooled-chassis should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 8000
#   - vlan id 3000
It '--config canitest.yml hpe-ex2500-1-liquid-cooled-chassis --auto --accept'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2500-1-liquid-cooled-chassis --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 8000"
  The line 7 of stderr should include " VLAN ID: 3000"
End

# Adding another hpe-ex2500-1-liquid-cooled-chassis should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 8001 (incremented by one)
#   - vlan id 3001 (incremented by one)
It '--config canitest.yml hpe-ex2500-1-liquid-cooled-chassis --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2500-1-liquid-cooled-chassis --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 8001"
  The line 7 of stderr should include " VLAN ID: 3001"
End

# Adding a hpe-ex2500-2-liquid-cooled-chassis should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 8000
#   - vlan id 3000
It '--config canitest.yml hpe-ex2500-2-liquid-cooled-chassis --auto --accept'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2500-2-liquid-cooled-chassis --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 8000"
  The line 7 of stderr should include " VLAN ID: 3000"
End

# Adding another hpe-ex2500-2-liquid-cooled-chassis should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 8001 (incremented by one)
#   - vlan id 3000 (incremented by one)
It '--config canitest.yml hpe-ex2500-2-liquid-cooled-chassis --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2500-2-liquid-cooled-chassis --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 8001"
  The line 7 of stderr should include " VLAN ID: 3001"
End

# Adding a hpe-ex2500-3-liquid-cooled-chassis should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 8000
#   - vlan id 3000
It '--config canitest.yml hpe-ex2500-3-liquid-cooled-chassis --auto --accept'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2500-3-liquid-cooled-chassis --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 8000"
  The line 7 of stderr should include " VLAN ID: 3000"
End

# Adding another hpe-ex2500-3-liquid-cooled-chassis should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 8001 (incremented by one)
#   - vlan id 3001 (incremented by one)
It '--config canitest.yml hpe-ex2500-3-liquid-cooled-chassis --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex2500-3-liquid-cooled-chassis --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 8001"
  The line 7 of stderr should include " VLAN ID: 3001"
End

# Adding a hpe-ex3000 should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 1000
#   - vlan id 3000
It '--config canitest.yml hpe-ex3000 --auto --accept'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex3000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 1000"
  The line 7 of stderr should include " VLAN ID: 3000"
End

# Adding another hpe-ex3000 should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 1001 (incremented by one)
#   - vlan id 3001 (incremented by one)
It '--config canitest.yml hpe-ex3000 --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex3000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 1001"
  The line 7 of stderr should include " VLAN ID: 3001"
End

# Adding a hpe-ex4000 should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 1002
#   - vlan id 3002
It '--config canitest.yml hpe-ex4000 --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex4000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 1002"
  The line 7 of stderr should include " VLAN ID: 3002"
End

# Adding another hpe-ex4000 should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
# given a system with zero cabinets, the added cabinet should have:
#   - cabinet number 1003 (incremented by one)
#   - vlan id 3003 (incremented by one)
It '--config canitest.yml hpe-ex4000 --auto --accept'
  When call bin/cani alpha add cabinet --config canitest.yml hpe-ex4000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest cabinet number and VLAN ID"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 1003"
  The line 7 of stderr should include " VLAN ID: 3003"
End

End
