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

# Helpers for the Nautobot round-trip specs (spec/integration/nautobot_rt_*).
#
# A spec is a short narrative of cani commands; every claim about state goes
# through the oracle in testdata/fixtures/nautobot/roundtrip: projections of
# both sides compared as values, and Nautobot's change ledger as a write
# detector. Snapshots live outside CANI_DIR so the environment builders that
# wipe it cannot take the evidence with them.

RT_DIR="${CANI_DIR:-/tmp/.cani}-rt"

# --- environment ------------------------------------------------------------

#shellcheck disable=SC2317
rt_begin() {
  setup_nautobot_env
  rm -rf "$RT_DIR"
  mkdir -p "$RT_DIR"
}

#shellcheck disable=SC2317
nautobot_unreachable() {
  ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
    "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
}

# Write a Nautobot config whose export section is given as key=value pairs
# (create_device_types=false merge=true ...). Every default is filled in so an
# export with a create_* flag off never falls into the interactive prompt.
#shellcheck disable=SC2317
rt_write_config() {
  _rt_create_device_types=true _rt_create_location_types=true _rt_create_module_types=true
  _rt_create_locations=true _rt_create_statuses=true _rt_create_roles=true
  _rt_merge=false _rt_dry_run=false
  for _rt_pair in "$@"; do
    case "$_rt_pair" in
      create_device_types=*)   _rt_create_device_types=${_rt_pair#*=} ;;
      create_location_types=*) _rt_create_location_types=${_rt_pair#*=} ;;
      create_module_types=*)   _rt_create_module_types=${_rt_pair#*=} ;;
      create_locations=*)      _rt_create_locations=${_rt_pair#*=} ;;
      create_statuses=*)       _rt_create_statuses=${_rt_pair#*=} ;;
      create_roles=*)          _rt_create_roles=${_rt_pair#*=} ;;
      merge=*)                 _rt_merge=${_rt_pair#*=} ;;
      dry_run=*)               _rt_dry_run=${_rt_pair#*=} ;;
      *) printf 'rt_write_config: unknown key %s\n' "$_rt_pair" >&2; return 1 ;;
    esac
  done
  cat >"$CANI_CONF" <<EOF
providers:
  nautobot:
    default_location: test-dc
    default_role: Compute
    default_status: Active
    token: "${NAUTOBOT_TOKEN}"
    url: ${NAUTOBOT_URL}
    import: {}
    export:
      create_device_types: ${_rt_create_device_types}
      create_location_types: ${_rt_create_location_types}
      create_module_types: ${_rt_create_module_types}
      create_locations: ${_rt_create_locations}
      create_statuses: ${_rt_create_statuses}
      create_roles: ${_rt_create_roles}
      merge: ${_rt_merge}
      dry_run: ${_rt_dry_run}
datastore: ${CANI_DS}
debug: false
types_dirs: []
types_repos:
  - https://github.com/netbox-community/devicetype-library.git
types_repo_pull: false
EOF
}

# --- commands ---------------------------------------------------------------

#shellcheck disable=SC2317
rt_cani() { cani alpha --config "$CANI_CONF" "$@"; }

# The CLI-built counterpart of `rt_seed`: dc > level > section, one rack, two
# DL325s (1000base-t ports only, so every interface exists in Nautobot) joined
# by one cable, and the same two-tenant IPAM layout. Every projected field is
# set explicitly so the oracle never has to know an export default.
#shellcheck disable=SC2317
rt_build_inventory() {
  rt_cani add location dc --name rt-dc --status Active &&
  rt_cani add location level --name rt-level --parent rt-dc --status Active &&
  rt_cani add location section --name rt-section --parent rt-level --status Active &&
  rt_cani add rack hpe-42u-800mmx1200mm-g2-enterprise-shock-rack \
    --location rt-section --name rt-rack --status Active &&
  rt_cani add device hpe-proliant-dl325-gen11-8sff --rack rt-rack --position 1 \
    --face front --name rt-node-1 --status Active --serial RT-001 &&
  rt_cani add device hpe-proliant-dl325-gen11-8sff --rack rt-rack --position 2 \
    --face front --name rt-node-2 --status Active --serial RT-002 &&
  rt_cani update device rt-node-1 --role Compute &&
  rt_cani update device rt-node-2 --role Compute &&
  rt_cani add cable hpe-3m-cat6-stp --a-device rt-node-1 --a-port 'Gig-E 1' \
    --b-device rt-node-2 --b-port 'Gig-E 1' --label rt-node-1-to-rt-node-2 \
    --color 00ff00 --status Connected &&
  rt_cani add vrf blue --namespace tenant-a --rd 65000:1 \
    --description 'blue VRF in tenant-a' --device rt-node-1 &&
  rt_cani add vrf blue --namespace tenant-b --rd 65000:2 \
    --description 'blue VRF in tenant-b' &&
  rt_cani add prefix 10.0.0.0/24 --namespace tenant-a --type network --status Active \
    --vrf tenant-a/blue --description 'network in tenant-a' &&
  rt_cani add prefix 10.0.0.0/24 --namespace tenant-b --type network --status Active \
    --description 'network in tenant-b' &&
  rt_cani add ip 10.0.0.5/24 --namespace tenant-a --type host --status Active \
    --dns-name host.tenant-a.example.test --description 'host in tenant-a' &&
  rt_cani add ip 10.0.0.5/24 --namespace tenant-b --type host --status Active \
    --dns-name host.tenant-b.example.test --description 'host in tenant-b' &&
  rt_cani add vlan 3100 --name rt-vlan --status Active --location rt-section &&
  rt_cani add prefix 10.31.0.0/24 --type network --status Active \
    --vlan rt-vlan --location rt-section
}

#shellcheck disable=SC2317
rt_py() { PYTHONPATH="$FIXTURES/nautobot" python3 -m roundtrip "$@"; }

#shellcheck disable=SC2317
rt_reset() { rt_py reset; }

#shellcheck disable=SC2317
rt_seed() { rt_py seed; }

#shellcheck disable=SC2317
rt_reset_and_seed() { rt_reset && rt_seed; }

# Run an export, keep its log at $RT_DIR/export.log and echo it, so a spec
# can assert on the summary and a later step can still read the file.
#shellcheck disable=SC2317
rt_export() {
  rt_cani export nautobot "$@" >"$RT_DIR/export.log" 2>&1
  _rt_rc=$?
  cat "$RT_DIR/export.log"
  return $_rt_rc
}

#shellcheck disable=SC2317
rt_import() {
  rt_cani import nautobot "$@" >"$RT_DIR/import.log" 2>&1
  _rt_rc=$?
  cat "$RT_DIR/import.log"
  return $_rt_rc
}

# --- oracle -----------------------------------------------------------------

# rt_snapshot nautobot NAME | rt_snapshot cani NAME  -> $RT_DIR/NAME.json
#shellcheck disable=SC2317
rt_snapshot() {
  case "$1" in
    nautobot) rt_py project nautobot "$RT_DIR/$2.json" ;;
    cani)     rt_py project cani "$CANI_DS" "$RT_DIR/$2.json" ;;
    *) printf 'rt_snapshot: side must be nautobot or cani\n' >&2; return 1 ;;
  esac
}

# Print "projections equal" or one line per difference (exit 1).
#shellcheck disable=SC2317
rt_diff() { rt_py diff "$RT_DIR/$1.json" "$RT_DIR/$2.json"; }

#shellcheck disable=SC2317
rt_assert_parity() {
  rt_snapshot nautobot parity-nautobot &&
  rt_snapshot cani parity-cani &&
  rt_diff parity-nautobot parity-cani
}

# Compare the datastore now with the snapshot `rt_snapshot cani NAME` took.
#shellcheck disable=SC2317
rt_assert_cani_unchanged() { rt_snapshot cani "$1.after" && rt_diff "$1" "$1.after"; }

# Compare Nautobot now with the snapshot `rt_snapshot nautobot NAME` took.
#shellcheck disable=SC2317
rt_assert_nautobot_unchanged() { rt_snapshot nautobot "$1.after" && rt_diff "$1" "$1.after"; }

#shellcheck disable=SC2317
rt_mark_changes() { rt_py changes >"$RT_DIR/changes.mark"; }

# Print "changes=N" followed by the N newest ledger entries since the mark.
#shellcheck disable=SC2317
rt_changes_since_mark() { rt_py changes "$(cat "$RT_DIR/changes.mark")"; }

# Datastore fingerprint; "absent" when no datastore exists yet.
#shellcheck disable=SC2317
rt_ds_checksum() {
  if [ -f "$CANI_DS" ]; then cksum <"$CANI_DS" | cut -d' ' -f1-2; else echo absent; fi
}

# Number of datastore records carrying a Nautobot external ID.
#shellcheck disable=SC2317
rt_stamp_count() { rt_py stamps "$CANI_DS"; }
