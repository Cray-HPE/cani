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


Describe 'cani add'
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

# Removes the db to start fresh 
# Not called for all tests, but when no duplicate uuids are expected this is called
cleanup(){ rm -rf testdb.json; }

It '--help'
  When call bin/cani add --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/help'
End

It '--database testdb.json'
  BeforeCall 'cleanup' # Remove the db to start fresh
  When call bin/cani add --database testdb.json --uuid abcdef12-3456-abcd-1234-abcdef123456
  The status should equal 1
  The line 1 of stdout should equal 'Error: No asset type specified.'
  The line 1 of stderr should include '"message":"testdb.json does not exist, creating default database"'
End

End
