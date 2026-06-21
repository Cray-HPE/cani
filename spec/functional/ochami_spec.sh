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

# Functional OpenCHAMI/Ochami provider CLI coverage.

#shellcheck disable=SC2317
ochami_missing_json_import() {
  setup_test_env
  bin/cani alpha import ochami --jsonfile "$CANI_DIR/missing-ochami.json" --config "$CANI_CONF"
}

Describe 'cani alpha import ochami'
  AfterAll 'teardown_test_env'

  It 'lists the JSON file flag'
    When call bin/cani alpha import ochami --help
    The status should equal 0
    The stdout should include '--jsonfile'
    The stdout should include 'Ochami JSON inventory file to import'
  End

  It 'reports a missing JSON file path'
    When call ochami_missing_json_import
    The status should equal 1
    The stderr should include 'failed to open Json file'
    The stderr should include 'missing-ochami.json'
  End
End
