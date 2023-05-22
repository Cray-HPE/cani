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
  When call bin/cani add blade --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/blade/help'
End

# Should create the config file if one does not exist
# No hardware type passed should show list of available hardware types
It '--config canitest.yml'
  BeforeCall 'cleanup'
  When call bin/cani add blade --config canitest.yml
  The status should equal 1
  The line 1 of stderr should include '"message":"canitest.yml does not exist, creating default config file"}'
  The line 2 of stderr should include 'Error: No hardware type provided: Choose from: [hpe-crayex-ex235a-compute-blade hpe-crayex-ex235n-compute-blade hpe-crayex-ex420-compute-blade hpe-crayex-ex425-compute-blade]'
End

# Config file exists now
# Passing invalid hardware type should fail
It '--config canitest.yml fake-hardware-type'
  When call bin/cani add blade --config canitest.yml fake-hardware-type
  The status should equal 1
  The line 1 of stderr should equal 'Error: Invalid hardware type: fake-hardware-type'
End

# Listing hardware types should show available hardware types
It '--config canitest.yml -L'
  When call bin/cani add blade --config canitest.yml add blade -L
  The status should equal 0
  The line 1 of stderr should equal "- hpe-crayex-ex235a-compute-blade"
  The line 2 of stderr should equal "- hpe-crayex-ex235n-compute-blade"
  The line 3 of stderr should equal "- hpe-crayex-ex420-compute-blade"
  The line 4 of stderr should equal "- hpe-crayex-ex425-compute-blade"
End

# Adding a valid hardware type should fail if no session is active
It '--config canitest.yml hpe-crayex-ex235a-compute-blade'
  When call bin/cani add blade --config canitest.yml hpe-crayex-ex235a-compute-blade
  The status should equal 1
  The line 1 of stderr should equal "Error: No active session.  Run 'session start' to begin"
End

# Adding a valid hardware type should succeed if a session is active
# Use a fixture instead of the one generated automatically by the test
It '--config canitest_valid_inactive.yml hpe-crayex-ex235n-compute-blade'
  BeforeCall 'canitest_valid_inactive'
  When call bin/cani add blade --config canitest_valid_inactive.yml hpe-crayex-ex235n-compute-blade
  The status should equal 1
  The line 1 of stderr should equal "Error: No active session.  Run 'session start' to begin"
End

# If a valid hardware type is passed and a session is active, it should succeed
# The JSON file should be updated with the new hardware type
It '--config canitest_valid_active.yml hpe-crayex-ex235n-compute-blade'
  BeforeCall 'canitest_valid_active'
  BeforeCall 'canitest_valid_empty_db'
  When call bin/cani add blade --config canitest_valid_active.yml hpe-crayex-ex235n-compute-blade
  The status should equal 0
  The line 1 of stderr should include '"message":"Added blade hpe-crayex-ex235n-compute-blade"}'
  The contents of file "canitest_valid_empty_db.json" should include "EX235N AMD NVIDIA accelerator blade (Grizzly Peak)"
End

# If a session is valid, but the datastore path is invalid, it should fail
It '--config canitest_invalid_provider.yml hpe-crayex-ex235n-compute-blade'
  BeforeCall 'canitest_valid_active'
  BeforeCall 'rm_canitest_valid_empty_db'
  When call bin/cani add blade --config canitest_valid_active.yml hpe-crayex-ex235n-compute-blade
  The status should equal 1
  The line 1 of stderr should equal "Error: Datastore './canitest_valid_empty_db.json' does not exist.  Run 'session start' to begin"
End

End
