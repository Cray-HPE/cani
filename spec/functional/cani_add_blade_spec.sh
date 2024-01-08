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
Describe 'cani add blade'

# help output should succeed and match the fixture
# a config file should be created if one does not exist
It '--help'
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade --help --config "$CANI_CONF"
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/blade/help'
  AfterCall The path "$CANI_CONF" should be exist
  AfterCall The path "$CANI_CONF" should be file
End

# Adding a blade withot a hardware type should fail
# it should list the available hardware types
It "--config $CANI_CONF"
  When call bin/cani alpha add blade csm --config "$CANI_CONF"
  The status should equal 1
  The line 1 of stderr should include 'Error: No hardware type provided: Choose from: [hpe-crayex-ex235a-compute-blade hpe-crayex-ex235n-compute-blade hpe-crayex-ex254n-compute-blade hpe-crayex-ex420-compute-blade hpe-crayex-ex425-compute-blade hpe-crayex-ex4252-compute-blade]'
End

# Adding a blade with an invalid hardware type should fail
It "--config $CANI_CONF fake-hardware-type"
  When call bin/cani alpha add blade csm --config "$CANI_CONF" fake-hardware-type
  The status should equal 1
  The line 1 of stderr should equal 'Error: Invalid hardware type: fake-hardware-type'
End

# Listing hardware types should show available hardware types
It "--config $CANI_CONF -L"
  When call bin/cani alpha add blade csm --config "$CANI_CONF" -L
  The status should equal 0
  The line 1 of stdout should equal "hpe-crayex-ex235a-compute-blade"
  The line 2 of stdout should equal "hpe-crayex-ex235n-compute-blade"
  The line 3 of stdout should equal "hpe-crayex-ex254n-compute-blade"
  The line 4 of stdout should equal "hpe-crayex-ex420-compute-blade"
  The line 5 of stdout should equal "hpe-crayex-ex425-compute-blade"
  The line 6 of stdout should equal "hpe-crayex-ex4252-compute-blade"
End

End 


Describe 'cani add blade (each blade type)'

# add each blade type as a parameter, determined dynamically from the current build's output of supported hardware
Parameters:dynamic
  mkdir -p "$CANI_DIR"
  cp "$SHELLSPEC_HELPERDIR/testdata/fixtures/cani/configs/canitest_valid_active.yml" "$CANI_CONF"
  cp "$FIXTURES"/cani/configs/canitestdb_valid_system_only.json "$CANI_DS"
  for blade in $(bin/cani --config "$SHELLSPEC_HELPERDIR/testdata/fixtures/cani/configs/canitest_valid_active.yml" alpha add blade csm -L); do
    %data "$blade"
  done
End

# Adding a blade should fail if no session is active
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 1 --blade 0 (no session)"
  BeforeCall use_inactive_session # session is inactive
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 1 --blade 0
  The status should equal 1
  The line 1 of stderr should include "No active session."
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore does not exist
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 1 --blade 0 (active session, no datastore)"
  BeforeCall use_active_session # session is active
  BeforeCall remove_datastore # datastore does not exist
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 1 --blade 0
  The status should equal 1
  The line 1 of stderr should include "Datastore '$CANI_DS' does not exist.  Run 'session init' to begin"
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is not set
#   - chassis flag is not set
#   - blade flag is not set
It "--config $CANI_CONF $1 (active session, datastore, no flags)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1"
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "blade", "cabinet", "chassis" not set'
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - chassis flag is not set
#   - blade flag is not set
It "--config $CANI_CONF $1 --cabinet 3000 (active session, datastore, some flags)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "blade", "chassis" not set'
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - blade flag is not set
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 0 (active session, datastore, some flags)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 0
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
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 1 --blade 0 (active session, datastore, all flags, no cabinet)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 1 --blade 0
  The status should equal 1
  The line 1 of stderr should equal 'Error: unable to find Cabinet at System:0->Cabinet:3000'
  The line 2 of stderr should equal "try 'list cabinet'"
End

# Adding a blade should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - chassis flag is set
#   - blade flag is set
#   - the cabinet exists
#   - the chassis does not exist
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 1234 --blade 0 (active session, datastore, all flags, no chassis)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_hpe_eia_cabinet_cabinet # deploy a valid datastore with one cabinet
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 1234 --blade 0
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
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 0 --blade 0 (happy path)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_hpe_eia_cabinet_cabinet # deploy a valid datastore with one cabinet and one blade
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 0 --blade 0
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
It "--config $CANI_CONF $1 --cabinet 3000 --chassis 0 --blade 0 (active session, datastore, all flags, existing hardware)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_hpe_eia_cabinet_cabinet_one_blade
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --cabinet 3000 --chassis 0 --blade 0
  The status should equal 1
  The line 1 of stderr should equal "Error: NodeBlade number 0 is already in use"
  The line 2 of stderr should equal "please re-run the command with an available NodeBlade number"
  The line 3 of stderr should equal "try 'cani alpha list blade'"
End

# blade suggestions should fail if there are no empty slots
It "--config $CANI_CONF $1 --auto --accept (no slots available)"
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$1" --auto --accept
  The status should equal 1
  The line 1 of stderr should equal 'Error: no available NodeBlade slots'
End

End



Describe 'cani add blade'

Parameters:dynamic
  # For each cabinet type
  for bld in $(bin/cani --config "$CANI_CONF" alpha add blade csm -L); do
    # add each blade type
    for cab in $(bin/cani --config "$CANI_CONF" alpha add cabinet csm -L); do
      cab_ds_fixture=$(echo "$cab" | awk '{print $1}' | tr '-' '_')
      # ordinals vary depending upon the cabinet
      if [ "$cab_ds_fixture" = "hpe_eia_cabinet" ]; then continue;fi
      if [ "$cab_ds_fixture" = "hpe_ex2000" ]; then cabinet=9000;chassis=1;blade=0;fi
      if [ "$cab_ds_fixture" = "hpe_ex2500_1_liquid_cooled_chassis" ]; then cabinet=8000;chassis=0;blade=0; fi
      if [ "$cab_ds_fixture" = "ex2500_2_liquid_cooled_chassis" ]; then cabinet=8000;chassis=0;blade=0; fi
      if [ "$cab_ds_fixture" = "ex2500_3_liquid_cooled_chassis" ]; then cabinet=8000;chassis=0;blade=0; fi
      if [ "$cab_ds_fixture" = "hpe_ex3000" ]; then cabinet=1000;chassis=0;blade=0; fi
      if [ "$cab_ds_fixture" = "hpe_ex4000" ]; then cabinet=1000;chassis=0;blade=0; fi
      # FIXME: does not work in github actions for some reason
      if [ "$cab_ds_fixture" = "my_custom_cabinet" ]; then continue; fi
      # these vars are used in the tests
      %data "$cab_ds_fixture" "$bld" "$cabinet" "$chassis" "$blade"
    done
  done
End

# check auto adding each blade type to each cabinet type using the dynamic matrix above
It "--config $CANI_CONF $2 --auto --accept (into $1)"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_"$1"_cabinet # deploy a valid datastore with one cabinet
  BeforeCall use_custom_hw_type
  When call bin/cani alpha add blade csm --config "$CANI_CONF" "$2" --auto --accept
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
