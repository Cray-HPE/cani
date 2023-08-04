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


Describe 'cani add blade'

# add each blade type as a parameter
Parameters:value "hpe-crayex-ex235a-compute-blade" "hpe-crayex-ex235n-compute-blade" "hpe-crayex-ex420-compute-blade" "hpe-crayex-ex425-compute-blade"

# help output should succeed and match the fixture
# a config file should be created if one does not exist
It '--help'
  BeforeCall remove_config # Remove the config to start fresh
  When call bin/cani alpha add blade --help --config canitest.yml
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/blade/help'
  AfterCall The path canitest.yml should be exist
  AfterCall The path canitest.yml should be file
End

# Adding a blade withot a hardware type should fail
# it should list the available hardware types
It '--config canitest.yml'
  When call bin/cani alpha add blade --config canitest.yml
  The status should equal 1
  The line 1 of stderr should include 'Error: No hardware type provided: Choose from: [hpe-crayex-ex235a-compute-blade hpe-crayex-ex235n-compute-blade hpe-crayex-ex420-compute-blade hpe-crayex-ex425-compute-blade]'
End

# Adding a blade with an invalid hardware type should fail
It '--config canitest.yml fake-hardware-type'
  When call bin/cani alpha add blade --config canitest.yml fake-hardware-type
  The status should equal 1
  The line 1 of stderr should equal 'Error: Invalid hardware type: fake-hardware-type'
End

# Listing hardware types should show available hardware types
It '--config canitest.yml -L'
  When call bin/cani alpha add blade --config canitest.yml -L
  The status should equal 0
  The line 1 of stderr should equal "- hpe-crayex-ex235a-compute-blade"
  The line 2 of stderr should equal "- hpe-crayex-ex235n-compute-blade"
  The line 3 of stderr should equal "- hpe-crayex-ex420-compute-blade"
  The line 4 of stderr should equal "- hpe-crayex-ex425-compute-blade"
End

# Adding a blade should fail if no session is active
It '--config canitest.yml hpe-crayex-ex235a-compute-blade --cabinet 3000 --chassis 1 --blade 0'
  BeforeCall use_inactive_session # session is inactive
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235a-compute-blade --cabinet 3000 --chassis 1 --blade 0
  The status should equal 1
  The line 1 of stderr should equal "Error: No active session.  Run 'session start' to begin"
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore does not exist
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 1 --blade 0'
  BeforeCall use_active_session # session is active
  BeforeCall remove_datastore # datastore does not exist
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 1 --blade 0
  The status should equal 1
  The line 1 of stderr should equal "Error: Datastore './canitestdb.json' does not exist.  Run 'session start' to begin"
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is not set
#   - chassis flag is not set
#   - blade flag is not set
It '--config canitest.yml hpe-crayex-ex235n-compute-blade'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "blade", "cabinet", "chassis" not set'
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - chassis flag is not set
#   - blade flag is not set
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "blade", "chassis" not set'
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - blade flag is not set
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 0'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 0
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "blade", not set'
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - chassis flag is set
#   - blade flag is set
#   - the cabinet does not exist
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 1 --blade 0'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 1 --blade 0
  The status should equal 1
  The line 1 of stderr should equal 'Error: unable to find Cabinet at System:0->Cabinet:3000'
  The line 2 of stderr should equal "try 'go run main.go alpha list cabinet'"
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - chassis flag is set
#   - blade flag is set
#   - the cabinet exists
#   - the chassis does not exist
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 1234 --blade 0'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_eia_cabinet # deploy a valid datastore with one cabinet
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 1234 --blade 0
  The status should equal 1
  The line 1 of stderr should equal 'Error: in order to add a NodeBlade, a Chassis is needed'
  The line 2 of stderr should equal "unable to find Chassis at System:0->Cabinet:3000->Chassis:1234"
End

# Adding a blade should succeed if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - chassis flag is set
#   - blade flag is set
#   - the cabinet exists
#   - the chassis exists
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 0 --blade 0'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_eia_cabinet # deploy a valid datastore with one cabinet
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 0 --blade 0
  The status should equal 0
  The line 2 of stderr should include "NodeBlade was successfully staged to be added to the system"
  The line 3 of stderr should include "UUID: "
  The line 4 of stderr should include "Cabinet: 3000"
  The line 5 of stderr should include "Chassis: 0"
  The line 6 of stderr should include "Blade: 0"
End

# (re-run the last command) Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - chassis flag is set
#   - blade flag is set
#   - the cabinet exists
#   - the chassis exists
#   - the blade already exists
It '--config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 0 --blade 0'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add blade --config canitest.yml hpe-crayex-ex235n-compute-blade --cabinet 3000 --chassis 0 --blade 0
  The status should equal 1
  The line 1 of stderr should equal "Error: NodeBlade number 0 is already in use"
  The line 2 of stderr should equal "please re-run the command with an available NodeBlade number"
  The line 3 of stderr should equal "try 'cani alpha list blade'"
End

# blade suggestions should fail if there are no empty slots
It "--config canitest.yml $1 --auto --accept"
  When call bin/cani alpha add blade --config canitest.yml "$1" --auto --accept
  The status should equal 1
  The line 1 of stderr should equal 'Error: no available NodeBlade slots'
End

End



Describe 'cani add blade'

Parameters:dynamic
  # For each cabinet type
  for cab in "use_valid_datastore_one_ex2000_cabinet" "use_valid_datastore_one_ex2500_1_cabinet" "use_valid_datastore_one_ex2500_2_cabinet" "use_valid_datastore_one_ex2500_3_cabinet" "use_valid_datastore_one_ex3000_cabinet" "use_valid_datastore_one_ex4000_cabinet"; do
    # add each blade type
    for bld in "hpe-crayex-ex235a-compute-blade" "hpe-crayex-ex235n-compute-blade" "hpe-crayex-ex420-compute-blade" "hpe-crayex-ex425-compute-blade"; do
      # ordinals vary depending upon the cabinet
      if [ "$cab" = "use_valid_datastore_one_ex2000_cabinet" ]; then cabinet=9000;chassis=1;blade=0; fi
      if [ "$cab" = "use_valid_datastore_one_ex2500_1_cabinet" ]; then cabinet=8000;chassis=0;blade=0; fi
      if [ "$cab" = "use_valid_datastore_one_ex2500_2_cabinet" ]; then cabinet=8000;chassis=0;blade=0; fi
      if [ "$cab" = "use_valid_datastore_one_ex2500_3_cabinet" ]; then cabinet=8000;chassis=0;blade=0; fi
      if [ "$cab" = "use_valid_datastore_one_ex3000_cabinet" ]; then cabinet=1000;chassis=0;blade=0; fi
      if [ "$cab" = "use_valid_datastore_one_ex4000_cabinet" ]; then cabinet=1000;chassis=0;blade=0; fi
      # these vars are used in the tests
      %data "$cab" "$bld" "$cabinet" "$chassis" "$blade"
    done
  done
End

# check auto adding each blade type to each cabinet type using the dynamic matrix above
It "--config canitest.yml $2 --auto --accept"
  BeforeCall use_active_session # session is active
  BeforeCall "$1" # deploy a valid datastore with one cabinet
  When call bin/cani alpha add blade --config canitest.yml "$2" --auto --accept
  The status should equal 0
  The line 1 of stderr should include 'Querying inventory to suggest cabinet, chassis, and blade for this NodeBlade'
  The line 2 of stderr should include "Suggested Cabinet number: $3"
  The line 3 of stderr should include "Suggested Chassis number: $4"
  The line 4 of stderr should include "Suggested NodeBlade number: $5"
  The line 6 of stderr should include 'NodeBlade was successfully staged to be added to the system'
  The line 7 of stderr should include 'UUID: '
  The line 8 of stderr should include "Cabinet: $3"
  The line 9 of stderr should include "Chassis: $4"
  The line 10 of stderr should include "Blade: $5"
End

End
