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

# ── module placement: spares, diagnostics, bay rules, port ownership ──
#
# Uses the CRUD inventory: rack "test-rack", device "test-device" with module
# bays PCIe1 and PCIe2, and "test-module", which has neither a parent device
# nor a location. Modules are the embedded hpe-2p-25gbe-pcie-nic type, whose
# ports are "Port 1" and "Port 2" on every instance, so two installed NICs
# carry identical port names and only module ownership tells them apart.
#
# Placement warnings are produced when the inventory is rebuilt after a
# mutation and are printed with --debug; a plain show loads silently.

cani_crud() {
  bin/cani alpha "$@" --config "$CANI_CONF"
}

# add_spare stores a NIC as a spare at test-site; placement warnings for the
# rest of the inventory are printed alongside because of --debug.
add_spare() {
  cani_crud add module hpe-2p-25gbe-pcie-nic --spare --location test-site --name spare-nic --debug
}

# spare_placement prints the spare's parent device and location UUIDs.
spare_placement() {
  cani_crud show module spare-nic -o json 2>/dev/null |
    jq -r '(if type == "array" then .[0] else . end) | "\(.parentDevice) \(.location != null and .location != "00000000-0000-0000-0000-000000000000")"'
}

# legacy_bay_env parents test-module on test-device without naming a bay, the
# shape inventories written before module bays were named have.
legacy_bay_env() {
  setup_crud_env
  dev=$(jq -r '.devices | to_entries[] | select(.value.name == "test-device") | .key' "$CANI_DS")
  jq --arg dev "$dev" '(.modules[] | select(.name == "test-module")).parentDevice = $dev' "$CANI_DS" > "$CANI_DS.tmp" &&
    mv "$CANI_DS.tmp" "$CANI_DS"
}

# install_two_nics puts one NIC in each bay of test-device and marks only
# nic-a's "Port 1" as a data interface.
install_two_nics() {
  cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay PCIe1 --name nic-a >/dev/null 2>&1 &&
  cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay PCIe2 --name nic-b >/dev/null 2>&1 &&
  cani_crud update interface --module nic-a --name 'Port 1' --role DataInterface
}

# port1_summary prints, for every interface named "Port 1": the count, the
# number of distinct owning modules, and how many carry the DataInterface role.
port1_summary() {
  cani_crud show interface -o json 2>/dev/null |
    jq -r '[.[] | select(.name == "Port 1")] | "\(length) \([.[].moduleId] | unique | length) \([.[] | select(.role == "DataInterface")] | length)"'
}

# module_count prints the number of modules in the datastore.
module_count() {
  jq '.modules | length' "$CANI_DS"
}

# module_names prints the sorted module names in the datastore on one line.
module_names() {
  jq -r '[.modules[].name] | sort | join(" ")' "$CANI_DS"
}

# module_bay prints the stored module bay name of the named module.
module_bay() {
  jq -r --arg name "$1" '.modules[] | select(.name == $name) | .moduleBayName' "$CANI_DS"
}

Describe 'module placement'
  After 'teardown_test_env'

  Describe 'add module --spare'
    Before 'setup_crud_env'

    It 'rejects --spare without --location'
      When call cani_crud add module hpe-2p-25gbe-pcie-nic --spare
      The status should equal 1
      The stderr should include '--spare requires --location'
    End

    It 'stores the module at the location with no parent device'
      When call add_spare
      The status should equal 0
      The stderr should include '1 module(s) added'
    End

    It 'records a nil parent and the location on the spare'
      add_spare >/dev/null 2>&1
      When call spare_placement
      The output should equal '00000000-0000-0000-0000-000000000000 true'
    End
  End

  Describe 'placement diagnostics with --debug'
    It 'warns about a module that is neither installed nor stored'
      setup_crud_env
      When call add_spare
      The status should equal 0
      The stderr should include 'module "test-module"'
      The stderr should include 'neither installed in a device nor stored at a location'
    End

    It 'warns when an installed module has no named bay and falls back to its name'
      legacy_bay_env
      When call add_spare
      The status should equal 0
      The stderr should include 'installed in "test-device" without a named module bay'
      The stderr should include 'the module name "test-module" is used as the bay'
    End
  End

  Describe 'module bay placement rules'
    Before 'setup_crud_env'

    It 'refuses a bay the parent does not declare and leaves the datastore unchanged'
      before=$(module_count)
      When call cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay PCIe9 --name nic-x
      The status should equal 1
      The stderr should include 'has no module bay "PCIe9"'
      The stderr should include 'PCIe1, PCIe2'
      The result of function module_count should equal "$before"
    End

    It 'refuses a second module authored into an occupied bay'
      cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay PCIe1 --name nic-a >/dev/null 2>&1
      When call cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay PCIe1 --name nic-b
      The status should equal 1
      The stderr should include 'module bay "PCIe1" on "test-device" is already occupied'
      The result of function module_names should equal 'nic-a test-module'
    End

    It 'stores a bay typed in another case under the declared name'
      cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay pcie2 --name nic-lower >/dev/null 2>&1
      When call module_bay nic-lower
      The output should equal 'PCIe2'
    End

    It 'refuses to remove a location that stores a spare module'
      cani_crud add module hpe-2p-25gbe-pcie-nic --spare --location removable-location --name stored-nic >/dev/null 2>&1
      When call cani_crud remove location removable-location --force
      The status should equal 1
      The stderr should include '1 spare module(s)'
    End
  End

  Describe 'module port ownership'
    Before 'setup_crud_env'

    It 'keeps same-named ports on sibling modules distinct'
      install_two_nics >/dev/null 2>&1
      When call port1_summary
      The output should equal '2 2 1'
    End

    It 'targets a port through its module'
      When call install_two_nics
      The status should equal 0
      The stderr should include 'Updated interface Port 1'
    End
  End
End
