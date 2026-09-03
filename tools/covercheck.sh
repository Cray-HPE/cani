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
# Runs the unit tests and fails when a package listed in the floors file has
# dropped below its recorded statement coverage. Point COVERAGE_FLOORS at an
# empty file to run the tests without enforcing anything.

set -eu

MODULE="github.com/Cray-HPE/cani"

# Run from the module root regardless of where the script was invoked, so both
# `go test ./...` and the default floors path resolve the same way every time.
repo_root="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

FLOORS="${COVERAGE_FLOORS:-tools/coverage-floors.txt}"

if [ ! -f "$FLOORS" ]; then
  echo "covercheck: floors file not found: $FLOORS" >&2
  exit 1
fi

output="$(mktemp)" || { echo "covercheck: mktemp failed" >&2; exit 1; }
floors="$(mktemp)" || { echo "covercheck: mktemp failed" >&2; exit 1; }
# shellcheck disable=SC2064
trap "rm -f '$output' '$floors'" EXIT INT TERM

status=0
go test -cover ./... >"$output" 2>&1 || status=$?
cat "$output"
if [ "$status" -ne 0 ]; then
  exit "$status"
fi

sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$FLOORS" >"$floors"

awk -v module="$MODULE" -v floorsfile="$floors" '
FILENAME == floorsfile {
    if (NF != 2 || $2 !~ /^[0-9]+(\.[0-9]+)?$/ || $2 + 0 > 100) {
        printf "covercheck: invalid coverage floor: %s (expected package and percentage from 0 to 100)\n", $0
        invalid = 1
        next
    }
    floor[$1] = $2 + 0
    order[++count] = $1
    next
}
/coverage:/ {
    pkg = ""
    pct = ""
    for (i = 1; i <= NF; i++) {
        if (index($i, module) == 1) {
            pkg = $i
        }
        if ($i ~ /^[0-9]+(\.[0-9]+)?%$/) {
            pct = substr($i, 1, length($i) - 1)
        }
    }
    if (pkg != "" && pct != "") {
        covered[pkg] = pct + 0
        seen[pkg] = 1
    }
}
END {
    if (invalid) {
        exit 1
    }
    failed = 0
    for (i = 1; i <= count; i++) {
        pkg = order[i]
        if (!seen[pkg]) {
            printf "  MISSING  %-52s no coverage reported\n", pkg
            failed = 1
            continue
        }
        if (covered[pkg] + 0.0001 < floor[pkg]) {
            printf "  DROPPED  %-52s %5.1f%% < %5.1f%% floor\n", pkg, covered[pkg], floor[pkg]
            failed = 1
        }
    }
    if (failed) {
        print ""
        print "coverage floors breached; raise coverage or, if the drop is intended,"
        print "update tools/coverage-floors.txt in the same commit."
        exit 1
    }
    printf "coverage floors met for %d package(s)\n", count
}
' "$floors" "$output"
