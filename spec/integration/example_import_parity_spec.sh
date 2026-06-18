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

# Integration coverage for pkg/provider/example/import behavior that crosses the
# CLI and datastore boundary.

#shellcheck disable=SC2317
example_yaml_import_summary() {
  setup_test_env
  bin/cani alpha import example --file "$FIXTURES/example/inventory.yaml" --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as datastore:
    inv = json.load(datastore)

location_id = "11111111-1111-1111-1111-111111111111"
rack_id = "22222222-2222-2222-2222-222222222222"
device_id = "33333333-3333-3333-3333-333333333333"
cable_id = "44444444-4444-4444-4444-444444444444"

locations = inv.get("locations", {})
racks = inv.get("racks", {})
devices = inv.get("devices", {})
cables = inv.get("cables", {})
rack = racks.get(rack_id, {})
device = devices.get(device_id, {})

print("counts=" + ",".join(f"{key}:{len(inv.get(key, {}))}" for key in ["locations", "racks", "devices", "cables"]))
print("keys=" + ",".join(str(key in values) for key, values in [(location_id, locations), (rack_id, racks), (device_id, devices), (cable_id, cables)]))
print("rack=" + "|".join([rack.get("name", ""), str(rack.get("uHeight", "")), rack.get("status", "")]))
print("device_id_match=" + str(device.get("id") == device_id))
PY
}

#shellcheck disable=SC2317
example_all_invalid_bom_summary() {
  setup_test_env
  _path="$CANI_DIR/example-all-invalid.csv"
  _stderr="$CANI_DIR/example-all-invalid.err"
  printf 'PartNumber,Description,Quantity\nP1,Widget,0\n' >"$_path"

  if ! bin/cani alpha import example --csv "$_path" --config "$CANI_CONF" >/dev/null 2>"$_stderr"; then
    cat "$_stderr"
    return 1
  fi

  cat "$_stderr"
  CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as datastore:
    inv = json.load(datastore)

print("counts=" + ",".join(f"{key}:{len(inv.get(key, {}))}" for key in ["locations", "racks", "devices", "cables"]))
PY
}

Describe 'INTEGRATION: example import parity'
  AfterAll 'teardown_test_env'

  It 'persists YAML file imports into the datastore'
    When call example_yaml_import_summary
    The status should equal 0
    The output should include 'counts=locations:1,racks:1,devices:1,cables:1'
    The output should include 'keys=True,True,True,True'
    The output should include 'rack=Rack-01|42|active'
    The output should include 'device_id_match=True'
  End

  It 'treats a BOM CSV with only invalid rows as a successful empty import'
    When call example_all_invalid_bom_summary
    The status should equal 0
    The output should include 'WARN: line 2: Quantity must be >= 1, got 0, skipping'
    The output should include 'No valid records found in CSV'
    The output should include 'Import completed successfully using provider example'
    The output should include 'counts=locations:0,racks:0,devices:0,cables:0'
  End

End
