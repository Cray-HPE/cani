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
"""Device custom fields for the re-import spec, edited as another operator would.

  define KEY                 create a text custom field on dcim.device
  set DEVICE KEY=VALUE       set one device's custom field value
  drop KEY                   delete the definition (and every value)
  show DATASTORE DEVICE KEY  print the datastore's value: DEVICE KEY=VALUE

Run as `python3 -m roundtrip.device_cf ...`. A standalone module, so it adds
no subcommand to roundtrip/__main__.py.
"""

import json
import sys
import urllib.parse

from seed_nautobot_api import NAUTOBOT_URL, _delete, api


def _field(key):
    found = api("GET", f"extras/custom-fields/?q={urllib.parse.quote(key)}").get("results", [])
    return next((field for field in found if field.get("key") == key), None)


def define(key):
    if not _field(key):
        api("POST", "extras/custom-fields/", {
            "key": key, "label": key, "type": "text", "content_types": ["dcim.device"],
        })
    print(f"defined {key}")


def set_value(device, assignment):
    key, value = assignment.split("=", 1)
    found = api("GET", f"dcim/devices/?name={urllib.parse.quote(device)}").get("results", [])
    if len(found) != 1:
        raise SystemExit(f"expected one device named {device}, found {len(found)}")
    api("PATCH", f"dcim/devices/{found[0]['id']}/", {"custom_fields": {key: value}})
    print(f"set {device} {key}={value}")


def drop(key):
    field = _field(key)
    if field:
        _delete(f"{NAUTOBOT_URL}/extras/custom-fields/{field['id']}/")
    print(f"dropped {key}")


def show(datastore, device, key):
    with open(datastore, encoding="utf-8") as handle:
        devices = json.load(handle).get("devices") or {}
    for record in devices.values():
        if record.get("name") == device:
            value = (record.get("customFields") or {}).get(key, "<unset>")
            print(f"{device} {key}={value}")
            return
    raise SystemExit(f"no device named {device} in {datastore}")


COMMANDS = {"define": define, "set": set_value, "drop": drop, "show": show}

if __name__ == "__main__":
    if len(sys.argv) < 2 or sys.argv[1] not in COMMANDS:
        raise SystemExit(__doc__)
    COMMANDS[sys.argv[1]](*sys.argv[2:])
