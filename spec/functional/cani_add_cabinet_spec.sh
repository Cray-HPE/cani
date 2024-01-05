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
Describe 'cani add cabinet'

# help output should succeed and match the fixture
# a config file should be created if one does not exist
It '--help'
  BeforeCall remove_config # Remove the config to start fresh
  When call bin/cani alpha add cabinet --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/cabinet/help'
  AfterCall The path "$CANI_CONF" should be exist
  AfterCall The path "$CANI_CONF" should be file
End

# Adding a cabinet withot a hardware type should fail
# it should list the available hardware types
It "--config $CANI_CONF"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF"
  The status should equal 1
  The line 1 of stderr should include 'Error: No hardware type provided: Choose from: hpe-eia-cabinet", "hpe-ex2000", "hpe-ex2500-1-liquid-cooled-chassis", "hpe-ex2500-2-liquid-cooled-chassis", "hpe-ex2500-3-liquid-cooled-chassis", "hpe-ex3000", "hpe-ex4000'
End

# Adding a cabinet with an invalid hardware type should fail
It "--config $CANI_CONF fake-hardware-type"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" fake-hardware-type
  The status should equal 1
  The line 1 of stderr should equal 'Error: Invalid hardware type: fake-hardware-type'
End

# Listing hardware types should show available hardware types
It "--config $CANI_CONF -L"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" -L
  The status should equal 0
  The line 1 of stdout should equal "hpe-eia-cabinet"
  The line 2 of stdout should equal "hpe-ex2000"
  The line 3 of stdout should equal "hpe-ex2500-1-liquid-cooled-chassis"
  The line 4 of stdout should equal "hpe-ex2500-2-liquid-cooled-chassis"
  The line 5 of stdout should equal "hpe-ex2500-3-liquid-cooled-chassis"
  The line 6 of stdout should equal "hpe-ex3000"
  The line 7 of stdout should equal "hpe-ex4000"
End

# Adding a cabinet should fail if no session is active
It "--config $CANI_CONF hpe-ex2000"
  BeforeCall use_inactive_session # session is inactive
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000
  The status should equal 1
  The line 1 of stderr should include "No active session."
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore does not exist
It "--config $CANI_CONF hpe-ex2000"
  BeforeCall use_active_session # session is active
  BeforeCall remove_datastore # datastore does not exist
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000
  The status should equal 1
  The line 1 of stderr should include "Datastore '$CANI_DS' does not exist.  Run 'session init' to begin"
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is not set
#   - vlan-id flag is not set
It "--config $CANI_CONF hpe-ex2000"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "cabinet", "vlan-id" not set'
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is not set
It "--config $CANI_CONF hpe-ex2000 --cabinet 1234"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --cabinet 1234
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "vlan-id" not set'
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is not set
#   - vlan-id flag is set
It "--config $CANI_CONF hpe-ex2000 --vlan-id 1234"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --vlan-id 1234
  The status should equal 1
  The line 1 of stderr should equal 'Error: required flag(s) "cabinet" not set'
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is set
#   - vlan-id flag is not within an acceptable range
It "--config $CANI_CONF hpe-ex2000 --cabinet 1234 --vlan-id 12345678"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --cabinet 1234 --vlan-id 12345678
  The status should equal 1
  The line 1 of stderr should include "Error: VLAN exceeds the provider's maximum range (3999).  Please choose a valid VLAN"
End

# Adding a cabinet should succeed if:
#   - a session is active
#   - a datastore exists
#   - cabinet flag is set
#   - vlan-id flag is set
#   - the cabinet does not exist
It "--config $CANI_CONF hpe-ex2000 --cabinet 1234 --vlan-id 1234"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_system_only # deploy a valid datastore
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --cabinet 1234 --vlan-id 1234
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
It "--config $CANI_CONF hpe-ex2000 --cabinet 1234 --vlan-id 1234"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --cabinet 1234 --vlan-id 1234
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
It "--config $CANI_CONF hpe-ex2000 --cabinet 4321 --vlan-id 1234"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --cabinet 4321 --vlan-id 1234
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
It "--config $CANI_CONF hpe-ex2000 --auto --vlan-id 1234"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --auto --vlan-id 1234
  The status should equal 1
  The line 1 of stderr should equal "Error: if any flags in the group [cabinet vlan-id] are set they must all be set; missing [cabinet]"
End

# Adding a cabinet should fail if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - cabinet flag is not set
#   - vlan-id flag is set 
It "--config $CANI_CONF hpe-ex2000 --auto --cabinet 4321"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --auto --cabinet 4321 
  The status should equal 1
  The line 1 of stderr should equal "Error: if any flags in the group [cabinet vlan-id] are set they must all be set; missing [vlan-id]"
End

# Adding a cabinet should succeed if:
#   - a session is active
#   - a datastore exists
#   - auto flag is set
#   - accept flag is set
It "--config $CANI_CONF hpe-ex2000 --auto --accept"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" hpe-ex2000 --auto --accept 
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest Cabinet"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: "
  The line 7 of stderr should include " VLAN ID: "
End
End

# matrix to check vlan and cabinet increment properly
Describe "(cabinet and vlan increment sequentially)"
BeforeAll use_active_session # session is active
BeforeAll use_valid_datastore_system_only # deploy a valid datastore

Parameters
  "hpe-eia-cabinet" 3000 1513
  "hpe-eia-cabinet" 3001 1514
  "hpe-ex2000" 9000 3000
  "hpe-ex2000" 9001 3001
  "hpe-ex2500-1-liquid-cooled-chassis" 8000 3002
  "hpe-ex2500-1-liquid-cooled-chassis" 8001 3003
  "hpe-ex2500-2-liquid-cooled-chassis" 8002 3004
  "hpe-ex2500-2-liquid-cooled-chassis" 8003 3005
  "hpe-ex2500-3-liquid-cooled-chassis" 8004 3006
  "hpe-ex2500-3-liquid-cooled-chassis" 8005 3007  
  "hpe-ex3000" 1000 3008
  "hpe-ex3000" 1001 3009
  "hpe-ex4000" 1002 3010
  "hpe-ex4000" 1003 3011
End

It "--config $CANI_CONF $1 --auto --accept"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" "$1" --auto --accept
  The line 1 of stderr should include " Querying inventory to suggest Cabinet"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: $2"
  The line 7 of stderr should include " VLAN ID: $3"
End
End

# matrix to check that the cabinet numbers are chosen in order when non-sequential cabinets exist
Describe "(setup non-sequential cabinet numbers)"
BeforeAll use_active_session # session is active
BeforeAll use_valid_datastore_system_only # deploy a valid datastore

Parameters
  "hpe-eia-cabinet" 3001 1513
  "hpe-eia-cabinet" 3003 1514
  "hpe-eia-cabinet" 3005 1515
  "hpe-eia-cabinet" 3006 1516
  "hpe-ex2000" 9000 3000
  "hpe-ex2000" 9003 3001
  "hpe-ex2000" 9005 3002
  "hpe-ex2000" 9006 3003
  "hpe-ex2500-1-liquid-cooled-chassis" 8000 3004
  "hpe-ex2500-1-liquid-cooled-chassis" 8003 3005
  "hpe-ex2500-1-liquid-cooled-chassis" 8005 3006
  "hpe-ex2500-1-liquid-cooled-chassis" 8006 3007
  "hpe-ex2500-2-liquid-cooled-chassis" 8009 3008
  "hpe-ex2500-2-liquid-cooled-chassis" 8011 3009
  "hpe-ex2500-2-liquid-cooled-chassis" 8012 3010
  "hpe-ex2500-2-liquid-cooled-chassis" 8013 3011
  "hpe-ex2500-3-liquid-cooled-chassis" 8015 3012
  "hpe-ex2500-3-liquid-cooled-chassis" 8016 3013
  "hpe-ex2500-3-liquid-cooled-chassis" 8017 3014
  "hpe-ex2500-3-liquid-cooled-chassis" 8018 3015
  "hpe-ex3000" 1000 3016
  "hpe-ex3000" 1003 3017
  "hpe-ex3000" 1005 3018
  "hpe-ex3000" 1006 3019
  "hpe-ex4000" 1007 3020
  "hpe-ex4000" 1009 3021
  "hpe-ex4000" 1011 3022
  "hpe-ex4000" 1013 3023
End

# setup a bunch of cabinets with incongruent cabinet numbers
It "--config $CANI_CONF $1 --cabinet $2 --vlan-id $3 --accept"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" "$1" --cabinet "$2" --vlan-id "$3" --accept
  The line 1 of stderr should include " was successfully staged to be added to the system"
  The line 2 of stderr should include " UUID: "
  The line 3 of stderr should include " Cabinet Number: $2"
  The line 4 of stderr should include " VLAN ID: $3"
End
End

# matrix to check that the cabinet numbers are chosen in order when non-sequential cabinets exist
Describe "(gaps in non-sequential cabinet numbers are filled when using --auto)"
Parameters
  "hpe-eia-cabinet" 3000 1517
  "hpe-eia-cabinet" 3002 1518
  "hpe-eia-cabinet" 3004 1519
  "hpe-eia-cabinet" 3007 1520
  "hpe-ex2000" 9001 3024
  "hpe-ex2000" 9002 3025
  "hpe-ex2000" 9004 3026
  "hpe-ex2000" 9007 3027
  "hpe-ex2500-1-liquid-cooled-chassis" 8001 3028
  "hpe-ex2500-1-liquid-cooled-chassis" 8002 3029
  "hpe-ex2500-1-liquid-cooled-chassis" 8004 3030
  "hpe-ex2500-1-liquid-cooled-chassis" 8007 3031
  "hpe-ex2500-2-liquid-cooled-chassis" 8008 3032
  "hpe-ex2500-2-liquid-cooled-chassis" 8010 3033
  "hpe-ex2500-2-liquid-cooled-chassis" 8014 3034
  "hpe-ex2500-2-liquid-cooled-chassis" 8019 3035
  "hpe-ex2500-3-liquid-cooled-chassis" 8020 3036
  "hpe-ex2500-3-liquid-cooled-chassis" 8021 3037
  "hpe-ex2500-3-liquid-cooled-chassis" 8022 3038
  "hpe-ex2500-3-liquid-cooled-chassis" 8023 3039
  "hpe-ex3000" 1001 3040
  "hpe-ex3000" 1002 3041
  "hpe-ex3000" 1004 3042
  "hpe-ex3000" 1008 3043
  "hpe-ex4000" 1010 3044
  "hpe-ex4000" 1012 3045
  "hpe-ex4000" 1014 3046
  "hpe-ex4000" 1015 3047
End

# setup a bunch of cabinets with incongruent cabinet numbers
It "--config $CANI_CONF $1 --auto --accept"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" "$1" --auto --accept
  The line 1 of stderr should include " Querying inventory to suggest Cabinet"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: $2"
  The line 7 of stderr should include " VLAN ID: $3"
End
End

# matrix to check that the vlan ids are chosen in order when non-sequential vlan ids exist
Describe "(setup non-sequential vlan ids)"
BeforeAll use_active_session # session is active
BeforeAll use_valid_datastore_system_only # deploy a valid datastore

Parameters
  "hpe-eia-cabinet" 3000 1514
  "hpe-eia-cabinet" 3001 1515
  "hpe-eia-cabinet" 3002 1517
  "hpe-eia-cabinet" 3003 1518
  "hpe-ex2000" 9000 3000
  "hpe-ex2000" 9001 3002
  "hpe-ex2000" 9002 3003
  "hpe-ex2000" 9003 3005
  "hpe-ex2500-1-liquid-cooled-chassis" 8000 3008
  "hpe-ex2500-1-liquid-cooled-chassis" 8001 3009
  "hpe-ex2500-1-liquid-cooled-chassis" 8002 3010
  "hpe-ex2500-1-liquid-cooled-chassis" 8003 3011
  "hpe-ex2500-2-liquid-cooled-chassis" 8004 3014
  "hpe-ex2500-2-liquid-cooled-chassis" 8005 3015
  "hpe-ex2500-2-liquid-cooled-chassis" 8006 3016
  "hpe-ex2500-2-liquid-cooled-chassis" 8007 3017
  "hpe-ex2500-3-liquid-cooled-chassis" 8008 3018
  "hpe-ex2500-3-liquid-cooled-chassis" 8009 3019
  "hpe-ex2500-3-liquid-cooled-chassis" 8010 3020
  "hpe-ex2500-3-liquid-cooled-chassis" 8011 3022
  "hpe-ex3000" 1000 3023
  "hpe-ex3000" 1001 3024
  "hpe-ex3000" 1002 3026
  "hpe-ex3000" 1003 3028
  "hpe-ex4000" 1004 3030
  "hpe-ex4000" 1005 3033
  "hpe-ex4000" 1006 3034
  "hpe-ex4000" 1007 3035
End

# setup a bunch of cabinets with incongruent cabinet numbers
It "--config $CANI_CONF $1 --cabinet $2 --vlan-id $3 --accept"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" "$1" --cabinet "$2" --vlan-id "$3" --accept
  The line 1 of stderr should include " was successfully staged to be added to the system"
  The line 2 of stderr should include " UUID: "
  The line 3 of stderr should include " Cabinet Number: $2"
  The line 4 of stderr should include " VLAN ID: $3"
End
End

# matrix to check that the vlan ids are chosen in order when non-sequential vlan ids exist
Describe "(gaps in non-sequential vlan ids are filled when using --auto)"
Parameters
  "hpe-eia-cabinet" 3004 1513
  "hpe-eia-cabinet" 3005 1516
  "hpe-eia-cabinet" 3006 1519
  "hpe-eia-cabinet" 3007 1520
  "hpe-ex2000" 9004 3001
  "hpe-ex2000" 9005 3004
  "hpe-ex2000" 9006 3006
  "hpe-ex2000" 9007 3007
  "hpe-ex2500-1-liquid-cooled-chassis" 8012 3012
  "hpe-ex2500-1-liquid-cooled-chassis" 8013 3013
  "hpe-ex2500-1-liquid-cooled-chassis" 8014 3021
  "hpe-ex2500-1-liquid-cooled-chassis" 8015 3025
  "hpe-ex2500-2-liquid-cooled-chassis" 8016 3027
  "hpe-ex2500-2-liquid-cooled-chassis" 8017 3029
  "hpe-ex2500-2-liquid-cooled-chassis" 8018 3031
  "hpe-ex2500-2-liquid-cooled-chassis" 8019 3032
  "hpe-ex2500-3-liquid-cooled-chassis" 8020 3036
  "hpe-ex2500-3-liquid-cooled-chassis" 8021 3037
  "hpe-ex2500-3-liquid-cooled-chassis" 8022 3038
  "hpe-ex2500-3-liquid-cooled-chassis" 8023 3039
  "hpe-ex3000" 1008 3040
  "hpe-ex3000" 1009 3041
  "hpe-ex3000" 1010 3042
  "hpe-ex3000" 1011 3043
  "hpe-ex4000" 1012 3044
  "hpe-ex4000" 1013 3045
  "hpe-ex4000" 1014 3046
  "hpe-ex4000" 1015 3047
End

# setup a bunch of cabinets with incongruent cabinet numbers
It "--config $CANI_CONF $1 --auto --accept"
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" "$1" --auto --accept
  The line 1 of stderr should include " Querying inventory to suggest Cabinet"
  The line 2 of stderr should include " Suggested cabinet number: "
  The line 3 of stderr should include " Suggested VLAN ID: "
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: $2"
  The line 7 of stderr should include " VLAN ID: $3"
End

It 'validates a custom hardware type appears in the list of supported hardware'
  BeforeCall use_active_session
  BeforeCall use_custom_hw_type
  BeforeCall use_valid_datastore_system_only
  When call bin/cani --config "$CANI_CONF" alpha add cabinet csm -L
  The status should equal 0
  The line 8 of stdout should equal 'my-custom-cabinet'
End

It "--config $CANI_CONF my-custom-cabinet --auto --accept"
  BeforeCall use_active_session
  BeforeCall use_custom_hw_type
  BeforeCall use_valid_datastore_system_only
  When call bin/cani alpha add cabinet csm --config "$CANI_CONF" my-custom-cabinet --auto --accept
  The status should equal 0
  The line 1 of stderr should include " Querying inventory to suggest Cabinet"
  The line 2 of stderr should include " Suggested cabinet number: 4321"
  The line 3 of stderr should include " Suggested VLAN ID: 1111"
  The line 4 of stderr should include " was successfully staged to be added to the system"
  The line 5 of stderr should include " UUID: "
  The line 6 of stderr should include " Cabinet Number: 4321"
  The line 7 of stderr should include " VLAN ID: 1111"
End

It 'validate cabinet is added'
  BeforeCall use_custom_hw_type
  When call bin/cani alpha list cabinet csm --config "$CANI_CONF"
  The status should equal 0
  The line 2 of stdout should include "staged"
  The line 2 of stdout should include "my-custom-cabinet"
  The line 2 of stdout should include "1111"
  The line 2 of stdout should include "System:0->Cabinet:4321"
End

End
