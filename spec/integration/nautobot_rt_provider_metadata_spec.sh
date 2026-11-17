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

# Provider metadata reaches Nautobot only when the export lists its key.
#
# Provider metadata belongs to the provider that wrote it, and an export sends
# an object's custom fields and nothing else. The export option
# provider_metadata_fields (--provider-metadata-field) sends one key of one
# provider's bucket as a custom field. This spec pins what a live Nautobot
# receives: nothing by default, the listed key under the name given, the
# provider's value in place of a custom field of that name, and a report for
# a custom field Nautobot has no definition of.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#   - RUN_EXTERNAL_TESTS=1 (make itest-nautobot)
#
# The provider value is CSM's node ID, which `update device --nid` writes to
# providerMetadata.csm.nid on any inventory. The custom field is spec-owned
# (cani_rt_ prefix) and dropped at the end, so no other spec sees it.
#
# Walkthrough:
#   a. Reset, build the CLI inventory, record node ID 1001 on rt-node-1 and
#      define cani_rt_node_id on devices. Export: Nautobot gets the
#      definition and no value.
#   b. Export with --merge and csm.nid listed as cani_rt_node_id: only
#      rt-node-1 is updated and Nautobot holds 1001 on it; rt-node-2, which
#      has no node ID, holds nothing. The datastore's custom fields are
#      untouched, and the same export again updates nothing.
#   c. Set the custom field itself to 2002 on rt-node-1. With the key listed
#      the export keeps sending 1001 and names the device and the key; with
#      no key listed it sends the custom field, 2002.
#   d. List the key under a name Nautobot has no definition of: the export
#      reports the device and the key.

CF_KEY=cani_rt_node_id
LISTED="csm.nid=$CF_KEY"

#shellcheck disable=SC2317
rt_cf() { PYTHONPATH="$FIXTURES/nautobot" python3 -m roundtrip.device_cf "$@"; }

# Record a CSM node ID on rt-node-1 and define the custom field on devices.
#shellcheck disable=SC2317
record_node_id() {
  rt_cani update device rt-node-1 --nid 1001 &&
    rt_cani add metadata custom-field "$CF_KEY" --type integer --content-types dcim.device
}

# The custom field's value on each node as Nautobot holds it: NAME=VALUE.
#shellcheck disable=SC2317
nautobot_values() {
  for _node in rt-node-1 rt-node-2; do
    nb_list_field "dcim/devices/?name=$_node" \
      "print('$_node=' + str(data['results'][0]['custom_fields'].get('$CF_KEY')))" || return 1
  done
}

#shellcheck disable=SC2317
drop_key() { rt_cf drop "$CF_KEY" >/dev/null 2>&1 || true; }

Describe 'INTEGRATION: Nautobot export of listed provider metadata'

  Skip if 'SKIP_EXTERNAL_TESTS is set' [ "${SKIP_EXTERNAL_TESTS:-0}" = "1" ]
  Skip if 'Nautobot is not reachable' nautobot_unreachable

  AfterAll 'drop_key'

  Describe 'with no key listed'
    # a. A provider's value and a custom field to carry it, not yet connected.
    It 'resets Nautobot'
      BeforeCall rt_begin
      When call rt_reset
      The status should equal 0
      The output should include 'Reset complete'
    End

    It 'builds the fixture inventory with the CLI'
      When call rt_build_inventory
      The status should equal 0
      The stderr should include 'Added prefix 10.31.0.0/24'
    End

    It 'records a CSM node ID and defines the custom field'
      When call record_node_id
      The status should equal 0
      The stderr should include "Added custom field \"$CF_KEY\""
    End

    It 'exports the definition'
      When call rt_export
      The status should equal 0
      The output should include 'Export completed successfully'
      The output should include "Created custom field: $CF_KEY"
      The output should include 'Created devices: 2'
      The output should not include 'did not arrive'
    End

    It 'sends no provider metadata'
      When call nautobot_values
      The status should equal 0
      The line 1 of output should equal 'rt-node-1=None'
      The line 2 of output should equal 'rt-node-2=None'
    End
  End

  Describe 'with the key listed'
    # b. Only the device whose CSM bucket holds the key has anything to send.
    It 'updates the device that holds the key'
      When call rt_export --merge --provider-metadata-field "$LISTED"
      The status should equal 0
      The output should include 'Export completed successfully'
      The output should include 'Updated devices: 1'
      The output should include '~ rt-node-1'
      The output should not include 'did not arrive'
    End

    It 'stores the provider value under the custom field'
      When call nautobot_values
      The status should equal 0
      The line 1 of output should equal 'rt-node-1=1001'
      The line 2 of output should equal 'rt-node-2=None'
    End

    It 'leaves the custom fields of the datastore alone'
      When call rt_cf show "$CANI_DS" rt-node-1 "$CF_KEY"
      The status should equal 0
      The output should equal "rt-node-1 $CF_KEY=<unset>"
    End

    It 'has nothing left to send'
      When call rt_export --merge --provider-metadata-field "$LISTED"
      The status should equal 0
      The output should include 'Export completed successfully'
      The output should not include 'Updated devices'
    End
  End

  Describe 'a custom field of the same name'
    # c. The listed key belongs to its provider, so its value is the one sent.
    It 'is set on the device'
      When call rt_cani update device rt-node-1 --custom-field "$CF_KEY=2002"
      The status should equal 0
      The stderr should include 'Updated device'
    End

    It 'gives way to the listed provider value, and the export says so'
      When call rt_export --merge --provider-metadata-field "$LISTED"
      The status should equal 0
      The output should include "device rt-node-1: custom field \"$CF_KEY\" comes from csm provider metadata"
      The output should not include 'Updated devices'
    End

    It 'is sent when no key is listed'
      When call rt_export --merge
      The status should equal 0
      The output should include 'Updated devices: 1'
      The output should not include 'provider metadata'
    End

    It 'is the value Nautobot holds afterwards'
      When call nautobot_values
      The status should equal 0
      The line 1 of output should equal 'rt-node-1=2002'
      The line 2 of output should equal 'rt-node-2=None'
    End
  End

  Describe 'a custom field Nautobot has no definition of'
    # d. Nautobot answers 2xx and drops the value; the export reports it.
    It 'is reported by device and key'
      When call rt_export --merge --provider-metadata-field csm.nid=cani_rt_undefined
      The status should equal 0
      The output should include 'device rt-node-1: custom field "cani_rt_undefined" did not arrive'
      The output should include 'Custom field values not accepted by Nautobot (no definition for the object type): 1'
    End
  End
End
