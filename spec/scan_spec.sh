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

It 'scan (with no args)'
  When call bin/cani scan
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "host", "password", "port", "username" not set'
End

It '--debug scan'
  When call bin/cani --debug scan
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "host", "password", "port", "username" not set'
End

It 'scan --host somehost'
  When call bin/cani scan --host somehost
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "password", "port", "username" not set'
End

It '--debug scan --host somehost'
  When call bin/cani --debug scan --host somehost
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "password", "port", "username" not set'
End

It 'scan --host somehost --password somepassword'
  When call bin/cani scan --host somehost --password somepassword
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "port", "username" not set'
End

It '--debug scan --host somehost --password somepassword'
  When call bin/cani --debug scan --host somehost --password somepassword
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "port", "username" not set'
End

It 'scan --host somehost --password somepassword --port 443'
  When call bin/cani scan --host somehost --password somepassword --port 443
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "username" not set'
End

It '--debug scan --host somehost --password somepassword --port 443'
  When call bin/cani --debug scan --host somehost --password somepassword --port 443
  The status should equal 1
  The lines of stderr should equal 14
  The line 1 of stderr should equal 'Error: required flag(s) "username" not set'
End

It 'scan --host somehost --password somepassword --port 443 --username someusername'
  When call bin/cani scan --host somehost --password somepassword --port 443 --username someusername
  The status should equal 1
  The lines of stderr should equal 2
  The line 2 of stderr should include 'dial tcp: lookup somehost'
End

It '--debug scan --host somehost --password somepassword --port 443 --username someusername'
  When call bin/cani --debug scan --host somehost --password somepassword --port 443 --username someusername
  The status should equal 1
  The lines of stderr should equal 2
  The line 2 of stderr should include 'dial tcp: lookup somehost'
End
