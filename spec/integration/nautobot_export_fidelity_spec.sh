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

# Integration test: fidelity of `cani export nautobot` for interfaces and
# module ports.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1
#
# Test flow:
#   1. Purge DCIM objects from Nautobot and build an inventory with the CLI:
#      dc > level > section, one rack, a DL380 (device port "HSN 0",
#      400gbase-x-qsfpdd) with a CX7 module (module port "HSN 0") and a
#      2-port NIC module (ports with MAC/label/description), and an XD670
#      with one H100 GPU module (nvlink / pcie-gen5-x16 ports).
#   2. Export to Nautobot and keep the log.
#   3. Read interfaces back from Nautobot: device and module port types and
#      attributes (MAC, label, description) must match the datastore. A module
#      port named like a device port is not carried yet and is recorded as a
#      Pending example, so implementing it turns it into a failure that asks
#      for its promotion.
#   4. Module ports whose type Nautobot lacks (nvlink, pcie-gen5-x16) are
#      skipped with a warning, counted in the summary, and never created.
#   5. Every interface spec is either present in Nautobot or named in a skip
#      line, and the summary's created plus skipped counts add up to the
#      number of specs.

Describe 'INTEGRATION: Nautobot export fidelity'

  nautobot_unreachable() {
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
  }

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  EXPORT_LOG="${CANI_DIR}/export.log"

  # --- helpers available inside this Describe scope ---

  cani_cfg() { cani alpha --config "$CANI_CONF" "$@"; }

  # Build the fixture inventory. One device port and one module port share
  # the name "HSN 0" on repro-dl380 by library definition (DL380 Gen11 and
  # the CX7 QSFP-DD adapter), which is the duplicate-name case.
  build_inventory() {
    cani_cfg add location dc --name repro-dc &&
    cani_cfg add location level --name repro-level --parent repro-dc &&
    cani_cfg add location section --name repro-section --parent repro-level &&
    cani_cfg add rack hpe-42u-800mmx1200mm-g2-enterprise-shock-rack \
      --location repro-section --name repro-rack &&
    cani_cfg add device hpe-dl380-gen-11 --rack repro-rack --position 10 --name repro-dl380 &&
    cani_cfg add module nvidia-connectx-7-ndr-infiniband-qsfpdd-pcie5 \
      --device repro-dl380 --bay PCIe1 --name cx7-a &&
    cani_cfg update interface --module cx7-a --name 'HSN 0' --mac aa:bb:cc:dd:ee:02 &&
    cani_cfg add module hpe-2p-25gbe-pcie-nic --device repro-dl380 --bay PCIe2 --name nic-a &&
    cani_cfg update interface --module nic-a --name 'Port 1' \
      --mac aa:bb:cc:dd:ee:01 --label 'Fabric A' --description 'module port' &&
    cani_cfg add device hpe-xd670 --rack repro-rack --position 20 --name repro-xd670 &&
    cani_cfg add module nvidia-h100-sxm-gpu --device repro-xd670 --bay 'GPU 0' --name gpu-0
  }

  build_fixture() {
    python3 "$FIXTURES/nautobot/purge_nautobot_dcim.py" && build_inventory
  }

  # Export and keep the combined log for later assertions.
  run_export() {
    cani_cfg export nautobot >"$EXPORT_LOG" 2>&1
    rc=$?
    cat "$EXPORT_LOG"
    return $rc
  }

  # Compare the sync summary against the datastore: created plus every
  # skipped-interfaces count must equal the number of interface specs.
  summary_accounting() {
    DS="$CANI_DS" LOG="$EXPORT_LOG" python3 - <<'PY'
import json
import os
import re

with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
with open(os.environ["LOG"], encoding="utf-8") as log_file:
    log = log_file.read()

specs = sum(len(d.get("interfaces", [])) for d in inventory["devices"].values())
specs += sum(len(m.get("interfaces", [])) for m in inventory.get("modules", {}).values())
created = sum(int(n) for n in re.findall(r"Created interfaces: (\d+)", log))
skipped = sum(int(n) for n in re.findall(r"Skipped interfaces[^:\n]*: (\d+)", log))
print(f"specs={specs} created={created} skipped={skipped} balanced={specs == created + skipped}")
PY
  }

  # Reconcile canidb interface specs against Nautobot for one parent device.
  # Every spec must be present remotely (counted by name) or be named in a
  # skip line of the export log. Prints "unaccounted=<n>" plus details.
  interface_accounting() {
    DEVICE="$1" DS="$CANI_DS" LOG="$EXPORT_LOG" python3 - <<'PY'
import collections
import json
import os
import re
import urllib.parse
import urllib.request

device_name = os.environ["DEVICE"]
with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
with open(os.environ["LOG"], encoding="utf-8") as log_file:
    log_lines = log_file.read().splitlines()

device = next(d for d in inventory["devices"].values() if d["name"] == device_name)
expected = collections.Counter(i["name"] for i in device.get("interfaces", []))
for module in inventory.get("modules", {}).values():
    if module.get("parentDevice") == device["id"]:
        expected.update(i["name"] for i in module.get("interfaces", []))

base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"]}
query = urllib.parse.urlencode({"device": device_name, "limit": 500})
request = urllib.request.Request(base_url + "dcim/interfaces/?" + query, headers=headers)
with urllib.request.urlopen(request) as response:
    remote = collections.Counter(i["name"] for i in json.load(response)["results"])

skip_pattern = re.compile(r"skip", re.IGNORECASE)
unaccounted = 0
for name, count in sorted(expected.items()):
    deficit = count - remote.get(name, 0)
    if deficit <= 0:
        continue
    logged = sum(1 for line in log_lines if skip_pattern.search(line) and name in line)
    missing = deficit - logged
    if missing > 0:
        unaccounted += missing
        print(f"missing={name!r} expected={count} remote={remote.get(name, 0)} logged_skips={logged}")
print(f"unaccounted={unaccounted}")
PY
  }

  # Print one Nautobot interface field for <device> <interface> <field>.
  nb_interface_field() {
    DEVICE="$1" IFACE="$2" FIELD="$3" python3 - <<'PY'
import json
import os
import urllib.parse
import urllib.request

base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"]}
query = urllib.parse.urlencode({"device": os.environ["DEVICE"], "name": os.environ["IFACE"], "depth": 1})
request = urllib.request.Request(base_url + "dcim/interfaces/?" + query, headers=headers)
with urllib.request.urlopen(request) as response:
    results = json.load(response)["results"]
if not results:
    print("<missing>")
    raise SystemExit(0)
value = results[0].get(os.environ["FIELD"])
if isinstance(value, dict):
    value = value.get("value", value.get("name", value))
if isinstance(value, list):
    value = ",".join(sorted(item.get("name", str(item)) if isinstance(item, dict) else str(item) for item in value))
print("" if value is None else value)
PY
  }

  Describe 'fixture'
    It 'purges Nautobot and builds the fixture inventory'
      BeforeCall setup_nautobot_env
      When call build_fixture
      The status should equal 0
      The output should include 'Purge complete'
      The stderr should include 'Added module'
    End

    It 'exports to Nautobot'
      When call run_export
      The status should equal 0
      The output should include 'Export completed successfully'
    End
  End

  Describe 'device interfaces'
    # Nautobot accepts 400gbase-x-qsfpdd; the exporter must not rewrite it to
    # 400gbase-x-osfp.
    It 'keeps the 400gbase-x-qsfpdd interface type'
      When call nb_interface_field repro-dl380 'HSN 0' type
      The output should equal '400gbase-x-qsfpdd'
    End

    It 'carries the device port label'
      When call nb_interface_field repro-dl380 'HSN 0' label
      The output should equal 'NDR InfiniBand CX7'
    End
  End

  # Module ports are created on the parent device because Nautobot does not
  # derive them from a module type created through the API.
  Describe 'module interfaces'
    It 'creates the NIC port on the parent device with its library type'
      When call nb_interface_field repro-dl380 'Port 1' type
      The output should equal '25gbase-x-sfp28'
    End

    It 'carries the module port MAC address'
      When call nb_interface_field repro-dl380 'Port 1' mac_address
      The output should equal 'AA:BB:CC:DD:EE:01'
    End

    It 'carries the module port label'
      When call nb_interface_field repro-dl380 'Port 1' label
      The output should equal 'Fabric A'
    End

    It 'carries the module port description'
      When call nb_interface_field repro-dl380 'Port 1' description
      The output should equal 'module port'
    End

    # The CX7 port "HSN 0" collides with the DL380 device port of the same
    # name and is dropped as a duplicate; only a module-owned interface
    # (Nautobot >= 2.3) can hold both.
    It 'keeps a module port named like a device port'
      Pending 'a module port whose name exists on the parent device is skipped'
      When call nb_interface_field repro-dl380 'HSN 0' mac_address
      The output should equal 'AA:BB:CC:DD:EE:02'
    End
  End

  # The H100 GPU module defines nvlink and pcie-gen5-x16 ports, which have no
  # Nautobot interface type; the exporter reports and skips them.
  Describe 'unsupported interface types'
    It 'warns about the port types Nautobot cannot store'
      When call cat "$EXPORT_LOG"
      The output should include 'Nautobot has no interface type "nvlink"'
      The output should include 'Nautobot has no interface type "pcie-gen5-x16"'
      The output should include 'Skipped interfaces (type not in Nautobot): 2'
    End

    It 'does not create the NVLink port'
      When call nb_interface_field repro-xd670 'NVLink 4.0' type
      The output should equal '<missing>'
    End
  End

  # Nothing may vanish silently: a spec that is not in Nautobot must be named
  # in a skip line, the summary must add up, and the run must say so.
  Describe 'accounting'
    It 'accounts for module ports whose type Nautobot lacks'
      When call interface_accounting repro-xd670
      The output should include 'unaccounted=0'
    End

    It 'accounts for a module port that shares its name with the parent device'
      When call interface_accounting repro-dl380
      The output should include 'unaccounted=0'
    End

    It 'names the skipped duplicate port in the summary'
      When call cat "$EXPORT_LOG"
      The output should include 'Skipped interface HSN 0 on module cx7-a'
      The output should include 'Skipped interfaces (name already on device): 1'
    End

    It 'balances created plus skipped interfaces against the datastore'
      When call summary_accounting
      The output should include 'balanced=True'
    End

    It 'warns that some objects were not exported as authored'
      When call grep -c 'not exported as authored' "$EXPORT_LOG"
      The output should equal '1'
    End
  End

End
