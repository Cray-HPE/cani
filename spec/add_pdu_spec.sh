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
Describe 'add pdu'
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

It '(with no args)'
  When call bin/cani add pdu
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add pdu called'
End

It '--debug add pdu'
  When call bin/cani --debug add pdu
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add pdu called'
  The lines of stderr should equal 1
  The line 1 of stderr should include '{"level":"debug","time":'
End

It '--help'
  When call bin/cani add pdu --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/pdu/help'
End

It '-C'
  When call bin/cani add pdu -C
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: 'C' in -C"
End

It '-C 1000'
  When call bin/cani add pdu -C 1000
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add pdu called'
End

It '-C junk'
  When call bin/cani add pdu -C junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-C, --cabinet" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--id'
  When call bin/cani add pdu --id
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --id"
End

It '--id 12'
  When call bin/cani add pdu --id 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add pdu called'
End

It '--port'
  When call bin/cani add pdu --port
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --port"
End

It '--port 12'
  When call bin/cani add pdu --port 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add pdu called'
End

It '--port junk'
  When call bin/cani add pdu --port junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-p, --port" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--list-supported-types'
  When call bin/cani add pdu --list-supported-types
  The status should equal 0
  The line 1 of stdout should equal 'add pdu called'
End

It '--type'
  When call bin/cani add pdu --type
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --type"
End

It '--type junk'
  When call bin/cani add pdu --type junk
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add pdu called'
End

End
