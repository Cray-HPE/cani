#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
  # use /tmp for consistent test dir that doesn't conflict with local dev
  # also helpful for cani config fixtures if this is a static, abs path
  setenv CANI_DIR="/tmp/.cani"
  setenv CANI_CONF="${CANI_DIR:=/tmp/.cani}/cani.yml"
  setenv CANI_DS="${CANI_DIR:=/tmp/.cani}/canidb.json"
  setenv CANI_LOG="${CANI_DIR:=/tmp/.cani}/canidb.log"
  setenv CANI_CUSTOM_HW_DIR="${CANI_DIR:=/tmp/.cani}/hardware-types"
  setenv CANI_CUSTOM_HW_CONF="${CANI_DIR:=/tmp/.cani}/hardware-types/my_custom_hw.yml"
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

# functions to deploy various fixtures with different scenarios

# deploys a config with session.active = true
use_active_session(){
  #shellcheck disable=SC2317
  mkdir -p "$(dirname "$CANI_CONF")"
  cp "$FIXTURES"/cani/configs/canitest_valid_active.yml "$CANI_CONF"
}

# deploys a config with session.active = false
use_inactive_session(){ 
  mkdir -p "$(dirname "$CANI_CONF")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitest_valid_inactive.yml "$CANI_CONF"
} 

use_custom_hw_type(){ 
  mkdir -p "$(dirname "$CANI_CUSTOM_HW_CONF")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/my_custom_hw.yml "$CANI_CUSTOM_HW_CONF"
} 

# deploys a datastore with one system only
use_valid_datastore_system_only(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_system_only.json "$CANI_DS"
}

# deploys a datastore with one eia cabinet (and child hardware)
use_valid_datastore_one_hpe_eia_cabinet_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_eia_only.json "$CANI_DS"
} 

# deploys a datastore with one ex2000 cabinet (and child hardware)
use_valid_datastore_one_hpe_ex2000_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex2000_only.json "$CANI_DS"
} 

# deploys a datastore with one ex2000 cabinet (and one blade)
use_valid_datastore_one_ex2000_one_blade(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex2000_one_blade.json "$CANI_DS"
} 

# deploys a datastore with one hpe_ex2500_1_liquid_cooled_chassis cabinet (and child hardware)
use_valid_datastore_one_hpe_ex2500_1_liquid_cooled_chassis_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex2500_1_only.json "$CANI_DS"
} 

# deploys a datastore with one hpe_ex2500_2_liquid_cooled_chassis cabinet (and child hardware)
use_valid_datastore_one_hpe_ex2500_2_liquid_cooled_chassis_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex2500_2_only.json "$CANI_DS"
} 

# deploys a datastore with one hpe_ex2500_3_liquid_cooled_chassis cabinet (and child hardware)
use_valid_datastore_one_hpe_ex2500_3_liquid_cooled_chassis_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex2500_3_only.json "$CANI_DS"
} 

# deploys a datastore with one hpe_ex3000 cabinet (and child hardware)
use_valid_datastore_one_hpe_ex3000_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex3000_only.json "$CANI_DS"
} 

# deploys a datastore with one hpe_ex4000 cabinet (and child hardware)
use_valid_datastore_one_hpe_ex4000_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_ex4000_only.json "$CANI_DS"
} 

# deploys a datastore with one hpe_ex4000 cabinet (and child hardware)
use_valid_datastore_one_my_custom_cabinet_cabinet(){ 
  mkdir -p "$(dirname "$CANI_DS")"
  #shellcheck disable=SC2317
  cp "$FIXTURES"/cani/configs/canitestdb_valid_my_custom_cabinet_only.json "$CANI_DS"
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
