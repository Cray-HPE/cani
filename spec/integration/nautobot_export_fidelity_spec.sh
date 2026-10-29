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
#      attributes (MAC, label, description) must match the datastore. Module
#      ports are owned by their module, so a module port named like a device
#      port is kept next to it rather than skipped.
#   4. Module ports whose type Nautobot lacks (nvlink, pcie-gen5-x16) are
#      skipped with a warning, counted in the summary, and never created.
#   5. Every interface spec is either present in Nautobot or named in a skip
#      line, and the summary's created plus skipped counts add up to the
#      number of specs.
#   6. Import Nautobot into an empty datastore: every module port stays on its
#      module and none is added to the parent device, although Nautobot 3.2
#      also names the root device on the port of an installed module.
#   7. Rename a module bay in Nautobot to a name the embedded library does not
#      declare, import again: the device records the bay as declared, the
#      module keeps it, and a re-export from that datastore creates neither a
#      bay nor a module.
#   8. Store a spare NIC, export it, then `swap module` the installed NIC for
#      it: a plain export reports the two pending moves, `--merge --dry-run`
#      previews them without writing, `--merge` moves both modules under their
#      recorded IDs and creates none, a re-export is idempotent, a bay move
#      made with `update module --bay` follows the same path, and an import of
#      the result places both modules where the datastore does.

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

  # Print one Nautobot interface field for <device> <interface> <field>
  # [device|module]; the owner picks between same-named device and module ports.
  nb_interface_field() {
    DEVICE="$1" IFACE="$2" FIELD="$3" OWNER="${4:-}" python3 - <<'PY'
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
if os.environ["OWNER"]:
    results = [r for r in results if (r.get("module") is None) == (os.environ["OWNER"] == "device")]
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

  IMPORT_CONF="${CANI_DIR}/import.yml"
  IMPORT_DS="${CANI_DIR}/import.json"

  # Import Nautobot into an empty datastore with the fixture's provider
  # settings and keep the log.
  run_import() {
    sed "s|^datastore:.*|datastore: ${IMPORT_DS}|" "$CANI_CONF" >"$IMPORT_CONF" &&
      cani alpha --config "$IMPORT_CONF" import nautobot \
        --default-location repro-section --default-status Active >"${CANI_DIR}/import.log" 2>&1
    rc=$?
    cat "${CANI_DIR}/import.log"
    return $rc
  }

  # Compare repro-dl380's ports in the imported datastore with the exported
  # one: count device-owned names the import added to the device, and the
  # "HSN 0" and "Port 1" ports on the device and on the module in each bay.
  imported_port_owners() {
    DS="$CANI_DS" IMPORTED="$IMPORT_DS" python3 - <<'PY'
import collections
import json
import os


def ports(path):
    with open(path, encoding="utf-8") as datastore:
        inventory = json.load(datastore)
    device = next(d for d in inventory["devices"].values() if d["name"] == "repro-dl380")
    own = collections.Counter(i["name"] for i in device.get("interfaces", []))
    modules = collections.Counter()
    for module in inventory.get("modules", {}).values():
        if module.get("parentDevice") == device["id"]:
            bay = module.get("moduleBayName", "").lower()
            modules.update((bay, i["name"]) for i in module.get("interfaces", []))
    return own, modules


exported_own, _ = ports(os.environ["DS"])
imported_own, imported_modules = ports(os.environ["IMPORTED"])
added = sum((imported_own - exported_own).values())
print(f"added_to_device={added} device_hsn0={imported_own['HSN 0']} "
      f"pcie1_hsn0={imported_modules[('pcie1', 'HSN 0')]} pcie2_port1={imported_modules[('pcie2', 'Port 1')]}")
PY
  }

  RENAMED_CONF="${CANI_DIR}/renamed.yml"
  RENAMED_DS="${CANI_DIR}/renamed.json"

  # rename_module_bay renames one module bay of a device in Nautobot and
  # prints the name Nautobot stored, so the import sees a bay name the
  # embedded library does not declare for the device type.
  rename_module_bay() {
    DEVICE="$1" FROM="$2" TO="$3" python3 - <<'PY'
import json
import os
import urllib.parse
import urllib.request

base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"], "Content-Type": "application/json"}
query = urllib.parse.urlencode({"parent_device": os.environ["DEVICE"], "name": os.environ["FROM"]})
with urllib.request.urlopen(urllib.request.Request(base_url + "dcim/module-bays/?" + query, headers=headers)) as response:
    bay = json.load(response)["results"][0]
body = json.dumps({"name": os.environ["TO"]}).encode()
request = urllib.request.Request(base_url + f"dcim/module-bays/{bay['id']}/", data=body, headers=headers, method="PATCH")
with urllib.request.urlopen(request) as response:
    print(f"renamed={json.load(response)['name']}")
PY
  }

  # Import Nautobot into a second empty datastore after the rename.
  run_import_renamed() {
    sed "s|^datastore:.*|datastore: ${RENAMED_DS}|" "$CANI_CONF" >"$RENAMED_CONF" &&
      cani alpha --config "$RENAMED_CONF" import nautobot \
        --default-location repro-section --default-status Active >"${CANI_DIR}/renamed-import.log" 2>&1
    rc=$?
    cat "${CANI_DIR}/renamed-import.log"
    return $rc
  }

  # imported_renamed_bays prints the bays the imported repro-dl380 declares
  # and the bay of each module imported into it, both sorted.
  imported_renamed_bays() {
    IMPORTED="$RENAMED_DS" python3 - <<'PY'
import json
import os

with open(os.environ["IMPORTED"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
device = next(d for d in inventory["devices"].values() if d["name"] == "repro-dl380")
declared = sorted(bay["name"] for bay in device.get("moduleBays", []))
placed = sorted(module.get("moduleBayName") or "<none>" for module in inventory.get("modules", {}).values()
                if module.get("parentDevice") == device["id"])
print(f"declared={','.join(declared)} modules={','.join(placed)}")
PY
  }

  # Re-export the renamed import and keep its log.
  reexport_renamed() {
    cani alpha --config "$RENAMED_CONF" export nautobot >"${CANI_DIR}/renamed-export.log" 2>&1
    rc=$?
    cat "${CANI_DIR}/renamed-export.log"
    return $rc
  }

  # nautobot_module_bays prints the number of module bays a device has in
  # Nautobot, which a re-export that honours the renamed bay leaves unchanged.
  nautobot_module_bays() {
    DEVICE="$1" python3 - <<'PY'
import json
import os
import urllib.parse
import urllib.request

base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"]}
query = urllib.parse.urlencode({"parent_device": os.environ["DEVICE"], "limit": 1})
with urllib.request.urlopen(urllib.request.Request(base_url + "dcim/module-bays/?" + query, headers=headers)) as response:
    print(f"module_bays={json.load(response)['count']}")
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
      When call nb_interface_field repro-dl380 'HSN 0' type device
      The output should equal '400gbase-x-qsfpdd'
    End

    It 'carries the device port label'
      When call nb_interface_field repro-dl380 'HSN 0' label device
      The output should equal 'NDR InfiniBand CX7'
    End
  End

  # Module ports are created on their module because Nautobot does not derive
  # them from a module type created through the API.
  Describe 'module interfaces'
    It 'creates the NIC port on its module with its library type'
      When call nb_interface_field repro-dl380 'Port 1' type module
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

    # The CX7 port "HSN 0" shares its name with a DL380 device port; module
    # ownership keeps both.
    It 'keeps a module port named like a device port'
      When call nb_interface_field repro-dl380 'HSN 0' mac_address module
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

    It 'does not skip a module port named like a device port'
      When call cat "$EXPORT_LOG"
      The output should not include 'Skipped interface HSN 0'
      The output should not include 'name already on device'
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

  # Nautobot 3.2 also names the root device on a port of an installed
  # module, so an import that grouped ports by device alone would give
  # repro-dl380 a second "HSN 0" and the NIC's "Port 1".
  Describe 'import'
    It 'imports the exported fixture into an empty datastore'
      When call run_import
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'keeps every module port on its module and off the parent device'
      When call imported_port_owners
      The output should equal 'added_to_device=0 device_hsn0=1 pcie1_hsn0=1 pcie2_port1=1'
    End
  End

  Describe 'import of a bay the library does not declare'
    It 'renames the NIC bay in Nautobot'
      When call rename_module_bay repro-dl380 PCIe2 'Slot B'
      The output should equal 'renamed=Slot B'
    End

    It 'imports the renamed bay without clearing it'
      When call run_import_renamed
      The status should equal 0
      The output should include 'Import completed successfully'
      The output should not include 'imported without a bay'
    End

    It 'declares the Nautobot bays on the device and keeps the module in the renamed one'
      When call imported_renamed_bays
      The output should equal 'declared=PCIe1,Slot B modules=PCIe1,Slot B'
    End

    It 're-exports the imported inventory without creating a module'
      When call reexport_renamed
      The status should equal 0
      The output should include 'Skipped modules (already exist): 3'
      The output should not include 'Created modules'
    End

    It 'leaves repro-dl380 with its two module bays'
      When call nautobot_module_bays repro-dl380
      The output should equal 'module_bays=2'
    End

    It 'restores the bay name'
      When call rename_module_bay repro-dl380 'Slot B' PCIe2
      The output should equal 'renamed=PCIe2'
    End
  End

  SWAP_LOG="${CANI_DIR}/swap-export.log"
  SWAPPED_CONF="${CANI_DIR}/swapped.yml"
  SWAPPED_DS="${CANI_DIR}/swapped.json"

  # export_swap runs an export with the given flags and keeps its log apart
  # from the fixture export's, whose accounting the earlier blocks assert.
  export_swap() {
    cani_cfg export nautobot "$@" >"$SWAP_LOG" 2>&1
    rc=$?
    cat "$SWAP_LOG"
    return $rc
  }

  # local_module_placement prints where the datastore places the named module.
  local_module_placement() {
    jq -r --arg name "$1" '
      . as $inv | .modules[] | select(.name == $name)
      | "parent=\($inv.devices[.parentDevice // ""].name // "-") " +
        "bay=\(if (.moduleBayName // "") == "" then "-" else .moduleBayName end) " +
        "location=\($inv.locations[.location // ""].name // "-")"' "$CANI_DS"
  }

  # nautobot_module_placement prints where Nautobot holds the named datastore
  # module, fetched through the module's recorded Nautobot ID.
  nautobot_module_placement() {
    MODULE="$1" DS="$CANI_DS" python3 - <<'PY'
import json
import os
import urllib.request

with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
module = next(m for m in inventory["modules"].values() if m["name"] == os.environ["MODULE"])
nautobot_id = module.get("externalIDs", {}).get("nautobot")
base_url = os.environ["NAUTOBOT_URL"].rstrip("/") + "/"
headers = {"Authorization": "Token " + os.environ["NAUTOBOT_TOKEN"]}
with urllib.request.urlopen(urllib.request.Request(base_url + f"dcim/modules/{nautobot_id}/?depth=1", headers=headers)) as response:
    remote = json.load(response)
bay = (remote.get("parent_module_bay") or {}).get("name") or "-"
location = (remote.get("location") or {}).get("name") or "-"
print(f"bay={bay} location={location}")
PY
  }

  # distinct_external_ids reports whether nic-a and nic-spare carry two
  # different recorded Nautobot IDs, which a bay-based match would cross-wire.
  distinct_external_ids() {
    jq -r '[.modules[] | select(.name == "nic-a" or .name == "nic-spare") | .externalIDs.nautobot]
      | "distinct=\(length == 2 and .[0] != null and .[1] != null and .[0] != .[1])"' "$CANI_DS"
  }

  # Import Nautobot into a third empty datastore after the swap.
  run_import_swapped() {
    sed "s|^datastore:.*|datastore: ${SWAPPED_DS}|" "$CANI_CONF" >"$SWAPPED_CONF" &&
      cani alpha --config "$SWAPPED_CONF" import nautobot \
        --default-location repro-section --default-status Active >"${CANI_DIR}/swapped-import.log" 2>&1
    rc=$?
    cat "${CANI_DIR}/swapped-import.log"
    return $rc
  }

  # imported_swap_placements prints where the import placed the modules that
  # carry nic-spare's and nic-a's Nautobot IDs; import names modules after
  # Nautobot's display text, so the ID is the only stable handle.
  imported_swap_placements() {
    DS="$CANI_DS" IMPORTED="$SWAPPED_DS" python3 - <<'PY'
import json
import os


def load(path):
    with open(path, encoding="utf-8") as datastore:
        return json.load(datastore)


source, imported = load(os.environ["DS"]), load(os.environ["IMPORTED"])
by_nautobot_id = {m.get("externalIDs", {}).get("nautobot"): m for m in imported["modules"].values()}


def describe(name):
    nautobot_id = next(m for m in source["modules"].values() if m["name"] == name)["externalIDs"]["nautobot"]
    module = by_nautobot_id.get(nautobot_id)
    if module is None:
        return f"{name}: <missing>"
    parent = imported["devices"].get(module.get("parentDevice") or "", {}).get("name", "-")
    location = imported["locations"].get(module.get("location") or "", {}).get("name", "-")
    return f"{name}: parent={parent} bay={module.get('moduleBayName') or '-'} location={location}"


print(describe("nic-spare") + " | " + describe("nic-a"))
PY
  }

  # A failed NIC is replaced by a spare with `swap module`; an export with
  # --merge then moves both modules in Nautobot through their recorded IDs, so
  # the spare is never matched to the module it replaces through the bay they
  # share in turn. A plain bay move follows the same path, and an import of
  # the result places both modules where the datastore does.
  Describe 'module swap'
    It 'stores a spare NIC at repro-section and exports it'
      cani_cfg add module hpe-2p-25gbe-pcie-nic --spare --location repro-section --name nic-spare --serial SP1 >/dev/null 2>&1
      When call export_swap
      The status should equal 0
      The output should include 'Created modules: 1'
      The output should include 'Skipped modules (already exist): 3'
    End

    It 'holds the spare at its location in Nautobot'
      When call nautobot_module_placement nic-spare
      The output should equal 'bay=- location=repro-section'
    End

    It 'previews the swap without changing the datastore'
      When call cani_cfg swap module nic-a nic-spare --dry-run
      The status should equal 0
      The output should include 'Would install nic-spare in repro-dl380 bay PCIe2'
      The output should include 'Would store nic-a at location repro-section'
    End

    It 'leaves nic-a installed after the preview'
      When call local_module_placement nic-a
      The output should equal 'parent=repro-dl380 bay=PCIe2 location=-'
    End

    It 'swaps the failed NIC for the spare in the datastore'
      When call cani_cfg swap module nic-a nic-spare
      The status should equal 0
      The stderr should include 'Swapped modules: nic-spare'
    End

    It 'installs the spare in PCIe2 locally'
      When call local_module_placement nic-spare
      The output should equal 'parent=repro-dl380 bay=PCIe2 location=-'
    End

    It 'stores the failed NIC at repro-section locally'
      When call local_module_placement nic-a
      The output should equal 'parent=- bay=- location=repro-section'
    End

    It 'reports the pending moves without --merge and moves nothing'
      When call export_swap
      The status should equal 0
      The output should include 'Skipped modules (placement differs; use --merge to move): 2'
      The output should include 'not exported as authored'
      The output should not include 'Moved module'
    End

    It 'leaves the failed NIC in PCIe2 in Nautobot after the plain export'
      When call nautobot_module_placement nic-a
      The output should equal 'bay=PCIe2 location=-'
    End

    It 'previews both moves with --merge --dry-run'
      When call export_swap --merge --dry-run
      The status should equal 0
      The output should include 'Would move module: nic-a (spare at location: repro-section)'
      The output should include 'Would move module: nic-spare (parent: repro-dl380, bay: PCIe2)'
    End

    It 'leaves the spare at its location in Nautobot after the dry run'
      When call nautobot_module_placement nic-spare
      The output should equal 'bay=- location=repro-section'
    End

    It 'moves both modules with --merge and creates none'
      When call export_swap --merge
      The status should equal 0
      The output should include 'Moved module: nic-a (spare at location: repro-section)'
      The output should include 'Moved module: nic-spare (parent: repro-dl380, bay: PCIe2)'
      The output should include 'Moved modules: 2'
      The output should not include 'Created modules'
    End

    It 'installs the spare in PCIe2 in Nautobot'
      When call nautobot_module_placement nic-spare
      The output should equal 'bay=PCIe2 location=-'
    End

    It 'stores the failed NIC at repro-section in Nautobot'
      When call nautobot_module_placement nic-a
      The output should equal 'bay=- location=repro-section'
    End

    It 'keeps the two modules under two Nautobot IDs'
      When call distinct_external_ids
      The output should equal 'distinct=true'
    End

    It 're-exports idempotently after the swap'
      When call export_swap --merge
      The status should equal 0
      The output should include 'Skipped modules (already exist): 4'
      The output should not include 'Moved module'
    End

    It 'moves a module between bays with update module --bay and --merge'
      cani_cfg update module nic-spare --bay PCIe3 >/dev/null 2>&1
      When call export_swap --merge
      The status should equal 0
      The output should include 'Moved module: nic-spare (parent: repro-dl380, bay: PCIe3)'
      The output should include 'Moved modules: 1'
    End

    It 'holds the moved module in PCIe3 in Nautobot'
      When call nautobot_module_placement nic-spare
      The output should equal 'bay=PCIe3 location=-'
    End

    It 'imports the swapped inventory into an empty datastore'
      When call run_import_swapped
      The status should equal 0
      The output should include 'Import completed successfully'
    End

    It 'places both modules where the datastore does'
      When call imported_swap_placements
      The output should equal 'nic-spare: parent=repro-dl380 bay=PCIe3 location=- | nic-a: parent=- bay=- location=repro-section'
    End
  End

End
