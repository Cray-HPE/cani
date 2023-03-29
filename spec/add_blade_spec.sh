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

It 'add blade (with no args)'
  When call bin/csminv add blade
  The status should equal 0
  The lines of stdout should equal 1
  The stdout should equal "add blade called"
End

It '--debug add blade'
  When call bin/csminv --debug add blade
  The status should equal 0
  The lines of stdout should equal 1
  The stdout should equal 'add blade called'
End

It 'add blade --status'
  When call bin/csminv add blade --status
  The status should equal 0
  The lines of stdout should equal 6
  The line 1 of stdout should satisfy match_colored_text 'OK cray CLI is initialized'
  The line 2 of stdout should satisfy match_colored_text 'OK DVS is on NMN or HMN'
  The line 3 of stdout should satisfy match_colored_text 'OK Use existing cabinet'
  The line 4 of stdout should satisfy match_colored_text 'OK Slingshot fabric is configured with the desired topology'
  The line 5 of stdout should satisfy match_colored_text 'OK SLS has the desired HSN configuration'
  The line 6 of stdout should satisfy match_colored_text 'OK Check HSN and link status'
End

It '--debug add blade --status'
  When call bin/csminv --debug add blade --status
  The status should equal 0
  The lines of stdout should equal 6
  The line 1 of stdout should satisfy match_colored_text 'OK cray CLI is initialized'
  The line 2 of stdout should satisfy match_colored_text 'OK DVS is on NMN or HMN'
  The line 3 of stdout should satisfy match_colored_text 'OK Use existing cabinet'
  The line 4 of stdout should satisfy match_colored_text 'OK Slingshot fabric is configured with the desired topology'
  The line 5 of stdout should satisfy match_colored_text 'OK SLS has the desired HSN configuration'
  The line 6 of stdout should satisfy match_colored_text 'OK Check HSN and link status'
End

It '--debug add blade blade1'
  When call bin/csminv --debug add blade blade1
  The status should equal 0
  The lines of stdout should equal 1
  The stdout should equal 'add blade called'
  The lines of stderr should equal 1
  The stderr should include '{"level":"debug","time":'
  The stderr should include '"message":"Added blade blade1"}'
End
