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
# Preflight check for the cani test tiers.  Reports which prerequisites are
# present and prints the exact command that fixes each one that is not.
#
# Prerequisites at or below the requested tier are required and cause a
# non-zero exit; anything above it is reported as a note so you can see what a
# higher tier will need later.
#
# Usage: tools/doctor.sh [unit|func|int|all]

set -u

SHELLSPEC_MIN="0.28.1"
CSM_REGISTRY="artifactory.algol60.net"
NAUTOBOT_HEALTH="http://localhost:8081/api/status/"
CSM_HEALTH="https://localhost:8443/apis/sls/v1/health"

TIER="${1:-func}"
case "$TIER" in
unit) want=1 ;;
func) want=2 ;;
int | all) want=3 ;;
*)
  echo "usage: $0 [unit|func|int|all]" >&2
  exit 2
  ;;
esac

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  c_reset=$(printf '\033[0m')
  c_green=$(printf '\033[32m')
  c_red=$(printf '\033[31m')
  c_yellow=$(printf '\033[33m')
  c_dim=$(printf '\033[2m')
  c_bold=$(printf '\033[1m')
else
  c_reset='' c_green='' c_red='' c_yellow='' c_dim='' c_bold=''
fi

missing=0

have() { command -v "$1" >/dev/null 2>&1; }

# report <tier> <label> <ok?0:1> <detail> <fix>
report() {
  if [ "$3" -eq 0 ]; then
    printf '  %s%-7s%s %-16s %s%s%s\n' \
      "$c_green" "ok" "$c_reset" "$2" "$c_dim" "$4" "$c_reset"
  elif [ "$1" -le "$want" ]; then
    printf '  %s%-7s%s %-16s %s\n' "$c_red" "missing" "$c_reset" "$2" "$5"
    missing=$((missing + 1))
  else
    printf '  %s%-7s%s %-16s %s%s%s\n' \
      "$c_yellow" "note" "$c_reset" "$2" "$c_dim" "$5" "$c_reset"
  fi
}

# check_cmd <tier> <label> <binary> <fix>
check_cmd() {
  if have "$3"; then
    report "$1" "$2" 0 "$(command -v "$3")" ""
  else
    report "$1" "$2" 1 "" "$4"
  fi
}

# Numeric field sort keeps this POSIX-safe; `sort -V` is not portable.
version_at_least() {
  [ -n "$1" ] || return 1
  [ "$(printf '%s\n%s\n' "$2" "$1" | sort -t. -k1,1n -k2,2n -k3,3n | head -n1)" = "$2" ]
}

printf '%scani doctor%s  tier: %s%s%s\n\n' \
  "$c_bold" "$c_reset" "$c_bold" "$TIER" "$c_reset"

# ── toolchain ────────────────────────────────────────────────────────────────
printf ' toolchain\n'

if have go; then
  go_want=$(awk '/^go /{print $2; exit}' go.mod 2>/dev/null)
  report 1 go 0 "$(go env GOVERSION 2>/dev/null) (go.mod wants ${go_want:-?})" ""
else
  report 1 go 1 "" "install Go: https://go.dev/dl/"
fi

check_cmd 1 git git "install git"
check_cmd 1 curl curl "install curl"

if [ "$(git config --get core.hooksPath 2>/dev/null)" = ".githooks" ]; then
  report 1 git-hooks 0 ".githooks" ""
else
  report 1 git-hooks 1 "" "make hooks"
fi

# ── shell tests ──────────────────────────────────────────────────────────────
printf '\n shell tests\n'

if have shellspec; then
  spec_version=$(shellspec --version 2>/dev/null | head -n1)
  if version_at_least "$spec_version" "$SHELLSPEC_MIN"; then
    report 2 shellspec 0 "$spec_version" ""
  else
    report 2 shellspec 1 "" \
      "found $spec_version, need >= $SHELLSPEC_MIN: make spec-clean spec-setup"
  fi
else
  report 2 shellspec 1 "" "make spec-setup"
fi

if [ "$(uname -s)" = "Darwin" ]; then
  check_cmd 2 gsed gsed "brew install gnu-sed (specs need GNU sed ranges)"
fi

# ── integration ──────────────────────────────────────────────────────────────
printf '\n integration\n'

check_cmd 3 docker docker "install Docker or Podman with a docker shim"

if have docker && docker compose version >/dev/null 2>&1; then
  report 3 docker-compose 0 "$(docker compose version --short 2>/dev/null)" ""
else
  report 3 docker-compose 1 "" "install the Docker Compose v2 plugin"
fi

check_cmd 3 python3 python3 "install Python 3.12+"
check_cmd 3 virtualenv virtualenv "pip3 install virtualenv"
check_cmd 3 jq jq "install jq (specs parse API responses with it)"
check_cmd 3 openssl openssl "install openssl (make csm-certs needs it)"

if [ -f venv/bin/activate ]; then
  report 3 venv 0 "venv/" ""
else
  report 3 venv 1 "" "make venv"
fi

# The registry key lands in the Docker config even when a credential helper
# holds the secret, so this detects a login without ever reading credentials.
docker_config="${DOCKER_CONFIG:-$HOME/.docker}/config.json"
if [ -f "$docker_config" ] && grep -q "$CSM_REGISTRY" "$docker_config" 2>/dev/null; then
  report 3 csm-registry 0 "$CSM_REGISTRY" ""
else
  report 3 csm-registry 1 "" \
    "docker login $CSM_REGISTRY (CSM simulator images are HPE-internal)"
fi

# ── services (informational: `make sim-up` starts these on demand) ───────────
printf '\n simulators %s(started on demand by make sim-up)%s\n' "$c_dim" "$c_reset"

if curl -sf -o /dev/null --max-time 3 "$NAUTOBOT_HEALTH" 2>/dev/null; then
  report 9 nautobot 0 "http://localhost:8081" ""
else
  report 9 nautobot 1 "" "not running: make nautobot-up"
fi

if curl -skf -o /dev/null --max-time 3 "$CSM_HEALTH" 2>/dev/null; then
  report 9 csm 0 "https://localhost:8443" ""
else
  report 9 csm 1 "" "not running: make csm-up"
fi

# ── summary ──────────────────────────────────────────────────────────────────
printf '\n'
if [ "$missing" -gt 0 ]; then
  printf '%s✖%s %d prerequisite(s) missing for tier %s.\n' \
    "$c_red" "$c_reset" "$missing" "$TIER"
  printf '  run the fixes above, then: make doctor TIER=%s\n' "$TIER"
  exit 1
fi

printf '%s✔%s tier %s is ready.\n' "$c_green" "$c_reset" "$TIER"
case "$TIER" in
unit) printf '  next: make utest\n' ;;
func) printf '  next: make test-fast   (higher tier: make doctor TIER=int)\n' ;;
*) printf '  next: make sim-up && make test\n' ;;
esac
