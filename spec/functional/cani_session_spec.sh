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
Describe 'cani session'

# help output should succeed and match the fixture
# a config file should be created if one does not exist
It '--help'
  BeforeCall remove_config # Remove the config to start fresh
  When call bin/cani alpha session --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/session/help'
  AfterCall The path "$CANI_CONF" should be exist
  AfterCall The path "$CANI_CONF" should be file
End

# Status should be INACTIVE if active: false
It "--config $CANI_CONF status"
  BeforeCall use_inactive_session # session is inactive
  When call bin/cani alpha session --config "$CANI_CONF" status
  The status should equal 0
  The line 1 of stderr should include "No active session"
  The line 2 of stderr should include "Session is INACTIVE for"
End

# Status should be ACTIVE if active: true
It "--config $CANI_CONF status"
  BeforeCall use_active_session # session is active
  When call bin/cani alpha session --config "$CANI_CONF" status
  The status should equal 0
  The stderr should include 'Session is ACTIVE for'
  The stderr should include "See $CANI_CONF for session details"
End

# Starting a session without passing a provider should fail
It "--config $CANI_CONF init"
  BeforeCall remove_config
  When call bin/cani alpha session --config "$CANI_CONF" init
  The status should equal 1
  The line 1 of stderr should equal 'Error: Need a provider.  Choose from: [csm]'
End

# Starting a session without passing a provider should fail
It "--config $CANI_CONF init fake"
  BeforeCall remove_config
  When call bin/cani alpha session --config "$CANI_CONF" init fake
  The status should equal 1
  The line 1 of stderr should equal 'Error: fake is not a valid provider.  Valid providers: [csm]'
End

# Starting a session should fail with:
#  - a valid proivder
#  - no connection to SLS
It "(timeout, no connection to provider) --config $CANI_CONF init csm"
  BeforeCall remove_config
  BeforeCall remove_datastore
  When call bin/cani alpha session --config "$CANI_CONF" init csm
  The status should equal 1
  The line 1 of stderr should include "$CANI_DS does not exist, creating default datastore"
  The line 2 of stderr should include 'No API Gateway token provided, getting one from provider '
  The line 3 of stderr should include 'https://api-gw-service.local/keycloak/realms/shasta/protocol/openid-connect/token'
  The stderr should include 'Failed to get token'
End

It 'initialize a session without a config file or datastore'
  BeforeCall remove_config
  BeforeCall remove_datastore
  BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config
  When call bin/cani alpha session --config "$CANI_CONF" init csm -S
  The status should equal 0
  The stderr should include 'Using simulation mode'
  The stderr should include 'Validated CANI inventory'
  The stderr should include 'Validated external inventory provider'
  # Verify the import logic reached out to SLS
  The stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'

  # Verify the import logic pushed changes into SLS
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

  # Verify the session has started
  The stderr should include 'Session is now ACTIVE with provider csm and datastore'

  # The config should get created
  The path "$CANI_CONF" should be exist
  The path "$CANI_CONF" should be file
End

It 'initialize a session and validate a custom hardware type'
  BeforeCall use_inactive_session
  BeforeCall use_valid_datastore_system_only
  BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config
  When call bin/cani alpha session --config "$CANI_CONF" init csm -S
  The status should equal 0
  The line 1 of stderr should include 'Using simulation mode'
  The stderr should include 'Validated CANI inventory'
  The stderr should include 'Validated external inventory provider'
  # Verify the import logic reached out to SLS
  The stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'

  # Verify the import logic pushed changes into SLS
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

  # Verify the session has started
  The stderr should include 'Session is now ACTIVE with provider csm and datastore'

  # The config should get created
  The path "$CANI_CONF" should be exist
  The path "$CANI_CONF" should be file

  # A hardware-types dir should be created
  The path "$CANI_CUSTOM_HW_DIR" should be exist
  The path "$CANI_CUSTOM_HW_DIR" should be directory
End

It 'initialize a session and validate a custom hardware type'
  BeforeCall use_inactive_session
  BeforeCall use_custom_hw_type
  BeforeCall use_valid_datastore_system_only
  BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config
  When call bin/cani alpha session --config "$CANI_CONF" init csm -S
  The status should equal 0
  The line 1 of stderr should include 'Using simulation mode'
  The stderr should include 'Validated CANI inventory'
  The stderr should include 'Validated external inventory provider'
  # Verify the import logic reached out to SLS
  The stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
  The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'

  # Verify the import logic pushed changes into SLS
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
  The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

  # Verify the session has started
  The stderr should include 'Session is now ACTIVE with provider csm and datastore'

  # The config should get created
  The path "$CANI_CONF" should be exist
  The path "$CANI_CONF" should be file

  # A hardware-types dir should be created
  The path "$CANI_CUSTOM_HW_DIR" should be exist
  The path "$CANI_CUSTOM_HW_DIR" should be directory
End

It 'initialize a session with a CIDR that overlaps k8s values'
  BeforeCall use_inactive_session
  BeforeCall use_valid_datastore_system_only
  BeforeCall "load_sls.sh testdata/fixtures/sls/invalid_networks_k8s_overlap.json" # simulator is running, load a specific SLS config
  When call bin/cani alpha session --config "$CANI_CONF" init csm -S
  The status should equal 1
  The line 1 of stderr should include 'Using simulation mode'
  The stderr should include 'Validated CANI inventory'
  # A session should not start if the CIDRs overlap that of k8s values set by CSI
  The stderr should include 'k8spodscidr 10.32.0.0/12 overlaps with BICAN 10.32.0.0/12'
  The stderr should include 'k8sservicescidr 10.16.0.0/12 overlaps with CAN 10.16.0.0/12'
End

End

# the single-provider fixture has one key with a user-defined value
# this key should be migrated to the new format, as opposed to creating new default values
Describe 'session migration:'
  cat_config(){
    cat "$CANI_CONF" >&2
  }

  # running any command with an older-style config should update the config file to the new format
  # 'migrated config validates group validates the results from this test
  It 'add a cabinet with a single-provider config'
    BeforeCall use_single_provider_session
    BeforeCall use_valid_datastore_system_only
    BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config
    When call bin/cani --config "$CANI_CONF" alpha add cabinet csm hpe-ex4000 --auto --accept
    The status should equal 0
    The stderr should include 'Translating single-provider config to multi-provider'

    # The config should get created
    The path "$CANI_CONF" should be exist
    The path "$CANI_CONF" should be file
    # the single-provider config should be renamed
    The path "$CANI_CONF_SINGLE" should be exist
    The path "$CANI_CONF_SINGLE" should be file
  End


  It 'check config for migrated field'
    When call cat_config
    The status should equal 0
    # the deprecated fixture contains a value of 'migrated'
    # check to see if it now exists in the converted config file
    The stderr should include 'migrated'
  End
End

Describe 'recreate session:'

  # setup the initial state with an active session
  It 'setup for test by initializing the session'
    BeforeCall remove_config
    BeforeCall remove_datastore
    BeforeCall "load_sls.sh testdata/fixtures/sls/valid_hardware_networks.json" # simulator is running, load a specific SLS config

    # Create the initial session
    When call bin/cani alpha session --config "$CANI_CONF" init csm -S
    The status should equal 0
    The stderr should include 'Using simulation mode'
    The stderr should include 'Validated CANI inventory'
    The stderr should include 'Validated external inventory provider'
    # Verify the import logic reached out to SLS
    The stderr should include 'GET https://localhost:8443/apis/sls/v1/dumpstate'
    The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/State/Components'
    The stderr should include 'GET https://localhost:8443/apis/smd/hsm/v2/Inventory/Hardware'

    # Verify the import logic pushed changes into SLS
    The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000'
    The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1'
    The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c1b0'
    The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3'
    The stderr should include 'PUT https://localhost:8443/apis/sls/v1/hardware/x9000c3b0'

    # Verify the session has started
    The stderr should include 'Session is now ACTIVE with provider csm and datastore'

    # The config should get created
    The path "$CANI_CONF" should be exist
    The path "$CANI_CONF" should be file
  End

  It 'session init with N answer'
    When call sh -c '\
        echo "N" | \
        bin/cani alpha session --config "$CANI_CONF" init csm -S '
    The status should equal 0
    The stdout should include 'Keep session active but overwrite the datastore'
    The stderr should include 'Using simulation mode'
    The stderr should not include 'Session is now ACTIVE with provider csm and datastore'

    # The config should still exist
    The path "$CANI_CONF" should be exist
    The path "$CANI_CONF" should be file
  End

  It 'session init with Y answer'
    When call sh -c '\
        echo "Y" | \
        bin/cani alpha session --config "$CANI_CONF" init csm -S '
    The status should equal 0
    The stdout should include 'Keep session active but overwrite the datastore'
    The stderr should include 'Using simulation mode'
    The stderr should include 'Session is now ACTIVE with provider csm and datastore'

    # The config should still exist
    The path "$CANI_CONF" should be exist
    The path "$CANI_CONF" should be file
  End

  It 'session init with force option'
    When call bin/cani alpha session --config "$CANI_CONF" init csm -S --force
    The status should equal 0
    The stdout should not include 'Keep session active but overwrite the datastore'
    The stderr should include 'Using simulation mode'
    The stderr should include 'Session is now ACTIVE with provider csm and datastore'

    # The config should still exist
    The path "$CANI_CONF" should be exist
    The path "$CANI_CONF" should be file
  End

  It 'session init with f option'
    When call bin/cani alpha session --config "$CANI_CONF" init csm -S -f
    The status should equal 0
    The stdout should not include 'Keep session active but overwrite the datastore'
    The stderr should include 'Using simulation mode'
    The stderr should include 'Session is now ACTIVE with provider csm and datastore'

    # The config should still exist
    The path "$CANI_CONF" should be exist
    The path "$CANI_CONF" should be file
  End

End
