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

# ── update orphans ──────────────────────────────────────────────────

Describe 'cani alpha update orphans'

  # ── help & flags ────────────────────────────────────────────────

  Describe '--help'
    It 'exits 0 and shows the description'
      When call bin/cani alpha update orphans --help
      The status should equal 0
      The stdout should include 'orphan'
    End

    Describe 'flags'
      Parameters:value --dry-run --apply-plan
      It "has $1 flag"
        When call bin/cani alpha update orphans --help
        The stdout should include "$1"
      End
    End
  End

  # ── apply-plan ──────────────────────────────────────────────────

  Describe '--apply-plan'
    Before 'setup_orphan_env'

    It 'applies a saved plan file'
      When call bin/cani alpha update orphans --apply-plan "$FIXTURES/cani/orphan_plan.json" --config "$CANI_CONF"
      The status should equal 0
      The stdout should include 'Applied'
      The stderr should include 'Applied plan'
    End

    It 'fails with a non-existent plan file'
      When call bin/cani alpha update orphans --apply-plan /tmp/nonexistent-plan.json --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'Error'
    End
  End

End
