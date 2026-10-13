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

"""Nautobot's change ledger as a write detector.

Nautobot records every create, update and delete made through the API in
extras/object-changes. A command that must not write (a dry run, an export
with nothing to do) is therefore checked by counting the ledger before and
after; the newest entries name whatever slipped through.
"""

from seed_nautobot_api import api

LEDGER = "extras/object-changes/"


def count():
    """Total number of recorded changes."""
    return int(api("GET", f"{LEDGER}?limit=1").get("count", 0))


def since(mark):
    """(number of changes since mark, newest-first one-line descriptions)."""
    delta = count() - mark
    if delta <= 0:
        return max(delta, 0), []
    page = api("GET", f"{LEDGER}?limit={delta}&sort=-time&depth=0")
    entries = [
        f"{row.get('action', {}).get('value', '?')} {row.get('changed_object_type', '?')} {row.get('object_repr', '')}"
        for row in page.get("results", [])
    ]
    return delta, entries
