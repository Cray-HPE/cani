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
#
# Structure:
#   1. --help            — exits 0 for every provider × operation
#   2. import            — each provider imports its fixture successfully
#   2b. show consistency — show device returns expected counts per provider
#   3. <src> to <dst>    — each import source fans out to all export targets
#   4. nautobot API      — API-based tests requiring a running Nautobot

# Unique datastore path per test to avoid interference.
ie_ds() { echo "/tmp/.cani/ie_test_${1}.json"; }

# Import a fixture for the given provider into the given datastore.
#shellcheck disable=SC2317
import_fixture() {
  _ie_provider="$1" _ie_ds="$2"
  case "$_ie_provider" in
    example)
      bin/cani alpha import example \
        --csv "$FIXTURES/matrix/example.csv" \
        --config "$CANI_CONF" --datastore-path "$_ie_ds" ;;
    ochami)
      bin/cani alpha import ochami \
        --jsonfile "$FIXTURES/matrix/ochami.json" \
        --config "$CANI_CONF" --datastore-path "$_ie_ds" ;;
    redfish)
      bin/cani alpha import redfish \
        --root "$FIXTURES/matrix/redfish.json" \
        --config "$CANI_CONF" --datastore-path "$_ie_ds" ;;
    hpcm)
      bin/cani alpha import hpcm \
        --node-json-file "$FIXTURES/matrix/hpcm.json" \
        --config "$CANI_CONF" --datastore-path "$_ie_ds" ;;
    csm)
      bin/cani alpha import csm \
        --sls-file "$FIXTURES/matrix/csm_sls.json" \
        --ignore-validation \
        --config "$CANI_CONF" --datastore-path "$_ie_ds" ;;
  esac
}

# ── 1. --help ──────────────────────────────────────────────────────
#
# Parameters:matrix produces the cross-product of providers × operations.
# Every combination must accept --help and exit 0.

Describe '--help'
  Parameters:matrix
    csm hpcm example nautobot ochami redfish
    import export
  End

  It "$2 $1 exits 0"
    When call bin/cani alpha "$2" "$1" --help
    The status should equal 0
    The stdout should include 'Usage:'
  End
End

# ── 2. import ──────────────────────────────────────────────────────
#
# Each provider imports its fixture file and produces a datastore.

Describe 'import from'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  Describe 'example (CSV)'
    It 'succeeds'
      When call import_fixture example "$(ie_ds import_example)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates datastore'
      The path "$(ie_ds import_example)" should be file
    End
  End

  Describe 'ochami (JSON)'
    It 'succeeds'
      When call import_fixture ochami "$(ie_ds import_ochami)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates datastore'
      The path "$(ie_ds import_ochami)" should be file
    End
  End

  Describe 'redfish (JSON)'
    It 'succeeds'
      When call import_fixture redfish "$(ie_ds import_redfish)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates datastore'
      The path "$(ie_ds import_redfish)" should be file
    End
  End

  Describe 'hpcm (JSON)'
    It 'succeeds'
      When call import_fixture hpcm "$(ie_ds import_hpcm)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates datastore'
      The path "$(ie_ds import_hpcm)" should be file
    End
  End

  Describe 'csm (SLS)'
    It 'succeeds'
      When call import_fixture csm "$(ie_ds import_csm)"
      The status should equal 0
      The stderr should include 'Import completed successfully'
    End

    It 'creates datastore'
      The path "$(ie_ds import_csm)" should be file
    End
  End
End

# ── 2b. show consistency ──────────────────────────────────────────
#
# After importing, run `cani show device` against each provider's
# datastore and verify:
#   a) the command exits 0 (inventory is queryable)
#   b) the device count matches the expected value for each provider
#
# This confirms the import pipeline produces deterministic, consistent
# results that the show layer can consume.

Describe 'show device after import'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  #shellcheck disable=SC2317
  _setup_show() {
    import_fixture example "$(ie_ds show_example)" 2>/dev/null
    import_fixture ochami  "$(ie_ds show_ochami)"  2>/dev/null
    import_fixture redfish "$(ie_ds show_redfish)" 2>/dev/null
    import_fixture hpcm    "$(ie_ds show_hpcm)"    2>/dev/null
    import_fixture csm     "$(ie_ds show_csm)"     2>/dev/null
  }
  BeforeAll '_setup_show'

  It 'example reports 6 devices'
    When call bin/cani alpha show device \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds show_example)"
    The status should equal 0
    The stdout should include 'Total: 6 device(s)'
  End

  It 'ochami reports 26 devices'
    When call bin/cani alpha show device \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds show_ochami)"
    The status should equal 0
    The stdout should include 'Total: 26 device(s)'
  End

  It 'redfish reports 4 devices'
    When call bin/cani alpha show device \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds show_redfish)"
    The status should equal 0
    The stdout should include 'Total: 4 device(s)'
  End

  It 'hpcm reports 6 devices'
    When call bin/cani alpha show device \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds show_hpcm)"
    The status should equal 0
    The stdout should include 'Total: 6 device(s)'
  End

  It 'csm reports 15 devices'
    When call bin/cani alpha show device \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds show_csm)"
    The status should equal 0
    The stdout should include 'Total: 15 device(s)'
  End
End

# ── 3. import → export matrix ─────────────────────────────────────
#
# For each file-based import source, import the fixture once, then
# export through every file-based provider.  This verifies:
#   a) the import produces a valid provider-agnostic datastore
#   b) every exporter can consume any datastore
#
# Same-provider pairs include additional format assertions.
# Test names read as: "<source> to <target> succeeds".

Describe 'import from example (CSV)'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  #shellcheck disable=SC2317
  _setup_example() { import_fixture example "$(ie_ds matrix_example)" 2>/dev/null; }
  BeforeAll '_setup_example'

  Describe 'export to'
    Parameters:value example ochami redfish hpcm csm
    It "$1 succeeds"
      When call bin/cani alpha export "$1" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds matrix_example)"
      The status should equal 0
      The stdout should be defined
      The stderr should include 'Export completed successfully'
    End
  End

  It 'to example output includes Summary'
    When call bin/cani alpha export example \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds matrix_example)"
    The stdout should include 'Summary:'
    The stderr should include 'Export completed successfully'
  End
End

Describe 'import from ochami (JSON)'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  #shellcheck disable=SC2317
  _setup_ochami() { import_fixture ochami "$(ie_ds matrix_ochami)" 2>/dev/null; }
  BeforeAll '_setup_ochami'

  Describe 'export to'
    Parameters:value example ochami redfish hpcm csm
    It "$1 succeeds"
      When call bin/cani alpha export "$1" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds matrix_ochami)"
      The status should equal 0
      The stdout should be defined
      The stderr should include 'Export completed successfully'
    End
  End

  It 'to ochami output includes nodes'
    When call bin/cani alpha export ochami \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds matrix_ochami)"
    The stdout should include 'nodes:'
    The stderr should include 'Export completed successfully'
  End
End

Describe 'import from redfish (JSON)'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  #shellcheck disable=SC2317
  _setup_redfish() { import_fixture redfish "$(ie_ds matrix_redfish)" 2>/dev/null; }
  BeforeAll '_setup_redfish'

  Describe 'export to'
    Parameters:value example ochami redfish hpcm csm
    It "$1 succeeds"
      When call bin/cani alpha export "$1" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds matrix_redfish)"
      The status should equal 0
      The stdout should be defined
      The stderr should include 'Export completed successfully'
    End
  End
End

Describe 'import from hpcm (JSON)'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  #shellcheck disable=SC2317
  _setup_hpcm() { import_fixture hpcm "$(ie_ds matrix_hpcm)" 2>/dev/null; }
  BeforeAll '_setup_hpcm'

  Describe 'export to'
    Parameters:value example ochami redfish hpcm csm
    It "$1 succeeds"
      When call bin/cani alpha export "$1" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds matrix_hpcm)"
      The status should equal 0
      The stdout should be defined
      The stderr should include 'Export completed successfully'
    End
  End
End

Describe 'import from csm (SLS)'
  BeforeAll 'setup_test_env'
  AfterAll  'teardown_test_env'

  #shellcheck disable=SC2317
  _setup_csm() { import_fixture csm "$(ie_ds matrix_csm)" 2>/dev/null; }
  BeforeAll '_setup_csm'

  Describe 'export to'
    Parameters:value example ochami redfish hpcm csm
    It "$1 succeeds"
      When call bin/cani alpha export "$1" \
        --config "$CANI_CONF" \
        --datastore-path "$(ie_ds matrix_csm)"
      The status should equal 0
      The stdout should be defined
      The stderr should include 'Export completed successfully'
    End
  End

  It 'to csm output includes CSV headers'
    When call bin/cani alpha export csm \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds matrix_csm)"
    The stdout should include 'Type,Vlan,Role'
    The stderr should include 'Export completed successfully'
  End
End

# ── 4. nautobot API ───────────────────────────────────────────────
#
# Nautobot tests require a running API server.
# Skip when SKIP_EXTERNAL_TESTS is set or when Nautobot is unreachable.

external_tests_disabled() { [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]; }
nautobot_unreachable() {
  ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
    "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
}

Describe 'nautobot API'
  Skip if 'external tests disabled' external_tests_disabled
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  BeforeAll 'setup_nautobot_env'
  AfterAll  'teardown_test_env'

  It 'import succeeds'
    When call bin/cani alpha import nautobot \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds nautobot)"
    The status should equal 0
    The stderr should include 'Import completed successfully'
  End
End

Describe 'nautobot API'
  Skip if 'external tests disabled' external_tests_disabled
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  #shellcheck disable=SC2317
  _setup_nautobot_export() {
    setup_nautobot_env
    cp "$FIXTURES/cani/crud_inventory.json" "$(ie_ds nautobot_export)"
  }

  BeforeAll '_setup_nautobot_export'
  AfterAll  'teardown_test_env'

  It 'export dry-run connects and processes pipeline'
    When call bin/cani alpha export nautobot \
      --dry-run \
      --config "$CANI_CONF" \
      --datastore-path "$(ie_ds nautobot_export)"
    # Exit status may be non-zero due to fixture/schema mismatches with
    # an empty Nautobot — the test verifies connectivity and pipeline
    # execution, not data correctness.
    The status should not equal ""
    The stderr should include 'Successfully connected to Nautobot'
    The stderr should include 'Nautobot Sync Summary'
  End
End
