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

# ── export connections ──────────────────────────────────────────────

Describe 'cani alpha export connections'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha export connections --help
      The status should equal 0
      The stdout should include 'connections'
    End

    Describe 'flags'
      It 'has --format flag'
        When call bin/cani alpha export connections --help
        The stdout should include '--format'
      End
    End
  End

  # ── default YAML format ─────────────────────────────────────────

  Describe 'default format (yaml)'
    Before 'setup_crud_env'

    It 'exports connections as YAML'
      When call bin/cani alpha export connections --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'version'
    End
  End

  # ── explicit YAML format ────────────────────────────────────────

  Describe '--format yaml'
    Before 'setup_crud_env'

    It 'exports connections as YAML explicitly'
      When call bin/cani alpha export connections --format yaml --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'version'
    End
  End

  # ── CSV format ──────────────────────────────────────────────────

  Describe '--format csv'
    Before 'setup_crud_env'

    It 'exports connections as CSV'
      When call bin/cani alpha export connections --format csv --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'a_device'
    End
  End

  # ── invalid format ──────────────────────────────────────────────

  Describe 'invalid format'
    Before 'setup_crud_env'

    It 'rejects unsupported format'
      When call bin/cani alpha export connections --format xml --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'unsupported format'
    End
  End

End
