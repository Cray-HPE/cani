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

# Functional Redfish provider CLI coverage.

#shellcheck disable=SC2317
redfish_invalid_service_root_import() {
  setup_test_env
  _path="$CANI_DIR/not-redfish-root.json"
  printf '{"not":"a service root"}\n' >"$_path"
  bin/cani alpha import redfish --root "$_path" --config "$CANI_CONF"
}

#shellcheck disable=SC2317
redfish_missing_root_import() {
  setup_test_env
  bin/cani alpha import redfish --root "$CANI_DIR/missing-redfish-root.json" --config "$CANI_CONF"
}

Describe 'cani alpha import redfish'
  AfterAll 'teardown_test_env'

  It 'lists the root file flag'
    When call bin/cani alpha import redfish --help
    The status should equal 0
    The stdout should include '--root'
    The stdout should include 'Path to Redfish ServiceRoot JSON file'
  End

  It 'rejects JSON that is not a ServiceRoot'
    When call redfish_invalid_service_root_import
    The status should equal 1
    The stderr should include 'JSON does not appear to be a Redfish ServiceRoot'
  End

  It 'reports a missing root file path'
    When call redfish_missing_root_import
    The status should equal 1
    The stderr should include 'reading file'
    The stderr should include 'missing-redfish-root.json'
  End
End
