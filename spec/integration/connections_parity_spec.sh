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

# Integration coverage for pkg/devicetypes/connections unit-test contracts.

#shellcheck disable=SC2317
write_connection_parity_csv() {
  _csv_path="$1"
  cat >"$_csv_path" <<'CSV'
a_device,a_port,a_mac,b_device,b_port,b_mac,type,label,color,length,length_unit,status
_defaults,,,,,,cat5,,blue,,m,Connected
test-device,Management,AA-BB-CC-DD-EE-01,test-device-2,Management,AA-BB-CC-DD-EE-02,,mgmt-link,,,,
test-device,GigabitEthernet0/0/1,,test-device-2,GigabitEthernet0/0/1,,cat6a,data-link,green,3,ft,Planned
CSV
}

#shellcheck disable=SC2317
print_connection_summary() {
  python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)

devices = {device.get("name"): device for device in inv.get("devices", {}).values()}
cables = {cable.get("label"): cable for cable in inv.get("cables", {}).values() if cable.get("label")}

def cable_line(label):
    cable = cables.get(label, {})
    length = cable.get("length", "")
    print(
        f"{label}="
        f"slug:{cable.get('slug', '')},"
        f"color:{cable.get('color', '')},"
        f"length:{length},"
        f"unit:{cable.get('lengthUnit', '')},"
        f"status:{cable.get('status', '')},"
        f"a:{cable.get('terminationAPort', '')},"
        f"b:{cable.get('terminationBPort', '')}"
    )

def interface_mac(device_name, interface_name):
    for iface in devices.get(device_name, {}).get("interfaces", []):
        if iface.get("name") == interface_name:
            return iface.get("macAddress", "")
    return ""

print("cable_count=" + str(len(inv.get("cables", {}))))
cable_line("mgmt-link")
cable_line("data-link")
print("test-device_management_mac=" + interface_mac("test-device", "Management"))
print("test-device-2_management_mac=" + interface_mac("test-device-2", "Management"))
PY
}

#shellcheck disable=SC2317
generated_star_summary() {
  setup_connections_env
  bin/cani alpha add connections generate star \
    --hub test-device-2 \
    --hub-ports Management \
    --spokes test-device \
    --spoke-port Management \
    --cable-type cat6a \
    --cable-color orange \
    >"$CANI_DIR/generated-star.yml" || return $?
  bin/cani alpha add connections "$CANI_DIR/generated-star.yml" --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS"], encoding="utf-8") as f:
    inv = json.load(f)

devices = {device.get("name"): device_id for device_id, device in inv.get("devices", {}).items()}
matches = []
for cable in inv.get("cables", {}).values():
    if (
        cable.get("terminationADevice") == devices.get("test-device")
        and cable.get("terminationBDevice") == devices.get("test-device-2")
        and cable.get("terminationAPort") == "Management"
        and cable.get("terminationBPort") == "Management"
        and cable.get("slug") == "cat6a"
        and cable.get("color") == "orange"
    ):
        matches.append(cable)

print("generated_star_links=" + str(len(matches)))
print("generated_star_status=" + (matches[0].get("status", "") if matches else ""))
PY
}

#shellcheck disable=SC2317
csv_defaults_summary() {
  setup_connections_env
  write_connection_parity_csv "$CANI_DIR/connections-parity.csv"
  bin/cani alpha add connections "$CANI_DIR/connections-parity.csv" --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  print_connection_summary
}

#shellcheck disable=SC2317
export_round_trip_summary() {
  setup_connections_env
  write_connection_parity_csv "$CANI_DIR/connections-parity.csv"
  bin/cani alpha add connections "$CANI_DIR/connections-parity.csv" --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  _exported_csv="/tmp/cani-connections-exported-$$.csv"
  bin/cani alpha export connections --format csv --config "$CANI_CONF" >"$_exported_csv" || return $?
  setup_connections_env
  bin/cani alpha add connections "$_exported_csv" --config "$CANI_CONF" >/dev/null 2>&1 || return $?
  rm -f "$_exported_csv"
  print_connection_summary
}

Describe 'INTEGRATION: connections ShellSpec parity'

  It 'applies generated star topology with cable defaults'
    When call generated_star_summary
    The status should equal 0
    The output should include 'generated_star_links=1'
    The output should include 'generated_star_status=Connected'
  End

  It 'persists CSV defaults overrides and endpoint MACs'
    When call csv_defaults_summary
    The status should equal 0
    The output should include 'cable_count=3'
    The output should include 'mgmt-link=slug:cat5,color:blue,length:,unit:m,status:Connected,a:Management,b:Management'
    The output should include 'data-link=slug:cat6a,color:green,length:3,unit:ft,status:Planned,a:GigabitEthernet0/0/1,b:GigabitEthernet0/0/1'
    The output should include 'test-device_management_mac=aa:bb:cc:dd:ee:01'
    The output should include 'test-device-2_management_mac=aa:bb:cc:dd:ee:02'
  End

  It 'replays exported connection CSV into a fresh inventory'
    When call export_round_trip_summary
    The status should equal 0
    The output should include 'cable_count=3'
    The output should include 'mgmt-link=slug:cat5,color:blue,length:,unit:m,status:Connected,a:Management,b:Management'
    The output should include 'data-link=slug:cat6a,color:green,length:3,unit:ft,status:Connected,a:GigabitEthernet0/0/1,b:GigabitEthernet0/0/1'
  End

End
