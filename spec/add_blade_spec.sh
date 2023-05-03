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

Describe 'add blade'
# Fixtures location ./spec/fixtures
FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"

# compare value to file content
fixture(){
  test "${fixture:?}" == "$( cat "$FIXTURES/$1" )"
}

It '(with no args)'
  When call bin/cani add blade
  The status should equal 0
  The lines of stdout should equal 1
  The stdout should equal "add blade called"
End

It '--debug'
  When call bin/cani --debug add blade
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
  The lines of stderr should equal 1
  The line 1 of stderr should include '"message":"Using default config file'
End


It '--help'
  When call bin/cani add blade --help
  The status should equal 0
  The stdout should satisfy fixture 'cani/add/blade/help'
End


It '-C'
  When call bin/cani add blade -C
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: 'C' in -C"
End

It '-C 1000'
  When call bin/cani add blade -C 1000
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '-C junk'
  When call bin/cani add blade -C junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-C, --cabinet" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--chassis'
  When call bin/cani add blade --chassis
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --chassis"
End

It '--chassis 12'
  When call bin/cani add blade --chassis 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--chassis junk'
  When call bin/cani add blade --chassis junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-c, --chassis" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--hmn-vlan'
  When call bin/cani add blade --hmn-vlan
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --hmn-vlan"
End

It '--hmn-vlan 12'
  When call bin/cani add blade --hmn-vlan 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--hmn-vlan junk'
  When call bin/cani add blade --hmn-vlan junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-v, --hmn-vlan" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--list-supported-types'
  When call bin/cani add blade --list-supported-types
  The status should equal 0
  The line 1 of stdout should equal 'add blade called'
End

It '--port'
  When call bin/cani add blade --port
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --port"
End

It '--port 12'
  When call bin/cani add blade --port 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--port junk'
  When call bin/cani add blade --port junk
  The status should equal 1
  The line 1 of stderr should equal 'Error: invalid argument "junk" for "-p, --port" flag: strconv.ParseInt: parsing "junk": invalid syntax'
End

It '--role'
  When call bin/cani add blade --role
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --role"
End

It '--role 12'
  When call bin/cani add blade --role 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--role junk'
  When call bin/cani add blade --role junk
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--sub-role'
  When call bin/cani add blade --sub-role
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --sub-role"
End

It '--sub-role 12'
  When call bin/cani add blade --sub-role 12
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--sub-role junk'
  When call bin/cani add blade --sub-role junk
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End

It '--type'
  When call bin/cani add blade --type
  The status should equal 1
  The line 1 of stderr should equal "Error: flag needs an argument: --type"
End

It '--type junk'
  When call bin/cani add blade --type junk
  The status should equal 0
  The lines of stdout should equal 1
  The line 1 of stdout should equal 'add blade called'
End
# It '--debug add blade --status'
#   When call bin/cani --debug add blade --status
#   The status should equal 0
#   The lines of stdout should equal 6
#   The line 1 of stdout should satisfy match_colored_text 'OK cray CLI is initialized'
#   The line 2 of stdout should satisfy match_colored_text 'OK DVS is on NMN or HMN'
#   The line 3 of stdout should satisfy match_colored_text 'OK Use existing cabinet'
#   The line 4 of stdout should satisfy match_colored_text 'OK Slingshot fabric is configured with the desired topology'
#   The line 5 of stdout should satisfy match_colored_text 'OK SLS has the desired HSN configuration'
#   The line 6 of stdout should satisfy match_colored_text 'OK Check HSN and link status'
# End

# It '--debug add blade blade1'
#   When call bin/cani --debug add blade blade1
#   The status should equal 0
#   The lines of stdout should equal 1
#   The stdout should equal 'add blade called'
#   The lines of stderr should equal 1
#   The stderr should include '{"level":"debug","time":'
#   The stderr should include '"message":"Added blade blade1"}'
# End
End
