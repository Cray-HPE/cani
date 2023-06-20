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


Describe 'cani add pdu'
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
  When call bin/cani add pdu --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/pdu/help'
End

End
