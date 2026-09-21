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

# ── remove device ───────────────────────────────────────────────────

#shellcheck disable=SC2317
setup_device_network_removal_env() {
  setup_crud_env
  bin/cani alpha add vrf shared-removal --device test-device --device test-device-2 --config "$CANI_CONF" >/dev/null 2>&1 || return
  bin/cani alpha add ip 10.0.0.1/24 --interface test-device:Management --interface test-device-2:Management --config "$CANI_CONF" >/dev/null 2>&1
}

#shellcheck disable=SC2317
remove_device_then_export() {
  bin/cani alpha remove device test-device --force --config "$CANI_CONF" || return
  bin/cani alpha export example --dry-run --config "$CANI_CONF" >/dev/null
}

Describe 'cani alpha remove device'
  Before 'setup_crud_env'

  Describe 'valid name'
    It 'removes a device by name'
      When call bin/cani alpha remove device test-device --force --config "$CANI_CONF"
      The status should equal 0
      The stderr should include 'Removed device'
    End
  End

  Describe 'invalid name'
    It 'rejects an unknown name'
      When call bin/cani alpha remove device nonexistent-name --force --config "$CANI_CONF"
      The status should equal 1
      The stderr should include 'no item found matching'
    End
  End

  Describe 'network references'
    Before 'setup_device_network_removal_env'

    It 'detaches shared IP and VRF assignments and leaves an exportable datastore'
      When call remove_device_then_export
      The status should equal 0
      The stderr should include 'Removed device'
      The contents of file "$CANI_DS" should not include '16e4c62b-237e-4977-8426-aaec65db017b'
      The contents of file "$CANI_DS" should include 'b7a1c3d4-5e6f-7890-abcd-ef1234567890'
      The contents of file "$CANI_DS" should include '10.0.0.1/24'
      The contents of file "$CANI_DS" should include 'shared-removal'
    End
  End

End
