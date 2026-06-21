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

# Integration coverage for pkg/visual behavior surfaced through `cani alpha show`
# with a persisted inventory.

#shellcheck disable=SC2317
setup_visual_parity_env() {
  setup_test_env
  python3 - <<'PY'
import json
import os

inventory = {
    "schemaVersion": "v1alpha2",
    "locations": {},
    "racks": {
        "00000000-0000-0000-0001-000000000001": {
            "id": "00000000-0000-0000-0001-000000000001",
            "name": "Rack-006U",
            "uHeight": 6,
            "devices": [
                "00000000-0000-0000-0002-000000000001",
                "00000000-0000-0000-0002-000000000002",
                "00000000-0000-0000-0003-000000000001",
            ],
            "status": "Active",
        }
    },
    "devices": {
        "00000000-0000-0000-0001-000000000001": {
            "id": "00000000-0000-0000-0001-000000000001",
            "name": "Rack-006U",
            "type": "rack",
            "children": [
                "00000000-0000-0000-0002-000000000001",
                "00000000-0000-0000-0002-000000000002",
                "00000000-0000-0000-0003-000000000001",
            ],
            "status": "Active",
        },
        "00000000-0000-0000-0002-000000000001": {
            "id": "00000000-0000-0000-0002-000000000001",
            "name": "Node-A",
            "type": "node",
            "model": "HPE DL360",
            "uHeight": 2,
            "status": "Active",
            "role": "compute",
            "parent": "00000000-0000-0000-0001-000000000001",
            "rack": "00000000-0000-0000-0001-000000000001",
            "rackPosition": 1,
            "interfaces": [
                {"id": "00000000-0000-0000-0004-000000000001", "name": "eth0", "type": "1000base-t"}
            ],
        },
        "00000000-0000-0000-0002-000000000002": {
            "id": "00000000-0000-0000-0002-000000000002",
            "name": "Node-B",
            "type": "node",
            "model": "HPE DL360",
            "uHeight": 1,
            "status": "Active",
            "role": "compute",
            "parent": "00000000-0000-0000-0001-000000000001",
            "rack": "00000000-0000-0000-0001-000000000001",
            "rackPosition": 4,
        },
        "00000000-0000-0000-0003-000000000001": {
            "id": "00000000-0000-0000-0003-000000000001",
            "name": "Leaf-1",
            "type": "switch",
            "model": "Aruba 8325",
            "uHeight": 1,
            "status": "Active",
            "role": "leaf",
            "parent": "00000000-0000-0000-0001-000000000001",
            "rack": "00000000-0000-0000-0001-000000000001",
            "rackPosition": 6,
            "interfaces": [
                {"id": "00000000-0000-0000-0005-000000000001", "name": "1/1/1", "type": "1000base-t"}
            ],
        },
    },
    "modules": {},
    "cables": {
        "00000000-0000-0000-0006-000000000001": {
            "id": "00000000-0000-0000-0006-000000000001",
            "slug": "cat6",
            "label": "node-a-to-leaf",
            "type": "cable",
            "cableType": "cat6",
            "status": "Connected",
            "terminationA": "00000000-0000-0000-0004-000000000001",
            "terminationB": "00000000-0000-0000-0005-000000000001",
            "terminationADevice": "00000000-0000-0000-0002-000000000001",
            "terminationBDevice": "00000000-0000-0000-0003-000000000001",
            "terminationAPort": "eth0",
            "terminationBPort": "1/1/1",
        }
    },
    "frus": {},
    "interfaces": {
        "00000000-0000-0000-0004-000000000001": {
            "id": "00000000-0000-0000-0004-000000000001",
            "name": "eth0",
            "interfaceType": "1000base-t",
            "deviceId": "00000000-0000-0000-0002-000000000001",
            "status": "Active",
        },
        "00000000-0000-0000-0005-000000000001": {
            "id": "00000000-0000-0000-0005-000000000001",
            "name": "1/1/1",
            "interfaceType": "1000base-t",
            "deviceId": "00000000-0000-0000-0003-000000000001",
            "status": "Active",
        },
    },
    "metadata": {},
}

with open(os.environ["CANI_DS"], "w", encoding="utf-8") as datastore:
    json.dump(inventory, datastore, indent=2)
PY
}

#shellcheck disable=SC2317
classic_rack_width_summary() {
  _output="$(bin/cani alpha show rack Rack-006U --format classic --no-color --config "$CANI_CONF")" || return $?
  VISUAL_OUTPUT="$_output" python3 - <<'PY'
import os
import sys

text = os.environ["VISUAL_OUTPUT"]
prefixes = ("┌", "│", "├", "└")
box_lines = [line for line in text.splitlines() if line.startswith(prefixes)]
widths = sorted({len(line) for line in box_lines})
print("widths=" + ",".join(str(width) for width in widths))
print("has_node_a=" + str("█ Node-A (2U)" in text))
print("has_continued=" + str("▓ (continued)" in text))
print("has_empty=" + str("░ [EMPTY]" in text))
print("has_cable=" + str("node-a-to-leaf [Node-A:eth0] ←→ [Leaf-1:1/1/1]" in text))
if widths != [80]:
    sys.exit(1)
PY
}

Describe 'INTEGRATION: visual output parity'
  Before 'setup_visual_parity_env'
  AfterAll 'teardown_test_env'

  It 'renders device tables with resolved rack names and U positions'
    When call bin/cani alpha show device --format table --config "$CANI_CONF"
    The status should equal 0
    The stdout should include 'NAME'
    The stdout should include 'RACK'
    The stdout should include 'Leaf-1'
    The stdout should include 'Rack-006U'
    The stdout should include 'Node-A'
    The stdout should include '1'
    The stdout should include 'Total: 4 device(s)'
  End

  It 'renders cable tables with resolved endpoint device names and ports'
    When call bin/cani alpha show cable --format table --config "$CANI_CONF"
    The status should equal 0
    The stdout should include 'A TERMINATION'
    The stdout should include 'B TERMINATION'
    The stdout should include 'node-a-to-leaf'
    The stdout should include 'Node-A:eth0'
    The stdout should include 'Leaf-1:1/1/1'
    The stdout should include 'Total: 1 cable(s)'
  End

  It 'renders classic rack rows with stable Unicode width and cable details'
    When call classic_rack_width_summary
    The status should equal 0
    The output should include 'widths=80'
    The output should include 'has_node_a=True'
    The output should include 'has_continued=True'
    The output should include 'has_empty=True'
    The output should include 'has_cable=True'
  End

  It 'renders tree connectors with roles interfaces empty U slots and cable leaves'
    When call bin/cani alpha show rack Rack-006U --format tree --no-color --with empty-us,cables,interfaces,roles --config "$CANI_CONF"
    The status should equal 0
    The stdout should include '■ (rack) Rack-006U'
    The stdout should include '├── ● (device) U6 Leaf-1 | Aruba 8325 | role:leaf'
    The stdout should include '│   └── ○ (interface) 1/1/1 1000base-t'
    The stdout should include '│       └── ═ (cable) node-a-to-leaf A:Node-A:eth0 → B:Leaf-1:1/1/1'
    The stdout should include '├── ● (device) U5 (empty)'
    The stdout should include '└── ● (device) U1-U0 Node-A | HPE DL360 | role:compute'
  End

  It 'renders routing labels for intra-rack cables'
    When call bin/cani alpha show rack Rack-006U --format routing --no-color -VV --labels --config "$CANI_CONF"
    The status should equal 0
    The stdout should include 'Symbols: D=device d=half M=modules *=cont ·=empty'
    The stdout should include '(showing all cables including intra-rack)'
    The stdout should include '→1(1/1/1:eth0)'
    The stdout should include '→6(eth0:1/1/1)'
  End

End
