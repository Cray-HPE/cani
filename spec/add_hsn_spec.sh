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
Describe 'add hsn'
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

It '(with no args)'
  When call bin/cani add hsn
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add hsn called'
End

It '--debug add hsn'
  When call bin/cani --debug add hsn
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add hsn called'
  The lines of stderr should equal 1
  The line 1 of stderr should include '{"level":"debug","time":'
End

It '--help'
  When call bin/cani add hsn --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/hsn/help'
End

It '-C'
  When call bin/cani add hsn -C
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: 'C' in -C"
End

It '-C 1000'
  When call bin/cani add hsn -C 1000
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add hsn called'
End

It '-C junk'
  When call bin/cani add hsn -C junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-C, --cabinet" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--port'
  When call bin/cani add hsn --port
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --port"
End

It '--port 12'
  When call bin/cani add hsn --port 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add hsn called'
End

It '--port junk'
  When call bin/cani add hsn --port junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-p, --port" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--list-supported-types'
  When call bin/cani add hsn --list-supported-types
  The status should equal 0
  The line 1 of stdout should equal 'add hsn called'
End

It '--type'
  When call bin/cani add hsn --type
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --type"
End

It '--type junk'
  When call bin/cani add hsn --type junk
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add hsn called'
End

End
