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

# ── import command ──────────────────────────────────────────────────

#shellcheck disable=SC2317
import_example_header_only_system_csv() {
  _path="$CANI_DIR/example-system-header-only.csv"
  printf 'Section,Name\n' >"$_path"
  bin/cani alpha import example --csv "$_path" --config "$CANI_CONF"
}

Describe 'cani alpha import'

  Describe '--help'
    It 'exits 0 and describes importing assets'
      When call bin/cani alpha import --help
      The status should equal 0
      The stdout should include 'Import assets into the inventory'
    End

    Describe 'flags'
      Parameters:value --phase --no-color --step
      It "has $1 flag"
        When call bin/cani alpha import --help
        The stdout should include "$1"
      End
    End

    It 'lists provider subcommands'
      When call bin/cani alpha import --help
      The stdout should include 'example'
    End
  End

  Describe 'example --help'
    Parameters:value --csv --file
    It "has $1 flag"
      When call bin/cani alpha import example --help
      The status should equal 0
      The stdout should include "$1"
    End
  End

  Describe 'example CSV validation'
    Before 'setup_test_env'
    After  'teardown_test_env'

    It 'rejects a BOM CSV missing the description column'
      When call bin/cani alpha import example --csv "$FIXTURES/example/missing_column.csv" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'failed to parse CSV: missing required column: Description'
    End

    It 'rejects a header-only BOM CSV'
      When call bin/cani alpha import example --csv "$FIXTURES/example/empty.csv" --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'failed to parse CSV: CSV must have a header row and at least one data row'
    End

    It 'routes a Section header to the system CSV parser and rejects header-only input'
      When call import_example_header_only_system_csv
      The status should equal 1
      The stderr should include 'failed to parse system CSV: system CSV must have a header row and at least one data row'
    End
  End

  Describe 'nautobot --help'
    Parameters:value --default-location --default-role --default-status
    It "has $1 flag"
      When call bin/cani alpha import nautobot --help
      The status should equal 0
      The stdout should include "$1"
    End
  End

End
