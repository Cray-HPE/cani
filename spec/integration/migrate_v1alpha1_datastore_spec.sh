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

Describe 'INTEGRATION:'

It 'migrates a v1alpha1 datastore to v1alpha4'
	BeforeCall 'setup_datastore_migration_env canitestdb_v1alpha1.json'
	When call bin/cani alpha show --config "$CANI_CONF"
	The status should equal 0
	The stdout should be present
	The stderr should include 'creating default config'
	The stderr should include 'Migrated datastore from v1alpha1 to v1alpha4'
	The file "$CANI_CONF" should be exist
	The file "$CANI_DS" should be exist
	The file "$CANI_DS.canisave" should be exist
	The contents of file "$CANI_DS" should include '"schemaVersion"'
	The contents of file "$CANI_DS" should include '"v1alpha4"'
	The contents of file "$CANI_DS" should include '"devices"'
	The contents of file "$CANI_DS" should include '"locations"'
	The contents of file "$CANI_DS" should include '"racks"'
	The contents of file "$CANI_DS" should not include '"Hardware"'
	The contents of file "$CANI_DS" should not include '"v1alpha1"'
End

It 'migrates released v1alpha3 additions to v1alpha4 without loss'
	When call v1alpha3_migration_summary
	The status should equal 0
	The stderr should include 'Migrated datastore from v1alpha3 to v1alpha4'
	The output should include 'backup_matches=True'
	The output should include 'schema=v1alpha4'
	The output should include 'device_fields=True'
	The output should include 'interface_spec_fields=True'
	The output should include 'interface_fields=True'
	The output should include 'vrf_fields=True'
End

End
