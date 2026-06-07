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

# ── add connections ─────────────────────────────────────────────────

Describe 'cani alpha add connections'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha add connections --help
      The status should equal 0
      The stdout should include 'connections'
    End

    Describe 'flags'
      Parameters:value --dry-run
      It "has $1 flag"
        When call bin/cani alpha add connections --help
        The stdout should include "$1"
      End
    End
  End

  # ── argument validation ─────────────────────────────────────────

  Describe 'validation'
    Before 'setup_connections_env'

    It 'shows help without a file argument'
      When call bin/cani alpha add connections --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'connections'
    End

    It 'fails with a non-existent file'
      When call bin/cani alpha add connections /tmp/nonexistent-file.yml --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'no such file or directory'
    End

    It 'fails with an invalid file extension'
      When call bin/cani alpha add connections /tmp/badfile.txt --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'Error'
    End

    It 'fails with YAML missing version field'
      When call bin/cani alpha add connections "$FIXTURES/cani/connections_noversion.yml" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include "missing required 'version' field"
    End
  End

  # ── YAML connections ────────────────────────────────────────────

  Describe 'YAML input'
    Before 'setup_connections_env'

    It 'creates cables from a YAML connection map'
      When call bin/cani alpha add connections "$FIXTURES/cani/connections_test.yml" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'cable(s) created'
    End
  End

  # ── CSV connections ─────────────────────────────────────────────

  Describe 'CSV input'
    Before 'setup_connections_env'

    It 'creates cables from a CSV file'
      When call bin/cani alpha add connections "$FIXTURES/cani/connections_test.csv" --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'cable(s) created'
    End
  End

  # ── dry-run mode ────────────────────────────────────────────────

  Describe '--dry-run'
    Before 'setup_connections_env'

    It 'shows proposed cables without persisting'
      When call bin/cani alpha add connections "$FIXTURES/cani/connections_test.yml" --dry-run --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'test-device'
    End
  End

End
