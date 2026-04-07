#
# MIT License
#
# (C) Copyright 2023, 2025 Hewlett Packard Enterprise Development LP
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
# shellcheck shell=sh

# Defining variables and functions here will affect all specfiles.
# Change shell options inside a function may cause different behavior,
# so it is better to set them here.
# set -eu

# This callback function will be invoked only once before loading specfiles.
spec_helper_precheck() {
  # Available functions: info, warn, error, abort, setenv, unsetenv
  # Available variables: VERSION, SHELL_TYPE, SHELL_VERSION
  : minimum_version "0.28.1"

  # Fixtures location ./spec/testdata/fixtures
  setenv FIXTURES="$SHELLSPEC_HELPERDIR/testdata/fixtures"
  # Make the built binary available as 'cani' in all subprocesses
  setenv PATH="$SHELLSPEC_HELPERDIR/../bin:$PATH"
  # use /tmp for consistent test dir that doesn't conflict with local dev
  # also helpful for cani config fixtures if this is a static, abs path
  setenv CANI_DIR="/tmp/.cani"
  setenv CANI_CONF="${CANI_DIR:=/tmp/.cani}/cani.yml"
  setenv CANI_DS="${CANI_DIR:=/tmp/.cani}/canidb.json"
  setenv CANI_LOG="${CANI_DIR:=/tmp/.cani}/canidb.log"

  # Nautobot integration test settings (must match cani_0.6.x.yml)
  setenv NAUTOBOT_URL="http://localhost:8081/api"
  setenv NAUTOBOT_TOKEN="0123456789abcdef0123456789abcdef01234567"

  # Skip tests that require external services (CSM API, mock servers, etc.)
  # Set SKIP_EXTERNAL_TESTS=1 to skip these tests
  : "${SKIP_EXTERNAL_TESTS:=0}"
	
  # On macOS, GNU sed (gsed) is required for GNU sed extensions (e.g., 0, address)
  # A wrapper at spec/support/bin/sed handles this transparently
  if [ "$(uname -s)" = "Darwin" ] && ! command -v gsed >/dev/null 2>&1; then
    warn "GNU sed (gsed) not found. Install with: brew install gnu-sed"
  fi
}

# This callback function will be invoked after a specfile has been loaded.
spec_helper_loaded() {
  :
}

# This callback function will be invoked after core modules has been loaded.
spec_helper_configure() {
  # Available functions: import, before_each, after_each, before_all, after_all
  : import 'support/custom_matcher'
}

# compare value to file content
# https://github.com/shellspec/shellspec/issues/295#issuecomment-1531834218
fixture(){
  #shellcheck disable=SC2317
  [ "${fixture:?}" = "$( cat "$FIXTURES/$1" )" ]
}

#shellcheck disable=SC2317
remove_config(){ rm -f "$CANI_CONF"; }
#shellcheck disable=SC2317
remove_datastore() { rm -f "$CANI_DS"; }
#shellcheck disable=SC2317
remove_log() { rm -f "$CANI_LOG"; }

# ---------- Test environment helpers ----------

# Create a clean test environment with an empty config and datastore.
# Uses /tmp/.cani to avoid conflicting with the user's local cani config.
#shellcheck disable=SC2317
setup_test_env() {
  rm -rf "$CANI_DIR"
  mkdir -p "$CANI_DIR"
  cp "$FIXTURES/cani/configs/test_config.yml" "$CANI_CONF"
  cp "$FIXTURES/cani/empty_inventory.json" "$CANI_DS"
}

# Create a test environment pre-loaded with the populated test-rack inventory.
#shellcheck disable=SC2317
setup_populated_env() {
  rm -rf "$CANI_DIR"
  mkdir -p "$CANI_DIR"
  cp "$FIXTURES/cani/configs/test_config.yml" "$CANI_CONF"
  cp "$FIXTURES/test-rack-inventory.json" "$CANI_DS"
}

# Remove the entire test directory.
#shellcheck disable=SC2317
teardown_test_env() {
  rm -rf "$CANI_DIR"
}

# Create a clean migration test environment by deploying a legacy config fixture.
# Usage: setup_migration_env cani_0.1.x.yml
#shellcheck disable=SC2317
setup_migration_env() {
  rm -rf "$CANI_DIR"
  mkdir -p "$CANI_DIR"
  cp "$FIXTURES/cani/configs/$1" "$CANI_CONF"
}

# Create a clean datastore migration test environment by deploying only a
# legacy datastore fixture.  No config is copied so cani must create one.
# Usage: setup_datastore_migration_env canitestdb_v1alpha1.json
#shellcheck disable=SC2317
setup_datastore_migration_env() {
  rm -rf "$CANI_DIR"
  mkdir -p "$CANI_DIR"
  cp "$FIXTURES/cani/legacy/$1" "$CANI_DS"
}

# Custom matcher used to find a string inside of a text containing ANSI escape codes.
# https://github.com/shellspec/shellspec/issues/278
match_colored_text() {
	# Source: https://unix.stackexchange.com/a/18979/348102
	sanitized_text="$(echo "${match_colored_text:?}" | perl -e '
		while (<>) {
			s/ \e[ #%()*+\-.\/]. |
				\r | # Remove extra carriage returns also
				(?:\e\[|\x9b) [ -?]* [@-~] | # CSI ... Cmd
				(?:\e\]|\x9d) .*? (?:\e\\|[\a\x9c]) | # OSC ... (ST|BEL)
				(?:\e[P^_]|[\x90\x9e\x9f]) .*? (?:\e\\|\x9c) | # (DCS|PM|APC) ... ST
				\e.|[\x80-\x9f] //xg;
				1 while s/[^\b][\b]//g;  # remove all non-backspace followed by backspace
			print;
		}
	')"
	echo "${sanitized_text}" | grep -q "$1"
}

# Custom matcher used to find a string inside of a text containing ANSI escape codes.
match_rich_text() {
	if [ -z "${1}" ]; then
		printf "ERROR: You cannot pass an empty string!\n" >&2;
		return 1;
	fi
	
	# shellcheck disable=SC2059
	rich_search_term="$(printf "$1")"
	match="$(echo "${match_rich_text:-}" | perl -ne "print if /\Q${rich_search_term}/")"

	[ -n "${match}" ]
}
