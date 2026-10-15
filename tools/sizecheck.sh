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
# Fails when a Go source file reaches the line budget from AGENTS.md: fewer
# than 300 lines for production code and fewer than 800 for _test.go files.
# vendor/ and generated files ("Code generated ... DO NOT EDIT") are skipped.
# A file listed in the exceptions file may stay at its recorded size but not
# grow; the gate notes entries that are back under budget so they can go.

set -eu

# Run from the module root regardless of where the script was invoked, so the
# paths in the exceptions file resolve the same way every time.
repo_root="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$repo_root"

ROOT="${SIZE_ROOT:-.}"
EXCEPTIONS="${SIZE_EXCEPTIONS:-tools/size-exceptions.txt}"
PROD_LIMIT="${SIZE_LIMIT:-300}"
TEST_LIMIT="${TEST_SIZE_LIMIT:-800}"

if [ ! -f "$EXCEPTIONS" ]; then
  echo "sizecheck: exceptions file not found: $EXCEPTIONS" >&2
  exit 1
fi

sizes="$(mktemp)" || { echo "sizecheck: mktemp failed" >&2; exit 1; }
caps="$(mktemp)" || { echo "sizecheck: mktemp failed" >&2; exit 1; }
# shellcheck disable=SC2064
trap "rm -f '$sizes' '$caps'" EXIT INT TERM

# One "<lines> <path>" row per hand-written Go file, paths relative to ROOT.
# find -exec … + is never invoked with zero files; GNU xargs would still run
# grep, which then reads stdin and reports "(standard input)" as a file.
(cd "$ROOT" && find . -name '*.go' -not -path '*/vendor/*' \
    -exec grep -L '^// Code generated .* DO NOT EDIT' {} + \
  | while IFS= read -r file; do
      printf '%s %s\n' "$(wc -l <"$file")" "${file#./}"
    done) >"$sizes"

sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$EXCEPTIONS" >"$caps"

awk -v prodLimit="$PROD_LIMIT" -v testLimit="$TEST_LIMIT" -v capsfile="$caps" -v exceptions="$EXCEPTIONS" '
FILENAME == capsfile {
    if (NF != 2 || $2 !~ /^[0-9]+$/) {
        printf "sizecheck: invalid size exception: %s (expected path and line count)\n", $0
        invalid = 1
        next
    }
    cap[$1] = $2 + 0
    next
}
{
    lines = $1 + 0
    path = $2
    limit = (path ~ /_test\.go$/) ? testLimit : prodLimit
    checked++
    if (path in cap) {
        seen[path] = 1
        if (lines > cap[path]) {
            printf "sizecheck: GREW %s: %d lines, exception allows %d\n", path, lines, cap[path]
            failed = 1
        } else if (lines < limit) {
            printf "sizecheck: note: %s is %d lines, under the %d-line budget; remove it from %s\n", path, lines, limit, exceptions
        }
        next
    }
    if (lines >= limit) {
        printf "sizecheck: OVER %s: %d lines (budget < %d)\n", path, lines, limit
        failed = 1
    }
}
END {
    if (invalid) {
        exit 1
    }
    excepted = 0
    for (path in cap) {
        excepted++
        if (!(path in seen)) {
            printf "sizecheck: STALE exception %s: no such file\n", path
            failed = 1
        }
    }
    if (failed) {
        exit 1
    }
    printf "sizecheck: %d file(s) within budget (%d excepted)\n", checked, excepted
}
' "$caps" "$sizes"
