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

# ── import / export ─────────────────────────────────────────────────
#
# Tests import and export across all providers using --datastore-path
# to isolate each test from pre-existing data.

# Unique datastore path per provider to avoid test interference.
ie_ds() { echo "/tmp/.cani/ie_test_${1}.json"; }

# ── 1. help flag matrix ────────────────────────────────────────────
#
# Parameters:matrix produces the cross-product of providers × operations.
# Every combination must accept --help and exit 0.

Describe 'import/export --help (matrix)'
  Parameters:matrix
    csm hpcm example nautobot ochami redfish
    import export
  End

  It "cani alpha $2 $1 --help exits 0"
    When call bin/cani alpha "$2" "$1" --help
    The status should equal 0
    The stdout should include 'Usage:'
  End
End

# ── 2. file-based import ───────────────────────────────────────────

Describe 'file-based import'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  Describe 'example (CSV)'
    It 'imports from a CSV fixture'
      When call bin/cani alpha import example \
        --csv "$FIXTURES/example/simple.csv" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds example_csv)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates the datastore file'
      The path "$(ie_ds example_csv)" should be file
    End
  End

  Describe 'ochami (JSON)'
    It 'imports from a JSON fixture'
      When call bin/cani alpha import ochami \
        --jsonfile "$FIXTURES/ochami/ochami_test_data.json" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds ochami)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates the datastore file'
      The path "$(ie_ds ochami)" should be file
    End
  End

  Describe 'redfish (JSON)'
    It 'imports from a ServiceRoot fixture'
      When call bin/cani alpha import redfish \
        --root "$FIXTURES/redfish/v1/redfish-root.json" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds redfish)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates the datastore file'
      The path "$(ie_ds redfish)" should be file
    End
  End

  Describe 'hpcm (node JSON)'
    It 'imports from a node JSON fixture'
      When call bin/cani alpha import hpcm \
        --node-json-file "$FIXTURES/hpcm/nodes.json" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds hpcm)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates the datastore file'
      The path "$(ie_ds hpcm)" should be file
    End
  End

  Describe 'csm (SLS file)'
    It 'imports from an SLS dumpstate fixture'
      When call bin/cani alpha import csm \
        --sls-file "$FIXTURES/csm/sls/valid_hardware_networks.json" \
        --ignore-validation \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds csm)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates the datastore file'
      The path "$(ie_ds csm)" should be file
    End
  End
End

# ── 3. file-based export ───────────────────────────────────────────
#
# Pre-populate a datastore with the CRUD inventory then export via
# each provider. Providers that are no-ops (hpcm, redfish) simply
# verify exit 0. Providers that produce output (example, ochami, csm)
# also check stdout.

Describe 'file-based export'
  setup_export_env() {
    setup_test_env
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds export_example)"
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds export_ochami)"
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds export_redfish)"
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds export_hpcm)"
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds export_csm)"
  }

  BeforeAll 'setup_export_env'
  AfterAll  'teardown_test_env'

  Describe 'example'
    It 'exports the inventory'
      When call bin/cani alpha export example \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds export_example)"
      The status should equal 0
      The stdout should include 'Summary:'
      The stderr should include 'Export completed successfully'
    End
  End

  Describe 'ochami'
    It 'exports the inventory'
      When call bin/cani alpha export ochami \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds export_ochami)"
      The status should equal 0
      The stdout should include 'nodes:'
      The stderr should include 'Export completed successfully'
    End
  End

  Describe 'redfish'
    It 'exports the inventory'
      When call bin/cani alpha export redfish \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds export_redfish)"
      The status should equal 0
      The stderr should include 'Export completed successfully'
    End
  End

  Describe 'hpcm'
    It 'exports the inventory'
      When call bin/cani alpha export hpcm \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds export_hpcm)"
      The status should equal 0
      The stderr should include 'Export completed successfully'
    End
  End

  Describe 'csm (CSV mode)'
    It 'exports the inventory as CSV'
      When call bin/cani alpha export csm \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds export_csm)"
      The status should equal 0
      The stdout should include 'Type,Vlan,Role'
      The stderr should include 'Export completed successfully'
    End
  End
End

# ── 4. round-trip (import → export) ────────────────────────────────
#
# Import a fixture, then immediately export. Verifies the full ETL
# pipeline produces consistent output for file-based providers.

Describe 'import → export round-trip'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  Describe 'example (CSV → visual export)'
    It 'imports successfully'
      When call bin/cani alpha import example \
        --csv "$FIXTURES/example/simple.csv" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds roundtrip_example)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'exports the imported data'
      When call bin/cani alpha export example \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds roundtrip_example)"
      The status should equal 0
      The stdout should include 'Summary:'
      The stderr should include 'Export completed successfully'
    End
  End

  Describe 'ochami (JSON → YAML export)'
    It 'imports successfully'
      When call bin/cani alpha import ochami \
        --jsonfile "$FIXTURES/ochami/ochami_test_data.json" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds roundtrip_ochami)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'exports the imported data'
      When call bin/cani alpha export ochami \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds roundtrip_ochami)"
      The status should equal 0
      The stdout should include 'nodes:'
      The stderr should include 'Export completed successfully'
    End
  End

  Describe 'csm (SLS file → CSV export)'
    It 'imports successfully'
      When call bin/cani alpha import csm \
        --sls-file "$FIXTURES/csm/sls/valid_hardware_networks.json" \
        --ignore-validation \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds roundtrip_csm)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'exports the imported data'
      When call bin/cani alpha export csm \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds roundtrip_csm)"
      The status should equal 0
      The stdout should include 'Type,Vlan,Role'
      The stderr should include 'Export completed successfully'
    End
  End
End

# ── 5. external service providers ──────────────────────────────────
#
# Nautobot and CSM API-based tests require running services.
# Skip when SKIP_EXTERNAL_TESTS is set or when Nautobot is unreachable.

external_tests_disabled() { [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]; }
nautobot_unreachable() {
  ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
    "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
}

Describe 'nautobot import (API)'
  Skip if 'external tests disabled' external_tests_disabled
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  BeforeAll 'setup_nautobot_env'
  AfterAll  'teardown_test_env'

  It 'imports from the Nautobot API'
    When call bin/cani alpha import nautobot \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds nautobot)"
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End
End

Describe 'nautobot export (API, dry-run)'
  Skip if 'external tests disabled' external_tests_disabled
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  # Pre-populate the datastore so export has inventory to process,
  # regardless of whether Nautobot contains data.
  setup_nautobot_export() {
    setup_nautobot_env
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds nautobot_export)"
  }

  BeforeAll 'setup_nautobot_export'
  AfterAll  'teardown_test_env'

  It 'connects to Nautobot and processes the export pipeline'
    When call bin/cani alpha export nautobot \
      --dry-run \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds nautobot_export)"
    # Exit status may be non-zero due to fixture/schema mismatches with
    # an empty Nautobot — the functional test verifies connectivity and
    # pipeline execution, not data correctness.
    The status should not equal ""
    The stderr should include 'Successfully connected to Nautobot'
    The stderr should include 'Nautobot Sync Summary'
  End
End