#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2022-2023 Hewlett Packard Enterprise Development LP
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
set -e
set -u

# SIM_REPO="${1:-../hms-simulation-environment}"
# SLS_DUMP="${2:-../hms-simulation-environment/configs/sls/no_hardware.json}"
TEST_TYPE="${1:-integration}"

# run a cani workflow, then inspect the CSM services 
for test in ./spec/"${TEST_TYPE}"/*
do
  if ! [ -d "${test}" ]; then
    # get the filename and extension
    filename=$(basename -- "$test")
    # extension="${filename##*.}"
    filename="${filename%.*}"
    # run shellspec to emulate human interation
    if [ "${DEBUG:-false}" = "true" ]; then
      set -x
      shellspec "${test}" -f tap --no-banner -x
      set +x
    else
      set -x
      shellspec "${test}" -f tap --no-banner
      set +x
    fi
    # strip the _spec.sh extension and replace it with .yml (an assertion file with the same name as the shellspec test)
    # run the tavern test to validate cani changed or didn't change the correct things within CSM services
    if [ -f ./spec/integration/tavern/test_"${filename%_spec}".tavern.yml ]; then
      set -x
      tavern-ci ./spec/integration/tavern/test_"${filename%_spec}".tavern.yml
      set +x
    fi
  fi 
done
