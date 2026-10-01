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

# ── interface identity across instances of one hardware type ───────
#
# Every instance added from a library type must own its own interface
# UUIDs. Instances created in one invocation (--qty N, %{FILL}) share the
# library template's Interfaces slice, so the first ID stamp leaks into
# every sibling and the inventory's interface index collapses them.

Describe 'interface identity per hardware instance'
  Before 'setup_crud_env'

  # Print "<owner-name> <interface-id>" for every interface spec embedded in
  # the devices (or modules) whose names start with the given prefix.
  interface_ids_for() {
    SECTION="$1" NAME_PREFIX="$2" DS="$CANI_DS" python3 - <<'PY'
import json, os

with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)

for item in inventory.get(os.environ["SECTION"], {}).values():
    if not item.get("name", "").startswith(os.environ["NAME_PREFIX"]):
        continue
    for iface in item.get("interfaces", []):
        print(item["name"], iface["name"], iface.get("id"))
PY
  }

  # Count interface IDs that appear under more than one owner.
  shared_interface_ids() {
    interface_ids_for "$1" "$2" | awk '{ owners[$NF] = owners[$NF] " " $1 } END {
      shared = 0
      for (id in owners) { n = split(owners[id], parts, " "); if (n > 1) shared++ }
      print shared
    }'
  }

  add_two_nics() {
    bin/cani alpha add module hpe-2p-25gbe-pcie-nic \
      --device test-device --qty 2 --prefix nic --config "$CANI_CONF" >/dev/null 2>&1
  }

  add_two_nodes() {
    bin/cani alpha add device cray-xd225v \
      --rack test-rack --qty 2 --prefix node --config "$CANI_CONF" >/dev/null 2>&1
  }

  Describe 'modules added with --qty'
    It 'adds two NIC modules to one device in a single invocation'
      When call bin/cani alpha add module hpe-2p-25gbe-pcie-nic \
        --device test-device --qty 2 --prefix nic --config "$CANI_CONF"
      The status should equal 0
      The stderr should include '2 module(s) added'
    End

    It 'gives each module its own interface UUIDs'
      BeforeCall add_two_nics
      When call shared_interface_ids modules nic
      The output should equal '0'
    End
  End

  Describe 'devices added with --qty'
    It 'gives each device its own interface UUIDs'
      BeforeCall add_two_nodes
      When call shared_interface_ids devices node
      The output should equal '0'
    End
  End

End
