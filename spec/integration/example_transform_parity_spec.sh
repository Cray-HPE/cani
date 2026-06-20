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

# Integration coverage for pkg/provider/example/transform contracts surfaced
# through the example import CLI.

#shellcheck disable=SC2317
example_matrix_transform_summary() {
  setup_test_env
  bin/cani alpha import example --csv "$FIXTURES/matrix/example.csv" --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os
from collections import Counter

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as f:
    inv = json.load(f)

locations = {loc.get("name"): loc for loc in inv.get("locations", {}).values()}
racks = {rack.get("name"): rack for rack in inv.get("racks", {}).values()}
devices = {device.get("name"): device for device in inv.get("devices", {}).values()}

def names(entries):
    return ",".join(entry.get("name", "") for entry in entries)

def device_line(name):
    device = devices.get(name, {})
    # device.Parent is the authoritative container FK; the derived rack/location
    # fields are rebuilt on load and not persisted, so resolve names via Parent.
    rack_obj = next((r for r in racks.values() if r.get("id") == device.get("parent")), {})
    rack = rack_obj.get("name", "")
    loc = next((loc_name for loc_name, loc in locations.items() if loc.get("id") == rack_obj.get("location")), "")
    return "|".join([
        name,
        device.get("slug", ""),
        device.get("role", ""),
        device.get("status", ""),
        device.get("serial", ""),
        str(device.get("rackPosition", "")),
        device.get("face", ""),
        rack,
        loc,
        "ifaces:" + str(len(device.get("interfaces", []))),
    ])

module_counts = Counter()
for module in inv.get("modules", {}).values():
    parent = next((name for name, device in devices.items() if device.get("id") == module.get("parentDevice")), "")
    if parent:
        module_counts[parent] += 1

cable_counts = Counter()
for cable in inv.get("cables", {}).values():
    key = "|".join([
        cable.get("slug", ""),
        cable.get("color", ""),
        str(cable.get("length", "")),
        cable.get("lengthUnit", ""),
    ])
    cable_counts[key] += 1

site = locations.get("matrix-site", {})
rack = racks.get("matrix-rack", {})
# rack.Devices is a derived reverse index rebuilt on load; count rack members
# from the authoritative device.Parent links instead.
rack_device_count = sum(1 for device in devices.values() if device.get("parent") == rack.get("id"))
print("counts=" + ",".join(f"{key}:{len(inv.get(key, {}))}" for key in ["locations", "racks", "devices", "modules", "cables"]))
print("roles=" + names(inv.get("metadata", {}).get("roles", [])))
print("location=" + "|".join([site.get("name", ""), site.get("locationType", ""), site.get("status", ""), ",".join(site.get("contentTypes", []))]))
print("rack=" + "|".join([rack.get("name", ""), rack.get("slug", ""), str(rack.get("uHeight", "")), rack.get("status", ""), site.get("id", "") == rack.get("location", "") and site.get("name", ""), "devices:" + str(rack_device_count)]))
print("device_gpu=" + device_line("matrix-gpu-01"))
print("device_service=" + device_line("matrix-serv-01"))
print("device_mgmt=" + device_line("matrix-mgmt-sw"))
print("device_hsn=" + device_line("matrix-hsn-sw"))
print("modules=" + ",".join(f"{name}:{module_counts[name]}" for name in sorted(module_counts)))
print("cables=" + ",".join(f"{key}:{cable_counts[key]}" for key in sorted(cable_counts)))
PY
}

#shellcheck disable=SC2317
example_system_transform_summary() {
  setup_test_env
  bin/cani alpha import example --csv "$FIXTURES/example/system.csv" --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os
from collections import Counter

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as f:
    inv = json.load(f)

devices = {device.get("name"): device for device in inv.get("devices", {}).values()}

def interface_mac(device_name, interface_name):
    for iface in devices.get(device_name, {}).get("interfaces", []):
        if iface.get("name") == interface_name:
            return iface.get("macAddress", "")
    return ""

def device_field(name, field):
    return str(devices.get(name, {}).get(field, ""))

cable_counts = Counter()
for cable in inv.get("cables", {}).values():
    key = "|".join([
        cable.get("slug", ""),
        cable.get("color", ""),
        str(cable.get("length", "")),
        cable.get("lengthUnit", ""),
        cable.get("terminationAPort", ""),
        cable.get("terminationBPort", ""),
    ])
    cable_counts[key] += 1

print("counts=" + ",".join(f"{key}:{len(inv.get(key, {}))}" for key in ["racks", "devices", "modules", "cables", "interfaces"]))
print("roles=" + ",".join(entry.get("name", "") for entry in inv.get("metadata", {}).get("roles", [])))
print("device_gh=" + "|".join([device_field("GH-x3701u34", "slug"), device_field("GH-x3701u34", "role"), device_field("GH-x3701u34", "rackPosition"), device_field("GH-x3701u34", "face"), interface_mac("GH-x3701u34", "iLO")]))
print("device_man=" + "|".join([device_field("MAN-x3701u48", "slug"), device_field("MAN-x3701u48", "role"), device_field("MAN-x3701u48", "rackPosition"), device_field("MAN-x3701u48", "face")]))
print("modules_for_gh=" + str(sum(1 for module in inv.get("modules", {}).values() if module.get("parentDevice") == devices.get("GH-x3701u34", {}).get("id"))))
print("cables=" + ",".join(f"{key}:{cable_counts[key]}" for key in sorted(cable_counts)))
PY
}

Describe 'INTEGRATION: example transform parity'
    AfterAll 'teardown_test_env'

  It 'transforms matrix system CSV into locations roles hardware and cables'
    When call example_matrix_transform_summary
    The status should equal 0
    The output should include 'counts=locations:1,racks:1,devices:6,modules:18,cables:6'
    The output should include 'roles=ComputeNode,ServiceNode,ManagementSwitch,HSNSwitch'
    The output should include 'location=matrix-site|site|Active|rack,device,module'
    The output should include 'rack=matrix-rack|hpe-48u-800mmx1200mm-g2-enterprise-shock-rack|48|Active|matrix-site|devices:6'
    The output should include 'device_gpu=matrix-gpu-01|hpe-xd670|ComputeNode|Active|SN-GPU-01|34|front|matrix-rack|matrix-site|ifaces:3'
    The output should include 'device_service=matrix-serv-01|hpe-proliant-dl380-gen11-8sff|ServiceNode|Active|SN-SERV-01|11|front|matrix-rack|matrix-site|ifaces:7'
    The output should include 'device_mgmt=matrix-mgmt-sw|hpe-aruba-2930f-48g-4sfp|ManagementSwitch|Active|SN-MGMT-SW|48|rear|matrix-rack|matrix-site|ifaces:52'
    The output should include 'device_hsn=matrix-hsn-sw|nvidia-infiniband-ndr-64-port-osfp-switch|HSNSwitch|Active|SN-HSN-SW|42|rear|matrix-rack|matrix-site|ifaces:65'
    The output should include 'modules=matrix-gpu-01:8,matrix-gpu-02:8,matrix-serv-01:1,matrix-serv-02:1'
    The output should include 'hpe-3m-cat6-stp|blue|3|m:4'
    The output should include 'hpe-ib-ndr-osfp-dac-cable|black|2|m:2'
  End

  It 'transforms example system CSV interface metadata and expanded connections'
    When call example_system_transform_summary
    The status should equal 0
    The output should include 'counts=racks:2,devices:6,modules:3,cables:6,interfaces:184'
    The output should include 'roles=ComputeNode,ManagementSwitch,HSNSwitch'
    The output should include 'device_gh=hpe-xd670|ComputeNode|34|front|aa:bb:cc:dd:ee:01'
    The output should include 'device_man=hpe-aruba-2930f-48g-4sfp|ManagementSwitch|48|rear'
    The output should include 'modules_for_gh=2'
    The output should include 'hpe-3m-cat6-stp|blue|3|m|iLO|1:1'
    The output should include 'hpe-3m-cat6-stp|blue|3|m|iLO|2:1'
    The output should include 'hpe-ib-ndr-osfp-dac-cable|green|2|m|HSN 0|1:1'
    The output should include 'hpe-ib-ndr-osfp-dac-cable|green|2|m|HSN 3|4:1'
  End

End
