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

# check-conventional-commit.sh validates a single commit message header
# against the Conventional Commits specification
# (https://www.conventionalcommits.org). It is the single source of truth
# shared by the commit-msg git hook and the Conventional Commits CI workflow.
#
# It deliberately depends on nothing but a POSIX shell, so contributors do
# not need Node, npm or any third-party tooling installed.

set -e
set -u

PROG=${0##*/}
HEADER=""

# Conventional Commit types accepted by this repository.
ALLOWED_TYPES="build chore ci docs feat fix perf refactor revert style test"

# Maximum allowed length of the header (first) line.
HEADER_MAX_LENGTH=100

# A literal carriage return, so a CRLF message is measured the same as an LF one.
CR=$(printf '\r')

usage() {
  cat <<EOF
usage: $PROG --file <path> | --message <string>

  --file <path>       read the commit message from <path> (commit-msg hook)
  --message <string>  validate the given message string (e.g. a PR title)
  -h, --help          show this help
EOF
}

print_examples() {
  cat <<EOF
Commit messages must follow Conventional Commits:

  <type>[optional scope][optional !]: <subject>

Allowed types: $ALLOWED_TYPES

Examples:
  feat: add ipam allocation command
  fix(csm): handle empty SLS response
  docs: clarify make hooks setup
  refactor(devicetypes)!: drop deprecated loader

Rules:
  - header must be $HEADER_MAX_LENGTH characters or fewer
  - type is required and lower-case
  - scope (if used) is lower-case and non-empty
  - subject is required and must not end with '.'

See .github/CONTRIBUTING.md for details.
EOF
}

fail() {
  printf 'commit message rejected: %s\n\n' "$1" >&2
  printf 'offending header:\n  %s\n\n' "$HEADER" >&2
  print_examples >&2
  exit 1
}

# extract_header reads a commit message on stdin and prints the first
# non-blank, non-comment line, which git treats as the header.
extract_header() {
  while IFS= read -r line || [ -n "$line" ]; do
    line=${line%"$CR"}
    case "$line" in
      '#'*) continue ;;
      '') continue ;;
      *)
        printf '%s\n' "$line"
        return 0
        ;;
    esac
  done
  return 0
}

# is_ignored succeeds for messages that tooling generates automatically and
# that should not be linted (merges, reverts, autosquash, initial commit).
is_ignored() {
  case "$1" in
    "Merge "*) return 0 ;;
    "Revert \""*) return 0 ;;
    "fixup! "*) return 0 ;;
    "squash! "*) return 0 ;;
    "amend! "*) return 0 ;;
    "Initial commit") return 0 ;;
    *) return 1 ;;
  esac
}

validate_header() {
  header=$1
  prefix=""
  subject=""
  type=""
  scope=""

  if is_ignored "$header"; then
    return 0
  fi

  if [ -z "$header" ]; then
    fail "the commit message is empty"
  fi

  if [ "${#header}" -gt "$HEADER_MAX_LENGTH" ]; then
    fail "header is ${#header} characters; must be $HEADER_MAX_LENGTH or fewer"
  fi

  case "$header" in
    *": "*) ;;
    *) fail "missing ': ' separator; expected '<type>: <subject>'" ;;
  esac

  prefix=${header%%: *}
  subject=${header#*: }

  # An optional trailing '!' marks a breaking change.
  case "$prefix" in
    *!) prefix=${prefix%!} ;;
  esac

  case "$prefix" in
    *"("*")")
      type=${prefix%%"("*}
      scope=${prefix#*"("}
      scope=${scope%")"}
      if [ -z "$scope" ]; then
        fail "scope is empty; write '<type>(<scope>): ...' or omit the parentheses"
      fi
      # A scope holding parentheses means the prefix nested them, e.g. 'feat(a)(b)'.
      case "$scope" in
        *[\(\)]*) fail "malformed scope in '$prefix'" ;;
      esac
      # Use the POSIX character class (not [A-Z]) so locale collation does
      # not cause lower-case letters to match an upper-case range.
      case "$scope" in
        *[[:upper:]]*) fail "scope '$scope' must be lower-case" ;;
      esac
      ;;
    *"("* | *")"*)
      fail "malformed scope in '$prefix'"
      ;;
    *)
      type=$prefix
      ;;
  esac

  if [ -z "$type" ]; then
    fail "type is required, e.g. 'feat: ...'"
  fi

  case " $ALLOWED_TYPES " in
    *" $type "*) ;;
    *) fail "type '$type' is not allowed; use one of: $ALLOWED_TYPES" ;;
  esac

  if [ -z "$subject" ]; then
    fail "subject is required after '<type>: '"
  fi

  case "$subject" in
    *.) fail "subject must not end with '.'" ;;
  esac

  return 0
}

main() {
  mode=""
  value=""

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --file)
        [ "$#" -ge 2 ] || { printf '%s: --file requires a value\n' "$PROG" >&2; exit 2; }
        mode="file"
        value=$2
        shift 2
        ;;
      --message)
        [ "$#" -ge 2 ] || { printf '%s: --message requires a value\n' "$PROG" >&2; exit 2; }
        mode="message"
        value=$2
        shift 2
        ;;
      -h | --help)
        usage
        exit 0
        ;;
      *)
        printf '%s: unknown argument: %s\n' "$PROG" "$1" >&2
        usage >&2
        exit 2
        ;;
    esac
  done

  case "$mode" in
    file)
      [ -f "$value" ] || { printf '%s: file not found: %s\n' "$PROG" "$value" >&2; exit 2; }
      HEADER=$(extract_header < "$value")
      ;;
    message)
      HEADER=$(printf '%s\n' "$value" | extract_header)
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac

  validate_header "$HEADER"
}

main "$@"
