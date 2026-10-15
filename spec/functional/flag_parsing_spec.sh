#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
#

# ── Flag parsing ────────────────────────────────────────────────────
# Behaviour of the flag parser itself, shown on `add device` because it has
# every flag shape; the rules hold for every command.

Describe 'cani flag parsing'
  Before 'setup_crud_env'

  Describe 'a value the flag refuses'
    It 'names the value and the flag'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --position abc --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'invalid argument "abc" for --position'
    End
  End

  Describe 'a flag written without its value'
    It 'does not swallow the flag that follows it'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --tag --status Active --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'flag needs an argument: --tag'
    End

    It 'does not swallow a trailing flag either'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --tag --status --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'flag needs an argument: --tag'
    End

    It 'still accepts a value that merely starts with a dash'
      When call bin/cani alpha add device cray-xd225v --rack test-rack --name -leading-dash --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Added device'
    End
  End
End
