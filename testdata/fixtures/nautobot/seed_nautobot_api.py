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

"""Shared HTTP and configuration helpers for Nautobot seed scripts."""

import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request

try:
    import yaml  # PyYAML
    _HAS_YAML = True
except ImportError:
    _HAS_YAML = False


def _load_from_config():
    """Read the Nautobot URL and token from the CANI config file."""
    config_path = os.environ.get("CANI_CONF", "")
    if not config_path or not os.path.isfile(config_path) or not _HAS_YAML:
        return None, None
    with open(config_path, encoding="utf-8") as config_file:
        config = yaml.safe_load(config_file)
    nautobot = (config or {}).get("providers", {}).get("nautobot", {})
    return nautobot.get("url"), nautobot.get("token")


_CONFIG_URL, _CONFIG_TOKEN = _load_from_config()
NAUTOBOT_URL = _CONFIG_URL or os.environ.get(
    "NAUTOBOT_URL", "http://localhost:8081/api"
)
NAUTOBOT_TOKEN = _CONFIG_TOKEN or os.environ.get(
    "NAUTOBOT_TOKEN", "0123456789abcdef0123456789abcdef01234567"
)


def api(method, path, data=None):
    """Make an API call and return its parsed JSON response."""
    url = f"{NAUTOBOT_URL}/{path}"
    body = json.dumps(data).encode() if data else None
    request = urllib.request.Request(
        url,
        data=body,
        method=method,
        headers={
            "Authorization": f"Token {NAUTOBOT_TOKEN}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
    )
    try:
        with urllib.request.urlopen(request) as response:
            return json.loads(response.read())
    except urllib.error.HTTPError as error:
        detail = error.read().decode()
        print(f"ERROR {error.code} {method} {url}: {detail}", file=sys.stderr)
        sys.exit(1)


def find_or_create(path, match_field, match_value, payload):
    """Return an existing object or create a new one."""
    value = urllib.parse.quote(str(match_value))
    results = api("GET", f"{path}?{match_field}={value}")
    if results.get("results"):
        return results["results"][0]
    return api("POST", path, payload)


def delete_all(path):
    """Delete every object at the given API list endpoint."""
    while True:
        results = api("GET", f"{path}?limit=50").get("results", [])
        if not results:
            return
        for obj in results:
            _delete(f"{NAUTOBOT_URL}/{path}{obj['id']}/")


def _delete(url):
    request = urllib.request.Request(
        url,
        method="DELETE",
        headers={
            "Authorization": f"Token {NAUTOBOT_TOKEN}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
    )
    try:
        urllib.request.urlopen(request)
    except urllib.error.HTTPError as error:
        if error.code != 404:
            detail = error.read().decode()
            print(f"WARN: DELETE {url}: {detail}", file=sys.stderr)