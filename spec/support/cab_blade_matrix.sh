#!/usr/bin/env sh
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

set -e
set -u

: "${CANI_CONF:-/tmp/.cani/cani.yml}"

for blade in $(bin/cani --config "$CANI_CONF" alpha add blade -L); do
  for cabinet in $(bin/cani --config "$CANI_CONF" alpha add cabinet -L); do
    if [ "$cabinet" = "hpe-eia-cabinet" ];then 
      continue
    else 
      # cab_ordinal=$(echo "$cabinet" | awk '{print $2}')
      cab_ds_fixture=$(echo "$cabinet" | awk '{print $1}' | tr '-' '_')
      # these vars are used in the tests
      echo %data "$cab_ds_fixture" "$blade"
    fi
  done
done