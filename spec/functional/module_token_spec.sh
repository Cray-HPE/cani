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

# ── {module} placeholder in module port names ───────────────────────
#
# Uses the fixture library in testdata/fixtures/cani/types/library: an
# "example-server" with module bays PCIe1 (position 1) and PCIe2 (position 2),
# and an "example-dual-port-nic" whose ports are "HSN {module}/0" and
# "HSN {module}/1", as in forge-devicetype-library. cani replaces {module}
# with the bay position, as Nautobot does.

setup_library_env() {
  setup_crud_env
  cp -R "$FIXTURES/cani/types/library" "$CANI_DIR/types"
}

cani_lib() {
  bin/cani alpha "$@" --config "$CANI_CONF" --types-dirs "$CANI_DIR/types"
}

add_server_nic() {
  cani_lib add device example-server --rack test-rack --position 32 --name srv-01 >/dev/null 2>&1 &&
  cani_lib add module example-dual-port-nic --device srv-01 --bay PCIe1 --name nic-01 >/dev/null 2>&1
}

# move_nic moves the NIC to PCIe2, then targets a port by its new name.
move_nic() {
  cani_lib update module nic-01 --bay PCIe2 >/dev/null 2>&1 &&
  cani_lib update interface --module nic-01 --name "HSN 2/1" --role DataInterface
}

Describe 'module {module} placeholder'
  Before 'setup_library_env' 'add_server_nic'
  After 'teardown_test_env'

  It 'names the ports after the bay position'
    When call cani_lib update interface --module nic-01 --name "HSN 1/0" --role DataInterface
    The status should equal 0
    The stderr should include 'Updated'
  End

  It 'leaves no port with the literal placeholder'
    When call cani_lib update interface --module nic-01 --name "HSN {module}/0" --role DataInterface
    The status should equal 1
    The stderr should include 'no interfaces matched'
  End

  It 'renames the ports when the module moves to another bay'
    When call move_nic
    The status should equal 0
    The stderr should include 'Updated'
  End
End
