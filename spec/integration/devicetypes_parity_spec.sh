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

# Integration coverage for pkg/devicetypes unit-test contracts that surface
# through CLI workflows.

#shellcheck disable=SC2317
write_local_type_library() {
  _types_root="$1"
  mkdir -p \
    "$_types_root/location-types" \
    "$_types_root/rack-types/LocalCo" \
    "$_types_root/device-types/LocalCo" \
    "$_types_root/module-types/LocalCo" \
    "$_types_root/cable-types/LocalCo"

  cat >"$_types_root/location-types/local-room.yaml" <<'YAML'
name: Local Room Type
slug: local-room
description: Local room loaded from an integration test
nestable: true
content_types:
  - rack
  - device
  - module
YAML

  cat >"$_types_root/rack-types/LocalCo/local-audit-rack.yaml" <<'YAML'
manufacturer: LocalCo
model: Local Audit Rack
slug: local-audit-rack
part_number: LOCAL-RACK-PN
type: rack
u_height: 18
YAML

  cat >"$_types_root/device-types/LocalCo/local-audit-node.yaml" <<'YAML'
manufacturer: LocalCo
model: Local Audit Node
slug: local-audit-node
part_number: LOCAL-NODE-PN
type: node
u_height: 1
interfaces:
  - name: mgmt0
    type: 1000base-t
YAML

  cat >"$_types_root/module-types/LocalCo/local-audit-nic.yaml" <<'YAML'
manufacturer: LocalCo
model: Local Audit NIC
slug: local-audit-nic
part_number: LOCAL-MOD-PN
type: nic
interfaces:
  - name: uplink0
    type: 1000base-t
YAML

  cat >"$_types_root/cable-types/LocalCo/local-audit-cable.yaml" <<'YAML'
manufacturer: LocalCo
model: Local Audit Cable
slug: local-audit-cable
part_number: LOCAL-CAB-PN
type: cable
cable_type: cat6
color: teal
length: 2
length_unit: m
YAML
}

#shellcheck disable=SC2317
print_local_type_inventory_summary() {
  python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)

locations = {item.get("name"): item for item in inv.get("locations", {}).values()}
racks = {item.get("name"): item for item in inv.get("racks", {}).values()}
devices = {item.get("name"): item for item in inv.get("devices", {}).values()}
modules = {item.get("name"): item for item in inv.get("modules", {}).values()}
cables = {item.get("label"): item for item in inv.get("cables", {}).values()}

loc = locations.get("LocalRoom", {})
rack = racks.get("LocalRack", {})
device = devices.get("LocalNode", {})
module = modules.get("LocalModule", {})
cable = cables.get("LocalCable", {})

loc_types = ",".join(loc.get("contentTypes", []))
device_interfaces = ",".join(iface.get("name", "") for iface in device.get("interfaces", []))

print("location=" + "|".join([
    loc.get("name", ""), loc.get("locationType", ""), str(loc.get("nestable", False)), loc_types,
]))
print("rack=" + "|".join([
    rack.get("name", ""), rack.get("slug", ""), rack.get("partNumber", ""),
    str(rack.get("uHeight", "")), loc.get("id", "") == rack.get("location", "") and loc.get("name", ""),
]))
print("device=" + "|".join([
    device.get("name", ""), device.get("slug", ""), device.get("partNumber", ""),
    device.get("model", ""), rack.get("id", "") == device.get("parent", "") and rack.get("name", ""),
  device.get("status", ""), device_interfaces,
]))
print("module=" + "|".join([
    module.get("name", ""), module.get("slug", ""), module.get("partNumber", ""),
    device.get("id", "") == module.get("parentDevice", "") and device.get("name", ""), module.get("status", ""),
]))
print("cable=" + "|".join([
    cable.get("label", ""), cable.get("slug", ""), cable.get("partNumber", ""),
    cable.get("status", ""), cable.get("cableType", ""), str(cable.get("length", "")), cable.get("lengthUnit", ""),
]))
PY
}

#shellcheck disable=SC2317
local_type_library_summary() {
  setup_test_env
  _types_root="$CANI_DIR/local-types"
  write_local_type_library "$_types_root"

  bin/cani --types-dirs "$_types_root" alpha add location local-room --name LocalRoom --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  bin/cani --types-dirs "$_types_root" alpha add LOCAL-RACK-PN --name LocalRack --location LocalRoom --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  _rack_id="$(CANI_DS="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)
for rack_id, rack in inv.get("racks", {}).items():
    if rack.get("name") == "LocalRack":
        print(rack_id)
        break
PY
)"
  bin/cani --types-dirs "$_types_root" alpha add LOCAL-NODE-PN --name LocalNode --parent "$_rack_id" --status Active --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  _device_id="$(CANI_DS="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)
for device_id, device in inv.get("devices", {}).items():
    if device.get("name") == "LocalNode":
        print(device_id)
        break
PY
)"
  bin/cani --types-dirs "$_types_root" alpha add LOCAL-MOD-PN --name LocalModule --parent "$_device_id" --status Active --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  bin/cani --types-dirs "$_types_root" alpha add LOCAL-CAB-PN --name LocalCable --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  print_local_type_inventory_summary
}

#shellcheck disable=SC2317
metadata_catalog_summary() {
  setup_crud_env
  bin/cani alpha add metadata role IntegrationRole --content-types dcim.device,dcim.rack --color aa1409 --description "Integration role" --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  bin/cani alpha add metadata status Maintenance --color 00aa00 --description "Maintenance window" --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  bin/cani alpha add metadata tag IntegrationTag --content-types dcim.device --color 00ff00 --description "Integration tag" --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)

metadata = inv.get("metadata", {})

def entry_line(kind, name):
    for entry in metadata.get(kind, []):
        if entry.get("name") == name:
            content_types = ",".join(entry.get("contentTypes", []))
            return "|".join([
                entry.get("name", ""), entry.get("color", ""),
                str(entry.get("weight", "")), content_types, entry.get("description", ""),
            ])
    return "missing"

print("role=" + entry_line("roles", "IntegrationRole"))
print("status=" + entry_line("statuses", "Maintenance"))
print("tag=" + entry_line("tags", "IntegrationTag"))
PY
}

#shellcheck disable=SC2317
orphan_plan_summary() {
  setup_orphan_env
  bin/cani alpha update orphans --apply-plan "$FIXTURES/cani/orphan_plan.json" --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  python3 - <<'PY'
import json
import os

device_id = "bbbb1111-2222-3333-4444-555566667777"
rack_id = "58f00e62-0b30-4435-9c78-98f5aa3649f1"
location_id = "f9e5d985-376f-4bb1-8249-a7e82647b7f9"

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)

device = inv.get("devices", {}).get(device_id, {})
rack = inv.get("racks", {}).get(rack_id, {})
interfaces = [
    iface for iface in inv.get("interfaces", {}).values()
    if iface.get("deviceId") == device_id
]

print("orphan_device=" + "|".join([
    device.get("parent", ""), device.get("rack", ""), device.get("location", ""),
    str(device.get("rackPosition", "")), device.get("face", ""),
]))
print("rack_contains_device=" + str(device_id in rack.get("devices", [])))
print("interface_count=" + str(len(interfaces)))
print("location_match=" + str(device.get("location") == location_id))
PY
}

Describe 'INTEGRATION: devicetypes ShellSpec parity'

  It 'loads local type directories and constructs each inventory kind'
    When call local_type_library_summary
    The status should equal 0
    The output should include 'location=LocalRoom|local-room|True|rack,device,module'
    The output should include 'rack=LocalRack|local-audit-rack|LOCAL-RACK-PN|18|LocalRoom'
    The output should include 'device=LocalNode|local-audit-node|LOCAL-NODE-PN|Local Audit Node|LocalRack|Active|mgmt0'
    The output should include 'module=LocalModule|local-audit-nic|LOCAL-MOD-PN|LocalNode|Active'
    The output should include 'cable=LocalCable|local-audit-cable|LOCAL-CAB-PN|Connected|cat6|2|m'
  End

  It 'persists metadata catalog entries with flags'
    When call metadata_catalog_summary
    The status should equal 0
    The output should include 'role=IntegrationRole|aa1409|1000|dcim.device,dcim.rack|Integration role'
    The output should include 'status=Maintenance|00aa00|||Maintenance window'
    The output should include 'tag=IntegrationTag|00ff00||dcim.device|Integration tag'
  End

  It 'applies an orphan resolve plan and rebuilds relationships'
    When call orphan_plan_summary
    The status should equal 0
    The output should include 'orphan_device=58f00e62-0b30-4435-9c78-98f5aa3649f1|58f00e62-0b30-4435-9c78-98f5aa3649f1|f9e5d985-376f-4bb1-8249-a7e82647b7f9|5|front'
    The output should include 'rack_contains_device=True'
    The output should include 'interface_count=1'
    The output should include 'location_match=True'
  End

End
