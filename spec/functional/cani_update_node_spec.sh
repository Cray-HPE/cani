#!/usr/bin/env sh
#
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
#
#TODO: parameterize
Describe 'cani update node'

# check auto adding each blade type to each cabinet type using the dynamic matrix above
It "validate blade exists"
  BeforeCall use_active_session # session is active
  BeforeCall use_valid_datastore_one_ex2000_one_blade # deploy a valid datastore with one cabinet
  When call bin/cani alpha list blade --config "$CANI_CONF"
  The status should equal 0
  The line 2 of stdout should include '1d87f035-6c89-4670-a86d-61bb09f1928c'
End

It "validate nodes exist"
  When call bin/cani alpha list node --config "$CANI_CONF"
  The status should equal 0
  The line 2 of stdout should include 'd28c3201-9dfa-4788-852f-86054cd99232'
  The line 3 of stdout should include '8ab879e0-f269-4c88-aaf4-4a345985d41e'
End

It "update node --config $CANI_CONF --uuid 8ab879e0-f269-4c88-aaf4-4a345985d41e --role Application --subrole Worker --nid 1234"
  When call bin/cani alpha update node --config "$CANI_CONF" --uuid 8ab879e0-f269-4c88-aaf4-4a345985d41e --role Application --subrole Worker --nid 1234
  The status should equal 0
  The line 1 of stderr should include 'Updated node'
End

It "validate node is updated"
  When call bin/cani alpha list node --config "$CANI_CONF"
  The status should equal 0
  The line 3 of stdout should include '8ab879e0-f269-4c88-aaf4-4a345985d41e'
  The line 3 of stdout should include 'Application'
  The line 3 of stdout should include 'Worker'
  The line 3 of stdout should include '1234'
End

End
