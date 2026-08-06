#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
# Enforces the 300-line budget for non-test Go files declared in AGENTS.md.
#
# The repo has pre-existing files over budget, so this is a ratchet rather than
# a hard gate: files listed in the baseline are allowed at their recorded size
# but may not grow, and any file not in the baseline must stay under the limit.
# Shrinking a baselined file below the limit is reported so the entry can be
# dropped.
#
# Usage: tools/check_file_size.sh [--update]
#   --update  rewrite the baseline from the current tree

set -e
set -u

LIMIT=300
BASELINE="tools/file_size_baseline.txt"

# Generated and vendored code is excluded: it is not hand-maintained.
list_files() {
  find pkg cmd internal -name '*.go' \
    -not -name '*_test.go' \
    -not -path 'pkg/nautobot/*' \
    -not -path 'internal/openapi/*' \
    | sort
}

current_sizes() {
  list_files | while read -r file; do
    printf '%s %s\n' "$(wc -l < "$file" | tr -d ' ')" "$file"
  done
}

if [ "${1:-}" = "--update" ]; then
  current_sizes | awk -v limit="$LIMIT" '$1 > limit { print $2, $1 }' > "$BASELINE"
  echo "wrote $BASELINE ($(wc -l < "$BASELINE" | tr -d ' ') entries)"
  exit 0
fi

if [ ! -f "$BASELINE" ]; then
  echo "missing $BASELINE; run tools/check_file_size.sh --update" >&2
  exit 1
fi

current_sizes | while read -r lines file; do
  allowed=$(awk -v f="$file" '$1 == f { print $2 }' "$BASELINE")

  if [ -z "$allowed" ]; then
    if [ "$lines" -gt "$LIMIT" ]; then
      echo "FAIL $file: $lines lines exceeds the $LIMIT-line budget" >&2
      echo fail >> "${TMPDIR:-/tmp}/cani_size_check.$$"
    fi
    continue
  fi

  if [ "$lines" -gt "$allowed" ]; then
    echo "FAIL $file: grew from $allowed to $lines lines; split it instead" >&2
    echo fail >> "${TMPDIR:-/tmp}/cani_size_check.$$"
  elif [ "$lines" -le "$LIMIT" ]; then
    echo "OK   $file is now $lines lines; drop it from $BASELINE" >&2
  fi
done

if [ -f "${TMPDIR:-/tmp}/cani_size_check.$$" ]; then
  failed=$(wc -l < "${TMPDIR:-/tmp}/cani_size_check.$$" | tr -d ' ')
  rm -f "${TMPDIR:-/tmp}/cani_size_check.$$"
  echo "$failed file(s) over budget" >&2
  exit 1
fi

echo "file size check passed"
