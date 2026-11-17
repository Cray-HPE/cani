#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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

#shellcheck disable=SC2317
v1alpha3_migration_summary() {
	setup_datastore_migration_env canitestdb_v1alpha3_v0.7.4.json
	_migration_fixture="$FIXTURES/cani/legacy/canitestdb_v1alpha3_v0.7.4.json"

	bin/cani alpha show --config "$CANI_CONF" >/dev/null || return $?

	if cmp -s "$_migration_fixture" "$CANI_DS.canisave"; then
		printf 'backup_matches=True\n'
	else
		printf 'backup_matches=False\n'
		return 1
	fi

	CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)

device = inventory["devices"]["10000000-0000-0000-0000-000000000001"]
interface_spec = device["interfaces"][0]
interface = inventory["interfaces"]["50000000-0000-0000-0000-000000000001"]
vrf = inventory["vrfs"]["40000000-0000-0000-0000-000000000001"]

expected_interface = {
    "tags": ["fabric"],
    "lag": "bond0",
    "mode": "tagged",
    "untaggedVlan": 100,
    "taggedVlans": [200, 300],
    "vrf": "management",
}
print("schema=" + inventory["schemaVersion"])
print("device_fields=" + str(
    device["assignedVlans"] == ["20000000-0000-0000-0000-000000000001"]
    and device["bmcParent"] == "30000000-0000-0000-0000-000000000001"
))
print("interface_spec_fields=" + str(all(
    interface_spec.get(key) == value for key, value in expected_interface.items()
)))
print("interface_fields=" + str(all(
    interface.get(key) == value for key, value in expected_interface.items()
)))
print("vrf_fields=" + str(vrf["name"] == "management" and vrf["rd"] == "65000:1"))
PY
}

# Loads a datastore the v1alpha9 writer produced from pairs typed with
# --metadata, and prints where each value is held after the upgrade.
#shellcheck disable=SC2317
v1alpha9_metadata_summary() {
	setup_datastore_migration_env canitestdb_v1alpha9_metadata.json
	_migration_fixture="$FIXTURES/cani/legacy/canitestdb_v1alpha9_metadata.json"

	bin/cani alpha show --config "$CANI_CONF" >/dev/null || return $?

	if cmp -s "$_migration_fixture" "$CANI_DS.canisave"; then
		printf 'backup_matches=True\n'
	else
		printf 'backup_matches=False\n'
		return 1
	fi

	CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as datastore:
    inventory = json.load(datastore)

devices = {device["name"]: device for device in inventory["devices"].values()}
locations = {location["name"]: location for location in inventory["locations"].values()}
rack = next(iter(inventory["racks"].values()))
left_typed = sorted(
    (record["name"], record["providerMetadata"]["nautobot"])
    for collection in ("devices", "racks", "locations")
    for record in inventory[collection].values()
    if "nautobot" in (record.get("providerMetadata") or {})
)

print("schema=" + inventory["schemaVersion"])
print("typed_role=" + devices["typed-role"]["role"])
print("kept_role=" + devices["test-device-2"]["role"])
print("device_custom_fields=" + json.dumps(devices["typed-role"]["customFields"], sort_keys=True))
print("conflict_custom_fields=" + json.dumps(devices["test-device-2"]["customFields"], sort_keys=True))
print("rack_custom_fields=" + json.dumps(rack["customFields"], sort_keys=True))
print("rack_keeps_csm_bucket=" + str(sorted(rack["providerMetadata"]) == ["csm"]))
print("location_custom_fields=" + json.dumps(locations["test-site"]["customFields"], sort_keys=True))
print("placed=" + str(devices["loose-dev"]["location"] == locations["removable-location"]["id"]))
print("racked_stays=" + str(devices["test-device"]["location"] == locations["test-site"]["id"]))
print("left_typed=" + json.dumps(left_typed, sort_keys=True))
PY
}

# Loads a datastore the v1alpha9 HPCM import wrote, imports the same nodes
# into a second datastore with this build, and prints whether every upgraded
# object holds its HPCM values where a fresh import puts them.
#shellcheck disable=SC2317
v1alpha9_hpcm_summary() {
	setup_datastore_migration_env canitestdb_v1alpha9_hpcm.json
	_migration_fixture="$FIXTURES/cani/legacy/canitestdb_v1alpha9_hpcm.json"
	_fresh_ds="$CANI_DIR/fresh.json"

	bin/cani alpha show --config "$CANI_CONF" >/dev/null || return $?
	bin/cani alpha import hpcm --node-json-file "$FIXTURES/hpcm/blade_system.json" \
		--config "$CANI_CONF" --datastore-path "$_fresh_ds" >/dev/null 2>&1 || return $?

	if cmp -s "$_migration_fixture" "$CANI_DS.canisave"; then
		printf 'backup_matches=True\n'
	else
		printf 'backup_matches=False\n'
		return 1
	fi

	CANI_DS_PATH="$CANI_DS" CANI_FRESH_PATH="$_fresh_ds" python3 - <<'PY'
import json
import os


def held(path):
    """Custom fields and provider metadata of every object, by kind and name."""
    with open(path, encoding="utf-8") as datastore:
        inventory = json.load(datastore)
    objects = {}
    for kind in ("locations", "racks", "devices", "modules", "frus"):
        for record in (inventory.get(kind) or {}).values():
            provider = json.loads(json.dumps(record.get("providerMetadata") or {}))
            # The two imports read the nodes from different paths.
            (provider.get("hpcm") or {}).pop("import_source", None)
            objects[kind + "/" + record["name"]] = {
                "customFields": record.get("customFields") or {},
                "providerMetadata": provider,
            }
    return inventory["schemaVersion"], objects


schema, upgraded = held(os.environ["CANI_DS_PATH"])
_, fresh = held(os.environ["CANI_FRESH_PATH"])
differing = sorted(key for key in set(upgraded) | set(fresh) if upgraded.get(key) != fresh.get(key))

print("schema=" + schema)
print("objects=" + str(len(upgraded)))
print("custom_fields_left=" + str(sum(len(entry["customFields"]) for entry in upgraded.values())))
print("differs_from_fresh_import=" + json.dumps(differing))
PY
}

Describe 'INTEGRATION:'

It 'migrates a v1alpha1 datastore to v1alpha10'
	BeforeCall 'setup_datastore_migration_env canitestdb_v1alpha1.json'
	When call bin/cani alpha show --config "$CANI_CONF"
	The status should equal 0
	The stdout should be present
	The stderr should include 'creating default config'
	The stderr should include 'Migrated datastore from v1alpha1 to v1alpha10'
	The file "$CANI_CONF" should be exist
	The file "$CANI_DS" should be exist
	The file "$CANI_DS.canisave" should be exist
	The contents of file "$CANI_DS" should include '"schemaVersion"'
	The contents of file "$CANI_DS" should include '"v1alpha10"'
	The contents of file "$CANI_DS" should include '"devices"'
	The contents of file "$CANI_DS" should include '"locations"'
	The contents of file "$CANI_DS" should include '"racks"'
	The contents of file "$CANI_DS" should not include '"Hardware"'
	The contents of file "$CANI_DS" should not include '"v1alpha1"'
End

It 'migrates released v1alpha3 additions to v1alpha10 without loss'
	When call v1alpha3_migration_summary
	The status should equal 0
	The stderr should include 'Migrated datastore from v1alpha3 to v1alpha10'
	The output should include 'backup_matches=True'
	The output should include 'schema=v1alpha10'
	The output should include 'device_fields=True'
	The output should include 'interface_spec_fields=True'
	The output should include 'interface_fields=True'
	The output should include 'vrf_fields=True'
End

It 'moves v1alpha9 provider metadata to the fields and custom fields that own it'
	When call v1alpha9_metadata_summary
	The status should equal 0
	The stderr should include 'Migrated datastore from v1alpha9 to v1alpha10'
	The stderr should include 'device "test-device-2"'
	The stderr should include 'provider metadata "role" was dropped: it was not the role export sent'
	The stderr should include 'device "test-device" ('
	The stderr should include 'provider metadata "location" was not applied'
	The stderr should include 'device "unplaced-dev"'
	The stderr should include 'provider metadata "location" names no location in the inventory'
	The output should include 'backup_matches=True'
	The output should include 'schema=v1alpha10'
	The output should include 'typed_role=Compute'
	The output should include 'kept_role=Storage'
	The output should include 'device_custom_fields={"asset_owner": "hpc-ops"}'
	The output should include 'conflict_custom_fields={"rack_unit_note": "front"}'
	The output should include 'rack_custom_fields={"cost_center": "cc-4411"}'
	The output should include 'rack_keeps_csm_bucket=True'
	The output should include 'location_custom_fields={"region": "emea"}'
	The output should include 'placed=True'
	The output should include 'racked_stays=True'
	The output should include 'left_typed=[["unplaced-dev", {"location": "dc-nowhere"}]]'
End

It 'moves what the v1alpha9 HPCM import stored as custom fields to the HPCM bucket'
	When call v1alpha9_hpcm_summary
	The status should equal 0
	The stderr should include 'Migrated datastore from v1alpha9 to v1alpha10'
	The stderr should include '9 custom field(s) the HPCM import wrote on modules, locations and inventory items moved to providerMetadata.hpcm'
	The output should include 'backup_matches=True'
	The output should include 'schema=v1alpha10'
	The output should include 'objects=10'
	The output should include 'custom_fields_left=0'
	The output should include 'differs_from_fresh_import=[]'
End

End
