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

# Integration test: two independent networks that reuse the same addresses
# stay separate when their Nautobot IPAM data is imported into CANI.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - SKIP_EXTERNAL_TESTS != 1
#
# Nautobot scopes prefix and address uniqueness by namespace, so tenant-a and
# tenant-b can each hold 10.0.0.0/24 with host 10.0.0.5/24. CANI must import
# two prefixes and two addresses, each address under its own network's prefix,
# instead of folding both networks into one record.

Describe 'INTEGRATION: overlapping networks in Nautobot IPAM namespaces'

  nautobot_unreachable() {
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
  }

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  cani_cfg() { cani alpha --config "$CANI_CONF" "$@"; }

  seed_overlapping_networks() {
    python3 "$FIXTURES/nautobot/purge_nautobot_dcim.py" &&
    PYTHONPATH="$FIXTURES/nautobot" python3 - <<'PY'
from seed_nautobot import get_status_id
from seed_nautobot_api import api, find_or_create

status = {"id": get_status_id("Active")}
for name in ("tenant-a", "tenant-b"):
    namespace = {"id": find_or_create("ipam/namespaces/", "name", name, {"name": name})["id"]}
    api("POST", "ipam/prefixes/", {
        "prefix": "10.0.0.0/24",
        "namespace": namespace,
        "status": status,
        "type": "network",
    })
    api("POST", "ipam/ip-addresses/", {
        "address": "10.0.0.5/24",
        "namespace": namespace,
        "status": status,
        "type": "host",
        "description": "host in " + name,
    })
print("Seeded 10.0.0.0/24 and 10.0.0.5/24 in tenant-a and tenant-b")
PY
  }

  # Print the namespace of every datastore prefix with the given CIDR.
  cani_prefix_namespaces() {
    CIDR="$1" DS="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["DS"], encoding="utf-8") as datastore:
    prefixes = (json.load(datastore).get("prefixes") or {}).values()
print(" ".join(sorted(
    p.get("namespace") or "<none>" for p in prefixes if p.get("prefix") == os.environ["CIDR"]
)))
PY
  }

  # Print "<description> -> <parent prefix namespace>" for every datastore
  # address with the given host.
  cani_host_parents() {
    HOST="$1" DS="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["DS"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)
prefixes = inventory.get("prefixes") or {}
pairs = []
for address in (inventory.get("ipAddresses") or {}).values():
    if address.get("host") != os.environ["HOST"]:
        continue
    parent = prefixes.get(address.get("parent") or "", {})
    pairs.append(address.get("description", "") + " -> " + (parent.get("namespace") or "<none>"))
print("; ".join(sorted(pairs)))
PY
  }

  It 'seeds 10.0.0.0/24 and host 10.0.0.5 in two Nautobot namespaces'
    BeforeCall setup_nautobot_env
    When call seed_overlapping_networks
    The status should equal 0
    The output should include 'Seeded 10.0.0.0/24'
  End

  It 'imports both networks into an empty datastore'
    When call cani_cfg import nautobot
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End

  It 'keeps one 10.0.0.0/24 per namespace instead of merging them'
    When call cani_prefix_namespaces 10.0.0.0/24
    The output should equal 'tenant-a tenant-b'
  End

  It 'keeps each 10.0.0.5 under the 10.0.0.0/24 of its own namespace'
    When call cani_host_parents 10.0.0.5
    The output should equal 'host in tenant-a -> tenant-a; host in tenant-b -> tenant-b'
  End

End
