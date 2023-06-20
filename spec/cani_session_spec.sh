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
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

# functions to deploy various fixtures with different scenarios
cleanup(){ rm -f canitest.*; }
canitest_valid_active(){ cp "$FIXTURES"/cani/configs/canitest_valid_active.yml .; }
canitest_valid_inactive(){ cp "$FIXTURES"/cani/configs/canitest_valid_inactive.yml  .; }
canitest_invalid_datastore_path(){ cp "$FIXTURES"/cani/configs/canitest_invalid_datastore_path.yml .; }
canitest_invalid_log_file_path(){ cp "$FIXTURES"/cani/configs/canitest_invalid_log_file_path.yml .; }
canitest_invalid_provider(){ cp "$FIXTURES"/cani/configs/canitest_invalid_provider.yml .; }
canitest_valid_empty_db(){ cp -f "$FIXTURES"/cani/configs/canitest_valid_empty_db.json .; }
canitest_invalid_empty_db(){ cp -f "$FIXTURES"/cani/configs/canitest_invalid_empty_db.json .; }
rm_canitest_valid_empty_db(){ rm -f canitest_valid_empty_db.json; }


It '--help'
  When call bin/cani session --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/session/help'
End

# Status with a valid config should suceed and show the status of the session
It '--config canitest_invalid_provider.yml status'
  BeforeCall 'canitest_valid_inactive'
  When call bin/cani session --config canitest_invalid_provider.yml status
  The status should equal 0
  The line 1 of stderr should include '"message":"See canitest_invalid_provider.yml for session details"'
  The line 2 of stderr should include '"message":"Session is INACTIVE"'
End

# Starting a session with a bad provider in a config should fail
It '--config canitest_invalid_provider.yml start'
  BeforeCall 'canitest_invalid_provider'
  When call bin/cani session --config canitest_invalid_provider.yml start
  The status should equal 1
  The line 1 of stderr should equal 'Error: fake is not a valid provider.  Valid providers: [csm]'
End

# Starting a session with a valid config should suceed and show the status of the session
It '--config canitest_valid_inactive.yml start'
  BeforeCall 'canitest_valid_inactive'
  BeforeCall 'canitest_valid_empty_db'
  When call bin/cani session --config canitest_valid_inactive.yml start
  The status should equal 0
  The line 1 of stderr should include '"message":"Session is now ACTIVE with provider csm and datastore'
End

# # TODO: Deal with interactive prompts
# It '--config canitest_valid_inactive.yml start'
#   BeforeCall 'canitest_valid_inactive'
#   When call bin/cani session --config canitest_valid_inactive.yml start
#   The status should equal 0
#   The line 1 stderr should include '"message":"Session is now ACTIVE with provider csm and datastore'
# End

# Status should show active after starting a session
It '--config canitest_valid_inactive.yml status (after starting)'
  When call bin/cani session --config canitest_valid_inactive.yml status
  The status should equal 0
  The line 1 of stderr should include '"message":"See canitest_valid_inactive.yml for session details"'
  The line 2 of stderr should include '"message":"Session is ACTIVE"'
End

# # TODO: Deal with interactive prompts
# Status should show active after starting a session
# It 'stop (after starting)'
#   When call bin/cani session stop
#   The status should equal 0
#   The line 1 stderr should include '"message":"Session is STOPPED"'
# End

# # TODO: Deal with interactive prompts
# It '--config canitest_valid_inactive.yml stop'
#   BeforeCall 'canitest_valid_inactive'
#   When call bin/cani session --config canitest_valid_inactive.yml stop
#   The status should equal 1
#   The line 1 stderr should include '"Session with provider 'csm' and datastore './canitest_valid_empty_db.json' is already STOPPED"'
# End

# Stopping a session with the --commit flag should succeed and not prompt the user for anything
It '--config canitest_valid_inactive.yml stop --commit'
  BeforeCall 'canitest_valid_inactive'
  When call bin/cani session --config canitest_valid_inactive.yml stop --commit
  The status should equal 1
  The line 1 of stderr should include "Session with provider 'csm' and datastore './canitest_valid_empty_db.json' is already STOPPED"
  The line 2 of stderr should include '"message":"Committing changes to session"'
End

End
