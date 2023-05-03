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
Describe 'add switch'
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

It '(with no args)'
  When call bin/cani add switch
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add switch called'
End

It '--debug add switch'
  When call bin/cani --debug add switch
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add switch called'
  The lines of stderr should equal 1
  The line 1 of stderr should include '{"level":"debug","time":'
End

It '--help'
  When call bin/cani add switch --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/switch/help'
End

It '--horizontal'
  When call bin/cani add switch --horizontal
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --horizontal"
End

It '--horizontal R'
  When call bin/cani add switch --horizontal R
  The status should equal 0
   The lines of stdout should equal 1
  The line 1 of stdout should equal 'add switch called'
End

It '--list-supported-types'
  When call bin/cani add switch --list-supported-types
  The status should equal 0
  The line 1 of stdout should equal 'add switch called'
End

It '--type'
  When call bin/cani add switch --type
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --type"
End

It '--type junk'
  When call bin/cani add switch --type junk
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add switch called'
End

End
