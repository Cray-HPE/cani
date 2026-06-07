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

# ── show metadata ───────────────────────────────────────────────────

Describe 'cani alpha show metadata'

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha show metadata --help
      The status should equal 0
      The stdout should include 'metadata'
    End
  End

  Describe 'with data'
    Before 'setup_crud_env'

    It 'shows metadata after adding a role'
      # First add a role, then show metadata
      When call sh -c "bin/cani alpha add metadata role show-test-role --config '$CANI_CONF' && bin/cani alpha show metadata --config '$CANI_CONF'"
      The status should equal 0
      The stdout should include 'show-test-role'
      The stderr should include 'Added role'
    End
  End

End
