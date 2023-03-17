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

Describe 'Hardware types'
  Parameters
    'Liquid-cooled cabinet'
    'Air-cooled cabinet'
    'Management network switch'
  End

  Todo "The user interface must allow for the addition and removal of $1"  
End

Todo 'Users must be able to add a single unit of hardware'
Todo 'Users must be able to bulk at several units of hardware'
Todo 'Users must be able to remove a single unit of hardware'
Todo 'Users must be able to bulk remove several units of hardware'
Todo 'Users must be able to extract the inventory on a running CSM machine to a portable format (the long term goal being that this inventory can be ingested to make reinstallation easier) (stretch goal)'
Todo 'For phase 1, only hardware added after CSM install is in scope'
Todo 'The user interface solution should not directly parse SHCD documents, but it must be assumed that admins may be using an SHCD when adding or removing hardware'
Todo 'Development must consider that while admins will ultimately have a need to query hardware inventory, system role/classification (e.g. return all management storage nodes or return all UAN nodes), system network information, or cabling/connectivity information, this data will likely not be a part of hardware inventory.  The design of future data structure changes and APIs should consider this carefully.'
