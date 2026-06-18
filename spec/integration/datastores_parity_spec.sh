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

# Integration coverage for pkg/datastores contracts surfaced through CLI use.

#shellcheck disable=SC2317
relative_datastore_summary() {
  setup_test_env
  _rel_dir="$CANI_DIR/relative-config"
  _rel_conf="$_rel_dir/cani.yml"
  _rel_ds="$_rel_dir/relative-canidb.json"
  mkdir -p "$_rel_dir"
  REL_CONF="$_rel_conf" FIXTURE_CONF="$FIXTURES/cani/configs/test_config.yml" python3 - <<'PY'
import os
from pathlib import Path

fixture = Path(os.environ["FIXTURE_CONF"])
target = Path(os.environ["REL_CONF"])
text = fixture.read_text()
text = text.replace("datastore: .cani/canidb.json", "datastore: relative-canidb.json")
target.write_text(text)
PY

  bin/cani --config "$_rel_conf" alpha add metadata tag RelativeStore >/dev/null 2>&1 || return $?

  REL_DS="$_rel_ds" DEFAULT_DS="$CANI_DS" python3 - <<'PY'
import json
import os
from pathlib import Path

rel = Path(os.environ["REL_DS"])
default = Path(os.environ["DEFAULT_DS"])
print("relative_exists=" + str(rel.exists()))
print("default_exists=" + str(default.exists()))
if rel.exists():
    with rel.open(encoding="utf-8") as f:
        inv = json.load(f)
    tags = [entry.get("name", "") for entry in inv.get("metadata", {}).get("tags", [])]
    mode = oct(rel.stat().st_mode & 0o777)
    print("relative_tags=" + ",".join(tags))
    print("relative_mode=" + mode)
PY
}

#shellcheck disable=SC2317
datastore_path_override_summary() {
  setup_test_env
  _override_ds="$CANI_DIR/override/nested/override.json"

  bin/cani --config "$CANI_CONF" --datastore-path "$_override_ds" alpha add metadata tag OverrideStore >/dev/null 2>&1 || return $?

  OVERRIDE_DS="$_override_ds" DEFAULT_DS="$CANI_DS" CANI_CONF_PATH="$CANI_CONF" python3 - <<'PY'
import json
import os
from pathlib import Path

override = Path(os.environ["OVERRIDE_DS"])
default = Path(os.environ["DEFAULT_DS"])
conf = Path(os.environ["CANI_CONF_PATH"])
print("override_exists=" + str(override.exists()))
print("default_exists=" + str(default.exists()))
print("config_has_override=" + str(str(override) in conf.read_text()))
if override.exists():
    with override.open(encoding="utf-8") as f:
        inv = json.load(f)
    tags = [entry.get("name", "") for entry in inv.get("metadata", {}).get("tags", [])]
    print("override_tags=" + ",".join(tags))
PY
}

#shellcheck disable=SC2317
metadata_migration_summary() {
  setup_test_env
  cat >"$CANI_DS" <<'JSON'
{
  "schemaVersion": "v1alpha2",
  "providerMetadata": {
    "nautobot": {
      "roles": [{"name": "LegacyRole", "contentTypes": ["dcim.device"]}],
      "statuses": [{"name": "Active", "color": "00ff00"}],
      "tags": [{"name": "legacy-tag"}]
    }
  },
  "locations": {},
  "racks": {},
  "devices": {},
  "modules": {},
  "cables": {},
  "frus": {},
  "interfaces": {}
}
JSON

  bin/cani alpha show metadata --config "$CANI_CONF" >/dev/null 2>&1 || return $?

  CANI_DS_PATH="$CANI_DS" python3 - <<'PY'
import json
import os

with open(os.environ["CANI_DS_PATH"], encoding="utf-8") as f:
    inv = json.load(f)
metadata = inv.get("metadata", {})
roles = [entry.get("name", "") for entry in metadata.get("roles", [])]
statuses = [entry.get("name", "") for entry in metadata.get("statuses", [])]
tags = [entry.get("name", "") for entry in metadata.get("tags", [])]
print("roles=" + ",".join(roles))
print("statuses=" + ",".join(statuses))
print("tags=" + ",".join(tags))
print("top_provider_metadata=" + str("providerMetadata" in inv))
PY
}

Describe 'INTEGRATION: datastores ShellSpec parity'

  It 'stores relative datastore paths next to the config file'
    When call relative_datastore_summary
    The status should equal 0
    The output should include 'relative_exists=True'
    The output should include 'default_exists=False'
    The output should include 'relative_tags=RelativeStore'
    The output should include 'relative_mode=0o600'
  End

  It 'uses datastore-path override without persisting it to config'
    When call datastore_path_override_summary
    The status should equal 0
    The output should include 'override_exists=True'
    The output should include 'default_exists=False'
    The output should include 'config_has_override=False'
    The output should include 'override_tags=OverrideStore'
  End

  It 'rejects unsupported datastore types through the CLI'
    BeforeCall setup_test_env
    When call bin/cani --config "$CANI_CONF" --datastore postgres alpha show
    The status should equal 1
    The stderr should include 'unsupported datastore type: postgres'
  End

  It 'surfaces invalid JSON datastore parse errors through the CLI'
    BeforeCall setup_test_env
    BeforeCall "printf '{not-json' > '$CANI_DS'"
    When call bin/cani alpha show --config "$CANI_CONF"
    The status should equal 1
    The stderr should include 'parsing inventory'
  End

  It 'migrates old inventory-level providerMetadata into typed metadata'
    When call metadata_migration_summary
    The status should equal 0
    The output should include 'roles=LegacyRole'
    The output should include 'statuses=Active'
    The output should include 'tags=legacy-tag'
    The output should include 'top_provider_metadata=False'
  End

End
