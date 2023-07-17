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


Describe 'cani session'

# help output should succeed and match the fixture
# a config file should be created if one does not exist
It '--help'
  BeforeCall remove_config # Remove the config to start fresh
  When call bin/cani alpha session --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/session/help'
  AfterCall The path canitest.yml should be exist
  AfterCall The path canitest.yml should be file
End

# Status should be INACTIVE if active: false
It '--config canitest.yml status'
  BeforeCall use_inactive_session # session is inactive
  When call bin/cani alpha session --config canitest.yml status
  The status should equal 0
  The line 1 of stderr should include 'See canitest.yml for session details'
  The line 2 of stderr should include 'Session is INACTIVE'
End


# Status should be ACTIVE if active: true
It '--config canitest.yml status'
  BeforeCall use_active_session # session is active
  When call bin/cani alpha session --config canitest.yml status
  The status should equal 0
  The line 1 of stderr should include 'See canitest.yml for session details'
  The line 2 of stderr should include 'Session is ACTIVE'
End


# Starting a session without passing a provider should fail
It '--config canitest.yml start'
  BeforeCall remove_config
  When call bin/cani alpha session --config canitest.yml start
  The status should equal 1
  The line 1 of stderr should equal 'Error: Need a provider.  Choose from: [csm]'
End

# Starting a session without passing a provider should fail
It '--config canitest.yml start fake'
  BeforeCall remove_config
  When call bin/cani alpha session --config canitest.yml start fake
  The status should equal 1
  The line 1 of stderr should equal 'Error: fake is not a valid provider.  Valid providers: [csm]'
End

# TODO: timeout is slow for tests; renable when simulator is hooked up in pipeline
# Starting a session should fail with:
#  - a valid proivder
#  - no connection to SLS
# It '--config canitest.yml start csm'
#   BeforeCall remove_config
#   BeforeCall remove_datastore
#   When call bin/cani alpha session --config canitest.yml start csm
#   The status should equal 1
#   The line 1 of stderr should include 'canidb.json does not exist, creating default datastore'
#   The line 2 of stderr should include 'No API Gateway token provided, getting one from provider '
#   The line 3 of stderr should include '/keycloak/realms/shasta/protocol/openid-connect/token'
# End

End
