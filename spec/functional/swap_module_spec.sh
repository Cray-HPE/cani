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

# ── swap module: replace an installed module with a spare ──────────
#
# Uses the CRUD inventory: device "test-device" with module bays PCIe1 and
# PCIe2, locations "test-site" and "removable-location". The embedded
# hpe-2p-25gbe-pcie-nic is installed in PCIe1 as nic-a and stored at test-site
# as nic-spare; the swap installs nic-spare in PCIe1 and stores nic-a.

cani_crud() {
  bin/cani alpha "$@" --config "$CANI_CONF"
}

# swap_env builds the CRUD environment with nic-a installed in PCIe1, the
# spare nic-spare at test-site, and a GPU spare of another hardware type.
swap_env() {
  setup_crud_env
  cani_crud add module hpe-2p-25gbe-pcie-nic --device test-device --bay PCIe1 --name nic-a >/dev/null 2>&1 &&
  cani_crud add module hpe-2p-25gbe-pcie-nic --spare --location test-site --name nic-spare --serial NICS0001 --asset-tag AT-SPARE >/dev/null 2>&1 &&
  cani_crud add module nvidia-h100-sxm-gpu --spare --location test-site --name gpu-spare >/dev/null 2>&1 &&
  cani_crud add cable hpe-cat5-rj45-4-3m-cable --label module-cable \
    --a-device nic-a --a-port 'Port 1' \
    --b-device test-device --b-port Management >/dev/null 2>&1
}

# bay_cabled_env adds a second cable whose A end names the parent device and
# the port the bay gives nic-a, the form `add cable --a-device <device>` and
# connection imports produce, next to the module-owned cable swap_env adds.
bay_cabled_env() {
  swap_env &&
  cani_crud add cable hpe-cat5-rj45-4-3m-cable --label bay-cable \
    --a-device test-device --a-port 'Port 2' \
    --b-device test-device-2 --b-port Management >/dev/null 2>&1
}

# module_placement prints "<parent device name or -> <bay or -> <location name or ->"
# for the named module, read from the datastore.
module_placement() {
  jq -r --arg name "$1" '
    . as $inv
    | .modules[] | select(.name == $name)
    | [
        ($inv.devices[.parentDevice // ""].name // "-"),
        (if .moduleBayName == "" or .moduleBayName == null then "-" else .moduleBayName end),
        ($inv.locations[.location // ""].name // "-")
      ] | join(" ")' "$CANI_DS"
}

# cable_a_end prints "<owner name> <port> <module owning the interface>" for
# the labelled cable's A end, read from the datastore: the device or module
# the end names, its port name, and the module whose interface terminationA
# resolves to ("-" when it is not a module port).
cable_a_end() {
  jq -r --arg label "$1" '
    . as $inv
    | .cables[] | select(.label == $label)
    | [
        ($inv.devices[.terminationADevice // ""].name // $inv.modules[.terminationADevice // ""].name // "-"),
        (.terminationAPort // "-"),
        (.terminationA as $iface
          | [$inv.modules[] | select(any(.interfaces[]?; .id == $iface)) | .name] | first // "-")
      ] | join(" ")' "$CANI_DS"
}

# ShellSpec's result-of-function subject takes no arguments.
placement_nic_a() { module_placement nic-a; }
placement_nic_spare() { module_placement nic-spare; }
placement_gpu_spare() { module_placement gpu-spare; }
module_cable_a_end() { cable_a_end module-cable; }
bay_cable_a_end() { cable_a_end bay-cable; }

Describe 'cani alpha swap module'
  Before 'swap_env'
  After 'teardown_test_env'

  Describe '--dry-run'
    It 'prints the resulting placements without saving'
      When call cani_crud swap module nic-a nic-spare --dry-run
      The status should equal 0
      The output should include 'Would install nic-spare in test-device bay PCIe1'
      The output should include 'Would store nic-a at location test-site'
      The result of function placement_nic_a should equal 'test-device PCIe1 -'
      The result of function placement_nic_spare should equal '- - test-site'
    End
  End

  Describe 'swap'
    It 'installs the spare in the bay and stores the replaced module at the parent location'
      When call cani_crud swap module nic-a nic-spare
      The status should equal 0
      The stderr should include 'Swapped modules: nic-spare'
      The stderr should include '1 cable(s) moved to the ports of nic-spare'
      The result of function placement_nic_spare should equal 'test-device PCIe1 -'
      The result of function placement_nic_a should equal '- - test-site'
    End

    It 'stores the replaced module at --location'
      When call cani_crud swap module nic-a nic-spare --location removable-location
      The status should equal 0
      The result of function placement_nic_a should equal '- - removable-location'
    End

    It 'names the spare by its serial number'
      When call cani_crud swap module nic-a nics0001
      The status should equal 0
      The result of function placement_nic_spare should equal 'test-device PCIe1 -'
      The result of function placement_nic_a should equal '- - test-site'
    End

    It 'names the spare by its asset tag in a dry run'
      When call cani_crud swap module nic-a AT-SPARE --dry-run
      The status should equal 0
      The output should include 'Would install nic-spare in test-device bay PCIe1'
      The result of function placement_nic_spare should equal '- - test-site'
    End

    It 'accepts the stored module as the spare of a later swap'
      cani_crud swap module nic-a nic-spare >/dev/null 2>&1
      When call cani_crud swap module nic-spare nic-a
      The status should equal 0
      The result of function placement_nic_a should equal 'test-device PCIe1 -'
      The result of function placement_nic_spare should equal '- - test-site'
    End
  End

  Describe 'refusals leave the datastore unchanged'
    It 'refuses a spare on the installed side'
      When call cani_crud swap module nic-spare nic-a
      The status should equal 1
      The stderr should include 'is not installed in a device'
      The result of function placement_nic_a should equal 'test-device PCIe1 -'
    End

    It 'refuses a spare of another hardware type'
      When call cani_crud swap module nic-a gpu-spare
      The status should equal 1
      The stderr should include 'swap like for like'
      The result of function placement_gpu_spare should equal '- - test-site'
    End

    It 'refuses an unknown location'
      When call cani_crud swap module nic-a nic-spare --location nowhere
      The status should equal 1
      The stderr should include 'location "nowhere" not found'
      The result of function placement_nic_a should equal 'test-device PCIe1 -'
    End

    It 'refuses an unknown module'
      When call cani_crud swap module nic-a ghost
      The status should equal 1
      The stderr should include 'no item found matching'
    End
  End

  Describe 'a cable addressed through the parent device'
    Before 'bay_cabled_env'
    After 'teardown_test_env'

    It 'keeps every cable on the bay ports, now on the spare'
      When call cani_crud swap module nic-a nic-spare
      The status should equal 0
      The stderr should include '2 cable(s) moved to the ports of nic-spare'
      The result of function placement_nic_spare should equal 'test-device PCIe1 -'
      The result of function placement_nic_a should equal '- - test-site'
      The result of function bay_cable_a_end should equal 'test-device Port 2 nic-spare'
      The result of function module_cable_a_end should equal 'nic-spare Port 1 nic-spare'
    End

    It 'reports the moved cables in --dry-run without saving'
      When call cani_crud swap module nic-a nic-spare --dry-run
      The status should equal 0
      The output should include 'Would move 2 cable(s) to the ports of nic-spare'
      The result of function bay_cable_a_end should equal 'test-device Port 2 nic-a'
      The result of function module_cable_a_end should equal 'nic-a Port 1 nic-a'
    End
  End
End
