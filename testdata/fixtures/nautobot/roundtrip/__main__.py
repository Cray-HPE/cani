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

"""Command-line entry point: ``python3 -m roundtrip <command> [args]``.

Commands:
  reset                          empty Nautobot of every object an export can create
  seed                           create the deterministic round-trip fixture
  project nautobot OUT           write the Nautobot projection to OUT
  project cani DATASTORE OUT     write the datastore projection to OUT
  diff A B                       compare two projections; exit 1 when they differ
  changes [MARK]                 print the object-change count, or the changes since MARK
  stamps DATASTORE               count the records that carry a Nautobot external ID
"""

import argparse
import json
import sys

from roundtrip import changes, projection, reset, seed
from roundtrip.cani_source import load_cani, stamp_count
from roundtrip.nautobot_source import fetch_nautobot


def _write(path, data):
    with open(path, "w", encoding="utf-8") as out:
        json.dump(data, out, indent=1, sort_keys=True)
        out.write("\n")


def _read(path):
    with open(path, encoding="utf-8") as src:
        return json.load(src)


def cmd_project(args):
    if args.side == "nautobot":
        _write(args.paths[0], projection.project_nautobot(fetch_nautobot()))
    else:
        _write(args.paths[1], projection.project_cani(load_cani(args.paths[0])))
    return 0


def cmd_diff(args):
    lines = projection.diff(_read(args.a), _read(args.b))
    if not lines:
        print("projections equal")
        return 0
    print("\n".join(lines))
    return 1


def cmd_changes(args):
    if args.mark is None:
        print(changes.count())
        return 0
    delta, entries = changes.since(args.mark)
    print(f"changes={delta}")
    for entry in entries:
        print(entry)
    return 0


def build_parser():
    parser = argparse.ArgumentParser(prog="roundtrip", description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("reset").set_defaults(run=lambda _args: reset.reset())
    sub.add_parser("seed").set_defaults(run=lambda _args: seed.seed())

    project = sub.add_parser("project")
    project.add_argument("side", choices=("nautobot", "cani"))
    project.add_argument("paths", nargs="+", help="OUT, or DATASTORE OUT for cani")
    project.set_defaults(run=cmd_project)

    diff = sub.add_parser("diff")
    diff.add_argument("a")
    diff.add_argument("b")
    diff.set_defaults(run=cmd_diff)

    change = sub.add_parser("changes")
    change.add_argument("mark", nargs="?", type=int)
    change.set_defaults(run=cmd_changes)

    stamps = sub.add_parser("stamps")
    stamps.add_argument("datastore")
    stamps.set_defaults(run=lambda args: print(stamp_count(load_cani(args.datastore))))
    return parser


def main(argv=None):
    args = build_parser().parse_args(argv)
    if args.command == "project":
        wanted = 1 if args.side == "nautobot" else 2
        if len(args.paths) != wanted:
            build_parser().error(f"project {args.side} takes {wanted} path(s)")
    return args.run(args) or 0


if __name__ == "__main__":
    sys.exit(main())
